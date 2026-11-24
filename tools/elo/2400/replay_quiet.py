"""Replay all adjudicated quiet defenses at game depth with identical history and cold search state."""
import argparse
import json
from pathlib import Path
import sys

import chess
import chess.engine

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "2200"))
from analyze_losses import VERIFY, decision, digest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("losses", "base", "candidate", "arbiter", "output"):
        parser.add_argument(name, type=Path)
    args = parser.parse_args()
    losses = json.loads(args.losses.read_text())
    if digest(args.arbiter) != losses["arbiter_sha256"]:
        raise ValueError("Unexpected arbiter")
    rows = []
    with (chess.engine.SimpleEngine.popen_uci(str(args.base.resolve())) as base,
          chess.engine.SimpleEngine.popen_uci(str(args.candidate.resolve())) as candidate,
          chess.engine.SimpleEngine.popen_uci(str(args.arbiter.resolve())) as arbiter):
        arbiter.configure({"Threads": 1, "Hash": 16, "UCI_LimitStrength": False})
        for loss in losses["losses"]:
            p = loss["turning_point"]
            if p is None or p["best_kind"] != "quiet":
                continue
            board = chess.Board(p["opening"])
            for move in p["prefix"]:
                board.push_uci(move)
            if board.fen() != p["fen"]:
                raise ValueError("Position history mismatch")
            grades = {p[k]["move"]: p[k]["cp"] for k in ("best", "played")}
            row = dict(round=loss["round"], color=loss["color"], ply=p["ply"], fen=p["fen"],
                       depth=p["game_depth"], reference=p["best"]["move"], played=p["played"]["move"])
            for name, engine in (("base", base), ("candidate", candidate)):
                result = engine.play(board, chess.engine.Limit(depth=p["game_depth"]),
                                     game=object(), info=chess.engine.INFO_ALL)
                if result.info["depth"] != p["game_depth"]:
                    raise ValueError("Incomplete search")
                uci = result.move.uci()
                if uci not in grades:
                    grades[uci] = decision(arbiter, board, VERIFY, result.move)["cp"]
                row[name] = dict(move=uci, grade=grades[uci], nodes=result.info["nodes"])
            rows.append(row)
    summary = dict(positions=len(rows), changes=sum(r["base"]["move"] != r["candidate"]["move"] for r in rows),
                   recovered=sum(r["base"]["grade"] <= -200 and r["candidate"]["grade"] >= -100 for r in rows),
                   regressed=sum(r["base"]["grade"] >= -100 and r["candidate"]["grade"] <= -200 for r in rows))
    for name in ("base", "candidate"):
        summary[name] = dict(nodes=sum(r[name]["nodes"] for r in rows),
                             references=sum(r[name]["move"] == r["reference"] for r in rows))
    with args.output.open("x") as out:
        json.dump(dict(inputs={n: digest(getattr(args,n)) for n in ("losses","base","candidate","arbiter")},
                       summary=summary, cases=rows), out, indent=2)
        out.write("\n")
    print(summary)


if __name__ == "__main__":
    main()
