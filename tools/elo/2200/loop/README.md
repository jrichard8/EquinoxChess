# Recherche du palier 2200+

Référence de jeu `baseline-1850` (binaire UCI `e5daafe`, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`). Son repère indépendant à 2200 est de 51,43 % sur 384 parties. [Mesure et diagnostic précédents](../README.md)

## Décision

Les matchs de développement utilisent `1+0,1` et chaque ouverture dans les deux couleurs. Premier essai : priorité au coup de table ou à la variante principale à la racine, sur les 128 premières paires de la suite diagnostique. Deuxième piste indépendante : améliorer l’évaluation positionnelle à partir des parties archivées avant cette boucle. Les binaires et les PGN de développement seront conservés pour les essais favorables comme défavorables.

La promotion demande un gain en match direct contre la base avec une borne basse positive (intervalle bilatéral de Hoeffding à 95 % par paire). Les lots tactiques v1, v2 et run6 mesurent séparément ce que le candidat gagne et perd. Pour le palier 2200+, utiliser **384 autres ouvertures** à couleurs inversées, sans arrêt anticipé ni choix de version sur ce lot. À cette cadence, la borne basse du score contre Stockfish 18 réglé à 2200 doit dépasser 50 %. Stockfish utilise un thread et 16 Mio de table.
