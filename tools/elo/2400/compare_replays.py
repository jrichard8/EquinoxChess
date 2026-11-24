"""Compare the full candidate with the window-only ablation on identical cold replays."""
import argparse
import json
from pathlib import Path


def compare(full, ranking):
    for key in ("losses", "base", "arbiter"):
        if full["inputs"][key] != ranking["inputs"][key]:
            raise ValueError(f"Different input: {key}")
    added_recoveries = added_losses = changed = 0
    base_nodes = ranking_nodes = full_nodes = 0
    for a, b in zip(full["cases"], ranking["cases"], strict=True):
        for key in ("round", "color", "fen", "reference", "depth", "base"):
            if a[key] != b[key]:
                raise ValueError(f"Inconsistent replay: {key}")
        window, verified = b["candidate"], a["candidate"]
        if window["move"] == verified["move"] and window["grade"] != verified["grade"]:
            raise ValueError("Different adjudication for the same move")
        changed += window["move"] != verified["move"]
        added_recoveries += window["grade"] <= -200 and verified["grade"] >= -100
        added_losses += window["grade"] >= -100 and verified["grade"] <= -200
        base_nodes += a["base"]["nodes"]
        ranking_nodes += window["nodes"]
        full_nodes += verified["nodes"]
    return dict(positions=len(full["cases"]), moves_changed_by_verification=changed,
                recovered_by_verification=added_recoveries, regressed_by_verification=added_losses,
                nodes_added_by_window_vs_base=ranking_nodes-base_nodes,
                nodes_added_by_verification_vs_window=full_nodes-ranking_nodes)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("full", type=Path)
    parser.add_argument("ranking", type=Path)
    args = parser.parse_args()
    print(json.dumps(compare(json.loads(args.full.read_text()), json.loads(args.ranking.read_text())), indent=2))
