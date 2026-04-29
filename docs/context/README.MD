# EquinoxChess Context

EquinoxChess est un moteur d'echecs classique ecrit en Go.

Ce dossier documente l'etat reel du projet au 2026-04-27, pas une vision cible abstraite.

## Etat Actuel

- Representation bitboard complete avec hash Zobrist incremental.
- Generation pseudo-legale et legale des coups.
- `make/unmake` et `null move` disponibles.
- Perft valide sur la position initiale, Kiwipete, en passant et roques.
- Recherche alpha-beta de type negamax avec iterative deepening.
- Quiescence search, aspiration windows, transposition table, killer moves, history heuristic, null-move pruning et LMR deja en place.
- Interface UCI fonctionnelle via `cmd/uci`.
- Petit serveur HTTP + UI web via `cmd/server`.

## Limites Observees

- La force pratique reste nettement insuffisante face a Stockfish limite a 1600 Elo.
- Le taux de nulles par pat est anormalement eleve, signe probable de mauvaise conversion des positions gagnantes.
- L'evaluation reste simple et tres materialiste.
- Le protocole de regression Elo n'est pas encore structure autour de suites fixes, SPRT et positions temoins.
- La documentation etait en retard sur plusieurs iterations du moteur.

## Priorites

- Re-aligner la documentation sur le code courant.
- Stabiliser la mesure de force avant toute nouvelle heuristique.
- Corriger les causes structurelles de pertes de force avant d'ajouter de nouvelles optimisations.
- Construire une roadmap incrementale vers 2000 Elo avec garde-fous de regression.

## Entrees Principales

- `cmd/uci/main.go` : moteur UCI pour cutechess et GUIs.
- `cmd/server/main.go` : API HTTP + interface web locale.
- `internal/board` : etat du plateau, FEN, hash, make/unmake.
- `internal/movegen` : generation de coups.
- `internal/search` : recherche, ordre des coups, evaluation.
- `internal/tt` : transposition table.
- `internal/uci` : parsing et rendu UCI.

## Commandes Utiles

Build UCI :

```bash
go build -o equinoxchess-uci cmd/uci/main.go
```

Tests :

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./...
```

Match rapide :

```bash
go build -o equinoxchess-uci cmd/uci/main.go && \
cutechess-cli \
  -engine cmd=./equinoxchess-uci name=EquinoxChess proto=uci \
  -engine cmd=stockfish name=SF1600 proto=uci option.UCI_LimitStrength=true option.UCI_Elo=1600 \
  -each tc=1+0.1 \
  -games 30
```
