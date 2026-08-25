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

## Tests

```bash
python3 -m unittest discover -s tools/kingzone -p 'test_*.py'
PYTHONPYCACHEPREFIX=/tmp/kingzone-pycache python3 -m compileall -q tools/kingzone
```
