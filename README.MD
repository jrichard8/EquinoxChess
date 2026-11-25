# MarsGopherChess

MarsGopherChess (formerly EquinoxChess) is a chess engine written in Go with a classical evaluation function. You can play against it in a browser or use it with any UCI-compatible chess GUI.

**The tournament engine uses `baseline-2200`.** It scored **60.16% against Stockfish 18 limited to 2200 Elo** at 1 second per game plus 0.1 seconds per move. Across 384 independent openings, each played with both colors, it recorded 430 wins, 64 draws and 274 losses. The estimated difference at this time control was +71.6 Elo (95% confidence interval +22.4 to +123.7). [Qualification games](tools/elo/2200/loop/README.md#2200-qualification) · [Rebuild the 2200+ baseline](tools/elo/2200/loop/README.md#rebuild-the-2200-baseline) · [Project status](status.md)

Against Stockfish at 2400, the baseline scored **38.67%** over 384 games. A candidate that rechecks a quiet defensive move scored **45.05%** on the same openings. Its head-to-head match against the baseline was inconclusive. [2400 campaign and games](tools/elo/2400/README.md)

Using only the wider search window, the candidate scored **46.22%** against the baseline and **42.97%** against Stockfish at 2400 on the same 192 openings. [Results and node counts](tools/elo/2400/window/README.md) · [Engine tournament entry](docs/tournois.md)

The previous baseline, `baseline-1850`, is available for comparison. The 2200+ baseline scored **69.01%** against it over 384 games. [1850 qualification](tools/elo/1850/README.md) · [First match against Stockfish at 2200](tools/elo/2200/README.md)

## Play in the browser

With Go 1.26 installed, start the server from the repository root, then open [localhost:8080](http://localhost:8080).

```sh
go run ./cmd/server
```

Drag a piece to play your move. The server keeps one game in memory, shared across browser tabs. The page loads jQuery and chessboard.js from CDNs.

| Request | Effect |
|---|---|
| `POST /newgame` | Resets the game to the starting position. |
| `POST /move` with `{"move":"e2e4"}` | Plays the requested move. |
| `POST /engine` with `{"depth":4}` | Searches for and plays a move, then returns `bestmove`, `score`, `depth` and `nodes`. |
| `GET /fen` | Returns the FEN, legal moves, side to move, and check and checkmate status. |

## Run the UCI engine

```sh
go build -o marsgopherchess-uci ./cmd/uci
./marsgopherchess-uci
```

Enter these commands in the engine's terminal to search for Black's reply to 1.e4. The engine responds with `uciok`, then `readyok`, then an `info` line followed by `bestmove`.

```text
uci
isready
position startpos moves e2e4
go depth 6
```

The engine also accepts FEN positions, a fixed time per move (`go movetime 500` for 500 ms), and UCI clocks with increments. Enter `quit` to exit.

## Development and testing

```sh
go test ./...
go test -tags searchtrace ./...
go test -tags 'searchtrace trace1850' ./...
```

- [Development guide](docs/developpement.md) — engine architecture, coding conventions and checks for each type of change.
- [Tactical suite v1](tools/tactics/README.md) covers regressions. [Tactical suite v2](tools/tactics/v2/README.md) tests harder quiet moves. Each has 32 positions and depth-6 results for the September baseline.
- [Elo match protocol](tools/elo/README.md) — compare two binaries by playing each opening with both colors.
- [Game analysis](tools/kingzone/README.md) and [search tracing](cmd/search_trace/README.md) — reconstruct positions from a PGN and inspect the engine's decisions.
