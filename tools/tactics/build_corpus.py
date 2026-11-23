"""Build the independent v1 fixture from new Stockfish games. Run from repo root."""

import argparse
import hashlib
import json
from pathlib import Path
from tempfile import TemporaryDirectory

import chess
import chess.engine
import chess.pgn

HERE = Path(__file__).resolve().parent
OPENINGS = HERE.parent / "elo/openings-v2.epd"
RUN6 = HERE.parents[1] / "matches/sf1600-run6-1+0.1-30.pgn"
QUOTA = 16
BUDGETS = (24000, 120000)
SEED = "equinox-tactics-v1"


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def games(path):
    with path.open() as stream:
        while (game := chess.pgn.read_game(stream)) is not None:
            if game.errors:
                raise ValueError(game.errors)
            yield game


def key(board):
    # Legal EP only. Ignore clocks, retain turn and castling rights.
    return " ".join(board.fen().split()[:4])


def positions(game):
    board = game.board()
    yield key(board)
    for move in game.mainline_moves():
        board.push(move)
        yield key(board)


def generate(engine, dest, excluded):
    with dest.open("x") as stream:
        for leg in range(2):
            for line in OPENINGS.read_text().splitlines():
                fen, name = line.split(' id "')
                name = name.removesuffix('";')
                number = int(name.split()[0])
                board = chess.Board(fen)
                if key(board) in excluded:
                    continue
                game = chess.pgn.Game()
                game.setup(board.copy())
                game.headers.update(
                    Event="Equinox tactical validation v1",
                    Round=str(number + leg * 64),
                    Opening=name,
                    White="SF18",
                    Black="SF18",
                )
                node = game
                for _ in range(80):
                    if board.is_game_over(claim_draw=True):
                        break
                    weak = board.turn == ((number + leg) % 2 == 1)
                    budget = 180 if weak else 5000
                    info = engine.analyse(
                        board, chess.engine.Limit(nodes=budget), game=object()
                    )
                    node = node.add_main_variation(info["pv"][0])
                    node.set_eval(info["score"])
                    board.push(node.move)
                game.headers["Result"] = board.result(claim_draw=True)
                if stream.tell():
                    stream.write("\n")
                print(game, file=stream)


def tactical_line(board, pv):
    board = board.copy(stack=False)
    for move in pv[:3]:
        if board.is_capture(move) or board.gives_check(move) or move.promotion:
            return True
        board.push(move)
    return False


def candidates(game, excluded, cohort):
    board = game.board()
    ranked = []
    for node in game.mainline():
        fen = board.fen()
        ply = board.ply() + 1  # absolute ply of the move about to be played
        estimate = node.eval().pov(board.turn).score()
        if (
            ply >= 17
            and key(board) not in excluded
            and estimate is not None
            and board.legal_moves.count() >= 2
        ):
            band = (-1300, -180) if cohort == "loss" else (-220, 900)
            if band[0] <= estimate <= band[1]:
                rank = hashlib.sha256(f"{SEED}|{fen}".encode()).hexdigest()
                ranked.append((rank, fen, ply, node.move.uci()))
        board.push(node.move)
    return sorted(ranked)


def reference(engine, board, budget, cohort):
    # A node limit may interrupt the next iteration between its two PVs.
    # Consume individual UCI lines so an older bound cannot stick to a later score.
    first = None
    pair = None
    with engine.analysis(
        board, chess.engine.Limit(nodes=budget), multipv=2, game=object()
    ) as analysis:
        for info in analysis:
            if not info.get("pv") or "score" not in info:
                continue
            exact = not (info.get("upperbound") or info.get("lowerbound"))
            if info["multipv"] == 1:
                first = info if exact else None
            elif (
                exact
                and first
                and info["depth"] == first["depth"]
                and info["pv"][0] != first["pv"][0]
            ):
                if pair is None or info["depth"] >= pair[0]["depth"]:
                    pair = first, info
                first = None
    if pair is None:
        return None
    top, second = pair
    best_cp, next_cp = (item["score"].pov(board.turn).score() for item in pair)
    if best_cp is None or next_cp is None:
        return None
    lo, hi = (-900, -300) if cohort == "loss" else (-100, 600)
    if (
        not lo <= best_cp <= hi
        or best_cp - next_cp < 150
        or not tactical_line(board, top["pv"])
    ):
        return None
    return {
        "nodes_budget": budget,
        "depth": top["depth"],
        "best": top["pv"][0].uci(),
        "cp": best_cp,
        "second": second["pv"][0].uci(),
        "second_cp": next_cp,
        "pv": [move.uci() for move in top["pv"][:5]],
    }


def select(engine, pgn, excluded):
    sources = list(games(pgn))
    # Remove positions shared by multiple generated games, including transpositions.
    seen, shared = set(), set()
    for game in sources:
        keys = set(positions(game))
        shared.update(seen & keys)
        seen.update(keys)
    excluded = excluded | shared
    counts = {"tactic": 0, "loss": 0}
    selected = []
    selected_openings = set()
    for game in sources:
        if game.headers["Opening"] in selected_openings:
            continue
        if min(counts.values()) == QUOTA:
            break
        number = int(game.headers["Round"])
        first = "loss" if number % 4 in (2, 3) else "tactic"
        for cohort in (first, "loss" if first == "tactic" else "tactic"):
            if counts[cohort] == QUOTA:
                continue
            for _, fen, ply, played in candidates(game, excluded, cohort):
                board = chess.Board(fen)
                low = reference(engine, board, BUDGETS[0], cohort)
                if low is None:
                    continue
                high = reference(engine, board, BUDGETS[1], cohort)
                if high is None or low["best"] != high["best"]:
                    continue
                selected.append(
                    {
                        "id": f"G{number:02d}-p{ply}",
                        "cohort": cohort,
                        "game": number,
                        "ply": ply,
                        "opening": game.headers["Opening"],
                        "fen": fen,
                        "played": played,
                        "reference": [low, high],
                    }
                )
                selected_openings.add(game.headers["Opening"])
                counts[cohort] += 1
                print(selected[-1]["id"], counts, flush=True)
                break
            else:
                continue
            break
    return sources, selected


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("stockfish", type=Path)
    parser.add_argument(
        "output", type=Path, help="new directory, or one containing only source.pgn"
    )
    args = parser.parse_args()
    args.output.mkdir(exist_ok=True, parents=True)
    pgn = args.output / "source.pgn"
    output = args.output / "positions.json"
    if output.exists():
        parser.error("positions.json already exists; choose another output directory")
    run6 = list(games(RUN6))
    excluded = {k for game in run6 for k in positions(game)}
    with chess.engine.SimpleEngine.popen_uci(str(args.stockfish.resolve())) as engine:
        engine.configure({"Hash": 16, "Threads": 1})
        if not pgn.exists():
            with TemporaryDirectory(dir=args.output) as temporary:
                generated = Path(temporary) / "source.pgn"
                generate(engine, generated, excluded)
                generated.rename(pgn)
        sources, selected = select(engine, pgn, excluded)
        data = {
            "version": 1,
            "reference_engine": engine.id["name"],
            "reference_sha256": digest(args.stockfish),
            "python_chess": chess.__version__,
            "opening_sha256": digest(OPENINGS),
            "run6_sha256": digest(RUN6),
            "run6_games": len(run6),
            "run6_unique_positions": len(excluded),
            "source_games": len(sources),
            "source_sha256": digest(pgn),
            "positions": selected,
        }
    if len(selected) != 2 * QUOTA:
        raise RuntimeError(f"Only {len(selected)} cases qualify; fixture not frozen")
    output.write_text(json.dumps(data, indent=2) + "\n")


if __name__ == "__main__":
    main()
