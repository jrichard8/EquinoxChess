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

Schema 4 reports the two canary outcomes independently at every replay depth:
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
