# The 1850+ baseline against Stockfish 2200

The `baseline-1850` tag points to `0839ea3`, with playing code from `91aeb4e`. This test uses the executable qualified against SF1850, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`. September is available at `f6871de` for historical comparisons. [1850+ qualification](../1850/README.md)

The next research cycle qualified `baseline-2200`. [Experiments, losses and independent match](loop/README.md#2200-qualification)

## November 24, 2026 measurement

**51.43% against Stockfish 2200** from 182 wins, 31 draws and 171 losses. The estimated difference was **+10.0 Elo**, with a two-sided 95% interval of **−58.7 to +79.4**. The baseline played close to even at this time control. The lower bound fell short of qualification. The previous benchmark scored 56.25% on the 64 v2 openings. The two suites are reported separately.

| Equinox color | Wins | Draws | Losses | Score |
|---|---:|---:|---:|---:|
| White | 94 | 13 | 85 | 52.34% |
| Black | 88 | 18 | 86 | 50.52% |

All 192 pairs finished. Equinox won both games on 46 openings and lost both on another 46. The score includes one Stockfish time forfeit in pair 99. Removing that entire pair gives 51.44%. Equinox finished without incident. [PGN](../../../matches/elo2200/baseline-sf2200.pgn) · [Manifest](../../../matches/elo2200/baseline-sf2200.pgn.meta.json)

The suite was frozen at `6632daa`. Its SHA-256 is `09a333e7b3ed17533558dde6f4a804180d9a1e22c083c717763adac80327435e`. None of the 192 openings occurs among the 178,552 distinct positions in the audited historical PGNs, the v1/v2 tactical suites or the earlier EPD suites. Each pair's PGN starting position was checked against the EPD. Cutechess starts at move 1 and resets the counters. For 78 openings, the fifty-move counter therefore differs from the EPD. Both colors and all diagnostics use the counters actually played.

## Turning points in the 171 losses

Adjudication found **92 losses** where one move turned a defensible position into a clearly worse one. The fast filter found 74. Checking each move in the remaining 97 losses at 80,000 nodes added 18. The 92 cases include 51 losses as White and 41 as Black. The median turning point was 36 plies out of book, with a median of 24 pieces still on the board.

- In **76 of 92** positions the adjudicator's best move was quiet, with no capture or check. Fourteen were captures and two were non-capturing checks. Equinox's move was itself quiet in 57 cases. These cases motivate studying quiet-move ordering and depth. Selection from losses limits what they say about the error rate across all quiet moves.
- **Another 71 losses** crossed −200 cp by a smaller margin without meeting the strict turning-point criteria. At the first crossing, the alternative was already below −100 cp in 66 cases. Median adjudicator scores were −152 cp for the best move and −239 cp for the move played.
- In the **remaining eight**, the adjudicator scored at least one decision at −200 cp or worse, without finding an Equinox move that crossed the threshold under the stated conditions. Attribution to a single move remains open. [Per-loss analysis, positions and lines](losses.json)

## Hypothesis selected for match testing

**Give the stored move first priority in practice.** At the root, try the previous iteration's principal variation move first when available. At other nodes, try the transposition table move first when available. Apply normal ordering to the remaining moves. The baseline puts the stored move in the first slot, but `selectBestMove` can immediately displace it. In seven of eight cases where the root table held the reference defense on the last observed iteration, the search tried another move first. [Move ordering in Negamax](../../../internal/search/search.go)

| Replay on the 92 positions | Defenses recovered | Defenses lost | Nodes | Difference |
|---|---:|---:|---:|---:|
| 1850+ baseline, depth D | — | — | 13,922,455 | — |
| One extra ply everywhere | 20 | 6 | 55,339,496 | +297.5% |
| Full depth for late root moves | 4 | 3 | 33,189,423 | +138.4% |
| **Search the stored move first** | **4** | **2** | **11,512,571** | **−17.3%** |

The columns count crossings between −200 and −100 cp relative to the baseline's cold replay, in both directions. The reference engine reproduced the game move at its recorded depth in **68 of 92** cases. In those 68, stored-move priority recovered four defenses and D+1 recovered 19. Across all 92, 59 defenses had received only a reduced root search. Deepening all of them was costly for a small net gain. Stored-move priority saved nodes in **76 of 92** cases, the main reason to test it. Cold replays on selected losses cannot establish an Elo gain. [Search results](search.json) · [Move adjudication](grades.json)

**Two checks at 320,000 nodes**, scored from Black's perspective with PGN move numbering. In pair **112**, Equinox played `11...Bxd4`. After `12.Re1+ Kd7 13.Nxd4`, Black had lost castling rights and the move scored −562 cp. Kingside castling scored −8 cp. At the game's depth of 7, stored-move priority found castling. In pair **45**, the normal cold replay found `18...d5` (+110 cp), while stored-move priority reverted to the losing game move, `18...e5` (−237 cp). Ordering also affects replies and search windows. [Control lines](examples.json)

The candidate then played `baseline-1850` at `1+0.1`, with paired openings, a pair-based interval and v1, v2 and run6 checks. The 192 openings above were used for development. Confirmation used 384 other frozen openings. [Research cycle and outcome](loop/README.md#2200-qualification)

## Protocol fixed before play

- **192 new openings**, eight plies from the starting position, both colors per opening and 384 games with no early stopping. The 1850 generator keeps its distribution and uses a new seed, `220020261124`. It excludes positions from previous EPD suites, all project match PGNs at `0839ea3` and the v1 source PGN. Comparison preserves side to move, castling rights and legal en passant captures and ignores counters. Distinct positions can belong to the same opening family.
- **1 second plus 0.1 seconds per move**, six games at once on an Apple M4 Pro. Stockfish 18 is limited to 2200 through UCI, with one thread and a 16 MiB hash. `cmd/elo_match` produces the PGN and manifest. Use the two-sided 95% Hoeffding interval per opening pair. The 2200 milestone requires a lower score bound above 50%.
- **Loss analysis** uses full-strength Stockfish with one thread and fixed node budgets. For each loss, find the first Equinox move that turns a defensible position (best move ≥ −100 cp) into a clearly worse one (played move ≤ −200 cp), dropping at least 150 cp. Verify the best and played moves from the same position at a higher budget. Retain losses without a confirmed turning point and report results by color. Then look for a testable search mechanism in the confirmed positions.

To regenerate the suite, use the Stockfish binary from the manifest, inputs from the baseline tag and the generator in the parent directory. The PGN list below is the one at `0839ea3`.

```sh
python3 tools/elo/1850/build_holdout.py /path/to/stockfish /tmp/2200.epd \
  --seed 220020261124 --exclude-epd tools/elo/1850/holdout.epd \
  --exclude-pgn matches/*.pgn matches/elo1850/*.pgn tools/tactics/v1/source.pgn
```

## Diagnostic method

The adjudicator resets for every search and keeps the last line with a score free of bound flags. For each Equinox move in a loss it searches freely at 5,000 nodes, then forces the played move if it differs from its choice. A broad filter (best ≥ −180 cp, played ≤ −125 cp, gap ≥ 100 cp) triggers two 80,000-node searches. Keep only the first turning point per game that meets the final thresholds. Scores are from Equinox's perspective. The file stores mate distance separately and encodes mate as ±32,000 for threshold comparisons. For each loss the filter misses, a second pass checks every decision at 80,000 nodes until a strict turning point or game end. It also records the first crossing from a best move above −200 cp to a played move at or below −200 cp. Adjudication is a finite-budget estimate.

A clockless replay preserves the move history and starts with an empty 1,048,576-entry table. Depth D is the PGN depth capped at 8. The probe compares four searches.

| Search | Question |
|---|---|
| Reference at D | Does a cold search reproduce the game move? The qualified binary must match the replay's move, depth and nodes. |
| Reference at D+1 | Is one extra ply enough to find the defense? |
| Full depth at the root | What happens when late move reductions are disabled only at the root, at the same nominal depths as D? |
| Stored-move priority | What happens if the engine searches the transposition table move first, or the principal variation move at the root? September and the 1850+ baseline put it at the front before reordering the moves. |

The two experimental controls run only in a copy of the engine. Stockfish rechecks their new moves at 80,000 nodes. A recovery improves the reference replay move from ≤ −200 cp to ≥ −100 cp. The reverse counts as a regression. These decisions come from engine losses, so their frequency describes the selected errors. Adopting a change requires a match against the 1850+ baseline.

### Reproduction

The PGN and manifest are in `matches/elo2200/`. Use the Stockfish binary from the manifest. Adjudication accepts the original 1850 baseline and its verified rebuild with Go 1.26.2 on macOS arm64. Use the pair number and Equinox's color to find each decision in the JSON files.

```sh
set -eu
diag=$(mktemp -d)
repo="$PWD"
sf=$(command -v stockfish)
uci="$diag/Base1850"
git clone --quiet "$(git remote get-url origin)" "$diag/engine"
git -C "$diag/engine" checkout --quiet --detach 91aeb4e
(cd "$diag/engine" && go build -trimpath -buildvcs=true -o "$uci" ./cmd/uci)
python3 tools/elo/2200/analyze_losses.py matches/elo2200/baseline-sf2200.pgn "$sf" "$diag/screen.json"
python3 tools/elo/2200/analyze_losses.py matches/elo2200/baseline-sf2200.pgn "$sf" "$diag/losses.json" \
  --exhaustive-misses "$diag/screen.json"
mkdir "$diag/src"
git archive baseline-1850 | tar -x -C "$diag/src"
(
  cd "$diag/src"
  git apply "$repo/tools/elo/2200/replay.patch"
  LOSS_INPUT="$diag/losses.json" LOSS_OUTPUT="$diag/search.json" \
    go test -count=1 -run '^TestLossReplay$' ./internal/search
)
python3 tools/elo/2200/judge_replay.py "$diag/losses.json" "$diag/search.json" \
  "$uci" "$sf" "$diag/grades.json"
```
