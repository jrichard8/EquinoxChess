"""Verify a completed paired PGN against the frozen EPD and its match manifest."""
import argparse
from collections import Counter, defaultdict
import hashlib
import json
from pathlib import Path
from statistics import median

import chess
import chess.pgn


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def audit(pgn, openings):
    meta = json.loads(pgn.with_suffix(".pgn.meta.json").read_text())
    if digest(pgn) != meta["pgn_sha256"] or digest(openings) != meta["openings_sha256"]:
        raise ValueError("PGN or EPD hash mismatch")
    count = meta["opening_count"]
    expected = [chess.Board(" ".join(s.split()[:4])+" 0 1").epd()
                for s in openings.read_text().splitlines()[:count]]
    if len(expected) != count or len(set(expected)) != count:
        raise ValueError("Wrong number of distinct openings")
    a, b = meta["engine_a_name"], meta["engine_b_name"]
    games = set()
    by_color = defaultdict(Counter)
    depths = defaultdict(list)
    with pgn.open() as stream:
        while (game := chess.pgn.read_game(stream)) is not None:
            round_no = int(game.headers["Round"])
            white, black = game.headers["White"], game.headers["Black"]
            key = (round_no, white)
            if game.errors or key in games or not 1 <= round_no <= count:
                raise ValueError(f"Invalid game {round_no}")
            games.add(key)
            if (white, black) not in ((a, b), (b, a)):
                raise ValueError(f"Incorrect colors in round {round_no}")
            if game.board().epd() != expected[round_no-1]:
                raise ValueError(f"Unexpected opening in round {round_no}")
            result = game.headers["Result"]
            if result not in ("1-0", "0-1", "1/2-1/2"):
                raise ValueError(f"Unfinished round {round_no}")
            by_color["white" if white == a else "black"][result] += 1
            for i, node in enumerate(game.mainline()):
                # Cutechess writes {score/depth 0.123s}; depth is absent for book moves.
                first = node.comment.split(maxsplit=1)
                if first and "/" in first[0]:
                    depth = first[0].rsplit("/", 1)[1]
                    if depth.isdigit():
                        depths[white if i % 2 == 0 else black].append(int(depth))
    if len(games) != 2*count:
        raise ValueError("Missing games")
    return dict(games=len(games), pairs=count, opening_order_and_colors_verified=True,
                results_by_color_of_a=by_color, median_depth={k: median(v) for k, v in depths.items()})


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("pgn", type=Path)
    parser.add_argument("openings", type=Path)
    args = parser.parse_args()
    print(json.dumps(audit(args.pgn, args.openings), indent=2))
