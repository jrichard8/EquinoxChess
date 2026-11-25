# Qualifying the 2200+ baseline

The cycle started from `baseline-1850` (UCI revision `91aeb4e`, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`). Its independent benchmark against the 2200 setting was 51.43% over 384 games. [Earlier measurement and diagnosis](../README.md)

<a name="confirmation-du-palier-2200"></a>

## 2200+ qualification

**60.16% against Stockfish 18 limited to 2200** on the 384 frozen `confirm-01.epd` openings with colors reversed (768 games). The engine recorded **430 wins, 64 draws and 274 losses**. The two-sided 95% Hoeffding interval by opening pair is **53.23% to 67.09%**, an estimated **+71.6 Elo [+22.4 to +123.7]** relative to the opponent setting. This clears the criterion fixed before play. The calibration applies to the 1 s + 0.1 s time control, hardware and opening distribution tested. The qualification tag is `baseline-2200`.

| Candidate color | Wins | Draws | Losses | Score |
|---|---:|---:|---:|---:|
| White | 218 | 34 | 132 | 61.20% |
| Black | 212 | 30 | 142 | 59.11% |

All pairs and starting positions match the EPD. The candidate won both games on 116 openings and lost both on 45. Both engines finished without time forfeits, illegal moves, disconnects or crashes. The binary was the one selected against the 1850 baseline. Confirmation used an Apple M4 Pro, cutechess-cli 1.4.0 and eight concurrent games, with at most two other development games running alongside them. Stockfish used one thread and a 16 MiB hash. [768 games](../../../../matches/elo2200/loop/interior-sf2200-confirm.pgn) · [Manifest with commands and checksums](../../../../matches/elo2200/loop/interior-sf2200-confirm.pgn.meta.json) · [Pair audit](confirmation-audit.json)

Go tests passed for the normal engine, the current tracer and historical replays with `searchtrace trace1850`. The UCI race check and `go vet` passed. Normal builds at `a43a8a3` and `8b75863` (historical test integration) agreed on moves, scores and nodes across all 77 UCI positions.

<a name="reconstruire-la-base-2200"></a>

## Rebuild the 2200+ baseline

The `baseline-2200` tag points to the qualification report (`18ce418`). The **InteriorTT** engine that played the 768 games is built from `a43a8a3`. Run the following with Go 1.26.2 on macOS arm64 (`CGO_ENABLED=1`, `GOARM64=v8.0`).

```sh
lab=$(mktemp -d)
git clone --quiet https://github.com/jrichard8/MarsGopherChess.git "$lab/source"
git -C "$lab/source" checkout --quiet --detach a43a8a34313c4c1fe717b7ccfbc5083f10ee44ed
(cd "$lab/source" && go build -trimpath -buildvcs=true -o "$lab/Base2200" ./cmd/uci)
go version -m "$lab/Base2200"
echo "ff4749fe461ab76b7ff383a6180e34d48bb2e15b00c24673b923ef876446648c  $lab/Base2200" | shasum -a 256 -c -
```

Go embeds the revision in the binary. Two independent clones produced this checksum with `vcs.modified=false`. The match manifest records the pre-rewrite binary's checksum, `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`. At D6 the rebuild matched that binary on moves, scores and nodes for all 77 v1, v2 and run6 positions, with two cold searches per position per executable. At D8 the replay matched the first move and node count on every one of the 192 openings in the 2400 campaign, totaling 100,004,748 baseline nodes. [Reference measurements](../../2400/window/opening-nodes.json)

Look up historical identifiers from manifests and tag annotations in the [Git revision map](../../../../docs/revisions-git.tsv). [UCI comparison script](compare_fixed_depth.py)

<a name="où-les-274-défaites-basculent"></a>

## Turning points in the 274 losses

After the match, full-strength SF18 reanalyzed the losing decisions. A strict turning point requires an alternative scoring ≥ −100 cp, a played move scoring ≤ −200 cp and a gap of at least 150 cp, from the side to move's perspective. The 5,000-node filter found 111 after verification at 80,000 nodes. A second pass at 80,000 nodes on unexplained losses added 42. Of the remaining 121 losses, 109 crossed −200 cp by a smaller margin and 12 reached an inferior position without an attributable crossing under this criterion. [274 losses, positions and lines](confirmation-losses.json) · [Analyzer](../analyze_losses.py)

Among the **153 strict turning points** (75 as White, 78 as Black), **123 reference defenses were quiet moves**. The others were 19 captures, 10 non-capturing checks and one promotion. The played move was quiet in 105 cases. The median turning point came 40 plies out of book, with 22 pieces and a game search depth of 7. For example, in pair 3 as Black, moving the king to d7 scored −263 cp while moving the rook to d8 held −58 cp. These frequencies describe the classified losses.

On 12 of the 123 cases, sampled at regular intervals in pair order, cold replay reproduced the game move 9 times. Going from D to D+1 recovered two positions under the thresholds and worsened one from −233 to −355 cp. D+1 found one exact reference defense, compared with zero at D. Work rose from 3,651,347 to 16,004,032 nodes (×4.38). **Next experiment.** Recheck the best rejected quiet root move at full depth when the first search reduced it. Count defenses recovered and lost and measure the extra nodes before another duel. Confirmation set 01 is now diagnostic data. Future qualification requires fresh holdout openings. [12 replays](confirmation-quiet-depth.json) · [Selection and adjudication](check_quiet_depth.py)

## Selection criteria

Development matches use `1+0.1` and both colors per opening. The first experiment prioritizes the table move or root principal variation move on the first 128 pairs of the diagnostic suite. An independent second approach fits the positional evaluation to games archived before this cycle. Preserve revisions, executable checksums and development PGNs for favorable and unfavorable trials.

Promotion requires a head-to-head gain over the baseline with a positive lower bound (two-sided 95% Hoeffding interval by pair). The v1, v2 and run6 suites track tactical gains and losses separately. For 2200+ qualification, use **384 other openings**, with colors reversed, no early stopping and no version selection on that set. At this time control the lower score bound against Stockfish 18 set to 2200 must exceed 50%. Stockfish uses one thread and a 16 MiB hash.

## First development trials

| Candidate | Result against baseline | v1 | v2 | run6 errors avoided | Nodes on 77 positions |
|---|---|---:|---:|---:|---:|
| 1850 baseline | Reference | 32/32 | 13/32 | 10/13 | 3,701,682 |
| Stored-move priority, `df9de88` | 119 W, 65 D, 72 L on 128 pairs. 59.18%, +64.5 Elo [−19.6, +157.1] | 32/32 | 13/32 | 9/13 | 2,845,694 |
| Phase-dependent tables, `794fcc7` | Separate match | 32/32 | 12/32 | 10/13 | 3,916,379 |
| Combined, `abfb9fe` | Separate matches | 32/32 | 10/32 | 10/13 | 3,097,958 |

With the stored move first, Game27_W_ply29 returns to `a2a3`. Phase-dependent tables reduced mean absolute error against SF18's static evaluation from 238.2 to 168.4 cp on 4,551 validation positions after fitting on 28,781 positions from other openings in the 1850 PGNs. However, v1 loss recognition fell from 15/16 to 10/16. Fitting also altered implicit material values. At full opening phase the individual a2-pawn contribution fell from +105 to −1 cp and the h2-pawn from +105 to −15 cp, before structure and mobility terms. Check for this drift in future fits. Generation used Python, python-chess 1.999 and NumPy 2.2.6. To reproduce, copy `tools/elo/2200/loop/static_eval` from `794fcc7` into a `baseline-1850` checkout, compile the Go command there, then pass that binary to `fit_tables.py STOCKFISH BASE_STATIC OUTPUT.go` from the candidate checkout. [Table fitting](tapered-fit.json) · [First match results](../../../../matches/elo2200/loop/hash-base.pgn.meta.json)

Search optimization `0f5cf3a` identifies a nonterminal position when a pawn can legally advance. On the 77 positions at UCI depth 6, two passes per engine per position agreed on move, score and nodes. The sum of per-position median times fell from 1.621 to 1.275 s (−21.4%). Passes alternate engine order, each in a fresh process on the match machine. [Paired timings](witness-parity.json)

The combined candidate includes all three changes. Per-position results are archived for [stored-move priority](hash-tactics.json), [phase-dependent tables](tapered-tactics.json) and [the combination](combined-tactics.json).

## Choosing a candidate for selection

The learned tables alone scored **38.28%** (84 W, 28 D, 144 L), or −83.0 Elo [−179.2, +2.0] against the baseline. There were no game incidents. Static scores were closer to Stockfish on validation positions, but the match led us to reject this global fit. [Games](../../../../matches/elo2200/loop/tapered-base.pgn) · [Manifest](../../../../matches/elo2200/loop/tapered-base.pgn.meta.json)

Two trials isolate the ordering regression on Game27_W_ply29. With the same historical evaluation and calculation optimizations, giving the principal variation priority at the root reintroduces `a2a3`. Giving the table move priority only below the root avoids the error again with `c4b3` (the baseline played `e1e2`). Both versions have the same game time budget. The tactical comparison uses fixed depth.

| Change | v1 | v2 | run6 errors avoided | run6 losses recognized | v1 + v2 + run6 nodes |
|---|---:|---:|---:|---:|---:|
| [Side-to-move null-move check](verified-tactics.json), `2a9ff6f` | 32/32 | 13/32 | 9/13 | 2/13 | 2,782,306 |
| [Null-window search on subsequent moves](scout-tactics.json), `c56ad71` | 32/32 | 13/32 | 9/13 | 3/13 | 2,897,200 |
| [Phase-dependent king placement](king-tactics.json), `36e2a0a` | 32/32 | 12/32 | 10/13 | 1/13 | 3,677,481 |
| [Root-only priority](RootPV-tactics.json), `6298632` | 32/32 | 13/32 | 9/13 | 2/13 | 3,426,891 |
| [Table move first below the root](InteriorTT-tactics.json), `a43a8a3` | 32/32 | 13/32 | 10/13 | 2/13 | 2,964,949 |

**InteriorTT** (`a43a8a3`) retains the 15 correct v1 loss alerts, three v2 false alarms and two recognized run6 losses. It combines a fast legal pawn-push check, direct occupancy calculation, clock management and stored-move priority below the root. With UCI clocks it can finish an iteration beyond the nominal budget, up to twice that budget. Two stable iterations lower the budget for starting the next. Explicit `movetime` retains its cap. Matches compare complete configurations and do not isolate each change's contribution.

InteriorTT's head-to-head uses all 192 `baseline.epd` pairs with no early stopping. If it passes selection, the same executable plays all 384 `confirm-01.epd` pairs against Stockfish 18 at 2200. The combined candidate with learned tables uses only the 64 historical v2 openings for its development match. To audit colors, checksums, legality and opening order, run `python3 tools/elo/2200/loop/audit_match.py MATCH.pgn SUITE.epd`.

The combined candidate scored 51.17% against SF2200 on the v2 openings (59 W, 13 D, 56 L, +8.1 Elo [−113.7, +132.1]). All 64 pairs completed without incident. [PGN](../../../../matches/elo2200/loop/combined-sf-dev.pgn) · [Manifest](../../../../matches/elo2200/loop/combined-sf-dev.pgn.meta.json)

All 384 confirmation starts are distinct. None occurs among the 219,695 positions in games, EPDs and tactical suites at `314d236` after normalizing side to move, castling and legal en passant, and dropping FEN counters. The generator seed is 220020261125. [Frozen suite audit](confirm-01-audit.json) · [Reproduce the checks](audit_openings.py)

## How InteriorTT changes fixed-depth search

Across 77 positions, work fell from 3,701,682 to **2,964,949 nodes** (−19.9%). The 32 v1 moves and 13 v2 solutions match the baseline. Two moves changed among the v2 misses. On run6 the same ten errors were avoided and two losses recognized. On `0JLoe`, `h5h6` took 3,885 nodes compared with 4,466. Game17_W_ply17 remained a tactical miss (`d1f3`), costing 11,701 extra nodes. [Per-position details](InteriorTT-tactics.json)

On the **92 previously identified turning points** against SF2200, replaying each move history to the game depth recovered four defenses and lost three relative to the cold baseline. A recovery goes from ≤ −200 to ≥ −100 cp under SF18 at 80,000 nodes, and a regression crosses the other way. Work fell from 13,922,455 to 12,227,917 nodes (−12.2%). In pair 112 as Black, InteriorTT castled instead of capturing on d4, improving from −548 to +1 cp. In pair 57 as White, it pushed to c5 instead of capturing on c4, falling from −79 to −429 cp. The sample contains only historical losses. [92 decisions](interior-defenses.json) · [Replay and adjudication](replay_defenses.py)

Full-strength SF18 rechecked the four changed moves hidden by the v2 and run6 totals. **Game25_W_ply49 is a regression** even though the old error `d4d5` is avoided. At one million nodes per search the baseline's rook move to g2 scores +153 cp. The candidate's king move to f3 scores −208 cp after …Re8, followed in the line by …Qf6+ and …Qe6. The adjudicator already preferred the rook at 80,000 and 320,000 nodes. On Game27, `c4b3` scores −38 cp versus −71 cp for `e1e2` at one million nodes. Both changed v2 moves still miss the reference first move. [Moves, scores and lines at all three budgets](interior-changed-moves.json)

The static evaluation cache (`d79da5d`) was also tested on the build with fast stalemate detection. Both passes on the 77 cold positions matched exactly on move, score and nodes. Per-position median times summed to 1.566 s with the cache and 1.598 s without it. The timing gain was small and its effect in games remains open. The selection candidate uses evaluation without this cache. [Measurements](cache-check.json) · [Alternating-pass protocol](compare_fixed_depth.py)

On Game25 the cold UCI engine chooses Kf3 at D5 and D6, then Rg2 at D7, D8 and D9. The baseline already played Rg2 at D5 and D6. After forcing Kf3, the candidate finds …Re8 at D5 but still evaluates the position at +442 cp for White. Even at D9 it chooses …Re8 and scores +349 cp. It returns to the right move at D7 before it correctly evaluates the losing line. This motivated the coordinated king-pressure trial below. [UCI replay at depths 4 through 9](game25-depths.json)

## Match selection

InteriorTT scored **69.01%** against the 1850 baseline on the 192 development openings (202 W, 126 D, 56 L). The estimated difference was **+139.1 Elo**, with a two-sided 95% interval of **+64.7 to +228.2**. As White it had 102 W, 65 D, 25 L. As Black it had 100 W, 61 D, 31 L. Median reported depth rose from 8 for the baseline to 9 for the candidate. There were no incidents, and all opening and color reversals passed the checks. The selected binary's SHA-256 was `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`. [Selection PGN](../../../../matches/elo2200/loop/interior-base.pgn) · [Manifest](../../../../matches/elo2200/loop/interior-base.pgn.meta.json)

The control against Clock isolates move ordering below the root. Across 64 pairs InteriorTT had 47 W, 46 D, 35 L (54.69%, +32.7 Elo [−87.2, +161.2]). The size of the ordering contribution remains uncertain. The +139.1 Elo against the baseline applies to the complete configuration. [Isolated duel](../../../../matches/elo2200/loop/interior-clock.pgn) · [Manifest](../../../../matches/elo2200/loop/interior-clock.pgn.meta.json)

## Follow-up experiment on Game25

Variant `54cccef` (KingAttack) starts from InteriorTT. It rewards attacks where a queen and a second piece control squares next to the king. The bonus grows with the number and weight of those contacts. The hypothesis and tactical measurements use only historical project positions. InteriorTT remains the only candidate tested on confirmation set 01.

At D6 the Game25 move changes from Kf3 to Qe1. SF18 scores Qe1 at +100, +129 and +132 cp at 80,000, 320,000 and one million nodes. At the last budget, Kf3 scores −208 cp and Rg2 +153 cp. The same 32 v1 and 13 v2 solutions are found. One v2 false loss alert disappears. Run6 retains ten errors avoided and two losses recognized. Work across 77 positions rises from 2,964,949 to 3,073,432 nodes (+3.7%). A development duel on 64 fixed v2 openings with reversed colors measures the effect in play. [Tactical results](attack-tactics.json) · [Qe1 adjudication](attack-game25.json)

On the 92 historical turning points KingAttack recovered five defenses and lost three against InteriorTT, searching 12,177,692 nodes versus 12,227,917. Against the 1850 baseline on the same suite it recovered six and regressed on one. [Decisions and costs](attack-defenses.json)

KingAttack scored **48.83%** against the qualified baseline over 128 games (37 W, 51 D, 40 L), or −8.1 Elo [−132.1 to +113.7]. There were no abnormal terminations. Game25 was recovered, but the duel did not establish a playing gain. The variant remains on `codex/2200-coordinated-attack`. **Decision at the end of the cycle.** Retain InteriorTT as the 2200+ baseline. The seven matches in the cycle totaled 2,048 games. [KingAttack duel](../../../../matches/elo2200/loop/attack-interior.pgn) · [Manifest](../../../../matches/elo2200/loop/attack-interior.pgn.meta.json)
