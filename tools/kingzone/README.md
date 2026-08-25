# Offline king-zone diagnostic

This tool replays an annotated PGN with `python-chess` and describes whether
geometric bishop, rook, or queen pressure on Equinox's king ring is associated
with later evaluation disagreement. It is diagnostic evidence, not causal
proof, and it does not modify or invoke the engine.

## Setup and run

```bash
python3 -m venv /tmp/equinox-kingzone-venv
/tmp/equinox-kingzone-venv/bin/pip install -r tools/kingzone/requirements.txt
/tmp/equinox-kingzone-venv/bin/python tools/kingzone/analyze.py \
  --pgn matches/sf1600-run6-1+0.1-30.pgn \
  --canaries tools/kingzone/canaries.csv \
  --engine EquinoxChess \
  --csv-out /tmp/kingzone-run6.csv \
  --json-out /tmp/kingzone-run6.json
```

Defaults are a 300 cp severe threshold and a 3000 cp mate cap. The mate cap is
used only in the explicitly separated sensitivity section. Primary continuous
statistics include only finite cp-vs-cp pairs. A positive mate reported by the
opponent is `loss_mate=true` and severe; a negative mate is not severe. Engine
mate scores and missing/unparseable scores are excluded with explicit counts.

The JSON pins SHA-256 hashes for the PGN, canary CSV, and analyzer. For the
versioned run6 source, verify the report records 30 games, 3281 plies, 15 games
with Equinox as White, 15 as Black, and all 13 exact-FEN canaries matched. Its
pinned SHA-256 is
`1c98886a75d26560326cbfae249742af76ea7dfe813e3d8e3fd66fa74cf0d4dc`.
The hash is also recorded in `inputs.pgn.sha256`; compare it when sharing or
repeating results rather than relying on a mutable filename.

## Features and interpretation

The king zone is the legal-board eight-square king ring; the king's own square
is excluded. For each enemy bishop, rook, and queen, the CSV records attacker
count, sum of ring squares hit, and unique union squares. `board.attacks()` is
occupancy-aware: the first friendly or enemy blocker is included and squares
beyond it are excluded. Pinned pieces still count geometrically. Detailed
x-rays, blocker identity, and contact-check classes are deliberately omitted in
this first diagnostic.

Unprefixed features describe the position before Equinox's move. `post_*`
describes the position after that move but before the opponent response, using
the same Equinox king as victim; `delta_*` is POST minus PRE. `placebo_*`
measures Equinox's pieces against the opposing king in the PRE position.
Selection must use PRE, non-check, midgame (`non_pawn_material > 2000`) results;
global, midgame, non-check, POST/delta, and placebo sections are descriptive
cross-checks.

The primary binary outcome is `severe`: a positive opponent mate or finite
`engine_cp + adverse_cp >= 300`. The signed finite sum is `loss_swing_cp`, and
`optimism_gap_cp=max(0, loss_swing_cp)` is retained only as a one-sided CSV
diagnostic. Per-feature output includes severe rates, raw risk difference,
Haldane-corrected odds/risk ratios, median signed-swing delta,
Pearson/Spearman correlation, exposed-game count, and deterministic 10,000-draw
game-cluster bootstrap percentile intervals (seed `20260824`). Canaries also
compare each target with the previous Equinox decision (ply - 2) using an exact
McNemar test. `midgame_noncheck_without_canaries` repeats the primary analysis
without the historically selected target rows.

Do not select more than one piece type from this experiment. A type is eligible
only if all of these hold in the PRE midgame/non-check analysis:

- at least 4 of 13 canaries are exposed;
- at least 10 games and 20 rows are exposed;
- raw severe risk difference is at least +10 percentage points;
- finite median signed-swing delta is at least +50 cp;
- the game-cluster bootstrap risk-difference 95% lower bound is positive;
- Spearman is positive globally and in the midgame/non-check subset.
- the same-type placebo is materially weaker and its risk-difference interval
  includes zero;

These thresholds guard against tiny or phase-confounded samples; meeting them
still does not establish that adding an evaluation term will improve Elo.

## Tactical reply diagnostic

Each retained Equinox decision also records the opponent's immediate PGN reply
from the POST position: UCI/SAN, capture, check, capture-check, quiet check,
promotion, and whether the reply position was already in check. Categories may
overlap; a `quiet check` is specifically a non-capture, non-promotion check so
that promotions remain independently visible.

`queen_net_danger` is fixed as PRE `queen_attackers -
placebo_queen_attackers`; positive values mean that more opposing queens touch
Equinox's king ring than Equinox queens touch the opposing ring. The primary
tactical subset remains midgame (`non_pawn_material > 2000`), PRE non-check,
with exact canary rows excluded, and then selects severe rows with positive net
queen danger.

The qsearch flag is theoretical eligibility, not a search trace. It mirrors the
current implementation: all legal evasions are eligible when the side to reply
is already in check; otherwise only captures pass the first filter. A
non-promotion capture may still be delta-pruned depending on alpha and the
stand-pat, which the PGN cannot reconstruct. Consequently a non-capturing
promotion is currently filtered before the later promotion exception, and an
ordinary quiet check is invisible. This does not prove that the recorded search
reached a quiescence frontier at the POST position.

Before inspecting the run6 result, the gate for proposing one targeted
quiet-check search experiment is fixed as all of:

- at least 20 primary severe/net-danger rows from at least 10 games;
- invisible quiet checks form a strict majority (more than 50%) of that subset;
- their rate exceeds the severe non-net-danger control by at least 15 percentage
  points.

Failure means no quiet-check search change is justified from this corpus. A
pass permits one isolated search experiment; it is still observational rather
than causal evidence.

## Tests

```bash
python3 -m unittest discover -s tools/kingzone -p 'test_*.py'
PYTHONPYCACHEPREFIX=/tmp/kingzone-pycache python3 -m compileall -q tools/kingzone
```
