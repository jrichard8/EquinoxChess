"""Measure a versioned Equinox UCI binary on the frozen tactical fixture."""

import argparse
import json
import platform
import subprocess
from collections import Counter
from pathlib import Path

import chess
import chess.engine
from build_corpus import (
    BUDGETS,
    HERE,
    OPENINGS,
    RUN6,
    digest,
    games,
    key,
    positions,
    tactical_line,
)

DEPTH = 6
REPEATS = 2


def load_fixture(path):
    data = json.loads(path.read_text())
    source = path.with_name("source.pgn")
    for file, field in (
        (source, "source_sha256"),
        (RUN6, "run6_sha256"),
        (OPENINGS, "opening_sha256"),
    ):
        if digest(file) != data[field]:
            raise ValueError(f"Input hash mismatch: {file}")
    excluded = {k for game in games(RUN6) for k in positions(game)}
    source_games = {int(g.headers["Round"]): g for g in games(source)}
    frequency = Counter(
        k for game in source_games.values() for k in set(positions(game))
    )
    cases = data["positions"]
    counts = Counter(c["cohort"] for c in cases)
    if counts != {"loss": 16, "tactic": 16} or len({c["game"] for c in cases}) != len(
        cases
    ):
        raise ValueError("Expected two cohorts of 16 independent games")
    if len({c["id"] for c in cases}) != len(cases) or len(
        {c["opening"] for c in cases}
    ) != len(cases):
        raise ValueError("Duplicate case identifier or opening")
    for c in cases:
        game = source_games[c["game"]]
        board = game.board()
        if key(board) in excluded or c["opening"] != game.headers["Opening"]:
            raise ValueError(f"{c['id']}: opening also appears in run6")
        for move in game.mainline_moves():
            if board.ply() + 1 == c["ply"]:
                break
            board.push(move)
        else:
            raise ValueError(f"{c['id']}: missing source ply")
        if (
            board.fen() != c["fen"]
            or move.uci() != c["played"]
            or key(board) in excluded
            or frequency[key(board)] != 1
        ):
            raise ValueError(f"{c['id']}: provenance or independence check failed")
        low, high = c["reference"]
        if low["best"] != high["best"] or [
            low["nodes_budget"],
            high["nodes_budget"],
        ] != list(BUDGETS):
            raise ValueError(f"{c['id']}: reference searches disagree")
        lo, hi = (-900, -300) if c["cohort"] == "loss" else (-100, 600)
        for ref in (low, high):
            pv = [chess.Move.from_uci(m) for m in ref["pv"]]
            replay = board.copy(stack=False)
            for move in pv:
                if move not in replay.legal_moves:
                    raise ValueError(f"{c['id']}: illegal reference variation")
                replay.push(move)
            if (
                not pv
                or pv[0].uci() != ref["best"]
                or not tactical_line(board, pv)
                or chess.Move.from_uci(ref["second"]) not in board.legal_moves
                or ref["second"] == ref["best"]
                or not lo <= ref["cp"] <= hi
                or ref["cp"] - ref["second_cp"] < 150
            ):
                raise ValueError(f"{c['id']}: reference criteria failed")
    return cases


def search(binary, fen):
    board = chess.Board(fen)
    with chess.engine.SimpleEngine.popen_uci(str(binary)) as engine:
        result = engine.play(
            board,
            chess.engine.Limit(depth=DEPTH),
            info=chess.engine.INFO_BASIC | chess.engine.INFO_SCORE,
        )
    info = result.info
    if (
        result.move not in board.legal_moves
        or info["depth"] != DEPTH
        or info["nodes"] <= 0
    ):
        raise RuntimeError(f"Incomplete or illegal result for {fen}")
    score = info["score"].pov(board.turn)
    return {
        "move": result.move.uci(),
        "cp": score.score(),
        "mate": score.mate(),
        "depth": info["depth"],
        "nodes": info["nodes"],
        "calls_loss": score <= chess.engine.Cp(-300),
    }


def build_info(binary):
    lines = subprocess.check_output(
        ["go", "version", "-m", str(binary)], text=True
    ).splitlines()
    settings = dict(
        line.strip()[len("build\t") :].split("=", 1)
        for line in lines[1:]
        if line.strip().startswith("build\t")
    )
    if "vcs.revision" not in settings:
        raise ValueError(
            "Missing Go VCS stamp. Build from a Git clone with -buildvcs=true."
        )
    return {
        "sha256": digest(binary),
        "go": lines[0].rsplit(": ", 1)[-1],
        "revision": settings["vcs.revision"],
        "modified": settings["vcs.modified"] == "true",
        "os": settings["GOOS"],
        "arch": settings["GOARCH"],
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=Path)
    parser.add_argument("output", type=Path)
    parser.add_argument("--fixture", type=Path, default=HERE / "v1/positions.json")
    args = parser.parse_args()
    if args.output.exists():
        parser.error("choose a new output file")
    binary = args.binary.resolve()
    provenance = build_info(binary)
    cases = load_fixture(args.fixture)
    results = []
    for case in cases:
        first = search(binary, case["fen"])
        if any(search(binary, case["fen"]) != first for _ in range(REPEATS - 1)):
            raise RuntimeError(f"Non-deterministic result for {case['id']}")
        best = case["reference"][-1]["best"]
        results.append(
            {
                "id": case["id"],
                "cohort": case["cohort"],
                "reference_move": best,
                "matches_reference": first["move"] == best,
                **first,
            }
        )
    summary = {}
    for cohort in ("tactic", "loss", "all"):
        subset = [r for r in results if cohort == "all" or r["cohort"] == cohort]
        summary[cohort] = {
            "n": len(subset),
            "matches_reference": sum(r["matches_reference"] for r in subset),
            "calls_loss": sum(r["calls_loss"] for r in subset),
            "nodes": sum(r["nodes"] for r in subset),
        }
    if digest(binary) != provenance["sha256"]:
        raise RuntimeError("Binary changed during measurement")
    report = {
        "fixture_sha256": digest(args.fixture),
        "binary": provenance,
        "python": platform.python_version(),
        "python_chess": chess.__version__,
        "depth": DEPTH,
        "identical_runs_per_case": REPEATS,
        "summary": summary,
        "results": results,
    }
    with args.output.open("x") as stream:
        stream.write(json.dumps(report, indent=2) + "\n")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
