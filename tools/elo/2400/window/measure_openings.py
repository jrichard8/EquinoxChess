"""Measure cold searches at the median baseline match depth on all frozen opening positions."""
import argparse
import hashlib
import json
import sys
from pathlib import Path
from statistics import median

import chess
import chess.engine

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))
from replay_binaries import matches_recorded_binary


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def measure(engine, board, depth):
    result = engine.play(board, chess.engine.Limit(depth=depth), game=object(),
                         info=chess.engine.INFO_ALL)
    if result.info["depth"] != depth or result.move not in board.legal_moves:
        raise ValueError(f"Incomplete search on {board.fen()}")
    return dict(move=result.move.uci(), nodes=result.info["nodes"])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("base", "candidate", "output"):
        parser.add_argument(name, type=Path)
    args = parser.parse_args()
    campaign = Path(__file__).resolve().parents[1]
    epd = campaign / "baseline.epd"
    hashes = json.loads((campaign / "ranking-2400.json").read_text())["inputs"]
    base_sha, candidate_sha = digest(args.base), digest(args.candidate)
    if not (matches_recorded_binary(base_sha, hashes["base"])
            and matches_recorded_binary(candidate_sha, hashes["candidate"])):
        raise ValueError("Executables differ from the frozen cold replay")
    if digest(epd) != json.loads((campaign / "openings-audit.json").read_text())["sha256"]:
        raise ValueError("Unexpected opening suite")
    depth = int(json.loads((campaign / "base-audit.json").read_text())["median_depth"]["Base2200"])
    rows = []
    with (chess.engine.SimpleEngine.popen_uci(str(args.base.resolve())) as base,
          chess.engine.SimpleEngine.popen_uci(str(args.candidate.resolve())) as candidate):
        for n, line in enumerate(epd.read_text().splitlines(), 1):
            board = chess.Board(" ".join(line.split()[:4]) + " 0 1")
            rows.append(dict(round=n, fen=board.fen(), base=measure(base, board, depth),
                             candidate=measure(candidate, board, depth)))
    diffs = [r["candidate"]["nodes"] - r["base"]["nodes"] for r in rows]
    summary = dict(positions=len(rows), changed_moves=sum(r["candidate"]["move"] != r["base"]["move"] for r in rows),
                   more_nodes=sum(d > 0 for d in diffs), fewer_nodes=sum(d < 0 for d in diffs),
                   equal_nodes=diffs.count(0), median_node_difference=median(diffs),
                   total_nodes={k: sum(r[k]["nodes"] for r in rows) for k in ("base", "candidate")})
    report = dict(depth=depth, openings_sha256=digest(epd), base_sha256=base_sha,
                  candidate_sha256=candidate_sha, summary=summary, cases=rows)
    with args.output.open("x") as stream:
        json.dump(report, stream, indent=2)
        stream.write("\n")
    print(summary)


if __name__ == "__main__":
    main()
