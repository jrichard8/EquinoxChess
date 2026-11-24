"""Rescore changed opening choices at two fixed Stockfish budgets, from the side to move."""
import argparse
import json
from pathlib import Path
import sys

import chess
import chess.engine

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "2200"))
from analyze_losses import decision, digest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("openings", "stockfish", "output"):
        parser.add_argument(name, type=Path)
    args = parser.parse_args()
    inputs = json.loads(args.openings.read_text())
    manifest = Path(__file__).resolve().parents[1] / "ranking-2400.json"
    if digest(args.stockfish) != json.loads(manifest.read_text())["inputs"]["arbiter"]:
        raise ValueError("Unexpected arbiter")
    rows = []
    with chess.engine.SimpleEngine.popen_uci(str(args.stockfish.resolve())) as engine:
        engine.configure({"Threads": 1, "Hash": 16, "UCI_LimitStrength": False})
        for case in inputs["cases"]:
            if case["base"]["move"] == case["candidate"]["move"]:
                continue
            board = chess.Board(case["fen"])
            scores = {}
            for budget in (80_000, 320_000):
                scores[budget] = {k: decision(engine, board, budget, chess.Move.from_uci(case[k]["move"]))
                                  for k in ("base", "candidate")}
            differences = [s["candidate"]["cp"] - s["base"]["cp"] for s in scores.values()]
            rows.append(dict(round=case["round"], fen=case["fen"], scores=scores,
                             gain_50cp_both=min(differences) >= 50, loss_50cp_both=max(differences) <= -50))
    summary = dict(changed=len(rows), gains=sum(r["gain_50cp_both"] for r in rows),
                   losses=sum(r["loss_50cp_both"] for r in rows))
    report = dict(inputs_sha256=digest(args.openings), arbiter_sha256=digest(args.stockfish),
                  summary=summary, cases=rows)
    with args.output.open("x") as stream:
        json.dump(report, stream, indent=2)
        stream.write("\n")
    print(summary)


if __name__ == "__main__":
    main()
