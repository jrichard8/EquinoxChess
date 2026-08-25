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

Every case is run with a fresh 65,536-entry TT at each requested depth. The
command repeats the same search without tracing and aborts unless best move,
score, completed depth, and node count match exactly. The JSON pins both input
CSV and executable hashes.

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
