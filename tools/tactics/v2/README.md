# Tactical suite v2 · finding a quiet move

This suite tests the **first move** on 32 puzzles from 32 Lichess games, with 16 decisions for each color. The side to move is out of check. The solution starts with a move that neither captures, gives check nor promotes. The [v1 suite](../README.md) remains the regression test for more immediate decisions.

## Selection before testing

1. **Source.** The builder reads the first 65,536 rows of the first Parquet file in the Lichess export at `461725966e`. It applies the first move of the published line to the source FEN to obtain the puzzle position. The second move is the solution. The puzzle data is distributed under CC0. [Pinned export](https://huggingface.co/datasets/Lichess/chess-puzzles/tree/461725966ea46898ec6481372a6326de3281bfbd) · [FEN and move convention](https://database.lichess.org/#puzzles)
2. **Filtering.** Require puzzle rating ≥ 2,000, `RatingDeviation` ≤ 100, popularity ≥ 80 and at least 50 attempts. The published line contains at least five plies from the decision, including a capture, check or promotion within the four plies after the quiet move. Sort by SHA-256 of `equinox-quiet-v2|` followed by the puzzle ID. The first 512 eligible cases, with one per game and position, form the [versioned pool](candidates.csv). The position key ignores FEN counters and keeps the side to move, castling rights and legal en passant captures. Exclude every position seen in the 30 run6 games or the 128 v1 source games.
3. **Difficulty and adjudication.** Stockfish 18 must miss the solution at 512 nodes. At depths 14 and 18 it must find the Lichess move and score it at least 150 centipawns above the runner-up each time. Both moves are evaluated at the requested depth with MultiPV 2, one thread, 16 MiB of hash, numeric scores without bound flags and fresh state for each search. The first 16 eligible cases per color form the [frozen suite](positions.json). Each case records the shallow search move and both deeper analyses.
4. **Measurement.** The primary metric is how many reference moves the engine finds at depth 6. Search each position twice with fresh UCI processes and no clock. Both runs must return the same legal move, score and node count at depth 6. Keep the individual results to identify gains and regressions. Selection targets positions with a clearly better quiet move that Stockfish's shallow search misses.

## September baseline · November 23, 2026

The suite was published on `origin/main` before its first measurement. Its revision in the current history is `259ec7e`. Both historical measurements used the same executable (source now at `f6871de`, SHA-256 `46caaa986c5e4ab0610790b012cad658c6578d6f3bf187fe29222111c1b0642b`). The 32 Lichess ratings range from 2,048 to 2,747 with a median of 2,345. Rebuilding from the pool reproduced the reference file byte for byte.

| Side to move | Reference first move found | Nodes in one pass |
|---|---:|---:|
| White | 8/16 | 1,166,903 |
| Black | 5/16 | 2,122,713 |
| **Total** | **13/32 (40.6%)** | **3,289,616** |

The two fresh searches agreed on every position. Of the 19 misses, 11 repeated Stockfish's 512-node choice. At depth 18, Stockfish favors the side to move in all 32 positions. Equinox scores three at −300 cp or worse, including `0afCF`, where it finds the reference move. [Per-position report](septembre.json)

Puzzle `00Nhl` illustrates the intended difficulty. Equinox immediately captures the rook on h7 (`h8h7`). Stockfish ranks that capture second at −176 cp. Its first choice, the knight move `g6f4`, scores +316 cp and invites the white queen to capture the knight on f4, where the e5-pawn can recapture. [View the position](https://lichess.org/training/00Nhl)

Count newly solved positions separately from any of the original 13 solutions lost. The v1 regression baseline found 32/32 moves and recognized 15/16 losses. The two scores show how difficult each selection was for the same binary at depth 6.

## Repeat a measurement

From the repository root, supply the UCI binary and an unused output path. The [build instructions](../README.md#reproduce-a-measurement-or-compare-a-candidate) rebuild the September engine with its embedded Git revision and set the shell variable `trial`. Run the following in the same shell.

```sh
python3 tools/tactics/measure.py "$trial/equinox" "$trial/v2.json" \
  --fixture tools/tactics/v2/positions.json
```

To rebuild the references from the versioned pool, use the Stockfish binary identified in the suite and the Python dependency of the [tactical harness](../README.md). The `pool` command also uses PyArrow (22.0.0 when the suite was built) to read the upstream Parquet file. Its URL and checksum are recorded in the suite.

```sh
python3 tools/tactics/quiet.py label "$(command -v stockfish)" \
  tools/tactics/v2/candidates.csv "$trial/positions-v2.json"
cmp tools/tactics/v2/positions.json "$trial/positions-v2.json"
```
