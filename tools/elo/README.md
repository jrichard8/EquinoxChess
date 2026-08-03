# Paired Elo harness

This harness is the primary gate for search and evaluation changes. It runs the
candidate directly against the baseline from fixed opening positions. Every
opening is played exactly twice with colors reversed, so opening and color
effects are controlled within each pair.

`openings-v1.epd` contains 32 unique positions after 8 plies. All positions
have White to move. The default run therefore contains 64 games.
`openings-v2.epd` extends it to 64 unique positions while preserving v1 as an
exact prefix. Use v2 explicitly for 128-game confirmation runs.

## Build two immutable engines

Build the baseline before editing the engine, then build the candidate after
the change. Keep the binaries outside the repository so a rebuild cannot
silently replace either side of the comparison.

```bash
go build -o /tmp/equinox-baseline ./cmd/uci
# Apply one isolated change, then:
go build -o /tmp/equinox-candidate ./cmd/uci
```

## Run the paired match

Engine A is the candidate and engine B is the baseline in the recommended
invocation:

```bash
go run ./cmd/elo_match \
  -engine-a /tmp/equinox-candidate \
  -engine-b /tmp/equinox-baseline \
  -name-a Candidate \
  -name-b Baseline \
  -pgn /tmp/equinox-candidate-vs-baseline.pgn
```

Defaults are `tools/elo/openings-v1.epd`, all 32 openings, `1+0.1`, and one
concurrent game. Useful explicit overrides are `-openings-count`, `-tc`,
`-concurrency`, and `-cutechess`. Repeat `-option-a Name=Value` or
`-option-b Name=Value` to configure UCI options for only that engine.

For example, this calibrates Equinox against Stockfish limited to Elo 1600,
with reproducible single-threaded and hash settings:

```bash
go run ./cmd/elo_match \
  -engine-a /tmp/equinox-candidate \
  -engine-b /path/to/stockfish \
  -name-a Equinox \
  -name-b Stockfish-1600 \
  -option-b UCI_LimitStrength=true \
  -option-b UCI_Elo=1600 \
  -option-b Threads=1 \
  -option-b Hash=16 \
  -openings tools/elo/openings-v2.epd \
  -pgn /tmp/equinox-vs-stockfish-1600.pgn
```

The runner refuses to overwrite either the PGN or its
`<match>.pgn.meta.json` sidecar. After a successful, structurally valid match,
the sidecar records the UTC timestamp, Git commit at match start when available,
absolute paths, full cutechess arguments, parameters, cutechess version, and
SHA-256 hashes of both engine binaries, cutechess, the opening suite, and the
PGN. The executable and opening hashes are captured before play and verified
again after play; the match is rejected if any input changed. No resign, draw,
or tablebase adjudication is enabled by the harness, and engines are kept alive
between games with `restart=off`.

## Re-analyse an existing PGN

```bash
go run ./cmd/matchstats \
  -pgn /tmp/equinox-candidate-vs-baseline.pgn \
  -engine-a Candidate \
  -engine-b Baseline
```

The analyser requires exactly two games per `Round`, the same valid six-field
opening FEN, and reversed engine colors. It reports W-D-L, score, color
breakdown, abnormal terminations by losing engine, Elo, and a bilateral 95%
Hoeffding confidence interval. The interval is computed from the mean score of
each opening pair, rather than treating the two correlated games as independent
observations. The Hoeffding bound is deliberately conservative and remains
non-degenerate even if every pair has the same result.

Decision rule: an interval entirely above 0 Elo is a gain; an interval entirely
below 0 is a loss; an interval containing 0 is inconclusive. An inconclusive
result is not evidence of improvement and should trigger more paired openings
from a larger versioned suite before promotion.

The paired candidate-vs-baseline match is the gate for accepting an isolated
change. A separate match against strength-limited Stockfish remains useful for
absolute calibration, but its batch-to-batch noise is not the primary change
detector.
