# Independent tactical validation suite · v1

The suite contains one decision per game and per opening. Use [positions.json](v1/positions.json) to compare engine versions. Each entry gives the FEN, game number and ply in the [source PGN](v1/source.pgn), the move played and both reference analyses.

## September baseline at depth 6

The suite was frozen at `769eb3f`. The historical UCI binary was built with Go 1.26.2 on macOS arm64. Its source is now at `f6871de`. Regeneration reproduced the 128-game PGN and 32-position fixture byte for byte. Both searches agreed on move, score, depth and nodes for all 32 positions. [Per-position results and checksums](v1/septembre.json)

| Group | Reference move found | Equinox evaluation ≤ −300 cp | Nodes in one pass |
|---|---:|---:|---:|
| 16 defensible or favorable positions | 16/16 | 0/16 false alarms | 993,345 |
| 16 losses according to Stockfish | 16/16 | 15/16 losses recognized | 486,877 |
| Total | 32/32 | 15 losses flagged | 1,480,222 |

On **G57-p59**, Equinox finds `Kxe1` (`f2e1`) and evaluates the position at −277 cp. Stockfish chooses the same move and scores it at −429 and −416 cp. The difference lies in how the engines evaluate the position after the best move.

Move agreement is already at its ceiling. All 32 reference moves appeared in the source games and 25 are immediate captures. This makes the suite useful for detecting regressions, tracking search cost and checking loss recognition. Measuring progress on harder tactical choices requires another blind sample. Run6 selected historical Equinox errors, so the two hit rates describe different populations.

## How the suite was built

Stockfish 18 generates games from the 64 eight-ply positions in [openings-v2.epd](../elo/openings-v2.epd). One side gets 180 nodes per move and the other 5,000. Colors are reversed for the second game of each opening. Generation ends after 80 additional plies or at the end of the game, so the PGN can contain unfinished games. Annotations identify candidate positions.

The builder excludes an opening found in run6, then discards any candidate occurring **at any ply** in run6 or in more than one game of the new PGN. It compares piece placement, side to move, castling rights and legal en passant captures, ignoring both FEN counters. Selection starts at ply 17. Within each game, SHA-256 and the seed `equinox-tactics-v1` determine the order. Once a position is selected, both games with that opening are closed to further selection.

Stockfish examines the best two moves (MultiPV 2) with one thread, a 16 MiB hash and fresh state at **24,000 and 120,000 nodes**. Both searches must agree on the first move, score it at least 150 centipawns above the runner-up and show a capture, check or promotion within the first three plies of the line. At each budget, the builder takes the last complete pair of lines at a common depth with numeric scores and no bound flags. Scores are from the side to move's perspective.

| Group | Positions | Best move score at both budgets |
|---|---:|---|
| `tactic` · defensible or favorable | 16 | −100 to +600 cp |
| `loss` · losing according to Stockfish | 16 | −900 to −300 cp |

Thresholds and labels use Stockfish scores. The first pass over 64 games produced only 13 eligible losses. The second added the reverse-color games and excluded positions shared by the two passes. The suite was frozen before Equinox searched it. These synthetic games use openings already known to the project's Elo protocol. Results describe the two selected groups at fixed depth.

<a name="relire-la-mesure-ou-comparer-un-candidat"></a>

## Reproduce a measurement or compare a candidate

From the repository root, with Go and the [game analyzer's Python environment](../kingzone/README.md#setup-and-run), build the reference revision in a clean clone and run the harness from the current repository.

```sh
trial=$(mktemp -d)
git clone --quiet "$(git remote get-url origin)" "$trial/reference"
git -C "$trial/reference" checkout --quiet --detach f6871deb67e1f8b8a9783b619010ca2551b8983b
(cd "$trial/reference" && go build -trimpath -buildvcs=true -o "$trial/equinox" ./cmd/uci)
python3 tools/tactics/measure.py "$trial/equinox" "$trial/september.json"
```

To measure a candidate, pass its UCI executable and a new output path to the last command. The harness verifies the input checksums, locates each FEN in the PGN and validates independence and reference moves. It starts a fresh UCI process for each position and repetition at depth 6 without clocks. Both passes must return the same legal move, score and node count at the requested depth. The report records the Git revision embedded by Go, the binary and suite checksums, and each position's result.

Match reports by position ID. For each group, count reference moves gained and lost, recognized losses (Equinox score ≤ −300 cp or a forced mate against it), false alarms in the defensible group, and nodes searched. Stockfish's score margins are estimates at finite search budgets. The move in the PGN identifies the decision and may itself be the best move.

## Rebuild the references

```sh
output=$(mktemp -d)
python3 tools/tactics/build_corpus.py "$(command -v stockfish)" "$output"
```

To reanalyze the saved games, copy `v1/source.pgn` into the output directory before running the command. Measuring a candidate with the frozen references requires no Stockfish installation. The suite records versions and checksums for the adjudicator, openings and PGNs. Use the same python-chess version and Stockfish binary when comparing regenerated checksums.
