# Référence 1850+ contre Stockfish 2200

Le tag `baseline-1850` pointe sur `536e375`, avec le code de jeu de `e5daafe`. La mesure emploie l’exécutable qualifié contre SF1850, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`. Septembre reste disponible à `819a8d5` pour les comparaisons historiques. [Qualification 1850+](../1850/README.md)

## Protocole fixé avant la première partie

- **192 nouvelles ouvertures**, huit demi-coups depuis la position initiale, deux couleurs par ouverture, 384 parties sans arrêt anticipé. Le générateur de 1850 conserve sa distribution et utilise une nouvelle graine, `220020261124`. Il écarte les positions des anciennes suites EPD et de tous les PGN de matchs du projet à `536e375`, ainsi que du PGN source de v1. La comparaison conserve trait, roques et prise en passant légale et ignore les compteurs. Des positions distinctes peuvent appartenir à la même famille d’ouverture.
- **1 s + 0,1 s par coup**, six parties simultanées sur Apple M4 Pro, Stockfish 18 limité par UCI à 2200, un thread et 16 Mio de table. Le PGN et son manifeste sont produits par `cmd/elo_match`. Lire l’intervalle bilatéral de Hoeffding à 95 % par paire. Le jalon 2200 demande une borne basse du score supérieure à 50 %.
- **Analyse des défaites** avec Stockfish à pleine force, un thread, budgets en nœuds. Pour chaque défaite, rechercher le premier coup d’Equinox qui fait passer une position encore défendable (meilleur coup ≥ −100 cp) à une position nettement inférieure (coup joué ≤ −200 cp), avec une perte d’au moins 150 cp. Vérifier le meilleur coup et le coup joué depuis la même position à budget accru. Conserver les défaites sans bascule confirmée et les résultats par couleur. Chercher ensuite un mécanisme de recherche vérifiable dans les positions confirmées.

Pour régénérer le lot, utiliser le Stockfish du manifeste et les entrées du tag de référence, avec le générateur de ce dossier parent. La liste des PGN ci-dessous est celle du dépôt à `536e375`.

```sh
python3 tools/elo/1850/build_holdout.py /chemin/stockfish /tmp/2200.epd \
  --seed 220020261124 --exclude-epd tools/elo/1850/holdout.epd \
  --exclude-pgn matches/*.pgn matches/elo1850/*.pgn tools/tactics/v1/source.pgn
```
