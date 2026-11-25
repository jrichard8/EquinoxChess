"""Check one extra ply on 12 systematically spaced quiet defenses from the completed confirmation."""
import argparse
import json
from pathlib import Path
import sys

import chess
import chess.engine

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE.parent))
from analyze_losses import VERIFY, decision, digest
sys.path.insert(0, str(HERE.parents[1]))
from replay_binaries import matches_recorded_binary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("engine", "stockfish", "output"):
        parser.add_argument(name, type=Path)
    args = parser.parse_args()
    data = json.loads((HERE / "confirmation-losses.json").read_text())
    manifest = json.loads((HERE.parents[3] / "matches/elo2200/loop/interior-sf2200-confirm.pgn.meta.json").read_text())
    if (not matches_recorded_binary(digest(args.engine), manifest["engine_a_sha256"])
            or digest(args.stockfish) != data["arbiter_sha256"] or data["pgn_sha256"] != manifest["pgn_sha256"]):
        raise ValueError("Unexpected inputs")
    quiet = [r for r in data["losses"] if r["turning_point"] and r["turning_point"]["best_kind"] == "quiet"]
    rows = []
    with (chess.engine.SimpleEngine.popen_uci(str(args.engine.resolve())) as engine,
          chess.engine.SimpleEngine.popen_uci(str(args.stockfish.resolve())) as judge):
        judge.configure({"Threads": 1, "Hash": 16, "UCI_LimitStrength": False})
        for i in range(12):
            r = quiet[i*(len(quiet)-1)//11]
            p = r["turning_point"]
            board = chess.Board(p["opening"])
            for move in p["prefix"]:
                board.push_uci(move)
            if board.fen() != p["fen"]:
                raise ValueError(r["round"])
            grades = {p[k]["move"]: p[k]["cp"] for k in ("best", "played")}
            searches = []
            for depth in (p["game_depth"], p["game_depth"]+1):
                play = engine.play(board, chess.engine.Limit(depth=depth), game=object(), info=chess.engine.INFO_ALL)
                uci = play.move.uci()
                if play.info["depth"] != depth:
                    raise ValueError("Incomplete search")
                if uci not in grades:
                    grades[uci] = decision(judge, board, VERIFY, play.move)["cp"]
                searches.append(dict(depth=depth, move=uci, nodes=play.info["nodes"], grade=grades[uci]))
            rows.append(dict(round=r["round"], color=r["color"], fen=p["fen"], game_move=p["played"]["move"],
                             best=p["best"]["move"], searches=searches))
    with args.output.open("x") as stream:
        json.dump(rows, stream, indent=2)
        stream.write("\n")


if __name__ == "__main__":
    main()
