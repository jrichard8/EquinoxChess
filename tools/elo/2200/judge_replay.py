"""Check the observational replay against the qualified binary and score alternative moves."""
import argparse
import json
import multiprocessing
from pathlib import Path

import chess
import chess.engine

from analyze_losses import VERIFY, decision, digest


def batch(job):
    binary, arbiter, cases = job
    with (chess.engine.SimpleEngine.popen_uci(binary) as equinox,
          chess.engine.SimpleEngine.popen_uci(arbiter) as stockfish):
        stockfish.configure({"Threads": 1, "Hash": 16, "UCI_LimitStrength": False})
        output = []
        for p, row in cases:
            b = chess.Board(p["opening"])
            for move in p["prefix"]:
                b.push_uci(move)
            assert b.fen() == p["fen"]
            baseline = row["Baseline"]
            result = equinox.play(b, chess.engine.Limit(depth=baseline["Depth"]), game=object(),
                                  info=chess.engine.INFO_ALL)
            if (result.move.uci(), result.info["nodes"], result.info["depth"]) != (
                baseline["Move"], baseline["Nodes"], baseline["Depth"]
            ):
                raise ValueError(f"Replay differs from qualified UCI: {row}")
            grades = {p[key]["move"]: p[key]["cp"] for key in ("best", "played")}
            for name in ("Baseline", "Deeper", "FullRoot", "HashFirst"):
                uci = row[name]["Move"]
                if uci not in grades:
                    grades[uci] = decision(stockfish, b, VERIFY, chess.Move.from_uci(uci))["cp"]
            output.append({"round": row["Round"], "color": row["Color"], "move_cp": grades})
    return output


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for key in ("losses", "replays", "binary", "arbiter", "output"):
        parser.add_argument(key, type=Path)
    parser.add_argument("--workers", type=int, default=4)
    args = parser.parse_args()
    if args.output.exists() or args.workers < 1:
        parser.error("choose a new output path and a positive worker count")
    losses = json.loads(args.losses.read_text())
    if (digest(args.binary) != "1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab"
            or digest(args.arbiter) != losses["arbiter_sha256"]):
        raise ValueError("Unexpected engine or arbiter")
    points = {(r["round"], r["color"]): r["turning_point"] for r in losses["losses"] if r["turning_point"]}
    cases = [(points.pop((r["Round"], r["Color"])), r) for r in json.loads(args.replays.read_text())]
    assert not points
    chunks = [(str(args.binary.resolve()), str(args.arbiter.resolve()), cases[i::args.workers])
              for i in range(args.workers)]
    with multiprocessing.Pool(args.workers) as pool:
        rows = [r for group in pool.map(batch, chunks) for r in group]
    rows.sort(key=lambda r: (r["round"], r["color"]))
    with args.output.open("x") as stream:
        json.dump(rows, stream, indent=2)
        stream.write("\n")


if __name__ == "__main__":
    main()
