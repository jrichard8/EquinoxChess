# Fenêtre élargie seule à la racine

## Protocole du 24 novembre 2026

Tester le binaire du rejeu à froid précédent, construit au commit `ff3ca1c` avec [le patch figé](../ranking-only.patch) et le tag `quietverify`. À partir de D5, la recherche réduite des coups calmes tardifs à la racine élargit sa fenêtre de 80 centipions. La boucle se termine après le dernier coup légal. La recherche normale à pleine profondeur d’un coup qui dépasse alpha reste celle de la base.

Les deux duels utilisent **les mêmes 192 ouvertures**, dans le même ordre et les deux couleurs, soit 384 parties chacun. Le premier oppose RankingOnly à la base qualifiée 2200+, le second à Stockfish 18 réglé à 2400 avec limitation activée, un thread et 16 Mio de table. Cadence 1 s + 0,1 s par coup et huit parties parallèles. Jouer les deux matchs intégralement. Le lot est celui des deux mesures SF2400 précédentes. Le duel direct antérieur du candidat avec revérification portait sur un autre lot.

Utiliser le score de match et l’intervalle bilatéral de Hoeffding à 95 % par paire. Le critère de sélection en duel direct est une borne basse supérieure à 50 %. Le score SF2400 sur ce lot de développement permet de comparer les trois moteurs. Une qualification 2400 demanderait de nouvelles ouvertures.

Compter les nœuds à profondeur fixée sur v1, v2, les canaris du run6 et les deux corpus de défenses calmes. Rapporter les solutions gagnées et perdues ainsi que les incidents de match. [Ouvertures](../baseline.epd) · [Expérience précédente](../README.md)
