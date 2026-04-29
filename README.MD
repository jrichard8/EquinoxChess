# EquinoxChess

## Build

```bash
go build -o equinoxchess-uci cmd/uci/main.go
```

## Tests

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./...
```

## Reference Gauntlet

Short regression run:

```bash
go build -o equinoxchess-uci cmd/uci/main.go && \
cutechess-cli \
  -engine cmd=./equinoxchess-uci name=EquinoxChess proto=uci \
  -engine cmd=stockfish name=SF1600 proto=uci option.UCI_LimitStrength=true option.UCI_Elo=1600 \
  -each tc=1+0.1 \
  -games 30
```

Short regression run with PGN output:

```bash
go build -o equinoxchess-uci cmd/uci/main.go && \
cutechess-cli \
  -engine cmd=./equinoxchess-uci name=EquinoxChess proto=uci \
  -engine cmd=stockfish name=SF1600 proto=uci option.UCI_LimitStrength=true option.UCI_Elo=1600 \
  -each tc=1+0.1 \
  -games 30 \
  -pgnout matches/sf1600-1+0.1-30.pgn
```

Longer validation run:

```bash
go build -o equinoxchess-uci cmd/uci/main.go && \
cutechess-cli \
  -engine cmd=./equinoxchess-uci name=EquinoxChess proto=uci \
  -engine cmd=stockfish name=SF1600 proto=uci option.UCI_LimitStrength=true option.UCI_Elo=1600 \
  -each tc=3+0.1 \
  -games 100
```
