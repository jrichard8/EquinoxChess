# Developing MarsGopherChess

## Follow a move through the engine

1. Both UCI commands and HTTP requests use `api.Game`, which holds the board and the search state. In UCI, `position` replaces the board while retaining the transposition table and search heuristics. `ucinewgame` resets the game. [Game API](../internal/api/game.go) · [UCI loop](../cmd/uci/main.go)
2. The board tracks twelve piece bitboards, occupancy, the side to move, castling rights, the en passant square, move counters and the Zobrist key. An undo stack lets the search make and unmake moves. Each move fits in 32 bits. For sliding pieces, the generator follows each ray to the first blocker. [Board](../internal/board) · [Move encoding](../internal/move) · [Move generation](../internal/movegen/movegen.go)
3. The single-threaded search uses iterative deepening, negamax with alpha-beta pruning, and aspiration windows. Below the root, it tries the transposition table move first. It orders the remaining moves using captures, killer moves and the history heuristic. Null-move pruning, late move reductions (LMR) and futility pruning limit the branches explored. At the horizon, quiescence searches captures when the king is safe and evasions when in check. [Search](../internal/search/search.go) · [Transposition table](../internal/tt)
4. Static evaluation combines material, piece-square tables, the bishop pair, pawn structure, passed pawns, mobility and the king's pawn shield. The experimental tools include king-pressure diagnostics and search reduction variants. [Evaluation](../internal/search/eval.go) · [Search tracer](../cmd/search_trace/README.md)

## Make and validate a change

Keep the search deterministic. Use concrete types, preallocated move arrays and a fixed-size transposition table. Keep allocation, logging and goroutines out of the search loop. For an optimization, measure allocations, elapsed time and nodes at fixed depth, then compare playing strength at equal time.

| Change | Check from the repository root |
|---|---|
| Board or move generation | `go test -count=1 ./internal/board ./internal/movegen ./internal/perft`. Perft counts legal leaf nodes. Expect 4,865,609 from the starting position at depth 5 and 4,085,603 from Kiwipete at depth 4. |
| Search or evaluation | `go test -count=1 ./...`. For per-position results, run `go test -count=1 -v ./internal/search -run 'TestSearchTacticalCanaryMetrics\|TestSearchAvoidsKnownStalematingMoves'`. Add a regression position or an invariant that exercises the change. |
| Instrumentation | `go test -count=1 -tags searchtrace ./...`. Replay historical search paths with `go test -count=1 -tags 'searchtrace trace1850' ./...`. In observation mode, the tracer checks the move, score, depth and node count against an untraced search. |
| UCI and static checks | `go test -race -count=1 ./cmd/uci` and `go vet ./...`. |
| Search cost | Run `go test -run '^$' -bench BenchmarkNegamax -benchmem ./internal/search` on the same machine for both builds. |
| Python analysis | Install the [analyzer dependencies](../tools/kingzone/README.md), then run `python3 -m unittest discover -s tools/kingzone -p 'test_*.py'`. |

Test one proposed strength improvement at a time. Freeze both executables, play each opening with colors reversed, and save the PGN and its manifest. Check the score by opening pair, its confidence interval and any abnormal terminations. Position tests show which decisions changed. The match measures the effect on play. [Match protocol and commands](../tools/elo/README.md)
