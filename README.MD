# EquinoxChess

EquinoxChess est un moteur d’échecs classique écrit en Go. Il se joue dans une interface compatible UCI ou dans le navigateur. Pour reprendre le développement, le [bilan du 23 novembre 2026](status.md) rassemble l’activité, la dernière mesure de force, les limites connues et les améliorations proposées.

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
- [Validation tactique indépendante](tools/tactics/README.md) — rejouer la référence de septembre sur 32 positions hors run6.
- [Protocole Elo](tools/elo/README.md) — comparer deux binaires sur les mêmes ouvertures, couleurs inversées.
- [Analyse des parties](tools/kingzone/README.md) et [traceur de recherche](cmd/search_trace/README.md) — reconstruire les positions diagnostiques depuis le PGN et examiner les décisions du moteur.
