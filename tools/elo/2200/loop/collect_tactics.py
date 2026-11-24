"""Collect fixed-depth UCI and run6 probe results in a compact per-position report."""
import argparse
import json
from pathlib import Path


def read(path):
    return json.loads(path.read_text())


def collect(v1, v2, run6):
    if v1["binary"] != v2["binary"]:
        raise ValueError("Different UCI binaries")
    cases = {}
    for suite, report in (("v1", v1), ("v2", v2)):
        if report["depth"] != 6 or len(report["results"]) != 32:
            raise ValueError(f"Incomplete {suite}")
        cases[suite] = {r["id"]: {k: r[k] for k in ("move", "cp", "mate", "nodes", "matches_reference", "calls_loss")}
                        for r in report["results"]}
    cases["run6"] = {}
    for case in run6["cases"]:
        run, = case["runs"]
        if run["completed_depth"] != 6 or not run["non_interference_verified"]:
            raise ValueError(case["name"])
        cases["run6"][case["name"]] = dict(move=run["best_move_uci"], score=run["score"],
                                          nodes=run["nodes"], avoids=not run["root_move_best"],
                                          calls_loss=run["score"] <= -300)
    if len(cases["run6"]) != 13:
        raise ValueError("Incomplete canaries")
    return dict(binary=v1["binary"], trace_sha256=run6["inputs"]["executable"]["sha256"],
                fixtures_sha256=[v1["fixture_sha256"], v2["fixture_sha256"], run6["inputs"]["csv"]["sha256"]],
                depth=6, cases=cases)


if __name__ == "__main__":
    p = argparse.ArgumentParser(description=__doc__)
    for name in ("v1", "v2", "run6", "output"):
        p.add_argument(name, type=Path)
    a = p.parse_args()
    with a.output.open("x") as out:
        json.dump(collect(read(a.v1), read(a.v2), read(a.run6)), out, indent=2)
        out.write("\n")
