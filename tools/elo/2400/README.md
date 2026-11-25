# Target 2400 · first benchmark and quiet defenses

## Result and decision

**The active baseline remains `baseline-2200`.** Rechecking a quiet defense scored 54.10% head-to-head and improved the observed score against SF2400 by 6.38 percentage points on the same openings. The head-to-head interval includes both a gain and a loss. The 2400 milestone is still open.

| Engine | Opponent | Wins / draws / losses | Score | Local Elo difference, 95% CI by pair |
|---|---|---:|---:|---:|
| 2200+ baseline | SF2400 | 131 / 35 / 218 | 38.67% | −80.1 [−156.6 to −10.6] |
| Candidate | 2200+ baseline | 98 / 81 / 77 | 54.10% | +28.6 [−55.4 to +116.0] |
| Candidate | SF2400 | 159 / 28 / 197 | 45.05% | −34.5 [−105.6 to +33.8] |

All 1,024 scheduled games were played. Against SF2400 on 192 openings, the candidate outscored the baseline on 71 pairs, tied on 67 and scored lower on 54, gaining 24.5 game points over 384 games. Each SF2400 match includes one Stockfish time forfeit. Excluding the affected pair from each match gives 38.35% and 45.03%. The head-to-head finished without incident. Eight games ran concurrently on the M4 Pro, which was also used for profiling and offline analysis. [Head-to-head games](../../../matches/elo2400/quiet-base.pgn) · [Head-to-head audit](direct-audit.json) · [Candidate against SF2400](../../../matches/elo2400/quiet-sf2400.pgn) · [Manifest](../../../matches/elo2400/quiet-sf2400.pgn.meta.json) · [Candidate audit](candidate-audit.json)

The candidate was built at `04efa19` with the `quietverify` tag. Commit `76b7fec` preserves the candidate, the rule test and the defense tracer. Candidate Go tests and current and historical trace checks passed. The next duel tested the wider window alone. On the 2400 losses, the additional recheck cost 1.69 million nodes and caused more regressions than recoveries at both adjudication budgets. [Results of both matches](window/README.md)

## Protocol fixed before play

The qualified baseline is `baseline-2200`, source at `a43a8a3` (historical match binary SHA-256 `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`). The new suite contains 192 openings, seed 240020261124, with the same balance criteria as previous suites. The generator excludes archived EPD and PGN positions. A second check also covers v1 and v2 tactical positions against the archive at `023a467`.

1. Freeze the suite in Git. Play it in full with both colors against Stockfish 18 with UCI_Elo=2400, strength limiting enabled, one thread and a 16 MiB hash. Use 1 second plus 0.1 seconds per move. Calculate the two-sided 95% Hoeffding interval by pair with no early stopping.
2. Test one candidate that rechecks at normal depth a quiet defense rejected by late move reduction at the root. From D5, widen the reduced-search window by 80 cp to rank rejected quiet moves. After the other moves, recheck at most one candidate per root pass if it remains within 80 cp of the best score. Exclude captures, promotions and checking moves.
3. Cold-replay the 123 quiet defenses from the 2200 qualification at their game depths. Adjudicate changed moves with full-strength SF18 at 80,000 nodes. Count crossings from ≤ −200 to ≥ −100 cp, reverse crossings and nodes. Also check v1, v2 and run6 at D6.
4. Play the candidate against the baseline on the first 128 openings of `tools/elo/2200/baseline.epd`, then against SF2400 on the same 192 openings used for the initial measurement. The head-to-head decides whether there is an established gain over the baseline. Keep the qualified baseline otherwise. Selecting a candidate on this suite would require a fresh opening holdout for 2400 qualification.

Matches use cutechess-cli. Each PGN and its manifest record openings, colors and executable checksums.

## First baseline measurement

On the 192 new openings with colors reversed, the baseline recorded **131 wins, 35 draws and 218 losses**, scoring **38.67%** against SF2400. The local difference was **−80.1 Elo** (95% CI −156.6 to −10.6). Stockfish incurred the only time forfeit, in round 78. Excluding the entire pair gives 38.35%. Median reported depths were 8 for Equinox and 14 for Stockfish. [Games](../../../matches/elo2400/base-sf2400.pgn) · [Manifest](../../../matches/elo2400/base-sf2400.pgn.meta.json) · [Opening, color and incident audit](base-audit.json)

## Tactical regressions at depth 6

| Suite | Baseline → candidate | Baseline → candidate nodes |
|---|---|---:|
| v1 | 32 → 32 solutions, 15 → 15 recognized losses | 816,043 → 842,517 |
| v2 | 13 → 13 solutions, 3 → 3 false loss alerts | 1,229,318 → 1,505,872 |
| run6 | 10 → 10 historical errors avoided, 2 → 2 recognized losses | 919,588 → 1,077,714 |

Across 77 positions cost rose from **2,964,949 to 3,426,103 nodes (+15.55%)**. The v2 total hides a trade-off. On `0ImFe` the candidate found `g3e5` (+252 cp versus −201 for `g5e7`). On `0eK7q` it abandoned `g8h8` (+568 cp) for `g8e8` (0 cp). The other two changes stayed near even on `0Pfb0` or improved a still-inferior position on `Game16_B_ply62`. Full-strength SF18 adjudicated all four changes at 80,000 and 320,000 nodes per forced move. Quoted scores use the second budget and the side to move's perspective. [Per-position comparison](quiet-tactics.json) · [Adjudications, depths and lines](tactic-changes.json)

## Testing the mechanism on losses at 2200

On the 123 historical quiet defenses, cold replay changed 20 decisions. Six positions improved from at most −200 cp to at least −100 cp, and four crossed in the other direction. Each engine found eight reference moves. Work rose from 36,730,886 to 42,391,554 nodes (+15.41%). Scores use full-strength SF18 at 80,000 nodes. Replay reconstructs the move history and clears the table. In a game the table and ordering data persist between moves. [Replay of both executables](quiet-2200.json)

At the final iteration in each of the 123 positions, the candidate triggered a recheck in 100 positions and selected the reference defense for rechecking in 13. In 73 positions the reference defense was searched only at reduced depth. The tracer reproduced the UCI binary's move and node count in every position. [Per-position traces](trace-2200.json)

To separate the effects of the wider window and the recheck, a control added after the first replay removed only the extra pass at the end of the root search. It played no match in this experiment. On the 123 positions, the window alone cost 3,136,341 extra nodes. Restoring the recheck added a further 2,524,327 nodes, changed 11 decisions and gave three recoveries and three regressions relative to the window alone. The counts include effects carried into later iterations. [Patch against 04efa19](ranking-only.patch) · [Replay without the final recheck](ranking-2200.json) · [Paired calculation](compare_replays.py)

To reproduce the control, apply the patch to a clean copy of `04efa19` and rebuild the UCI executable with the same options as `QuietVerify`. Pass the loss report, baseline binary, test binary, Stockfish and a new output path to `replay_quiet.py`. Compare with `python3 tools/elo/2400/compare_replays.py tools/elo/2400/quiet-2200.json tools/elo/2400/ranking-2200.json`. Build `cmd/quiet_trace` from `76b7fec` with tags `searchtrace quietverify` and supply the loss report followed by the full replay.

## Defenses in losses to SF2400

In 120 of the 218 losses, full-strength SF18 confirmed an opportunity where the best move scored at least −100 cp, the played move at most −200 cp, and the gap at least 150 cp. Filtering at 5,000 nodes and verifying at 80,000 found 90 cases. Checking every decision in the remaining 128 losses at 80,000 nodes added 30. Of the 120 reference moves, 96 were quiet, 22 captures and two non-capturing checks. Among the remaining 98 losses, 95 crossed −200 cp without meeting all three turning-point criteria. [Analysis of the 218 losses, lines and positions](base-losses.json)

At the game depths of the 96 quiet defenses, the candidate changed 12 decisions, recovering three positions and worsening one by the replay thresholds. Both engines found seven reference moves. Work rose from 30,091,189 to 31,141,813 nodes (+3.49%). At final depth the recheck triggered in 79 positions and selected the reference move in 19. In 48 of 96 positions the reference defense was searched only at reduced depth. In 20 others, the ordinary search already examined it at normal depth, and SF18 still scored the chosen move at −200 cp or worse. [Cold replays](quiet-2400.json) · [Per-position traces](trace-2400.json)

| Cold comparison on the 96 positions | Recoveries / regressions at 80k | Recoveries / regressions at 320k | Equinox node difference |
|---|---:|---:|---:|
| Window alone vs baseline | 6 / 2 | 4 / 2 | −641,485 |
| Full candidate vs baseline | 3 / 1 | 2 / 1 | +1,050,624 |
| Adding the recheck to the wider window | 2 / 4 | 2 / 3 | +1,692,109 |

The second adjudication covered the 17 positions where at least two versions chose different moves, using the same history and 320,000 nodes per forced move. The net result pointed in the same direction for each comparison. Threshold crossings are not additive across rows. The window alone was selected for a timed duel. [Control without final recheck](ranking-2400.json) · [Second adjudication and lines](quiet-2400-deeper.json)

## Go engine performance

On the Apple M4 Pro with Go 1.26.2, the baseline's D7 search visited 432,453 middlegame nodes in 97.7 ms and 33,693 endgame nodes in 7.28 ms, or 4.42 and 4.63 million nodes/s respectively. The single-threaded benchmark reuses a one-million-entry UCI table and clears it before each repetition. Both unprofiled measurements reported 0 bytes and 0 allocations per operation. Matches were running on the same machine. [Benchmark output](go-bench.txt)

In the middlegame CPU profile, 66.75% of samples were in quiescence including its callees, and 18.58% in evaluation including its callees. These percentages can overlap. Sliding move generation and quiescence capture ordering accounted for 8.60% and 7.71% of self time respectively. These are Go optimization targets. The two positions measure neither the full UCI process nor the difference a native implementation of the same algorithm might make. [CPU profile](go-cpu.txt)

```sh
GOMAXPROCS=1 go test ./internal/search -run '^$' -bench '^BenchmarkColdRoot$' -benchmem -benchtime=4s
GOMAXPROCS=1 go test ./internal/search -run '^$' -bench '^BenchmarkColdRoot$/middlegame$' -benchtime=8s -cpuprofile /tmp/equinox-root.pprof -o /tmp/equinox-root.test
go tool pprof -top /tmp/equinox-root.test /tmp/equinox-root.pprof
```

## Replay the matches

From the repository root, these commands rebuild both executables in clean trees and write games to a temporary directory. Both revisions are available in the published history of `main`. Match manifests record the original binary checksums. The [rebuild check](../2200/loop/README.md#rebuild-the-2200-baseline) gives the baseline checksum for the current Git history. Individual timed game results can vary.

```sh
remote=$(git remote get-url origin)
lab=$(mktemp -d)
git clone --quiet "$remote" "$lab/base"
git -C "$lab/base" checkout --quiet --detach a43a8a3
(cd "$lab/base" && go build -trimpath -buildvcs=true -o "$lab/Base2200" ./cmd/uci)
git clone --quiet "$remote" "$lab/research"
git -C "$lab/research" checkout --quiet --detach 04efa19
(cd "$lab/research" && go build -trimpath -buildvcs=true -tags quietverify -o "$lab/QuietVerify" ./cmd/uci)

play_sf2400() {
  go run ./cmd/elo_match -engine-a "$lab/$1" -name-a "$1" \
    -engine-b "$(command -v stockfish)" -name-b SF2400 \
    -option-b UCI_LimitStrength=true -option-b UCI_Elo=2400 \
    -option-b Threads=1 -option-b Hash=16 \
    -openings tools/elo/2400/baseline.epd -openings-count 192 \
    -tc 1+0.1 -concurrency 8 -pgn "$lab/$1-sf2400.pgn"
}
play_sf2400 Base2200
go run ./cmd/elo_match -engine-a "$lab/QuietVerify" -name-a QuietVerify \
  -engine-b "$lab/Base2200" -name-b Base2200 -openings tools/elo/2200/baseline.epd \
  -openings-count 128 -tc 1+0.1 -concurrency 8 -pgn "$lab/quiet-base.pgn"
play_sf2400 QuietVerify
```
