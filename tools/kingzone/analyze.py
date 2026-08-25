#!/usr/bin/env python3
"""Offline king-zone exposure analysis for annotated paired-match PGNs."""

from __future__ import annotations

import argparse
import csv
import hashlib
import json
import math
import random
import re
from dataclasses import dataclass
from decimal import Decimal, InvalidOperation, ROUND_HALF_UP
from pathlib import Path
from typing import Any, Iterable, Sequence

import chess
import chess.pgn


PIECE_FEATURES = (
    ("bishop", chess.BISHOP),
    ("rook", chess.ROOK),
    ("queen", chess.QUEEN),
)
NON_PAWN_VALUES = {
    chess.KNIGHT: 320,
    chess.BISHOP: 330,
    chess.ROOK: 500,
    chess.QUEEN: 900,
}
BOOTSTRAP_REPLICATIONS = 10_000
BOOTSTRAP_SEED = 20260824
TACTICAL_REPLY_FIELDS = (
    "reply_is_capture",
    "reply_is_check",
    "reply_is_capture_check",
    "reply_is_quiet_check",
    "reply_is_promotion",
    "reply_qsearch_visible",
    "reply_qsearch_invisible",
    "reply_is_quiet_check_qsearch_invisible",
    "reply_is_quiet_noncheck",
)
TACTICAL_STAT_FIELDS = ("reply_position_in_check", *TACTICAL_REPLY_FIELDS)
TACTICAL_MIN_ROWS = 20
TACTICAL_MIN_GAMES = 10
TACTICAL_QUIET_CHECK_MAJORITY_MIN_EXCLUSIVE = 0.50
TACTICAL_QUIET_CHECK_EXCESS_MIN = 0.15
SCORE_RE = re.compile(
    r"^\s*(?P<sign>[+-]?)(?:(?P<mate>[Mm])(?P<distance>\d+)|(?P<cp>\d+(?:\.\d+)?))(?=/|\s|,|$)"
)


@dataclass(frozen=True)
class ParsedScore:
    cp: int | None
    is_mate: bool
    mate_distance: int | None = None
    mate_sign: int = 0


def parse_score(comment: str, mate_cap_cp: int) -> ParsedScore | None:
    """Parse cutechess' leading pawn or mate score from a move comment."""
    match = SCORE_RE.match(comment or "")
    if not match:
        return None
    sign = -1 if match.group("sign") == "-" else 1
    if match.group("mate"):
        distance = int(match.group("distance"))
        return ParsedScore(None, True, distance, sign)
    try:
        pawns = Decimal(match.group("cp"))
    except (InvalidOperation, TypeError):
        return None
    cp = int((pawns * 100).quantize(Decimal("1"), rounding=ROUND_HALF_UP))
    return ParsedScore(sign * cp, False)


def king_ring(board: chess.Board, color: chess.Color) -> int:
    """Return the geometric eight-square king ring, excluding the king square."""
    king_square = board.king(color)
    if king_square is None:
        return 0
    return chess.BB_KING_ATTACKS[king_square]


def king_zone_features(board: chess.Board, king_color: chess.Color) -> dict[str, int | bool]:
    """Count occupancy-aware geometric B/R/Q attacks into the king ring.

    ``Board.attacks`` includes the first blocker (friendly or enemy) and stops
    beyond it. Pins are deliberately ignored: this is geometric pressure.
    """
    ring = king_ring(board, king_color)
    enemy = not king_color
    result: dict[str, int | bool] = {}
    all_hits = 0
    total_attackers = 0
    total_hit_sum = 0

    for name, piece_type in PIECE_FEATURES:
        attackers = 0
        hit_sum = 0
        union = 0
        for square in board.pieces(piece_type, enemy):
            hits = int(board.attacks(square)) & ring
            if hits:
                attackers += 1
                hit_sum += hits.bit_count()
                union |= hits
        result[f"{name}_attackers"] = attackers
        result[f"{name}_hit_sum"] = hit_sum
        result[f"{name}_union_hits"] = union.bit_count()
        total_attackers += attackers
        total_hit_sum += hit_sum
        all_hits |= union

    result["total_attackers"] = total_attackers
    result["total_hit_sum"] = total_hit_sum
    result["total_union_hits"] = all_hits.bit_count()
    king_square = board.king(king_color)
    result["king_in_check"] = bool(
        king_square is not None and board.is_attacked_by(enemy, king_square)
    )
    return result


def non_pawn_material(board: chess.Board) -> int:
    return sum(
        len(board.pieces(piece_type, color)) * value
        for piece_type, value in NON_PAWN_VALUES.items()
        for color in (chess.WHITE, chess.BLACK)
    )


def load_canaries(path: Path) -> dict[str, str]:
    by_fen: dict[str, str] = {}
    with path.open(newline="", encoding="utf-8") as handle:
        reader = csv.DictReader(handle)
        if reader.fieldnames != ["canary_name", "fen"]:
            raise ValueError("canary CSV must have exactly: canary_name,fen")
        for line, record in enumerate(reader, start=2):
            name = (record.get("canary_name") or "").strip()
            fen = (record.get("fen") or "").strip()
            if not name or not fen:
                raise ValueError(f"empty canary at line {line}")
            chess.Board(fen)  # Validate, but preserve exact spelling for matching.
            if fen in by_fen:
                raise ValueError(f"duplicate canary FEN at line {line}")
            by_fen[fen] = name
    return by_fen


def _player_for_turn(headers: chess.pgn.Headers, turn: chess.Color) -> str:
    return headers.get("White" if turn == chess.WHITE else "Black", "")


def classify_tactical_reply(board: chess.Board, move: chess.Move) -> dict[str, Any]:
    """Classify a legal immediate reply and its current qsearch eligibility.

    Equinox's quiescence admits every legal evasion when already in check.
    Otherwise it admits captures only before delta pruning; its later promotion
    exception does not rescue non-capturing promotions from the preceding
    capture filter.
    """
    if not board.is_legal(move):
        raise ValueError(f"illegal tactical reply {move.uci()} in {board.fen()}")
    position_in_check = board.is_check()
    is_capture = board.is_capture(move)
    is_check = board.gives_check(move)
    is_promotion = move.promotion is not None
    qsearch_visible = position_in_check or is_capture
    if position_in_check:
        visibility_reason = "evasion"
    elif is_capture:
        visibility_reason = "capture"
    else:
        visibility_reason = "filtered_non_capture"
    is_quiet_check = is_check and not is_capture and not is_promotion
    return {
        "reply_position_in_check": position_in_check,
        "reply_is_capture": is_capture,
        "reply_is_check": is_check,
        "reply_is_capture_check": is_capture and is_check,
        "reply_is_quiet_check": is_quiet_check,
        "reply_is_promotion": is_promotion,
        "reply_qsearch_visible": qsearch_visible,
        "reply_qsearch_invisible": not qsearch_visible,
        "reply_qsearch_visibility_reason": visibility_reason,
        "reply_is_quiet_check_qsearch_invisible": is_quiet_check and not qsearch_visible,
        "reply_is_quiet_noncheck": not is_capture and not is_check and not is_promotion,
    }


def analyze_pgn(
    pgn_path: Path,
    canaries: dict[str, str],
    engine: str,
    swing_threshold_cp: int,
    mate_cap_cp: int,
) -> tuple[list[dict[str, Any]], dict[str, int]]:
    rows: list[dict[str, Any]] = []
    counts = {
        "games": 0,
        "plies": 0,
        "engine_games_white": 0,
        "engine_games_black": 0,
        "engine_moves": 0,
        "rows": 0,
        "severe": 0,
        "reference_mates": 0,
        "skipped_engine_mate": 0,
        "skipped_engine_score": 0,
        "skipped_no_response": 0,
        "skipped_response_score": 0,
    }

    with pgn_path.open(encoding="utf-8") as handle:
        game_index = 0
        while game := chess.pgn.read_game(handle):
            game_index += 1
            counts["games"] += 1
            counts["engine_games_white"] += int(game.headers.get("White") == engine)
            counts["engine_games_black"] += int(game.headers.get("Black") == engine)
            board = game.board()
            nodes = list(game.mainline())
            counts["plies"] += len(nodes)
            for index, node in enumerate(nodes):
                mover = _player_for_turn(game.headers, board.turn)
                if mover == engine:
                    counts["engine_moves"] += 1
                    engine_score = parse_score(node.comment, mate_cap_cp)
                    if engine_score is None:
                        counts["skipped_engine_score"] += 1
                    elif engine_score.is_mate:
                        counts["skipped_engine_mate"] += 1
                    elif index + 1 >= len(nodes):
                        counts["skipped_no_response"] += 1
                    else:
                        response_score = parse_score(nodes[index + 1].comment, mate_cap_cp)
                        if response_score is None:
                            counts["skipped_response_score"] += 1
                        else:
                            fen = board.fen()
                            move = node.move
                            loss_swing = (
                                engine_score.cp + response_score.cp
                                if response_score.cp is not None
                                else None
                            )
                            optimism_gap = max(0, loss_swing) if loss_swing is not None else None
                            loss_mate = response_score.is_mate and response_score.mate_sign > 0
                            severe = loss_mate or (
                                loss_swing is not None and loss_swing >= swing_threshold_cp
                            )
                            pre_features = king_zone_features(board, board.turn)
                            placebo_features = king_zone_features(board, not board.turn)
                            post_board = board.copy(stack=False)
                            post_board.push(move)
                            post_features = king_zone_features(post_board, board.turn)
                            reply_move = nodes[index + 1].move
                            reply_features = classify_tactical_reply(post_board, reply_move)
                            row: dict[str, Any] = {
                                "game_index": game_index,
                                "round": game.headers.get("Round", ""),
                                "engine_color": "white" if board.turn == chess.WHITE else "black",
                                "ply": board.ply() + 1,
                                "fen": fen,
                                "move_uci": move.uci(),
                                "move_san": board.san(move),
                                "reply_move_uci": reply_move.uci(),
                                "reply_move_san": post_board.san(reply_move),
                                "engine_cp": engine_score.cp,
                                "adverse_cp": response_score.cp,
                                "ref_is_mate": response_score.is_mate,
                                "ref_mate_distance": response_score.mate_distance,
                                "ref_mate_sign": response_score.mate_sign,
                                "loss_swing_cp": loss_swing,
                                "loss_mate": loss_mate,
                                "optimism_gap_cp": optimism_gap,
                                "severe": severe,
                                "non_pawn_material": non_pawn_material(board),
                                "canary_name": canaries.get(fen, ""),
                            }
                            row.update(pre_features)
                            for key, value in post_features.items():
                                row[f"post_{key}"] = value
                                if key != "king_in_check":
                                    row[f"delta_{key}"] = int(value) - int(pre_features[key])
                            for key, value in placebo_features.items():
                                row[f"placebo_{key}"] = value
                            row["queen_net_danger"] = (
                                int(row["queen_attackers"])
                                - int(row["placebo_queen_attackers"])
                            )
                            row.update(reply_features)
                            rows.append(row)
                            counts["rows"] += 1
                            counts["severe"] += int(row["severe"])
                            counts["reference_mates"] += int(response_score.is_mate)
                board.push(node.move)
    return rows, counts


def average_ranks(values: Sequence[float | int]) -> list[float]:
    order = sorted(range(len(values)), key=lambda index: values[index])
    ranks = [0.0] * len(values)
    start = 0
    while start < len(order):
        end = start + 1
        while end < len(order) and values[order[end]] == values[order[start]]:
            end += 1
        rank = (start + 1 + end) / 2.0
        for position in range(start, end):
            ranks[order[position]] = rank
        start = end
    return ranks


def pearson(xs: Sequence[float | int], ys: Sequence[float | int]) -> float | None:
    if len(xs) != len(ys) or len(xs) < 2:
        return None
    mean_x = sum(xs) / len(xs)
    mean_y = sum(ys) / len(ys)
    dx = [value - mean_x for value in xs]
    dy = [value - mean_y for value in ys]
    denominator = math.sqrt(sum(value * value for value in dx) * sum(value * value for value in dy))
    if denominator == 0:
        return None
    return sum(x * y for x, y in zip(dx, dy)) / denominator


def spearman(xs: Sequence[float | int], ys: Sequence[float | int]) -> float | None:
    if len(xs) != len(ys) or len(xs) < 2:
        return None
    return pearson(average_ranks(xs), average_ranks(ys))


def _rounded(value: float | None) -> float | None:
    return None if value is None else round(value, 6)


def median(values: Sequence[float | int]) -> float | None:
    if not values:
        return None
    ordered = sorted(values)
    middle = len(ordered) // 2
    if len(ordered) % 2:
        return float(ordered[middle])
    return (ordered[middle - 1] + ordered[middle]) / 2


def percentile(values: Sequence[float], probability: float) -> float | None:
    if not values:
        return None
    ordered = sorted(values)
    position = (len(ordered) - 1) * probability
    lower = math.floor(position)
    upper = math.ceil(position)
    if lower == upper:
        return ordered[lower]
    weight = position - lower
    return ordered[lower] * (1 - weight) + ordered[upper] * weight


def _effect(rows: Sequence[dict[str, Any]], feature: str) -> tuple[float | None, float | None]:
    exposed = [row for row in rows if row[feature] > 0]
    unexposed = [row for row in rows if row[feature] <= 0]
    risk_difference = None
    if exposed and unexposed:
        risk_difference = (
            sum(bool(row["severe"]) for row in exposed) / len(exposed)
            - sum(bool(row["severe"]) for row in unexposed) / len(unexposed)
        )
    exposed_swings = [row["loss_swing_cp"] for row in exposed if row["loss_swing_cp"] is not None]
    unexposed_swings = [row["loss_swing_cp"] for row in unexposed if row["loss_swing_cp"] is not None]
    median_delta = None
    if exposed_swings and unexposed_swings:
        median_delta = median(exposed_swings) - median(unexposed_swings)
    return risk_difference, median_delta


def cluster_bootstrap(
    rows: Sequence[dict[str, Any]],
    feature: str,
    replications: int = BOOTSTRAP_REPLICATIONS,
    seed: int = BOOTSTRAP_SEED,
) -> dict[str, Any]:
    by_game: dict[int, list[dict[str, Any]]] = {}
    for row in rows:
        by_game.setdefault(row["game_index"], []).append(row)
    games = sorted(by_game)
    games_exposed = sum(any(row[feature] > 0 for row in by_game[game]) for game in games)
    if not games:
        return {
            "replications": replications,
            "seed": seed,
            "games": 0,
            "games_exposed": 0,
            "risk_difference_ci95": [None, None],
            "median_delta_cp_ci95": [None, None],
        }
    rng = random.Random(seed)
    risk_differences: list[float] = []
    median_deltas: list[float] = []
    for _ in range(replications):
        sampled: list[dict[str, Any]] = []
        for _ in games:
            sampled.extend(by_game[rng.choice(games)])
        risk_difference, median_delta = _effect(sampled, feature)
        if risk_difference is not None:
            risk_differences.append(risk_difference)
        if median_delta is not None:
            median_deltas.append(median_delta)
    return {
        "replications": replications,
        "seed": seed,
        "games": len(games),
        "games_exposed": games_exposed,
        "risk_difference_ci95": [
            _rounded(percentile(risk_differences, 0.025)),
            _rounded(percentile(risk_differences, 0.975)),
        ],
        "median_delta_cp_ci95": [
            _rounded(percentile(median_deltas, 0.025)),
            _rounded(percentile(median_deltas, 0.975)),
        ],
    }


def feature_stats(
    rows: Sequence[dict[str, Any]], feature: str, with_bootstrap: bool = False
) -> dict[str, Any]:
    exposed = [row for row in rows if row[feature] > 0]
    unexposed = [row for row in rows if row[feature] <= 0]
    exposed_severe = sum(bool(row["severe"]) for row in exposed)
    unexposed_severe = sum(bool(row["severe"]) for row in unexposed)
    a = exposed_severe
    b = len(exposed) - exposed_severe
    c = unexposed_severe
    d = len(unexposed) - unexposed_severe
    odds_ratio = ((a + 0.5) * (d + 0.5)) / ((b + 0.5) * (c + 0.5))
    exposed_risk = (a + 0.5) / (len(exposed) + 1)
    unexposed_risk = (c + 0.5) / (len(unexposed) + 1)
    finite = [row for row in rows if row["loss_swing_cp"] is not None]
    values = [row[feature] for row in finite]
    swings = [row["loss_swing_cp"] for row in finite]
    risk_difference, median_delta = _effect(rows, feature)
    result = {
        "rows": len(rows),
        "exposed_rows": len(exposed),
        "unexposed_rows": len(unexposed),
        "exposed_severe": exposed_severe,
        "unexposed_severe": unexposed_severe,
        "severe_rate_exposed": _rounded(exposed_severe / len(exposed) if exposed else None),
        "severe_rate_unexposed": _rounded(unexposed_severe / len(unexposed) if unexposed else None),
        "risk_difference": _rounded(risk_difference),
        "haldane_odds_ratio": _rounded(odds_ratio),
        "haldane_risk_ratio": _rounded(exposed_risk / unexposed_risk),
        "finite_rows": len(finite),
        "median_loss_swing_exposed_cp": _rounded(median([
            row["loss_swing_cp"] for row in exposed if row["loss_swing_cp"] is not None
        ])),
        "median_loss_swing_unexposed_cp": _rounded(median([
            row["loss_swing_cp"] for row in unexposed if row["loss_swing_cp"] is not None
        ])),
        "median_delta_cp": _rounded(median_delta),
        "pearson_vs_loss_swing": _rounded(pearson(values, swings)),
        "spearman_vs_loss_swing": _rounded(spearman(values, swings)),
    }
    if with_bootstrap:
        result["cluster_bootstrap"] = cluster_bootstrap(rows, feature)
    return result


def descriptive_stats(rows: Sequence[dict[str, Any]], prefix: str = "") -> dict[str, Any]:
    base_features = [
        "bishop_attackers",
        "bishop_hit_sum",
        "bishop_union_hits",
        "rook_attackers",
        "rook_hit_sum",
        "rook_union_hits",
        "queen_attackers",
        "queen_hit_sum",
        "queen_union_hits",
        "total_attackers",
        "total_hit_sum",
        "total_union_hits",
    ]
    features = [prefix + feature for feature in base_features]
    return {
        feature: feature_stats(
            rows,
            feature,
            prefix in ("", "placebo_") and feature.endswith("_attackers"),
        )
        for feature in features
    }


def sensitivity_stats(
    rows: Sequence[dict[str, Any]], swing_threshold_cp: int, caps: Iterable[int]
) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for cap in sorted(set(caps)):
        adjusted = []
        for row in rows:
            copy = dict(row)
            adverse_cp = row["adverse_cp"]
            if row["ref_is_mate"]:
                adverse_cp = row["ref_mate_sign"] * cap
            copy["optimism_gap_cp"] = max(0, row["engine_cp"] + adverse_cp)
            copy["severe"] = copy["optimism_gap_cp"] >= swing_threshold_cp
            adjusted.append(copy)
        result[str(cap)] = {
            "diagnostic_only": True,
            "capped_severe": sum(bool(row["severe"]) for row in adjusted),
        }
    return result


def tactical_reply_group_stats(rows: Sequence[dict[str, Any]]) -> dict[str, Any]:
    counts = {
        field: sum(bool(row[field]) for row in rows)
        for field in TACTICAL_STAT_FIELDS
    }
    visibility_reasons = ("evasion", "capture", "filtered_non_capture")
    reason_counts = {
        reason: sum(row["reply_qsearch_visibility_reason"] == reason for row in rows)
        for reason in visibility_reasons
    }
    return {
        "rows": len(rows),
        "games": len({row["game_index"] for row in rows}),
        "severe_rows": sum(bool(row["severe"]) for row in rows),
        "counts": counts,
        "rates": {
            field: _rounded(count / len(rows) if rows else None)
            for field, count in counts.items()
        },
        "games_by_feature": {
            field: len({row["game_index"] for row in rows if row[field]})
            for field in TACTICAL_STAT_FIELDS
        },
        "qsearch_visibility_reason_counts": reason_counts,
        "qsearch_visibility_reason_rates": {
            reason: _rounded(count / len(rows) if rows else None)
            for reason, count in reason_counts.items()
        },
    }


def tactical_reply_section(rows: Sequence[dict[str, Any]]) -> dict[str, Any]:
    net_danger = [row for row in rows if row["queen_net_danger"] > 0]
    non_net_danger = [row for row in rows if row["queen_net_danger"] <= 0]
    return {
        "all": tactical_reply_group_stats(rows),
        "net_queen_danger": tactical_reply_group_stats(net_danger),
        "non_net_queen_danger": tactical_reply_group_stats(non_net_danger),
        "severe_net_queen_danger": tactical_reply_group_stats(
            [row for row in net_danger if row["severe"]]
        ),
        "severe_non_net_queen_danger": tactical_reply_group_stats(
            [row for row in non_net_danger if row["severe"]]
        ),
    }


def tactical_reply_decision(rows: Sequence[dict[str, Any]]) -> dict[str, Any]:
    primary = [
        row for row in rows
        if row["severe"] and row["queen_net_danger"] > 0
    ]
    control = [
        row for row in rows
        if row["severe"] and row["queen_net_danger"] <= 0
    ]
    feature = "reply_is_quiet_check_qsearch_invisible"
    primary_count = sum(bool(row[feature]) for row in primary)
    control_count = sum(bool(row[feature]) for row in control)
    primary_rate = primary_count / len(primary) if primary else None
    control_rate = control_count / len(control) if control else None
    rate_excess = (
        primary_rate - control_rate
        if primary_rate is not None and control_rate is not None
        else None
    )
    primary_games = len({row["game_index"] for row in primary})
    coverage_pass = len(primary) >= TACTICAL_MIN_ROWS and primary_games >= TACTICAL_MIN_GAMES
    majority_pass = (
        primary_rate is not None
        and primary_rate > TACTICAL_QUIET_CHECK_MAJORITY_MIN_EXCLUSIVE
    )
    excess_pass = (
        rate_excess is not None
        and rate_excess >= TACTICAL_QUIET_CHECK_EXCESS_MIN
    )
    recommend = coverage_pass and majority_pass and excess_pass
    return {
        "primary_definition": (
            "severe PRE midgame/non-check rows without canaries and with "
            "queen_net_danger > 0"
        ),
        "control_definition": (
            "severe PRE midgame/non-check rows without canaries and with "
            "queen_net_danger <= 0"
        ),
        "target_feature": feature,
        "primary_rows": len(primary),
        "primary_games": primary_games,
        "primary_feature_count": primary_count,
        "primary_feature_rate": _rounded(primary_rate),
        "control_rows": len(control),
        "control_games": len({row["game_index"] for row in control}),
        "control_feature_count": control_count,
        "control_feature_rate": _rounded(control_rate),
        "rate_excess_vs_control": _rounded(rate_excess),
        "thresholds": {
            "minimum_primary_rows": TACTICAL_MIN_ROWS,
            "minimum_primary_games": TACTICAL_MIN_GAMES,
            "primary_rate_min_exclusive": TACTICAL_QUIET_CHECK_MAJORITY_MIN_EXCLUSIVE,
            "rate_excess_vs_control_min_inclusive": TACTICAL_QUIET_CHECK_EXCESS_MIN,
        },
        "criteria": {
            "coverage_pass": coverage_pass,
            "strict_majority_pass": majority_pass,
            "excess_vs_control_pass": excess_pass,
        },
        "recommend_targeted_quiet_check_search_experiment": recommend,
        "interpretation_limit": (
            "Immediate reply eligibility under qsearch rules; not evidence that "
            "the recorded search reached a qsearch frontier at this position."
        ),
    }


def tactical_canaries(rows: Sequence[dict[str, Any]]) -> list[dict[str, Any]]:
    fields = (
        "canary_name",
        "game_index",
        "ply",
        "move_san",
        "reply_move_san",
        "loss_swing_cp",
        "loss_mate",
        "severe",
        "queen_net_danger",
        "reply_position_in_check",
        *TACTICAL_REPLY_FIELDS,
        "reply_qsearch_visibility_reason",
    )
    return [
        {field: row[field] for field in fields}
        for row in rows
        if row["canary_name"]
    ]


def canary_coverage(rows: Sequence[dict[str, Any]], canaries: dict[str, str]) -> dict[str, Any]:
    matched = sorted({row["canary_name"] for row in rows if row["canary_name"]})
    all_names = sorted(canaries.values())
    feature_coverage = {}
    for feature in ("bishop_attackers", "rook_attackers", "queen_attackers", "total_attackers"):
        names = sorted(
            {
                row["canary_name"]
                for row in rows
                if row["canary_name"] and row[feature] > 0
            }
        )
        feature_coverage[feature] = {"count": len(names), "names": names}
    return {
        "total": len(all_names),
        "matched_count": len(matched),
        "matched": matched,
        "unmatched": sorted(set(all_names) - set(matched)),
        "features": feature_coverage,
    }


def _mcnemar_exact(discordant_target: int, discordant_control: int) -> float:
    total = discordant_target + discordant_control
    if total == 0:
        return 1.0
    tail = sum(math.comb(total, k) for k in range(min(discordant_target, discordant_control) + 1))
    return min(1.0, 2 * tail / (2**total))


def canary_controls(rows: Sequence[dict[str, Any]]) -> dict[str, Any]:
    by_game_ply = {(row["game_index"], row["ply"]): row for row in rows}
    pairs = []
    for target in rows:
        if not target["canary_name"]:
            continue
        control = by_game_ply.get((target["game_index"], target["ply"] - 2))
        if control is not None:
            pairs.append((target, control))
    features = {}
    for feature in ("bishop_attackers", "rook_attackers", "queen_attackers", "total_attackers"):
        both = target_only = control_only = neither = 0
        for target, control in pairs:
            target_exposed = target[feature] > 0
            control_exposed = control[feature] > 0
            if target_exposed and control_exposed:
                both += 1
            elif target_exposed:
                target_only += 1
            elif control_exposed:
                control_only += 1
            else:
                neither += 1
        features[feature] = {
            "both_exposed": both,
            "target_only": target_only,
            "control_only": control_only,
            "neither": neither,
            "mcnemar_exact_p": _rounded(_mcnemar_exact(target_only, control_only)),
        }
    return {
        "pair_count": len(pairs),
        "paired_canaries": sorted(target["canary_name"] for target, _ in pairs),
        "control_definition": "same game, previous engine decision (ply - 2)",
        "features": features,
    }


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def write_csv(path: Path, rows: Sequence[dict[str, Any]]) -> None:
    fieldnames = [
        "game_index", "round", "engine_color", "ply", "fen", "move_uci", "move_san",
        "reply_move_uci", "reply_move_san",
        "engine_cp", "adverse_cp", "ref_is_mate", "ref_mate_distance", "ref_mate_sign",
        "loss_swing_cp", "loss_mate", "optimism_gap_cp", "severe",
        "non_pawn_material", "queen_net_danger",
        "king_in_check", "post_king_in_check", "reply_position_in_check",
        *TACTICAL_REPLY_FIELDS,
        "reply_qsearch_visibility_reason",
        "bishop_attackers", "bishop_hit_sum", "bishop_union_hits",
        "rook_attackers", "rook_hit_sum", "rook_union_hits",
        "queen_attackers", "queen_hit_sum", "queen_union_hits",
        "total_attackers", "total_hit_sum", "total_union_hits",
    ]
    for prefix in ("post_", "delta_", "placebo_"):
        for name, _ in PIECE_FEATURES:
            fieldnames.extend([
                f"{prefix}{name}_attackers", f"{prefix}{name}_hit_sum", f"{prefix}{name}_union_hits"
            ])
        fieldnames.extend([
            f"{prefix}total_attackers", f"{prefix}total_hit_sum", f"{prefix}total_union_hits"
        ])
    fieldnames.append("placebo_king_in_check")
    fieldnames.append("canary_name")
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=fieldnames, lineterminator="\n")
        writer.writeheader()
        writer.writerows(rows)


def build_report(
    rows: Sequence[dict[str, Any]],
    counts: dict[str, int],
    pgn_path: Path,
    canary_path: Path,
    canaries: dict[str, str],
    engine: str,
    swing_threshold_cp: int,
    mate_cap_cp: int,
) -> dict[str, Any]:
    midgame = [row for row in rows if row["non_pawn_material"] > 2000]
    noncheck = [row for row in rows if not row["king_in_check"]]
    midgame_noncheck = [
        row for row in rows if row["non_pawn_material"] > 2000 and not row["king_in_check"]
    ]
    midgame_noncheck_without_canaries = [
        row for row in midgame_noncheck if not row["canary_name"]
    ]
    return {
        "schema_version": 2,
        "inputs": {
            "pgn": {"path": str(pgn_path), "sha256": sha256_file(pgn_path)},
            "canaries": {"path": str(canary_path), "sha256": sha256_file(canary_path)},
            "analyzer_sha256": sha256_file(Path(__file__).resolve()),
        },
        "parameters": {
            "engine": engine,
            "swing_threshold_cp": swing_threshold_cp,
            "mate_cap_cp": mate_cap_cp,
            "midgame_non_pawn_material_min_exclusive": 2000,
            "bootstrap_replications": BOOTSTRAP_REPLICATIONS,
            "bootstrap_seed": BOOTSTRAP_SEED,
            "queen_net_danger_definition": (
                "queen_attackers - placebo_queen_attackers in PRE position"
            ),
            "qsearch_visibility_definition": (
                "all legal evasions when reply position is in check; otherwise captures "
                "admitted before delta pruning"
            ),
        },
        "counts": counts,
        "canary_coverage": canary_coverage(rows, canaries),
        "canary_controls": canary_controls(rows),
        "global": descriptive_stats(rows),
        "midgame": descriptive_stats(midgame),
        "noncheck": descriptive_stats(noncheck),
        "midgame_noncheck": descriptive_stats(midgame_noncheck),
        "midgame_noncheck_without_canaries": descriptive_stats(
            midgame_noncheck_without_canaries
        ),
        "placebo": {
            "global": descriptive_stats(rows, "placebo_"),
            "midgame_noncheck": descriptive_stats(midgame_noncheck, "placebo_"),
        },
        "post_delta_descriptive": {
            "global_post": descriptive_stats(rows, "post_"),
            "global_delta": descriptive_stats(rows, "delta_"),
            "midgame_noncheck_post": descriptive_stats(midgame_noncheck, "post_"),
            "midgame_noncheck_delta": descriptive_stats(midgame_noncheck, "delta_"),
        },
        "tactical_replies": {
            "global": tactical_reply_section(rows),
            "midgame_noncheck": tactical_reply_section(midgame_noncheck),
            "midgame_noncheck_without_canaries": tactical_reply_section(
                midgame_noncheck_without_canaries
            ),
            "canaries": tactical_canaries(rows),
            "decision": tactical_reply_decision(midgame_noncheck_without_canaries),
        },
        "mate_cap_sensitivity": sensitivity_stats(
            rows, swing_threshold_cp, (1500, 3000, mate_cap_cp)
        ),
    }


def parse_args(argv: Sequence[str] | None = None) -> argparse.Namespace:
    directory = Path(__file__).resolve().parent
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pgn", required=True, type=Path)
    parser.add_argument("--canaries", type=Path, default=directory / "canaries.csv")
    parser.add_argument("--engine", default="EquinoxChess")
    parser.add_argument("--csv-out", required=True, type=Path)
    parser.add_argument("--json-out", required=True, type=Path)
    parser.add_argument("--swing-threshold-cp", type=int, default=300)
    parser.add_argument("--mate-cap-cp", type=int, default=3000)
    args = parser.parse_args(argv)
    if args.swing_threshold_cp < 0:
        parser.error("--swing-threshold-cp must be non-negative")
    if args.mate_cap_cp <= 0:
        parser.error("--mate-cap-cp must be positive")
    return args


def main(argv: Sequence[str] | None = None) -> int:
    args = parse_args(argv)
    canaries = load_canaries(args.canaries)
    rows, counts = analyze_pgn(
        args.pgn,
        canaries,
        args.engine,
        args.swing_threshold_cp,
        args.mate_cap_cp,
    )
    write_csv(args.csv_out, rows)
    report = build_report(
        rows,
        counts,
        args.pgn,
        args.canaries,
        canaries,
        args.engine,
        args.swing_threshold_cp,
        args.mate_cap_cp,
    )
    args.json_out.parent.mkdir(parents=True, exist_ok=True)
    with args.json_out.open("w", encoding="utf-8") as handle:
        json.dump(report, handle, indent=2, sort_keys=True, allow_nan=False)
        handle.write("\n")
    print(
        f"kingzone: {counts['games']} games, {counts['plies']} plies, "
        f"{counts['rows']} rows, {counts['severe']} severe, "
        f"{report['canary_coverage']['matched_count']}/{report['canary_coverage']['total']} canaries"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
