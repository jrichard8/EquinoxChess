# Recherche du palier 2200+

Référence de jeu `baseline-1850` (binaire UCI `e5daafe`, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`). Son repère indépendant à 2200 est de 51,43 % sur 384 parties. [Mesure et diagnostic précédents](../README.md)

## Décision

Les matchs de développement utilisent `1+0,1` et chaque ouverture dans les deux couleurs. Premier essai : priorité au coup de table ou à la variante principale à la racine, sur les 128 premières paires de la suite diagnostique. Deuxième piste indépendante : améliorer l’évaluation positionnelle à partir des parties archivées avant cette boucle. Les binaires et les PGN de développement seront conservés pour les essais favorables comme défavorables.

La promotion demande un gain en match direct contre la base avec une borne basse positive (intervalle bilatéral de Hoeffding à 95 % par paire). Les lots tactiques v1, v2 et run6 mesurent séparément ce que le candidat gagne et perd. Pour le palier 2200+, utiliser **384 autres ouvertures** à couleurs inversées, sans arrêt anticipé ni choix de version sur ce lot. À cette cadence, la borne basse du score contre Stockfish 18 réglé à 2200 doit dépasser 50 %. Stockfish utilise un thread et 16 Mio de table.

## Premiers essais de développement

| Candidat | Résultat contre la base | v1 | v2 | Erreurs run6 évitées | Nœuds sur les 77 positions |
|---|---|---:|---:|---:|---:|
| Base 1850 | référence | 32/32 | 13/32 | 10/13 | 3 701 682 |
| Priorité au coup mémorisé, `85af2f1` | 119 V, 65 N, 72 D sur 128 paires. 59,18 %, écart +64,5 Elo [−19,6 ; +157,1] | 32/32 | 13/32 | 9/13 | 2 845 694 |
| Tables par phase, `89e0f2e` | match séparé | 32/32 | 12/32 | 10/13 | 3 916 379 |
| Combinaison, `b1a844f` | matchs séparés | 32/32 | 10/32 | 10/13 | 3 097 958 |

Avec le coup mémorisé en premier, Game27_W_ply29 rejoue `a2a3`. Les tables par phase réduisent l’erreur absolue moyenne face à l’évaluation statique de SF18 de 238,2 à 168,4 cp sur 4 551 positions de validation, après ajustement sur 28 781 positions d’autres ouvertures issues des PGN 1850. La reconnaissance des pertes v1 passe toutefois de 15/16 à 10/16. La génération des tables utilise Python, python-chess 1.999 et NumPy 2.2.6. Pour reproduire, copier le répertoire `tools/elo/2200/loop/static_eval` du commit `89e0f2e` dans un checkout de `baseline-1850`, y compiler la commande Go, puis passer le binaire ainsi obtenu à `fit_tables.py STOCKFISH BASE_STATIC SORTIE.go` depuis le checkout du candidat. [Ajustement](tapered-fit.json) · [Mesure du premier match](../../../../matches/elo2200/loop/hash-base.pgn.meta.json)

L’optimisation de recherche `3324544` reconnaît une position non terminale lorsqu’un pion peut avancer légalement. Sur les 77 positions à profondeur 6 en UCI, deux passages par position et par moteur ont reproduit coup, score et nœuds. La somme des temps médians par position passe de 1,621 à 1,275 s (−21,4 %). Les passages alternent l’ordre des moteurs, chacun dans un processus neuf, sur la machine utilisée pour les matchs. [Mesures appariées](witness-parity.json)

La combinaison réunit les trois changements. Les analyses par position sont archivées pour [la priorité au coup mémorisé](hash-tactics.json), [les tables par phase](tapered-tactics.json) et [la combinaison](combined-tactics.json).
