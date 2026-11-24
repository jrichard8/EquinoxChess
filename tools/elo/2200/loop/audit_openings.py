"""Audit the confirmation positions against the repository as it stood before the loop."""
import argparse
import hashlib
import io
import json
from pathlib import Path
import subprocess

import chess
import chess.pgn


def epd_positions(data):
    return {chess.Board(" ".join(s.split()[:4])+" 0 1").epd() for s in data.splitlines() if s.strip()}


def historical_positions(ref):
    files = subprocess.check_output(["git", "ls-tree", "-r", "--name-only", ref], text=True).splitlines()
    old = set()
    for path in files:
        if not (path.endswith((".epd", ".pgn")) or path in
                ("tools/tactics/v1/positions.json", "tools/tactics/v2/positions.json")):
            continue
        data = subprocess.check_output(["git", "show", f"{ref}:{path}"]).decode()
        if path.endswith(".pgn"):
            stream = io.StringIO(data)
            while (game := chess.pgn.read_game(stream)) is not None:
                if game.errors:
                    raise ValueError(f"Invalid historical game in {path}")
                board = game.board()
                old.add(board.epd())
                for move in game.mainline_moves():
                    board.push(move)
                    old.add(board.epd())
        elif path.endswith(".json"):
            old.update(chess.Board(row["fen"]).epd() for row in json.loads(data)["positions"])
        else:
            old.update(epd_positions(data))
    return old


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("epd", type=Path)
    parser.add_argument("--archive", default="c5f6da9", help="historical Git revision to exclude")
    args = parser.parse_args()
    ref = subprocess.check_output(["git", "rev-parse", args.archive + "^{commit}"], text=True).strip()
    raw = args.epd.read_text()
    new = epd_positions(raw)
    old = historical_positions(ref)
    overlap = sorted(new & old)
    print(json.dumps(dict(archive_at=ref, unique_historical_positions=len(old),
                          positions=len(new), overlap=overlap,
                          sha256=hashlib.sha256(args.epd.read_bytes()).hexdigest()), indent=2))
    if overlap or len(new) != len(raw.splitlines()):
        raise SystemExit("Overlapping or duplicate openings")
