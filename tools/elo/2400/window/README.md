# A wider window alone at the root

## Results from 768 games

**MarsGopherChess retains the 2200+ baseline.** The wider window alone missed the head-to-head selection criterion. On the same openings against SF2400 it outscored the baseline and scored below the variant with a recheck. Both new matches finished without incidents or forfeits.

| Opponent | Wider window wins / draws / losses | Score | Elo difference at 1+0.1 s, 95% CI |
|---|---:|---:|---:|
| 2200+ baseline | 105 / 145 / 134 | 46.22% | −26.3 [−96.8 to +42.1] |
| Stockfish 18 set to 2400 | 152 / 26 / 206 | 42.97% | −49.2 [−121.7 to +19.3] |

Against SF2400 the historical baseline scored 38.67% and the recheck variant 45.05%. The window alone gained **4.30 percentage points** over the baseline (77 pairs better, 58 tied, 57 worse) and trailed the recheck variant by **2.08 points** (70 better, 51 tied, 71 worse). The first difference becomes +4.58 points after excluding the pair affected by Stockfish's time forfeit in the historical baseline match. The window-only engine reached a median depth of 9 in both matches. [Head-to-head PGN](../../../../matches/elo2400/window-base.pgn) · [SF2400 PGN](../../../../matches/elo2400/window-sf2400.pgn) · [Head-to-head audit](base-audit.json) · [SF2400 audit](sf2400-audit.json) · [Pairs compared with baseline](paired-base.json) · [Pairs compared with recheck](paired-quiet.json)

The search savings on quiet defenses did not carry over to the openings. The window used 2.13% fewer nodes on the 96 SF2400 defenses and **30.40% more** when searching from the 192 openings at D8. The run6 trace confirmed 345 reduced searches in the wider window across 11 of the 13 final iterations, with no added recheck pass.

## November 24, 2026 protocol

Test the binary used in the previous cold replay, built at `04efa19` with the [frozen patch](../ranking-only.patch) and the `quietverify` tag. Starting at D5, widen the reduced-search window for late quiet root moves by 80 centipawns. The loop ends after the last legal move. A move exceeding alpha receives the baseline's normal full-depth search.

Use **the same 192 openings** for both duels, in the same order with both colors, for 384 games each. First play RankingOnly against the qualified 2200+ baseline, then against Stockfish 18 at 2400 with strength limiting enabled, one thread and a 16 MiB hash. Use 1 second plus 0.1 seconds per move with eight games in parallel. Play both matches in full. The two previous SF2400 measurements used this suite. The earlier head-to-head with the recheck candidate used a different suite.

Use the match score and the two-sided 95% Hoeffding interval by opening pair. Head-to-head selection requires a lower bound above 50%. The SF2400 scores on this development set allow all three engines to be compared. Qualification at 2400 would require new openings.

Count fixed-depth nodes on v1, v2, the run6 canaries and both quiet-defense sets. Report solutions gained and lost and any match incidents. [Openings](../baseline.epd) · [Previous experiment](../README.md)

## Fixed-depth cost

| Positions | Depth | Baseline nodes | Window-only nodes | Change | Observed quality |
|---|---:|---:|---:|---:|---|
| 192 match openings | 8 | 100,004,748 | 130,408,283 | +30.40% | 36 first moves changed |
| v1, 32 positions | 6 | 816,043 | 834,306 | +2.24% | 32 → 32 solutions, 15 → 15 recognized losses |
| v2, 32 positions | 6 | 1,229,318 | 1,402,340 | +14.07% | 13 → 12 solutions, 3 → 3 false loss alerts |
| run6, 13 canaries | 6 | 919,588 | 1,066,327 | +15.96% | 10 → 10 historical errors avoided, 2 → 2 recognized losses |
| 123 defenses from the SF2200 set | Game depth | 36,730,886 | 39,867,227 | +8.54% | 7 recoveries, 5 regressions |
| 96 defenses from the SF2400 set | Game depth | 30,091,189 | 29,449,704 | −2.13% | 6 recoveries, 2 regressions |

Across the 77 v1, v2 and run6 positions, the window cost **338,024 extra nodes (+11.40%)**. On the openings, D8 was the baseline's median depth in its first SF2400 measurement. Each search started with an empty table and the PGN starting FEN. The node count rose on 158 openings, fell on 32 and matched on two. Of the 36 changed first moves, full-strength SF18 preferred the candidate by at least 50 cp in **two cases** and the baseline by that margin in **three**, at both 80,000 and 320,000 nodes per forced move. The other 31 changes did not clear the threshold in the same direction at both budgets. [Nodes and moves by opening](opening-nodes.json) · [Adjudications](opening-grades.json)

The last two rows use the archived replays, whose executable checksums match the game binaries. A recovery means the baseline move scored at most −200 cp and the candidate move at least −100 cp. A regression crosses in the other direction. The table uses SF18 at 80,000 nodes. On the 96 SF2400 defenses, the 320,000-node check found four recoveries and two regressions. [SF2200 replay](../ranking-2200.json) · [SF2400 replay](../ranking-2400.json) · [320,000-node check](../quiet-2400-deeper.json)

## Tactical gains and regressions

Five moves changed across the 77 positions. At 320,000 nodes per forced move on `Game25_W_ply49`, `g1g2` replaced `e2f3`, improving from −242 to +143 cp. On `Game30_B_ply26`, `h7h5` replaced `c6e5`, falling from −171 to −439 cp. The historical run6 counters did not distinguish these changes. The lost v2 solution was `0eK7q`, where `g8g7` scored −75 cp versus +568 for `g8h8`. On `Game16_B_ply62` the score moved from −289 to −216 cp and on `0Pfb0` from +1 to −35. Both adjudication budgets agreed on the direction of all five changes. [Three-suite results](tactics.json) · [Moves, scores and lines](tactic-changes.json)

## Reproduction

The match manifest records the historical `RankingOnly` SHA-256, `1323dd9cc0c1c8eafa0b311f53e3b69314acd91004db3648a8c5f0a7e6ff93e7`. On macOS arm64 with Go 1.26.2, the rebuild below yields `f2a2806e4901c28513755ade3580624ac00c3f89f0c2cd6d7bd008141adf58b3`. The commands apply the [window-only patch](../ranking-only.patch) in a linked worktree. On the verified rebuild, `go version -m` reported `(devel)` and omitted the VCS fields. Use the checksum above to identify this executable.

From the repository root, first run the [Base2200 rebuild commands](../../2200/loop/README.md#rebuild-the-2200-baseline), then run the following in the same shell.

```sh
git -C "$lab/source" worktree add --quiet --detach "$lab/ranking" 04efa19
git -C "$lab/ranking" apply "$PWD/tools/elo/2400/ranking-only.patch"
(cd "$lab/ranking" && go build -trimpath -buildvcs=true -tags quietverify -o "$lab/RankingOnly" ./cmd/uci)
python3 tools/elo/2400/window/measure_openings.py "$lab/Base2200" "$lab/RankingOnly" "$lab/opening-nodes.json"
python3 tools/elo/2400/window/grade_openings.py "$lab/opening-nodes.json" "$(command -v stockfish)" "$lab/opening-grades.json"
```
