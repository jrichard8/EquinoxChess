"""Build or verify the frozen quiet-move puzzle fixture independently of Equinox."""

import argparse
import csv
import hashlib
import json
from collections import Counter
from pathlib import Path

import chess
import chess.engine
from build_corpus import HERE, RUN6, digest, games, key, positions

V1_SOURCE = HERE / "v1/source.pgn"
PARQUET_SHA = "5352e3bff4d6163851c07a0cc922b1ff72875cc6f75d994bdfc3272a37ca9e87"
DATASET = "https://huggingface.co/datasets/Lichess/chess-puzzles/resolve/461725966ea46898ec6481372a6326de3281bfbd/data/train-00000-of-00003.parquet"
FIELDS = [
    "PuzzleId",
    "GameId",
    "FEN",
    "Moves",
    "Rating",
    "RatingDeviation",
    "Popularity",
    "NbPlays",
    "Themes",
]
DEPTHS = (14, 18)
QUOTA = 16


def exclude_previous():
    return {k for pgn in (V1_SOURCE, RUN6) for g in games(pgn) for k in positions(g)}


def puzzle(row):
    if (
        int(row["Rating"]) < 2000
        or int(row["RatingDeviation"]) > 100
        or int(row["Popularity"]) < 80
        or int(row["NbPlays"]) < 50
    ):
        return None
    moves = row["Moves"].split()
    if len(moves) < 6:
        return None
    board = chess.Board(row["FEN"])
    board.push_uci(moves[0])  # The published FEN precedes the opponent's setup move.
    solution = board.parse_uci(moves[1])
    if (
        board.is_check()
        or board.is_capture(solution)
        or board.gives_check(solution)
        or solution.promotion
    ):
        return None
    line = board.copy(stack=False)
    line.push(solution)
    tactic_follows = False
    for index, uci in enumerate(moves[2:]):
        move = line.parse_uci(uci)
        if index < 4 and (
            line.is_capture(move) or line.gives_check(move) or move.promotion
        ):
            tactic_follows = True
        line.push(move)
    return (board, solution.uci()) if tactic_follows else None


def export_pool(parquet, output):
    import pyarrow.parquet as pq  # Needed only to recover the pool from the upstream shard.

    if digest(parquet) != PARQUET_SHA:
        raise ValueError("Unexpected upstream shard")
    rows = next(pq.ParquetFile(parquet).iter_batches(batch_size=65536)).to_pylist()
    if len(rows) != 65536:
        raise ValueError("Incomplete source prefix")
    rows.sort(
        key=lambda r: hashlib.sha256(
            ("equinox-quiet-v2|" + r["PuzzleId"]).encode()
        ).digest()
    )
    seen = exclude_previous()
    game_ids = set()
    with output.open("x", newline="") as stream:
        writer = csv.DictWriter(
            stream, fieldnames=FIELDS, extrasaction="ignore", lineterminator="\n"
        )
        writer.writeheader()
        for row in rows:
            parsed = puzzle(row)
            game_id = row["GameId"].split("/")[0].split("#")[0]
            if parsed is None or game_id in game_ids or key(parsed[0]) in seen:
                continue
            seen.add(key(parsed[0]))
            game_ids.add(game_id)
            row["Themes"] = " ".join(row["Themes"] or [])
            writer.writerow(row)
            if len(game_ids) == 512:
                break
    if len(game_ids) != 512:
        raise ValueError("Incomplete candidate pool")


def rank_moves(engine, board, depth):
    # Take fresh UCI lines at this exact depth. Aggregated info may retain an old bound.
    exact = {}
    with engine.analysis(
        board, chess.engine.Limit(depth=depth), game=object(), multipv=2
    ) as analysis:
        for info in analysis:
            if info.get("depth") != depth or not info.get("pv") or "score" not in info:
                continue
            rank = info["multipv"]
            exact.pop(rank, None)
            cp = info["score"].pov(board.turn).score()
            if cp is not None and not (
                info.get("lowerbound") or info.get("upperbound")
            ):
                exact[rank] = cp, [m.uci() for m in info["pv"][:7]]
    if set(exact) != {1, 2}:
        return None
    (best_cp, pv), (second_cp, alternative) = exact[1], exact[2]
    if pv[0] == alternative[0] or best_cp - second_cp < 150:
        return None
    return {
        "depth": depth,
        "best": pv[0],
        "cp": best_cp,
        "pv": pv,
        "second": alternative[0],
        "second_cp": second_cp,
    }


def build(stockfish, pool, output):
    counts = Counter()
    selected = []
    with (
        pool.open() as stream,
        chess.engine.SimpleEngine.popen_uci(str(stockfish.resolve())) as engine,
    ):
        engine.configure({"Threads": 1, "Hash": 16})
        for row in csv.DictReader(stream):
            board, best = puzzle(row)
            side = "white" if board.turn else "black"
            if counts[side] == QUOTA:
                continue
            shallow = engine.play(
                board, chess.engine.Limit(nodes=512), game=object()
            ).move.uci()
            if shallow == best:
                continue
            refs = []
            for depth in DEPTHS:
                ref = rank_moves(engine, board, depth)
                if not ref or ref["best"] != best:
                    break
                refs.append(ref)
            if len(refs) == len(DEPTHS):
                counts[side] += 1
                selected.append(
                    {
                        "id": row["PuzzleId"],
                        "game": row["GameId"],
                        "cohort": side,
                        "fen": board.fen(),
                        "rating": int(row["Rating"]),
                        "shallow_move": shallow,
                        "reference": refs,
                    }
                )
                print(row["PuzzleId"], dict(counts), flush=True)
            if len(selected) == 2 * QUOTA:
                break
        name = engine.id["name"]
    if len(selected) != 2 * QUOTA:
        raise ValueError(f"Insufficient qualifying cases: {counts}")
    data = {
        "version": 2,
        "dataset": DATASET,
        "parquet_sha256": PARQUET_SHA,
        "pool_sha256": digest(pool),
        "v1_source_sha256": digest(V1_SOURCE),
        "run6_sha256": digest(RUN6),
        "reference_engine": name,
        "reference_sha256": digest(stockfish),
        "python_chess": chess.__version__,
        "positions": selected,
    }
    with output.open("x") as stream:
        stream.write(json.dumps(data, indent=2) + "\n")


def validate(path, data):
    pool = path.with_name("candidates.csv")
    for file, field in [
        (pool, "pool_sha256"),
        (V1_SOURCE, "v1_source_sha256"),
        (RUN6, "run6_sha256"),
    ]:
        if digest(file) != data[field]:
            raise ValueError(f"Input hash mismatch: {file}")
    with pool.open() as stream:
        rows = {r["PuzzleId"]: r for r in csv.DictReader(stream)}
    cases = data["positions"]
    if len(rows) != 512 or Counter(c["cohort"] for c in cases) != {
        "white": QUOTA,
        "black": QUOTA,
    }:
        raise ValueError("Unexpected v2 pool or cohort counts")
    seen = exclude_previous()
    game_ids = set()
    for case in cases:
        row = rows[case["id"]]
        board, best = puzzle(row)
        game = row["GameId"].split("/")[0].split("#")[0]
        if (
            game in game_ids
            or key(board) in seen
            or board.fen() != case["fen"]
            or case["game"] != row["GameId"]
            or case["rating"] != int(row["Rating"])
            or case["cohort"] != ("white" if board.turn else "black")
            or board.parse_uci(case["shallow_move"]).uci() == best
        ):
            raise ValueError(f"Provenance or independence failed: {case['id']}")
        game_ids.add(game)
        seen.add(key(board))
        if [r["depth"] for r in case["reference"]] != list(DEPTHS):
            raise ValueError("Missing reference depths")
        for ref in case["reference"]:
            if (
                ref["best"] != best
                or ref["pv"][0] != best
                or ref["cp"] - ref["second_cp"] < 150
                or board.parse_uci(ref["second"]).uci() == best
            ):
                raise ValueError(f"Invalid reference: {case['id']}")
            line = board.copy(stack=False)
            for move in ref["pv"]:
                line.push_uci(move)
    return cases


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="action", required=True)
    pool_parser = sub.add_parser("pool")
    pool_parser.add_argument("parquet", type=Path)
    pool_parser.add_argument("output", type=Path)
    label_parser = sub.add_parser("label")
    label_parser.add_argument("stockfish", type=Path)
    label_parser.add_argument("pool", type=Path)
    label_parser.add_argument("output", type=Path)
    args = parser.parse_args()
    if args.action == "pool":
        export_pool(args.parquet, args.output)
    else:
        build(args.stockfish, args.pool, args.output)
