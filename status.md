# MarsGopherChess · project status as of November 24, 2026

## MarsGopherChess · wider-window result

The engine is now called MarsGopherChess. On the 192 openings with colors reversed, the wider window alone scored 46.22% head-to-head against the baseline and 42.97% against SF2400. All 768 games finished without incident. On the openings it searched 30.40% more nodes at D8, and it lost one v2 solution. The 2200+ baseline remains the tournament entry. [Results and games](tools/elo/2400/window/README.md) · [Tournament entry guide](docs/tournois.md)

## Target 2400 · first benchmark

On 192 new openings with colors reversed, the 2200+ baseline scored **38.67% against Stockfish 18 set to 2400**. Selectively rechecking a quiet defense scored **45.05%** on the same openings and **54.10%** head-to-head against the baseline (95% CI −55.4 to +116.0 Elo). The qualified baseline remains active. Replay on 96 quiet defenses from the 2400 losses led to a timed trial of the wider window alone. Adding the final recheck cost 1.69 million nodes and produced two recoveries and four regressions under 80,000-node adjudication. [Report, PGNs and Go engine profile](tools/elo/2400/README.md)

## Active baseline · 2200+ qualification

On November 24, `baseline-2200` cleared the criterion set for Stockfish 18 limited to 2200 at 1 second plus 0.1 seconds per move. It recorded **430 wins, 64 draws and 274 losses** on 384 openings played with both colors, scoring 60.16% [53.23% to 67.09%]. The local difference was +71.6 Elo [+22.4 to +123.7]. The openings were frozen before testing. There were no game incidents. [Report, PGN and manifest](tools/elo/2200/loop/README.md#2200-qualification)

The selected engine (`a43a8a3`) detects nonterminal positions faster, adjusts time allocation to completed iterations and prioritizes the table move below the root. It scored 69.01% against `baseline-1850` over 384 games. At D6 it searched 19.9% fewer nodes on the 77 tactical positions while retaining 32 v1 solutions, 13 v2 solutions and ten historical run6 mistakes avoided. Game25 illustrates a limitation of that last metric. The new king move loses the advantage. The coordinated-attack variant recovered Game25 but scored 48.83% head-to-head over 128 games and remains experimental. [Per-position details and rejected trials](tools/elo/2200/loop/README.md)

Of 274 confirmation losses, 153 had a verified turning point. In 123 the reference defense was a quiet move. On a sample of 12, one extra ply recovered two defenses at 4.38 times the nodes. These observations motivated the first 2400 trial, rechecking a quiet move that had received a reduced root search. [Loss analysis](tools/elo/2200/loop/README.md#turning-points-in-the-274-losses)

## Previous 1850+ qualification

Candidate `91aeb4e` builds on September (`f6871de`), speeds up mobility and evaluation, and orders captures in quiescence. It scored 92.19% against September on the project's 64 openings with colors reversed. On 192 additional openings frozen before the match it scored 74.09% against Stockfish 18 set to 1850 at 1 second plus 0.1 seconds per move. The estimated local difference was +182.5 Elo, with a two-sided 95% interval of +102.1 to +286.6. Against the 2200 setting it scored 56.25% over 128 games, with an interval compatible with either a gain or a deficit. The 1850 baseline subsequently scored 51.43% on 192 independent pairs against that setting. [Protocol, results and PGN](tools/elo/1850/README.md)

## September baseline

The engine runs via UCI and in the browser. The clearest gain came from allocating search time. At depth 6, Equinox avoided ten of 13 historical regression-suite errors but scored only two of those positions at −300 centipawns or worse. On the v1 regression suite the same revision chose all 32 reference moves and recognized 15 of the 16 positions Stockfish classified as losing. On v2, which requires a difficult quiet first move, it found 13 of 32 solutions.

## Code revisions

| Reference | Contents |
|---|---|
| `8611e7a`, August 5 | Engine revision rebuilt for calibration against Stockfish 18 limited to 1800. Uses a 15 ms margin for time allocation with increment. |
| `90ea1da`, August 31 | Starting point for the September experiments. Includes offline analysis and targeted diagnosis of opponent replies. |
| **`f6871de`**, frozen November 23 | September baseline for resuming development. Includes the schema 11 tracer, its tests, instructions and experiment log. The checks below apply to this revision. |
| Local stash `786d539` | Earlier experiment combining several king-pressure terms. Later knight and queen experiments are documented separately. |

## Activity and decisions

- **April — making results interpretable.** Fixed stalemate handling at the search horizon, best-move tracking, the fallback on interruption and mate scores. The sixth 30-game match against SF1600 finished with 12 wins, no draws and 18 losses. Its positions still serve as regression tests. [Run6 PGN](matches/sf1600-run6-1+0.1-30.pgn)
- **July to early August — time management.** Timed search passed the old depth-4 cap. The engine retains search state between `position` commands and checks the fifty-move counter before reusing a transposition table bound. Two 64-game sets against the old time allocation gave the new increment budget 73.05%. The 15 ms margin was then retained, with zero time forfeits in its 128-game A/B test. [Paired match protocol](tools/elo/README.md)
- **August — testing evaluation hypotheses.** Penalties of 20 centipawns for knight and then queen pressure on the king zone scored 46.88% and 48.44% in their longer A/B tests. Both were rejected. Queen pressure remains a way to locate difficult positions in the PGNs. [Position analysis](tools/kingzone/README.md)
- **Late August and September — explaining a search error.** Traces covered 19 non-capturing checking replies and the 13 regression positions. Searching selected normally reduced checks at full depth raised errors avoided from 9 to 11 on the 19 cases, but reintroduced `d2b4` on `Game25_ply35`. The general `D5D3D1CheckChain` rule was rejected before match play. The latest probe compares isolated subtrees and separates the effects of score, ordering and the transposition table. [Detailed experiment log and decisions](docs/archives/journal-avril-septembre-2026.md)

## Earlier calibration on August 5

The August 5 matches used strength-limited Stockfish 18 at **1 second plus 0.1 seconds per move**. Each set contained 64 openings played with both colors. The 95% Hoeffding interval uses the opening pair as its unit.

| Opponent | Wins – draws – losses | Equinox score | Estimated Elo difference and interval | Time forfeits |
|---|---:|---:|---:|---|
| SF18 limited to 1600 | 79 – 2 – 47 | 62.50% | +88.7 [−31.2, +235.2] | 0 |
| SF18 limited to 1800 | 59 – 7 – 62 | 48.83% | −8.1 [−132.1, +113.7] | 1 by Stockfish |

The engine played near even against SF18-1800 at this time control. The interval still allows a substantial difference in either direction. An absolute rating or extrapolation to slower play would require further measurements. The August figures come from the experiment log. The repository contains the six historical April sets and the November paired matches with their manifests. [August 5 entry](docs/archives/journal-avril-septembre-2026.md#update-2026-08-05)

## Checks when resuming development

On macOS arm64 with Go 1.26.2, the standard and `searchtrace` Go suites passed without cached test results. `go vet ./...`, UCI race tests and all 14 analyzer Python tests also passed. All six anti-stalemate positions passed. At depth 6 the two tactical metrics reproduced exactly 10/13 errors avoided and 2/13 losses recognized. The test also protects each previous success individually. [Validation commands](docs/developpement.md) · [Tactical test](internal/search/search_anti_blunder_test.go)

Both entry points were tried with a rebuilt binary. After 1.e4, the UCI engine returned `b8c6` at depth 6. The server served the page, reported 20 legal starting moves, accepted `e2e4`, played `g8f6` at depth 4 and returned the move to White.

Run6 PGN replay reproduced all 30 games, 1,530 usable decisions and the September CSV SHA-256 (`7fa89590…`). At depth 6, the `-probe-d5-reduced-full` probe reproduced 107 comparisons on 23 positions. On all 32 cases, the observed search exactly matched the untraced move, score, depth and nodes. On the 19 primary replies it avoided nine errors and recognized seven losses. [Replay the probe](cmd/search_trace/README.md)

## Tactical benchmark independent of run6

The v1 suite was frozen at `769eb3f` before measuring `f6871de`. Its 32 positions come from 32 new Stockfish games and 32 distinct openings. None appears among the 3,108 distinct run6 positions, even after ignoring FEN counters. At depth 6 the engine found **16/16 reference moves in defensible positions and 16/16 in losing positions**. It recognized **15/16 losses**, with no false alarms in the 16 defensible positions. Two fresh-process passes agreed exactly, at 1,480,222 nodes each. [Suite, protocol and limitations](tools/tactics/README.md) · [Per-position results](tools/tactics/v1/septembre.json)

## Difficult quiet moves · v2 suite

The v2 suite was published at `259ec7e` before testing the September engine. On 32 puzzles from 32 other games with no overlap with run6 or the v1 source games, `f6871de` found **13/32 first moves** at depth 6 (8/16 as White, 5/16 as Black). Two passes agreed on every position and each searched 3,289,616 nodes. The same binary reproduced its v1 report byte for byte. [v2 selection and results](tools/tactics/v2/README.md)

## November 23 review and follow-up

| Priority | Finding | Action and decision criterion |
|---|---|---|
| 1 · Compare changes against frozen baselines | September found 32/32 moves on v1 and 13/32 quiet moves on v2. | Replay both suites at depth 6. On v2 count gains among the 19 misses and regressions among the 13 successes. On v1 also track the 15/16 recognized losses. Compare nodes and retain the run6 canaries as a historical control. |
| 2 · Preserve evidence for future matches | The tracer is versioned. For the August matches the repository contains results and checksums but lacks the PGNs and manifests. | Since November 24 each qualification match has a verified PGN and manifest. Scores were recalculated from the archived PGNs. [November games](matches/elo1850/) |
| 3 · Quiet promotions in quiescence, experiment closed | The candidate fixed the control endgames but lost one v2 solution (13/32 → 12/32) and increased work by 39.29% over the 77 positions. | Keep September as the reference for this experiment. The measurements and isolated candidate are available for further analysis. [Results by suite and position](tools/tactics/experiments/quiet-promotions/README.md) |
| 4 · Cover real usage | UCI tests mainly exercise position and clock handling. The HTTP server shares mutable game state across requests and its package has no tests. | Add an end-to-end UCI scenario from setting the position through best move and into the next game. For HTTP, check board consistency under concurrent requests and serialize shared game access. Validate under `-race`. |

The `D5D3D1ScoreOnlyShadow` probe proposed in September remains an option for isolating the transposition table's effect on that specific branch.
