# Targeted search trace

`search_trace` is an offline, opt-in diagnostic. It replays the 19 primary
invisible quiet-check replies selected by `tools/kingzone/analyze.py` plus all
13 exact canaries. Normal engine searches keep `Search.Trace == nil`; the
trace hooks do not change evaluation, ordering, pruning, reductions, or qsearch
rules.

```bash
go build -tags searchtrace -o /tmp/equinox-search-trace ./cmd/search_trace
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-search-trace-depth4-6-run6-20260825.json \
  -depths 4,6
```

Targeted counterfactuals require an explicit mode and exact case allowlist:

```bash
target_cases=Game07_ply149,Game07_ply177,Game07_ply183,Game17_ply95,Game25_ply35,Game25_ply69,Game30_ply52
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-target-reply-plus-one.json \
  -depths 6 \
  -counterfactual-mode plus-one \
  -counterfactual-cases "$target_cases"
```

`full-depth` removes the target reply's normal LMR reduction; `plus-one`
searches one ply beyond that normal full depth. Both can affect only a named
exact `RootMove/ReplyMove` pair at ply 1 of the requested final iteration that
gives check. `full-depth` still requires current LMR eligibility. `plus-one`
is reserved for replies preselected as baseline-LMR and continues to depth
normal+1 if its own causal effects later make the exact reply a killer or
otherwise change its ordering status. Earlier iterations, unnamed cases,
every other checking move, and all non-checking replies remain observational.
Counterfactual runs are repeated with fresh searches and TTs and must match
exactly.

Every case is run with a fresh 65,536-entry TT at each requested depth. In
observational mode, the command repeats the same search without tracing and
requires best move, score, completed depth, and node count to match exactly.
In counterfactual mode, it repeats the targeted trace with another fresh TT and
requires the result and continuation records to match exactly. The JSON pins
both input CSV and executable hashes.

Schema 11 can additionally trace every move of one exact deeper branch. The
bridge case must also be the sole or one of the explicitly selected
`full-depth` cases:

```bash
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-target-bridge.json \
  -depths 6 \
  -counterfactual-mode full-depth \
  -counterfactual-cases Game25_ply69 \
  -bridge-case Game25_ply69 \
  -bridge-line a3a4,b7c7,c4b3,c7c3,b3a2,b4b3
```

The command validates that the line begins with the corpus root/reply pair,
that every move is legal, and that the final bridge move is a quiet check.
Intermediate moves use the `bridge_step` role and the final move uses
`bridge_move`; normal generation, ordering, legality, pruning, LMR, TT and
qsearch events then identify the first point where the exact path disappears.
This mode remains observational unless a separately documented exact
counterfactual is enabled.

One intermediate quiet check on that exact line can be exempted from LMR with
an additional explicit flag:

```bash
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-target-nested-check-nolmr.json \
  -depths 6 \
  -counterfactual-mode full-depth \
  -counterfactual-cases Game25_ply69 \
  -bridge-case Game25_ply69 \
  -bridge-line a3a4,b7c7,c4b3,c7c3,b3a2,b4b3 \
  -bridge-no-lmr-move c7c3
```

The command requires the named move to occur exactly once between the root
reply and final bridge move, and verifies that it is legal, quiet, and gives
check. The counterfactual applies only when the exact preceding path and final
requested iteration match and the move is normally LMR-eligible. It changes
the reduced child depth to the normal child depth without adding a ply;
earlier iterations, other paths, other checks, TT, futility, qsearch, and the
normal engine remain unchanged. The run exposes the application through
`bridge_no_lmr_applied` and `negamax.bridge_step_forced_full_depth`.

The exact final quiet check can additionally bypass only a futility decision
that would otherwise prune it. This mode deliberately depends on the exact
intermediate NoLMR gate so that the experiment changes the two already
observed links of one causal chain and nothing else:

```bash
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-target-nested-bridge-nofutility.json \
  -depths 6 \
  -counterfactual-mode full-depth \
  -counterfactual-cases Game25_ply69 \
  -bridge-case Game25_ply69 \
  -bridge-line a3a4,b7c7,c4b3,c7c3,b3a2,b4b3 \
  -bridge-no-lmr-move c7c3 \
  -bridge-no-futility-move b4b3
```

`-bridge-no-futility-move` must equal the validated final bridge move and
requires `-bridge-no-lmr-move`. It has no effect unless the normal depth-1
futility predicate is true after the complete exact prefix. The event
`negamax.bridge_move_forced_no_futility` records that actual bypass; the move
then follows the unchanged legality, terminal, full-depth and qsearch paths.

Schema 11 also supports a corpus-wide structural probe. The seven exact target
replies previously observed under LMR are first searched full-depth; every
remaining-depth-3 quiet check that is still LMR-eligible is then replayed at
normal depth in an isolated shadow:

```bash
target_cases=Game07_ply149,Game07_ply177,Game07_ply183,Game17_ply95,Game25_ply35,Game25_ply69,Game30_ply52
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-d3-d1-chain-probe.json \
  -depths 6 \
  -counterfactual-mode full-depth \
  -counterfactual-cases "$target_cases" \
  -probe-d3-check-chains
```

Each shadow receives value copies of the board and search heuristics plus a
deep copy of the TT taken immediately before the baseline reduced search. Its
score and nodes are discarded by the live search. The report records every D3
probe and every legal quiet check that the unchanged depth-1 futility predicate
would prune two plies later. Main best move, score, depth and node count remain
subject to the normal non-interference or deterministic counterfactual check.

Once that structural motif has been established beyond one exact line, the
coupled final-iteration candidate can be measured with:

```bash
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-general-d3-d1.json \
  -depths 6 \
  -counterfactual-mode full-depth \
  -counterfactual-cases "$target_cases" \
  -general-d3-d1-checks
```

The generalized counterfactual searches every normally LMR-eligible quiet
check at remaining depth 3 with child depth 2. A depth-1 futility exemption is
enabled only while searching that exact D3 subtree, exactly two plies below,
and only for a legal quiet check that meets the unchanged futility predicate.
It does not extend the search, alter TT/qsearch/evaluation, or affect earlier
iterations. All 32 candidate runs are repeated with independent searches and
TTs and must match exactly.

The allowlist-free `D5D3D1CheckChain` experiment adds the outer depth-5 stage:

```bash
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-general-d5-d3-d1.json \
  -depths 6 \
  -general-d5-d3-d1-checks
```

This mode requires the default `observational` target-reply mode and rejects
all case allowlists, bridge/probe modes, and `-general-d3-d1-checks`. In the
requested final iteration it searches every normally LMR-eligible quiet check
at remaining depths 5 and 3 with normal child depths 4 and 2. The depth-5
stage is an outer NoLMR rule; the depth-1 NoFutility rule remains strictly
coupled to a depth-3 check forced by this same candidate exactly two plies
above. It adds no extension and does not alter TT, qsearch, evaluation, or
earlier iterations. Events use `d5_check_forced_full_depth`,
`d3_check_forced_full_depth`, and `d1_check_forced_no_futility`.

One exact depth-5 path can be ablated from that generalized candidate for a
diagnostic leave-one-out run:

```bash
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260825.csv \
  -output /tmp/equinox-d5-ablation-qe2-qh1.json \
  -depths 6 \
  -general-d5-d3-d1-checks \
  -d5-ablation-case Game25_ply35 \
  -d5-ablation-line d1e2,h3h1
```

Both ablation flags are required together. The case must exist, the line must
contain exactly one legal root move followed by one legal quiet check, and the
general D5→D3→D1 candidate must remain enabled. Only that exact D5 occurrence
keeps its baseline LMR sequence (`depth 5 -> reduced child 3`, null-window
search, and an ordinary full re-search if it fails high); every other D5 and
the generalized D3→D1 coupling stay active. The event
`d5_check_ablation_lmr_preserved` records the actual application, including
the normal reduction and child depth. This is an attribution mechanism, not a
production allowlist.

The observation-only `D5ReducedFullShadow` mode compares every quiet check
that the normal final-iteration search really reduces at remaining depth 5:

```bash
/tmp/equinox-search-trace \
  -input /tmp/equinox-kingzone-tactical-run6-20260909.csv \
  -output /tmp/equinox-d5-reduced-full-shadow.json \
  -depths 6 \
  -probe-d5-reduced-full
```

Each occurrence starts from one exact child board, TT, killer and history
snapshot. Three isolated views are recorded: the normal reduced depth-3 null
window, a fresh full depth-4 window, and the normal sequential reduced search
followed by a full re-search only when the reduced score exceeds alpha. The
report preserves scores, alpha crossings, PVs, TT hits/cutoffs and writes,
first ordered moves, killer/history differences, plus every D3 probe and
D3-to-D1 futility chain present in the fresh full subtree. Shadow state and
nodes never return to the main search. The command repeats the traced search
with a fresh TT, requires identical shadow records, then requires the main
move, score, depth and node count to equal a third untraced search exactly.

This mode requires `observational`, no case allowlist and no bridge, D3 probe,
generalized candidate or D5 ablation. It is an attribution probe, not an LMR,
TT, ordering, pruning or evaluation change.

The report also includes the two canary outcomes independently at every replay
depth:
`bad_move_avoided` means the selected move differs from the historical
Equinox move stored in the corpus; `loss_recognized` means the root score is
at most `-300 cp` from the side-to-move POV. `historical_bad_move_selected`
is the complement of avoidance. The legacy `root_move_best` field is retained
and has the same value as `historical_bad_move_selected`.

Each searched target reply also publishes `continuations` and one
`final_continuation` for the decisive pass. It records reduced/full mode,
parent and child windows, returned score, the best explored UCI line after the
reply, and the propagated qsearch/terminal leaf. `score_root_pov` and
`leaf_score_root_pov` normalize the two native node scores back to the root
side. With no extensions, `child_depth` is also the number of main-search
plies after the reply before qsearch; qsearch captures may extend the reported
leaf farther.

`final_root_move_results` records every legal root move in the decisive pass,
including its pre-search alpha/beta window and whether the returned score is
exact, an upper bound (fail-low), or a lower bound (fail-high). `root_margin`
compares the historical bad root move with its best competitor and retains the
corresponding qualification: the reported centipawn difference may be exact,
a lower bound, an upper bound, or indeterminate. This avoids treating normal
alpha-beta fail-low values as exact move scores.

Events distinguish root and reply moves and record iterative depth, aspiration
or full-window pass, pseudo-legal and ordered indices, legal index, TT/PV/killer
priority, futility, null-move and beta cutoffs, LMR reduction/full re-search,
terminal handling, and qsearch generation/filter/delta-pruning/search. TT probe
events also publish key/halfmove/depth eligibility, bound type, stored score,
and stored move.

Use `summary.final_pass` for causal interpretation: it is the full-window
re-search when aspiration failed, otherwise the aspiration pass. Aggregates are
case-level booleans for that decisive pass. `reply_ever_qsearch_*` deliberately
covers all iterative depths and proves only what happens when the immediate
reply node itself reaches qsearch. `target_line_qsearch_entered_final` means a
descendant after the recorded reply reached qsearch; it does not show that a
later reference continuation was present there.
