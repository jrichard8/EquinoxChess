"""Compare scores against the same opponent, using both colors of each shared opening."""
import argparse
from collections import defaultdict
import json
from pathlib import Path
import sys

import chess.pgn

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "2200" / "loop"))
from audit_match import digest


def scores(pgn):
    meta = json.loads(pgn.with_suffix(".pgn.meta.json").read_text())
    if digest(pgn) != meta["pgn_sha256"]:
        raise ValueError(f"Invalid PGN digest for {pgn}")
    pairs = defaultdict(dict)
    abnormal = set()
    with pgn.open() as stream:
        while headers := chess.pgn.read_headers(stream):
            key = int(headers["Round"]), headers["FEN"]
            white, black = headers["White"], headers["Black"]
            color = white == meta["engine_a_name"]
            if {white, black} != {meta["engine_a_name"], meta["engine_b_name"]} or color in pairs[key]:
                raise ValueError(f"Unexpected pairing in round {key[0]}")
            points = {"1-0": 1, "0-1": 0, "1/2-1/2": 0.5}[headers["Result"]]
            pairs[key][color] = points if color else 1-points
            if headers.get("Termination", "").strip().lower() not in ("", "normal"):
                abnormal.add(key)
    if len(pairs) != meta["opening_count"] or any(len(pair) != 2 for pair in pairs.values()):
        raise ValueError(f"Incomplete opening pairs in {pgn}")
    return meta, {key: sum(pair.values()) for key, pair in pairs.items()}, abnormal


def summarize(reference, candidate):
    a, before, first_incidents = scores(reference)
    b, after, second_incidents = scores(candidate)
    for k in ("engine_b_sha256", "engine_b_name", "cutechess_sha256", "time_control", "concurrency",
              "openings_sha256", "opening_count"):
        if a[k] != b[k]:
            raise ValueError(f"Match manifests differ on {k}")
    def uci_options(meta):
        return sorted(arg for arg in meta["cutechess_args"] if arg.startswith("option."))
    if uci_options(a) != uci_options(b) or before.keys() != after.keys():
        raise ValueError("Different engine options or opening pairs")
    changes = {key: after[key] - value for key, value in before.items()}
    clean = before.keys() - (first_incidents | second_incidents)
    return dict(reference=a["engine_a_name"], candidate=b["engine_a_name"],
                pgn_sha256=[a["pgn_sha256"], b["pgn_sha256"]], pairs=len(before),
                better=sum(v>0 for v in changes.values()), equal=sum(v==0 for v in changes.values()),
                worse=sum(v<0 for v in changes.values()), points_difference=sum(changes.values()),
                score_percentage_points=100*sum(changes.values())/(2*len(changes)),
                clean_pairs=len(clean), clean_percentage_points=100*sum(changes[k] for k in clean)/(2*len(clean)) if clean else None)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("reference", type=Path)
    parser.add_argument("candidate", type=Path)
    args = parser.parse_args()
    print(json.dumps(summarize(args.reference, args.candidate), indent=2))
