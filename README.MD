# Owl-Gopher

Owl-Gopher (anciennement EquinoxChess) est un moteur d’échecs classique écrit en Go, jouable en UCI et dans le navigateur. **La base `baseline-2200` marque 60,16 % contre Stockfish 18 réglé à 2200**, à 1 seconde + 0,1 seconde par coup. Sur 384 ouvertures indépendantes jouées dans les deux couleurs, elle obtient 430 victoires, 64 nulles et 274 défaites. L’écart estimé à cette cadence est de +71,6 Elo [IC 95 % +22,4 à +123,7]. [Qualification et parties](tools/elo/2200/loop/README.md#confirmation-du-palier-2200) · [Bilan du projet](status.md)

Face au réglage 2400, la base marque **38,67 %** sur 384 parties. Le candidat qui revérifie une défense calme atteint **45,05 %** sur le même lot. Son gain en duel direct reste incertain. [Campagne 2400, parties et décisions](tools/elo/2400/README.md)

La version précédente `baseline-1850` reste disponible pour les comparaisons. Le match direct donne 69,01 % à la nouvelle base sur 384 parties. [Qualification 1850](tools/elo/1850/README.md) · [Première mesure à 2200](tools/elo/2200/README.md)

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
go build -o owl-gopher-uci ./cmd/uci
./owl-gopher-uci
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
go test -tags 'searchtrace trace1850' ./...
```

- [Guide de développement](docs/developpement.md) — architecture, conventions et contrôles à exécuter selon le changement.
- [Lot tactique v1](tools/tactics/README.md) pour les régressions et [lot v2](tools/tactics/v2/README.md) pour les coups calmes difficiles. Chaque lot comprend 32 positions avec une référence de septembre à profondeur 6.
- [Protocole Elo](tools/elo/README.md) — comparer deux binaires sur les mêmes ouvertures, couleurs inversées.
- [Analyse des parties](tools/kingzone/README.md) et [traceur de recherche](cmd/search_trace/README.md) — reconstruire les positions diagnostiques depuis le PGN et examiner les décisions du moteur.
