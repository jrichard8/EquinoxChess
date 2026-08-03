# EquinoxChess

UCI chess engine written in Go, with a built-in HTTP server and a small web UI to play directly against the engine in a browser.

## Current strength

Measured at `1+0.1` against Stockfish capped at the named Elo:

| Opponent | Score | W-D-L | Elo estimate |
|----------|-------|-------|--------------|
| SF18-1600 (paired v2, 128 games, 2026-07-31) | `62.5%` | `78-4-46` | `+88.7 +/- 61.8` |
| SF1600 (run6, 2026-04-29) | `40.0%` | `12-0-18` | `-70.4 +/- 133.7` |

The current nominal calibration is therefore around `1690` at this very fast
time control. That is an estimate, not a certified absolute rating: the paired
Hoeffding interval remains inconclusive, and the run contained three Equinox
time forfeits versus two for Stockfish. The detailed protocol, uncertainty and
current development priorities are in [status.md](status.md).

## What's implemented

### Board and move generation

- Bitboard representation (`12` piece bitboards + occupancies)
- Magic-bitboard-style sliding piece attacks
- Make / unmake move with full undo state, en passant, castling, promotions
- Zobrist hashing for the transposition table and repetition detection
- Halfmove clock (`50`-move rule) and threefold repetition draw detection

### Search

- Negamax with alpha-beta
- Iterative deepening with aspiration windows (`+/- 50` cp)
- Transposition table with Zobrist key, exact / alpha / beta flags, move hint
- Null-move pruning (`R = 2`, depth `>= 5`, side has non-pawn material, score below mate threshold)
- Late Move Reductions (depth `>= 3`, after `3` moves tried, reduction `1/2/3` by depth)
- Futility pruning at depth `1`
- Move ordering : TT move, MVV-LVA on captures (king-attacker valued at `10000` so it tries last), promotions ranked alongside high captures, killer moves (`2` per ply), history heuristic with decay
- PV move tracking (preserved on fail-high at the root, fixed in Phase 1)
- Mate score band : scores within `MateScore - MateThreshold` are treated as mate distances, displayed as `score mate N` over UCI

### Quiescence

- Stand-pat with beta cut
- Captures and capture-promotions (with futility check on the captured piece)
- Forced full search of evasions when in check (no stand-pat)
- Stalemate recognized as draw at the qsearch frontier

### Evaluation

- Material : `P=100`, `N=320`, `B=330`, `R=500`, `Q=900`, `K=0`
- Piece-square tables for all pieces (mirrored for black)
- Bishop pair bonus `+30`
- Doubled pawns penalty `-8` per extra pawn on a file
- Isolated pawns penalty `-10` per pawn
- Mobility : pseudo-legal move count weighted per piece type
- Passed pawns bonus, scaling with rank from `+10` (rank 2) up to `+100` (rank 7)
- Simple king-shield safety in middlegame (disabled when non-pawn material drops below `2000`)

### UCI

- `uci`, `isready`, `ucinewgame`, `position startpos|fen ... moves ...`
- `go depth N`, `go movetime ms`
- `info depth ... score cp|mate ... nodes ...`, `bestmove ...`
- Fallback move if the search panics or is stopped before the first depth completes

### Tests and tooling

- Per-package unit tests (`board`, `movegen`, `perft`, `search`, `uci`)
- Perft suite for movegen correctness
- Two regression canaries on real PGN positions :
  - `TestSearchAvoidsKnownStalematingMoves` : `6` FEN where Equinox used to play stalemating moves with a winning eval (Phase 1 fix, all green)
  - `TestSearchSeesLossingTacticAtSwingPly` : `13` FEN where Equinox missed a tactic in run6 (Phase 2 work in progress)

## Build

```bash
go build -o equinoxchess-uci cmd/uci/main.go
```

## Run the web UI (play in a browser)

The HTTP server in `cmd/server` serves the UI from `web/` and exposes a small JSON API the page calls under the hood.

```bash
go run cmd/server/main.go
```

Then open http://localhost:8080 in any browser. The page shows a chessboard, an evaluation bar and a move log. Make a move by drag-and-drop ; the engine answers automatically.

The HTTP API exposed by the server :

| Method | Path | Body | Returns |
|--------|------|------|---------|
| `POST` | `/newgame` | — | `200` |
| `POST` | `/move` | `{"move": "e2e4"}` | `200`, applies the move (rejects illegal moves with `400`) |
| `POST` | `/engine` | `{"depth": 4}` | `{"bestmove": "...", "score": ..., "depth": ..., "nodes": ...}`, applies the engine's move |
| `GET` | `/fen` | — | `{"fen": "...", "legal_moves": [...], "in_check": ..., "checkmate": ..., "side_to_move": ...}` |

Quick smoke test from the shell :

```bash
curl -s http://localhost:8080/fen
curl -s -X POST http://localhost:8080/newgame
curl -s -X POST -H 'Content-Type: application/json' -d '{"depth":4}' http://localhost:8080/engine
```

## Run the UCI engine (cutechess-cli, Arena, etc.)

```bash
go build -o equinoxchess-uci cmd/uci/main.go
./equinoxchess-uci
```

Speak UCI on stdin :

```
uci
isready
position startpos
go depth 6
```

## Tests

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./...
```

## Reference gauntlet

For candidate-vs-baseline validation, use the fixed, color-swapped opening
protocol in [tools/elo/README.md](tools/elo/README.md). The Stockfish commands
below remain useful for absolute strength calibration.

Short regression run :

```bash
go build -o equinoxchess-uci cmd/uci/main.go && \
cutechess-cli \
  -engine cmd=./equinoxchess-uci name=EquinoxChess proto=uci \
  -engine cmd=stockfish name=SF1600 proto=uci option.UCI_LimitStrength=true option.UCI_Elo=1600 \
  -each tc=1+0.1 \
  -games 30
```

Same with PGN output (used for canari extraction and trajectory tracking) :

```bash
go build -o equinoxchess-uci cmd/uci/main.go && \
cutechess-cli \
  -engine cmd=./equinoxchess-uci name=EquinoxChess proto=uci \
  -engine cmd=stockfish name=SF1600 proto=uci option.UCI_LimitStrength=true option.UCI_Elo=1600 \
  -each tc=1+0.1 \
  -games 30 \
  -pgnout matches/sf1600-1+0.1-30.pgn
```

Longer validation run :

```bash
go build -o equinoxchess-uci cmd/uci/main.go && \
cutechess-cli \
  -engine cmd=./equinoxchess-uci name=EquinoxChess proto=uci \
  -engine cmd=stockfish name=SF1600 proto=uci option.UCI_LimitStrength=true option.UCI_Elo=1600 \
  -each tc=3+0.1 \
  -games 100
```

## Project layout

```
cmd/
  uci/        UCI binary (cutechess, Arena, ...)
  server/     HTTP server hosting the web UI
  list_moves/ debug tool : print legal moves for a FEN
  search_test/debug tool : run a search and print bestmove/score/nodes
internal/
  board/      bitboards, FEN parsing, make/unmake, repetition, zobrist
  move/       move encoding (from/to/moving/captured/promo/flags)
  movegen/    pseudo-legal and legal generators, magic attacks
  perft/      perft suite for movegen correctness
  search/     negamax, qsearch, evaluation, ordering, transposition table
  tt/         transposition table backing store
  uci/        UCI <-> internal move conversion
  zobrist/    Zobrist key tables
  api/        Game wrapper used by the HTTP server
web/          static UI (chessboard.js, single page, no build step)
matches/      archived PGN of regression gauntlets (run3 ... run6)
docs/         design notes and historical context
status.md     current status, canaris, leveraged fixes, next steps
```
