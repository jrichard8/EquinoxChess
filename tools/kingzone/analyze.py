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
                            row: dict[str, Any] = {
                                "game_index": game_index,
                                "round": game.headers.get("Round", ""),
                                "engine_color": "white" if board.turn == chess.WHITE else "black",
                                "ply": board.ply() + 1,
                                "fen": fen,
                                "move_uci": move.uci(),
                                "move_san": board.san(move),
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
        "engine_cp", "adverse_cp", "ref_is_mate", "ref_mate_distance", "ref_mate_sign",
        "loss_swing_cp", "loss_mate", "optimism_gap_cp", "severe",
        "non_pawn_material", "king_in_check", "post_king_in_check",
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
        "schema_version": 1,
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
