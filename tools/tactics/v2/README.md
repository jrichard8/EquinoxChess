# Lot v2 · trouver un coup calme

Cette suite mesure le choix du **premier coup** sur 32 problèmes provenant de 32 parties Lichess, avec 16 décisions de chaque couleur. Dans chaque position, le joueur au trait est hors échec. La solution commence par un coup sans prise, échec ni promotion. Le [lot v1](../README.md) reste le témoin de régression sur les décisions plus immédiates.

## Sélection avant la mesure

1. **Source.** Le constructeur prend les 65 536 premières lignes du premier fichier Parquet de l’export Lichess figé à la révision `461725966e`. Il applique le premier coup de la ligne publiée à la FEN source pour obtenir la position du problème. Le deuxième coup est la solution à trouver. Les données de problèmes sont publiées sous CC0. [Export daté](https://huggingface.co/datasets/Lichess/chess-puzzles/tree/461725966ea46898ec6481372a6326de3281bfbd) · [Convention FEN et coups](https://database.lichess.org/#puzzles)
2. **Présélection.** Cote du problème ≥ 2 000, incertitude de la cote ≤ 100 (champ `RatingDeviation`), popularité ≥ 80 et au moins 50 tentatives. La ligne publiée contient au moins cinq demi-coups à partir de la décision, dont une prise, un échec ou une promotion dans les quatre demi-coups suivant le coup calme. L’ordre est fixé par le SHA-256 de `equinox-quiet-v2|` suivi de l’identifiant du problème. Les 512 premiers cas admissibles, à raison d’un par partie et par position, forment le [réservoir versionné](candidates.csv). La clé de position ignore les compteurs FEN et conserve le trait, le roque et la prise en passant légale. Toute position rencontrée dans les 30 parties du run6 ou les 128 parties sources de v1 est écartée.
3. **Difficulté et arbitrage.** Stockfish 18 doit d’abord manquer la solution à 512 nœuds. À profondeur 14 puis 18, il doit retrouver le coup Lichess à chaque recherche et le créditer d’au moins 150 centipions d’avance sur le deuxième. Les deux coups sont évalués à la profondeur demandée avec MultiPV 2, un thread, 16 Mio de table, scores numériques sans indication de borne et état neuf à chaque recherche. Les 16 premiers cas admissibles par couleur constituent le [lot figé](positions.json). Chaque cas garde le coup du contrôle court et les deux analyses profondes.
4. **Mesure.** L’indicateur principal est le nombre de coups de référence trouvés à profondeur 6. Un processus UCI neuf est lancé deux fois par position, sans horloge. Les deux passages doivent rendre le même coup légal, le même score et le même nombre de nœuds à profondeur 6. On conserve les résultats individuels pour comparer les gains et les régressions d’un candidat. L’échantillonnage cible les positions où un coup calme se détache et échappe à la recherche courte de Stockfish.

## Référence de septembre · 23 novembre 2026

Le lot a été publié sur `origin/main` avant sa première mesure. Sa révision dans l’historique actuel est `259ec7e`. Les deux mesures historiques v1 et v2 ont utilisé le même exécutable (sources actuelles à `f6871de`, SHA-256 `46caaa986c5e4ab0610790b012cad658c6578d6f3bf187fe29222111c1b0642b`). Les 32 problèmes ont des cotes Lichess de 2 048 à 2 747, de médiane 2 345. La reconstruction depuis le réservoir a reproduit le fichier des références à l’octet près.

| Trait | Premier coup de référence trouvé | Nœuds sur un passage |
|---|---:|---:|
| Blancs | 8/16 | 1 166 903 |
| Noirs | 5/16 | 2 122 713 |
| **Ensemble** | **13/32 (40,6 %)** | **3 289 616** |

Les deux recherches neuves par position concordent. Sur les 19 écarts de choix, 11 reproduisent le coup du contrôle Stockfish à 512 nœuds. À profondeur 18, Stockfish évalue les 32 positions en faveur du camp au trait. Equinox en évalue trois à −300 cp ou moins, y compris `0afCF`, où il trouve pourtant le coup de référence. [Rapport par position](septembre.json)

Le problème `00Nhl` illustre le type de décision recherché. Equinox prend immédiatement la tour en h7 (`h8h7`). Stockfish classe cette prise deuxième, à −176 cp. Son premier choix, le coup de cavalier `g6f4`, vaut +316 cp et invite la dame blanche à prendre le cavalier en f4, où le pion e5 peut reprendre. [Position sur l’échiquier](https://lichess.org/training/00Nhl)

Lors d’une comparaison, compter séparément les solutions nouvellement trouvées et celles perdues parmi les 13 déjà acquises. Le lot v1 reste la référence pour les régressions, avec 32/32 coups et 15/16 pertes reconnues. La comparaison des deux scores décrit la difficulté des deux sélections pour le même binaire à profondeur 6.

## Rejouer une mesure

Depuis la racine du dépôt, passer le binaire UCI et un chemin de sortie libre. Le [protocole de compilation](../README.md#relire-la-mesure-ou-comparer-un-candidat) précise comment reconstruire le moteur de septembre avec son identifiant Git.

```sh
python3 tools/tactics/measure.py "$essai/equinox" "$essai/v2.json" \
  --fixture tools/tactics/v2/positions.json
```

Pour reconstruire les références depuis le réservoir suivi par Git, utiliser le binaire Stockfish identifié dans le lot et la dépendance Python du [banc tactique](../README.md). La commande `pool` utilise également PyArrow (22.0.0 lors de la construction) pour relire le Parquet amont dont l’URL et l’empreinte figurent dans le lot.

```sh
python3 tools/tactics/quiet.py label "$(command -v stockfish)" \
  tools/tactics/v2/candidates.csv "$essai/positions-v2.json"
cmp tools/tactics/v2/positions.json "$essai/positions-v2.json"
```
