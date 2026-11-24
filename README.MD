# EquinoxChess

EquinoxChess est un moteur d’échecs classique écrit en Go. Il se joue dans une interface compatible UCI ou dans le navigateur. À 1 seconde + 0,1 seconde par coup, la version du 24 novembre 2026 obtient 74,09 % sur 384 parties contre Stockfish 18 réglé à 1850. Le bilan de reprise conserve la référence de septembre et les résultats des versions successives. [Calibration, intervalles et parties](tools/elo/1850/README.md) · [Bilan du projet](status.md)

La base de développement est `baseline-1850`. Sa première mesure indépendante contre Stockfish 2200 donne 51,43 % sur 384 parties à la même cadence. [Diagnostic des défaites et prochaine amélioration](tools/elo/2200/README.md)

## Jouer dans le navigateur

Depuis la racine du dépôt, avec Go 1.26 installé, lancer le serveur puis ouvrir [localhost:8080](http://localhost:8080).

```sh
go run ./cmd/server
```

Déplacer une pièce sur l’échiquier pour jouer contre le moteur. Le serveur gère une partie en mémoire, partagée entre les onglets. La page charge jQuery et chessboard.js depuis des CDN.

| Requête | Effet |
|---|---|
| `POST /newgame` | Repart de la position initiale. |
| `POST /move` avec `{"move":"e2e4"}` | Joue le coup du joueur. |
| `POST /engine` avec `{"depth":4}` | Cherche et joue un coup, puis renvoie `bestmove`, `score`, `depth` et `nodes`. |
| `GET /fen` | Renvoie la position FEN, les coups légaux, le trait et les indicateurs d’échec et de mat. |

## Utiliser le moteur UCI

```sh
go build -o equinoxchess-uci ./cmd/uci
./equinoxchess-uci
```

Dans le terminal du moteur, saisir les commandes suivantes. Le moteur répond successivement par `uciok`, `readyok`, puis par une ligne `info` et un `bestmove`.

```text
uci
isready
position startpos moves e2e4
go depth 6
```

Le moteur accepte aussi une position FEN, un temps par coup (`go movetime 500`) ou les pendules UCI avec incréments. Utiliser `quit` pour terminer.

## Développer et mesurer

```sh
go test ./...
go test -tags searchtrace ./...
```

- [Guide de développement](docs/developpement.md) — architecture, conventions et contrôles à exécuter selon le changement.
- [Lot tactique v1](tools/tactics/README.md) pour les régressions et [lot v2](tools/tactics/v2/README.md) pour les coups calmes difficiles. Chaque lot comprend 32 positions avec une référence de septembre à profondeur 6.
- [Protocole Elo](tools/elo/README.md) — comparer deux binaires sur les mêmes ouvertures, couleurs inversées.
- [Analyse des parties](tools/kingzone/README.md) et [traceur de recherche](cmd/search_trace/README.md) — reconstruire les positions diagnostiques depuis le PGN et examiner les décisions du moteur.
