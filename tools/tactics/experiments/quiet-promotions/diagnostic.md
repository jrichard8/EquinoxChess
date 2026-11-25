# Two effects of non-capturing promotions in quiescence

**September (`f6871de`) remains the baseline for this experiment.** Both positions were replayed at depth 6 against the experiment's only other revision, `5361155`. On Game17 the candidate extends capture sequences in quiescence. On 0JLoe a promotion changes root move ordering and a reduced-depth search then fails to distinguish the two pawn pushes. All four instrumented searches reproduced the move, score and node count from the [paired comparison](paired.json).

## Game17_W_ply17 · accounting for 4,253,226 extra nodes

| Node visits across iterations 1–6 | September | Candidate | Difference |
|---|---:|---:|---:|
| Main search | 53,407 | 53,407 | 0 |
| Quiescence before any non-capturing promotion admitted by the candidate | 2,734,644 | 2,640,291 | −94,353 |
| Quiescence from the first such promotion, including its child | 0 | 4,347,579 | +4,347,579 |
| **Total** | **2,788,051** | **7,041,277** | **+4,253,226** |

Both engines entered quiescence **41,694 times** from main search. The ordered sequence of main-search visits had the same fingerprint, including the alpha-beta windows (`49d84c88663f124423d6634698567037136de60da6338deaed5895627132b3ff`). Both engines chose `d1f3` with identical scores throughout the six iterations, reaching +67 cp at depth 6.

**The promoting pawn can start in the middle of the board.** In the most expensive subtree reached after a first promotion, main search plays `d1f3 b8d7 a2a3 e7d6 g1h1`. Quiescence continues `c5d4 c3d5 d4e3 f3f6 e3d2 f6d8`. Black's c5-pawn has made three captures and reached d2. The candidate then searches `d2d1q`, which opens a subtree of **25,786 nodes** in this occurrence alone. September filters the non-capturing push here. This is one line explored during search.

The candidate visits **198,871 first promotions** outside check along their respective paths. The probe counts each subtree once, even when it contains further promotions. Black promotions on d1 and b1 account for 87.6% of those 4,347,579 nodes. Queen promotions account for 2,563,766 (59.0%) and the three underpromotions combined for 1,783,813. Most of the work comes from the continuations: **3,785,952** nodes are reached by a capture and **788,280** by an evasion, with overlap when a capture answers a check. Outside check, promotions are exempt from delta pruning. In check, the engine searches legal evasions. Other quiescence visits fall by 94,353, offsetting part of the increase.

## 0JLoe · why h5h6 loses priority

The suite's Stockfish reference gives `h5h6 g7h6 g5h6 e7f8 e5f6 f8g8 h6h7`. The first move sacrifices the h-pawn so that the g-pawn can recapture on h6. The scores below are Equinox's, in centipawns from White's perspective. [Position and reference analysis](../../v2/positions.json)

1. **The first difference occurs on a third move, `g5g6`.** At depth 1, after `g5g6 f3h5 g6f7 h5f7`, the candidate can still play `a7a8q` in quiescence. In the window actually searched, `g5g6` rises from −90 to −29. At depth 2 it stays at the −54 bound for September but reaches −39 for the candidate. That improves alpha and earns a history bonus in the candidate. Both engines still finish both iterations on `f5f6` at −4.
2. **The bonus changes which move is tried first.** At depth 3, after the four immediate promotions on a8, September ranks `h5h6` and `f5f6` at 1,002 ahead of `g5g6` at 1,000 and tries `h5h6` first. In the candidate all three score 1,002 and `g5g6` appears first in the list. During its search the ordering score of `f5f6` rises to 1,003, ahead of `h5h6`. The latter returns to the lead at depth 4. At depth 5 history again favors `f5f6` by 1,013 to 1,008, then by 1,019 to 1,005 at depth 6.
3. **The reduced search accepts a shallow tie.** At depths 5 and 6, `f5f6` sets alpha to +26. The candidate then probes `h5h6` with late move reduction (LMR) at child depth 3 both times. Black's window is [−27, −26]. The transposition table already contains an exact −26 score **at depth 3**, with `g7h6` as the reply. The probe therefore returns +26 for White in **one node**. A full-depth re-search requires a score strictly above alpha. The tie leaves `f5f6` in place.

| Full-depth, open-window control | September, h5h6 | Candidate, h5h6 | September, f5f6 | Candidate, f5f6 |
|---|---:|---:|---:|---:|
| Root depth 5, child depth 4 | +46 | +46 | +26 | +26 |
| Root depth 6, child depth 5 | **+38** | **+38** | +26 | +26 |

Each control starts from a copy of the board, table and heuristics taken just before the reduced search of the move in question. It uses the revision's search rules and opens the window at full depth **for that root move**. Reductions still apply inside the subtree. At depth 6 the candidate's control spends 2,538 nodes on `h5h6` and recovers +38. The control result is discarded before the observed search continues. The +26 used in play explains why the engine kept `f5f6`. It came from a shallower search of `h5h6` than the requested depth.

## Reproduce the observations

From the repository root, with Go 1.26 and both revisions available locally, these commands apply the [offline probe](diagnostic.patch) only to temporary copies. Each position starts with a fresh table: 65,536 entries for Game17 and 1,048,576 for 0JLoe. The SHA-256 fingerprint covers the Zobrist keys, halfmove counters, plies, depths and windows at every main-search entry. Control searches have their own counter and fingerprint.

```sh
set -eu
probe="$PWD/tools/tactics/experiments/quiet-promotions/diagnostic.patch"
replay=$(mktemp -d)
for revision in f6871de 5361155; do
  mkdir "$replay/$revision"
  git archive "$revision" | tar -x -C "$replay/$revision"
  (
    cd "$replay/$revision"
    git apply "$probe"
    go test -count=1 -run '^TestQuietPromotionDiagnosis$' -v ./internal/search
  ) > "$replay/$revision.log"
done
echo "$replay"
```
