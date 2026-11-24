"""Check exact UCI parity and time 77 cold searches per engine in alternating order."""
import argparse
import csv
import hashlib
import json
from pathlib import Path
from statistics import median
from time import perf_counter_ns

import chess
import chess.engine

ROOT = Path(__file__).resolve().parents[4]


def cases():
    for suite in ("v1", "v2"):
        for r in json.loads((ROOT / "tools/tactics" / suite / "positions.json").read_text())["positions"]:
            yield r["id"], r["fen"]
    with (ROOT / "tools/kingzone/canaries.csv").open() as stream:
        for r in csv.DictReader(stream):
            yield r["canary_name"], r["fen"]


def measure(binary, fen):
    with chess.engine.SimpleEngine.popen_uci(str(binary)) as engine:
        engine.ping()
        start = perf_counter_ns()
        play = engine.play(chess.Board(fen), chess.engine.Limit(depth=6), info=chess.engine.INFO_ALL)
        elapsed = perf_counter_ns() - start
        if play.info["depth"] != 6:
            raise ValueError("Incomplete search")
        return (play.move.uci(), str(play.info["score"].relative), play.info["nodes"]), elapsed


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("candidate", "reference", "output"):
        parser.add_argument(name, type=Path)
    args = parser.parse_args()
    binaries = (args.candidate.resolve(), args.reference.resolve())
    if binaries[0].name == binaries[1].name:
        parser.error("use distinct binary names")
    rows = []
    for n, (key, fen) in enumerate(cases()):
        times = {p.name: [] for p in binaries}
        signature = None
        for repeat in range(2):
            order = binaries if (n + repeat) % 2 else binaries[::-1]
            for p in order:
                observed, elapsed = measure(p, fen)
                if signature is not None and observed != signature:
                    raise ValueError(f"Search differs on {key}: {observed} versus {signature}")
                signature = observed
                times[p.name].append(elapsed)
        rows.append(dict(id=key, times_ns=times, signature=signature))
    report = dict(sha256={p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in binaries},
                  seconds={p.name: sum(median(r["times_ns"][p.name]) for r in rows)/1e9 for p in binaries},
                  positions=rows)
    with args.output.open("x") as stream:
        json.dump(report, stream, indent=2)
        stream.write("\n")


if __name__ == "__main__":
    main()
