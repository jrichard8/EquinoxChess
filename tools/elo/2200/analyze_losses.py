"""Locate and independently rescore the first confirmed turning point per lost game."""
import argparse
import hashlib
import io
import json
import multiprocessing
import re
from collections import Counter
from pathlib import Path

import chess
import chess.engine
import chess.pgn

SCREEN = 5_000
VERIFY = 80_000
NAME = "Equinox1850"


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def decision(engine, board, nodes, only=None):
    last = None
    with engine.analysis(board, chess.engine.Limit(nodes=nodes), game=object(),
                         root_moves=[only] if only else None) as analysis:
        for iteration in analysis:
            if ("score" in iteration and iteration.get("pv")
                    and not iteration.get("lowerbound") and not iteration.get("upperbound")):
                last = iteration
    if last is None:
        raise RuntimeError(f"No completed score for {board.fen()}")
    info = last
    score = info["score"].pov(board.turn)
    pv = info["pv"]
    return {"move": pv[0].uci(), "cp": score.score(mate_score=32000),
            "mate": score.mate(), "depth": info["depth"],
            "pv": [m.uci() for m in pv[:8]]}


def scores(engine, board, played, nodes):
    best = decision(engine, board, nodes)
    actual = best if best["move"] == played.uci() else decision(engine, board, nodes, played)
    return best, actual


def kind(board, move):
    if move.promotion:
        return "promotion"
    if board.is_capture(move):
        return "capture"
    if board.gives_check(move):
        return "quiet_check"
    return "quiet"


def analyze(engine, game, exhaustive):
    color = game.headers["White"] == NAME
    board = game.board()
    row = {"round": int(game.headers["Round"]), "color": chess.COLOR_NAMES[color],
           "end": game.end().comment, "game_plies": len(list(game.mainline_moves())),
           "screened_decisions": 0, "verified_decisions": 0, "turning_point": None}
    if exhaustive:
        row["first_crossing"] = None
        row["already_inferior_ply"] = None
    prefix = []
    for ply, node in enumerate(game.mainline(), 1):
        played = node.move
        if board.turn == color:
            verify = exhaustive
            if not exhaustive:
                row["screened_decisions"] += 1
                best, actual = scores(engine, board, played, SCREEN)
                verify = (best["cp"] >= -180 and actual["cp"] <= -125
                          and best["cp"] - actual["cp"] >= 100)
            if verify:
                row["verified_decisions"] += 1
                best, actual = scores(engine, board, played, VERIFY)
                if exhaustive:
                    if best["cp"] <= -200 and row["already_inferior_ply"] is None:
                        row["already_inferior_ply"] = ply
                    if best["cp"] > -200 >= actual["cp"] and row["first_crossing"] is None:
                        row["first_crossing"] = {"ply": ply, "fen": board.fen(),
                                                 "best": best, "played": actual}
                if best["cp"] >= -100 and actual["cp"] <= -200 and best["cp"] - actual["cp"] >= 150:
                    ref = chess.Move.from_uci(best["move"])
                    depth = re.search(r"/(\d+)\s", node.comment)
                    row["turning_point"] = {
                        "ply": ply, "fen": board.fen(), "opening": game.board().fen(),
                        "prefix": prefix, "played": actual, "best": best,
                        "game_depth": int(depth[1]) if depth else None,
                        "game_annotation": node.comment, "in_check": board.is_check(),
                        "pieces": len(board.piece_map()), "played_kind": kind(board, played),
                        "best_kind": kind(board, ref), "played_san": board.san(played),
                        "best_san": board.san(ref),
                    }
                    return row
        board.push(played)
        prefix.append(played.uci())
    return row


def batch(job):
    binary, games, exhaustive = job
    with chess.engine.SimpleEngine.popen_uci(binary) as engine:
        engine.configure({"Threads": 1, "Hash": 16, "UCI_LimitStrength": False})
        return [analyze(engine, chess.pgn.read_game(io.StringIO(game)), exhaustive) for game in games]


def combine_passes(original, exhaustive_rows, first_pass_sha256):
    extra = {(r["round"], r["color"]): r for r in exhaustive_rows}
    losses = []
    for first in original["losses"]:
        if first["turning_point"]:
            losses.append(dict(first, source="screen"))
        else:
            second = extra.pop((first["round"], first["color"]))
            losses.append(dict(second, source="exhaustive",
                               screened_decisions=first["screened_decisions"],
                               first_pass_verifications=first["verified_decisions"]))
    if extra:
        raise ValueError("Unexpected supplemental decisions")
    return dict(original, losses=losses, first_pass_sha256=first_pass_sha256)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("pgn", type=Path)
    parser.add_argument("stockfish", type=Path)
    parser.add_argument("output", type=Path)
    parser.add_argument("--workers", type=int, default=4)
    parser.add_argument("--exhaustive-misses", type=Path,
                        help="Recheck every move in previously unclassified defeats directly at 80000 nodes")
    args = parser.parse_args()
    if args.output.exists() or args.workers < 1:
        parser.error("choose a new output path and a positive worker count")
    meta = json.loads(Path(str(args.pgn) + ".meta.json").read_text())
    if meta["pgn_sha256"] != digest(args.pgn) or meta["engine_b_sha256"] != digest(args.stockfish):
        raise ValueError("PGN or arbiter differs from match manifest")
    previous = None
    if args.exhaustive_misses:
        original = json.loads(args.exhaustive_misses.read_text())
        if original["pgn_sha256"] != meta["pgn_sha256"] or original["arbiter_sha256"] != meta["engine_b_sha256"]:
            raise ValueError("Different first-pass inputs")
        previous = {(r["round"], r["color"]) for r in original["losses"] if not r["turning_point"]}
    games = []
    outcomes = Counter()
    with args.pgn.open() as stream:
        while (game := chess.pgn.read_game(stream)) is not None:
            if game.errors or {game.headers["White"], game.headers["Black"]} != {NAME, "SF2200"}:
                raise ValueError(game.headers)
            result = game.headers["Result"]
            if result not in ("1-0", "0-1", "1/2-1/2"):
                raise ValueError(result)
            loss = result == ("0-1" if game.headers["White"] == NAME else "1-0")
            outcomes["loss" if loss else "draw" if result == "1/2-1/2" else "win"] += 1
            key = (int(game.headers["Round"]), chess.COLOR_NAMES[game.headers["White"] == NAME])
            if loss and (previous is None or key in previous):
                games.append(str(game))
    if outcomes.total() != 2 * meta["opening_count"]:
        raise ValueError(outcomes)
    chunks = [(str(args.stockfish.resolve()), games[i::args.workers], previous is not None) for i in range(args.workers)]
    with multiprocessing.Pool(args.workers) as pool:
        rows = [row for group in pool.map(batch, chunks) for row in group]
    rows.sort(key=lambda r: (r["round"], r["color"]))
    assert len({(r["round"], r["color"]) for r in rows}) == (outcomes["loss"] if previous is None else len(previous))
    output = {"pgn_sha256": meta["pgn_sha256"], "arbiter_sha256": meta["engine_b_sha256"],
              "screen_nodes": SCREEN, "verification_nodes": VERIFY, "outcomes": outcomes, "losses": rows}
    if previous is not None:
        output = combine_passes(original, rows, digest(args.exhaustive_misses))
    with args.output.open("x") as stream:
        json.dump(output, stream, indent=2)
        stream.write("\n")


if __name__ == "__main__":
    main()
