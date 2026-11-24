# Objectif 2400 · première mesure et défenses calmes

## Protocole fixé avant les parties

La base est le moteur qualifié `baseline-2200`, binaire de match du commit `13800c8` (SHA-256 `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`). Le nouveau lot compte 192 ouvertures, graine 240020261124, avec les mêmes critères d’équilibre que les lots précédents. Le générateur exclut les EPD et les positions des PGN archivés. Un deuxième contrôle vérifie aussi les positions tactiques v1 et v2 contre l’archive `ddc5924`.

1. Figer le lot dans Git. Le jouer intégralement, dans les deux couleurs, contre Stockfish 18 avec UCI_Elo=2400, limitation activée, un thread et 16 Mio de table. Cadence 1 s + 0,1 s par coup. Calculer l’intervalle bilatéral de Hoeffding à 95 % par paire, sans arrêt anticipé.
2. Tester un seul candidat qui vérifie à profondeur normale une défense calme écartée par la réduction des coups tardifs à la racine. Dès D5, utiliser une fenêtre réduite élargie de 80 cp pour classer les coups calmes rejetés. Après les autres coups, revérifier au plus un candidat par passage racine, à condition qu’il reste à moins de 80 cp du meilleur score. Exclure prises, promotions et coups donnant échec.
3. Rejouer à froid les 123 défenses calmes de la qualification 2200 à leur profondeur de partie. Arbitrer les coups modifiés avec SF18 à pleine force et 80 000 nœuds. Compter les passages de ≤ −200 à ≥ −100 cp, les passages inverses et les nœuds. Contrôler aussi v1, v2 et le run6 à D6.
4. Opposer directement le candidat à la base sur les 128 premières ouvertures de `tools/elo/2200/baseline.epd`, puis mesurer sa performance face à SF2400 sur les mêmes 192 ouvertures que la mesure initiale. Ces deux résultats ont des rôles distincts. Garder la base qualifiée en l’absence de gain direct établi. Toute sélection sur ce lot demanderait une confirmation 2400 sur de nouvelles ouvertures avant de qualifier une version.

Les matchs emploient cutechess-cli. Le PGN et le manifeste de chaque match conservent les ouvertures, les couleurs et les empreintes des exécutables.
