# Calibration against Stockfish 18 limited to 1850

Qualification requires a score above 50% against Stockfish 18 set to 1850 at 1 second plus 0.1 seconds per move, with the lower bound of the two-sided 95% interval also above 50%. The [match protocol](../README.md) uses one observation per opening pair. This scale depends on the opponent, its settings, the hardware and the time control.

## November 24, 2026 results

**The full candidate cleared the 1850 milestone on the 192 confirmation openings.** It scored 74.09%, an estimated +182.5 Elo relative to Stockfish 18 limited to 1850. Against the 2200 setting it scored 56.25%, with an interval spanning even play.

| Match (first engine's perspective) | Pairs | W–D–L | Score | Elo difference [95% CI] |
|---|---:|---:|---:|---:|
| September – SF1850 | 64 | 59–4–65 | 47.66% | −16.3 [−141.6, +104.7] |
| Faster evaluation – September | 64 | 96–19–13 | 82.42% | +268.4 [+111.0, +887.2] |
| Faster evaluation – SF1850 | 64 | 90–4–34 | 71.88% | +163.0 [+34.1, +360.6] |
| Full candidate – September | 64 | 111–14–3 | 92.19% | +428.8 [+192.8, +∞] |
| Full candidate – faster evaluation | 64 | 66–38–24 | 66.41% | +118.4 [−4.0, +280.2] |
| Full candidate – faster evaluation, confirmation | 128 | 144–70–42 | 69.92% | +146.5 [+55.5, +262.5] |
| **Full candidate – SF1850, confirmation** | **192** | **277–15–92** | **74.09%** | **+182.5 [+102.1, +286.6]** |
| Full candidate – SF2200 | 64 | 66–12–50 | 56.25% | +43.7 [−75.7, +174.8] |

Against September, median depth on the first move out of book rose from 4 to 7. As White, the candidate completed a deeper iteration on that move in **all 64 openings**. Against SF1850 on the holdout, it had 139 wins, 6 draws and 47 losses as White, and 138 wins, 9 draws and 45 losses as Black.

The PGNs and manifests record the openings, names, results, time control and verified checksums. Engine revisions are `f6871de` (September), `c9ff718` (mobility only, timed), `6b24f99` (faster evaluation) and `91aeb4e` (full candidate). Commit `2c43fdf` updates expected tracer node counts and adds a capture-ordering test. Checks for moves, scores, depths, paths and tracer non-interference pass. [Games and manifests](../../../matches/elo1850/) · [1850 confirmation PGN](../../../matches/elo1850/capt-sf1850-holdout.pgn)

The eight matches comprise 1,408 games with no time forfeits, illegal moves, disconnects or crashes. They ran on an Apple M4 Pro, with six concurrent games in the main matches, three for the first candidate duel and the 2200 benchmark, and at most nine across matches. Equinox searches on one thread. Each Stockfish used one thread and a 16 MiB hash. Intervals use the project's paired calculation. The upper Elo bound can be infinite when the upper score bound reaches 100%.

## Selection and confirmation

- Compare two successive configurations with September on the 64 `openings-v2.epd` openings, with colors reversed. The first speeds up evaluation. The second adds capture ordering in quiescence.
- Check v1, v2 and the 13 run6 canaries at depth 6. To confirm the full configuration, require a head-to-head gain over September under the project's interval criterion, retention of the previously found reference moves, and passing tests.
- Test one selected configuration against Stockfish 1850 on this directory's **192 new openings**. Play all 384 games and calculate the interval across the 192 pairs without early stopping. Report this confirmation separately from selection matches.

Capture ordering scored 66.41% against the already optimized version over 64 pairs, with an interval of −4 to +280 Elo. To isolate its contribution, a second duel used the **first 128 openings** of the new suite (256 games, with the stopping point fixed before play). Capture ordering gained +146.5 Elo at this time control (95% CI +55.5 to +262.5). The two duels were analyzed separately. The Stockfish 2200 benchmark reused the 64 v2 openings.

## Opening selection

The suite was frozen at `b2c150f` before its first game with Equinox. Starting from the initial position, the generator plays eight plies, choosing among five Stockfish lines at 3,000 nodes that are within 60 cp of the best move. A 16,000-node check keeps positions between −60 and +60 cp. The seed is 18502200. The generator rejects final positions already in the previous suite or the new set, ignoring both FEN counters. Different lines can belong to the same opening family.

Stockfish used one thread and a 16 MiB hash. The EPD SHA-256 is `590061cbd1ff2801ffac0e97e872835d13afc2a02119e9b9398e81284f6bcfac`. [Generator](build_holdout.py) · [Openings](holdout.epd)

## Search work at depth 6

| Suite | September and faster evaluation | With capture ordering | Tactical result for all three builds |
|---|---:|---:|---|
| v1 | 1,480,222 nodes | 897,840 | 32/32 moves, 15/16 losses flagged |
| v2 | 3,289,616 | 1,558,741 | 13/32 moves, 3 false alarms |
| run6 | 7,772,789 | 1,245,101 | 10/13 errors avoided, 2 losses flagged |
| **Total** | **12,542,627** | **3,701,682** | **77 identical moves and scores** |

The first configuration counts mobility directly on bitboards, reuses sliding-piece attacks to detect checks and avoids repeating the same evaluation for every capture. The second searches the most valuable captures first in quiescence when out of check, breaking ties on victim value by trying the least valuable attacker first. In check it still searches all evasions in September order. Non-capturing promotions are still filtered out when out of check. On Game17, the node count falls from 2,788,051 to 159,809 for the same `d1f3` at +67 cp. On 0JLoe, all three play `h5h6` at +38 cp. [Per-position measurements and checksums](tactics.json)

On a systematic subsample of 12 v1/v2 positions, the median of four depth-6 passes fell from **1.428 s** (September) to **0.398 s** (mobility alone), **0.289 s** (full evaluation optimization) and **0.146 s** (with capture ordering). Each position used a fresh UCI process, timed from the end of `isready` to the best move. The four engines alternated order between passes. Timings were taken on the M4 Pro during a match to help explain fixed-time performance. [Timing measurements](timings.json)

## Replay a match

Build each executable from its listed revision in a clean clone with `go build -trimpath -buildvcs=true -o /path/to/output ./cmd/uci`. The September match binary was rebuilt byte for byte from `f6871de` with `-buildvcs=false`. Its SHA-256 is `e84d7f9af0e2ae4912fb1f26b2cf9152e76361cf385adf8dbe976afb89d99830`. The tactical suites used a build with embedded Git metadata. Their reports record that historical executable's SHA-256.

From the repository root, run the following with two prebuilt executables and a new PGN path. The harness checks the binary and opening checksums and the game pairings. Timed match results can vary between runs.

```sh
go run ./cmd/elo_match -engine-a /path/to/equinox -name-a Captures \
  -engine-b /path/to/stockfish -name-b SF1850 \
  -option-b UCI_LimitStrength=true -option-b UCI_Elo=1850 \
  -option-b Threads=1 -option-b Hash=16 \
  -openings tools/elo/1850/holdout.epd -tc 1+0.1 -concurrency 6 \
  -pgn /tmp/equinox-confirmation.pgn
```

Recalculate statistics from the archived PGN with `go run ./cmd/matchstats -pgn matches/elo1850/capt-sf1850-holdout.pgn -engine-a Captures -engine-b SF1850`.
