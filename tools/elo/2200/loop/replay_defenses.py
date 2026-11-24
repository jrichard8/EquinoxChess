"""Replay a candidate on the frozen 92 SF2200 loss positions at the base engine's depth."""
import argparse
import json
from pathlib import Path
import sys

import chess
import chess.engine

DATA = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(DATA))
from analyze_losses import VERIFY, decision, digest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("engine", "stockfish", "output"):
        parser.add_argument(name, type=Path)
    args = parser.parse_args()
    losses = json.loads((DATA / "losses.json").read_text())
    if digest(args.stockfish) != losses["arbiter_sha256"]:
        raise ValueError("Different arbiter")
    positions = {(r["round"], r["color"]): r["turning_point"] for r in losses["losses"] if r["turning_point"]}
    grades = {(r["round"], r["color"]): r["move_cp"] for r in json.loads((DATA / "grades.json").read_text())}
    rows = []
    with (chess.engine.SimpleEngine.popen_uci(str(args.engine.resolve())) as engine,
          chess.engine.SimpleEngine.popen_uci(str(args.stockfish.resolve())) as arbiter):
        arbiter.configure({"Threads": 1, "Hash": 16, "UCI_LimitStrength": False})
        for r in json.loads((DATA / "search.json").read_text()):
            k = (r["Round"], r["Color"])
            p = positions[k]
            board = chess.Board(p["opening"])
            for uci in p["prefix"]:
                board.push_uci(uci)
            if board.fen() != p["fen"]:
                raise ValueError(k)
            base = r["Baseline"]
            played = engine.play(board, chess.engine.Limit(depth=base["Depth"]), game=object(),
                                 info=chess.engine.INFO_ALL)
            if played.info["depth"] != base["Depth"]:
                raise ValueError("Incomplete search")
            move = played.move.uci()
            grade = grades[k].get(move)
            if grade is None:
                grade = decision(arbiter, board, VERIFY, played.move)["cp"]
            rows.append(dict(round=k[0], color=k[1], depth=base["Depth"], base_move=base["Move"],
                             base_grade=grades[k][base["Move"]], base_nodes=base["Nodes"],
                             move=move, grade=grade, nodes=played.info["nodes"]))
    if len(rows) != len(positions):
        raise ValueError("Incomplete diagnostic")
    with args.output.open("x") as out:
        json.dump(dict(engine_sha256=digest(args.engine), arbiter_sha256=digest(args.stockfish),
                       cases=rows), out, indent=2)
        out.write("\n")


if __name__ == "__main__":
    main()
