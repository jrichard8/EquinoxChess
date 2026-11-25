# Non-capturing promotions at the search horizon · November 23, 2026

**The candidate was rejected in this form.** Quiescence now sees a winning promotion and a winning underpromotion in the control positions. Across the three frozen suites it loses one v2 solution, gains none and searches 39.29% more nodes overall. The September engine remains the playing baseline for this experiment.

## Isolated change and direct checks

Candidate `5361155` searches all four non-capturing promotions in quiescence when out of check. Both engines use the same move generation, check evasions, stand pat and existing exemption for promotions in delta pruning. The candidate is on `codex/qsearch-quiet-promotions`.

| Control position, quiescence only | September → candidate score (cp, side to move) | Nodes |
|---|---:|---:|
| White promotes to a queen, `a7a8q` | +220 → +881 | 1 → 5 |
| Mirrored position for Black, `a2a1q` | +220 → +881 | 1 → 5 |
| Rook underpromotion `c7c8r` avoids stalemating with a queen | +250 → +526 | 1 → 6 |
| Capture control, `a7b8q` | +613 → +613 | 5 → 5 |

## Depth-6 comparison

| Suite | September → candidate | Newly solved / lost | September nodes | Candidate nodes | Difference |
|---|---|---:|---:|---:|---:|
| v1, 32 positions | Reference moves 32 → 32. Losses recognized 15/16 → 15/16, false alarms 0 → 0. | 0 / 0 | 1,480,222 | 1,574,941 | +94,719 (+6.40%) |
| v2, 32 quiet moves | Reference moves 13 → 12. White 8/16 → 7/16, Black 5/16 → 5/16. Both versions raise three false alarms in the same positions. | 0 / 1 | 3,289,616 | 3,400,898 | +111,282 (+3.38%) |
| run6, 13 canaries | Historical mistakes avoided 10 → 10. Losses recognized 2 → 2. | 0 / 0 | 7,772,789 | 12,495,002 | +4,722,213 (+60.75%) |
| **Total, 77 positions** | | **0 / 1** | **12,542,627** | **17,470,841** | **+4,928,214 (+39.29%)** |

The v2 regression is **`0JLoe`**. September found `h5h6` (+38 cp). The candidate chooses `f5f6` (+26 cp). At depth 18, Stockfish scored `h5h6` at +522 cp and its second choice, `e5d4`, at +60 cp. On **`0Np2W`**, the only other changed decision, both versions miss `g2g5`. September plays `g2g3` and the candidate plays `f3f2`. The other 75 first moves match.

One position dominates the cost. On **`Game17_W_ply17`**, both engines repeat the historical error `d1f3` and report +67 cp. The candidate searches 7,041,277 nodes against 2,788,051, accounting for **90.1% of the extra run6 nodes**. In v1, `G18-p26` accounts for 81,075 of the suite's 94,719 extra nodes. These counts measure work at fixed depth. No Elo match was played for this experiment.

The two-position diagnosis traces the black promotions responsible for the Game17 cost and the reduced-depth cutoff on 0JLoe. [Traces, control search and reproduction](diagnostic.md)

## Provenance and reproduction

- The historical binaries correspond to current revisions `f6871de` and `5361155`. Both were built from clean trees with Go 1.26.2 on macOS arm64. [paired.json](paired.json) records executable and input checksums and all 77 paired results. The `cp` and `mate` fields are from the side to move's perspective.
- For v1 and v2, run [measure.py](../../measure.py) with each binary and both frozen `positions.json` files. Two fresh UCI processes per position must return the same move, score, depth and node count. Compare rebuilt binaries by position. The versioned reports record the historical binaries' Git metadata. The UCI table has 1,048,576 entries.
- For run6, rebuild the [September tactical CSV](../../../kingzone/README.md) from the PGN (full CSV SHA-256 `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`). Keep the 13 rows with a nonempty `canary_name`. Build each [tracer](../../../../cmd/search_trace/README.md) with `-tags searchtrace` and run it in observation mode with `-depths 6 -tt-size 65536 -expected-primary 0`. Each traced result was checked against a fresh untraced search. The table counts one main search per position. Recognized losses include scores ≤ −300 cp and forced mates against the side to move.
- The candidate passes the standard Go suite and `go vet -tags searchtrace ./...`. With `-tags searchtrace`, three historical tests fail on their frozen node counts (`TestSearchTraceBridgeLocatesNestedLMRTTCutoff`, `TestSearchTraceD5BranchAblationsStayOnExactPath` and `TestSearchTraceD5ReducedFullShadowIsDeterministicAndNonInterfering`). All three pass on September. Their expected values were retained.
