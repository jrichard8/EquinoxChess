"""Generate the 1850 confirmation openings with python-chess and single-thread Stockfish."""

import argparse
import hashlib
from pathlib import Path
import random

import chess
import chess.engine

HERE = Path(__file__).resolve().parent
SEED = 18502200
COUNT = 192


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("stockfish", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    seen = {
        " ".join(line.split()[:4])
        for line in (HERE.parent / "openings-v2.epd").read_text().splitlines()
    }
    rng = random.Random(SEED)
    positions = []
    with chess.engine.SimpleEngine.popen_uci(str(args.stockfish)) as engine:
        engine.configure({"Threads": 1, "Hash": 16})
        while len(positions) < COUNT:
            position = chess.Board()
            game = object()
            for _ in range(8):
                lines = engine.analyse(position, chess.engine.Limit(nodes=3000), multipv=5, game=game)
                best = lines[0]["score"].relative.score(mate_score=32000)
                choices = [
                    line["pv"][0] for line in lines
                    if best - line["score"].relative.score(mate_score=32000) <= 60
                ]
                position.push(rng.choice(choices))
            fen = position.fen()
            key = " ".join(fen.split()[:4])
            if key in seen:
                continue
            info = engine.analyse(position, chess.engine.Limit(nodes=16000), game=game)
            if abs(info["score"].white().score(mate_score=32000)) > 60:
                continue
            seen.add(key)
            positions.append(f'{fen} id "holdout-{len(positions) + 1:03d}";')
    with args.output.open("x") as stream:
        stream.write("\n".join(positions) + "\n")
    print(hashlib.sha256(args.output.read_bytes()).hexdigest())


if __name__ == "__main__":
    main()
