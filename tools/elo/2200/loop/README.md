# Recherche du palier 2200+

Référence de jeu `baseline-1850` (binaire UCI `e5daafe`, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`). Son repère indépendant à 2200 est de 51,43 % sur 384 parties. [Mesure et diagnostic précédents](../README.md)

## Décision

Les matchs de développement utilisent `1+0,1` et chaque ouverture dans les deux couleurs. Premier essai : priorité au coup de table ou à la variante principale à la racine, sur les 128 premières paires de la suite diagnostique. Deuxième piste indépendante : améliorer l’évaluation positionnelle à partir des parties archivées avant cette boucle. Les révisions, empreintes des binaires et PGN de développement seront conservés pour les essais favorables comme défavorables.

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

## Choix du candidat de sélection

Les tables apprises seules terminent à **38,28 %** (84 V, 28 N, 144 D), soit −83,0 Elo [−179,2 ; +2,0] contre la base. Aucun incident de jeu. Les scores statiques se rapprochent de ceux de Stockfish sur les positions de validation, mais le match conduit à écarter cet ajustement global. [Parties](../../../../matches/elo2200/loop/tapered-base.pgn) · [Manifeste](../../../../matches/elo2200/loop/tapered-base.pgn.meta.json)

Deux essais isolent la régression de l’ordre des coups sur Game27_W_ply29. Avec la même évaluation historique et les mêmes optimisations de calcul, donner la priorité à la variante principale à la racine suffit à réintroduire `a2a3` ; donner la priorité au coup de table uniquement sous la racine évite de nouveau cette erreur avec `c4b3` (la base jouait `e1e2`). Les deux versions ont le même budget de temps en partie. La comparaison tactique utilise une profondeur fixe.

| Changement | v1 | v2 | run6 évités | Pertes run6 reconnues | Nœuds v1 + v2 + run6 |
|---|---:|---:|---:|---:|---:|
| [Vérification du coup nul du côté au trait](verified-tactics.json), `2f72aa3` | 32/32 | 13/32 | 9/13 | 2/13 | 2 782 306 |
| [Recherche à fenêtre nulle sur les coups suivants](scout-tactics.json), `d595fcb` | 32/32 | 13/32 | 9/13 | 3/13 | 2 897 200 |
| [Position du roi selon la phase](king-tactics.json), `d231770` | 32/32 | 12/32 | 10/13 | 1/13 | 3 677 481 |
| [Priorité à la racine seule](RootPV-tactics.json), `7ecbdf9` | 32/32 | 13/32 | 9/13 | 2/13 | 3 426 891 |
| [Priorité au coup de table sous la racine](InteriorTT-tactics.json), `13800c8` | 32/32 | 13/32 | 10/13 | 2/13 | 2 964 949 |

Le candidat **InteriorTT** (`13800c8`) garde les 15 alertes correctes de v1, les 3 fausses alertes de v2 et les deux pertes identifiées du run6. Il réunit la preuve rapide d’un coup de pion légal, le calcul direct des occupations, la gestion de l’horloge et la priorité au coup mémorisé sous la racine. Avec une horloge UCI, il peut finir une itération au-delà du budget nominal, dans la limite du double. Deux itérations stables abaissent le budget de démarrage de la suivante. Le délai explicite `movetime` conserve son plafond. Les matchs opposent des configurations complètes ; leurs résultats ne quantifient pas séparément le gain de chaque changement.

Le match direct d’InteriorTT porte sur les 192 paires de `baseline.epd`, sans arrêt anticipé. S’il passe le critère de sélection, le même exécutable joue les 384 paires de `confirm-01.epd` contre Stockfish 18 réglé à 2200. Le match de développement de la combinaison avec les tables utilise uniquement les 64 ouvertures v2 historiques. Le contrôle des couleurs, des empreintes, de la légalité et de l’ordre des ouvertures se rejoue avec `python3 tools/elo/2200/loop/audit_match.py MATCH.pgn LOT.epd`.

La combinaison avec les tables termine à 51,17 % contre SF2200 sur les ouvertures v2 (59 V, 13 N, 56 D, écart +8,1 Elo [−113,7 ; +132,1]). Les 64 paires sont complètes, sans incident. [PGN](../../../../matches/elo2200/loop/combined-sf-dev.pgn) · [Manifeste](../../../../matches/elo2200/loop/combined-sf-dev.pgn.meta.json)

Les 384 départs de confirmation sont tous distincts. Aucun ne figure parmi les 219 695 positions des parties, EPD et lots tactiques du dépôt au commit `c5f6da9`, après normalisation du trait, des droits de roque et de la prise en passant et retrait des compteurs FEN. Le générateur utilise la graine 220020261125. [Audit du lot figé](confirm-01-audit.json) · [Vérification reproductible](audit_openings.py)

## Ce que change InteriorTT à profondeur fixe

Sur les 77 positions, le travail descend de 3 701 682 à **2 964 949 nœuds** (−19,9 %). Les 32 coups de v1 et les 13 solutions trouvées sur v2 sont identiques à ceux de la base. Deux coups changent parmi les échecs de v2. Sur le run6, les dix mêmes erreurs sont évitées et les deux mêmes pertes reconnues. `0JLoe` conserve `h5h6` avec 3 885 nœuds contre 4 466. Game17_W_ply17 reste un échec tactique (`d1f3`) avec un surcoût de 11 701 nœuds. [Détail par position](InteriorTT-tactics.json)

Sur les **92 bascules de défaites** déjà identifiées contre SF2200, le rejeu de l’historique de chaque position à la profondeur de la partie donne 4 défenses récupérées et 3 perdues par rapport à la base rejouée à froid. Le seuil de sauvetage est le passage de ≤ −200 à ≥ −100 cp chez l’arbitre SF18 (80 000 nœuds). La régression est le passage inverse. Le travail descend de 13 922 455 à 12 227 917 nœuds (−12,2 %). Dans la paire 112 avec les Noirs, InteriorTT roque au lieu de prendre en d4 et passe de −548 à +1 cp. Dans la paire 57 avec les Blancs, il pousse en c5 au lieu de prendre en c4 et passe de −79 à −429 cp. Cette sélection porte uniquement sur d’anciennes défaites. [92 décisions](interior-defenses.json) · [Rejeu et arbitrage](replay_defenses.py)

Les quatre changements de coup masqués par les totaux v2 et run6 ont été rejugés avec SF18 à pleine force. **Game25_W_ply49 est une régression** malgré l’évitement de l’ancienne faute `d4d5`. À un million de nœuds par recherche, la tour en g2 de la base vaut +153 cp. Le roi en f3 du candidat vaut −208 cp après la réponse …Te8, suivie dans la variante de …Df6+ et …De6. À 80 000 et 320 000 nœuds, l’arbitre préfère déjà la tour. Sur Game27, `c4b3` vaut −38 cp contre −71 cp pour `e1e2` à un million de nœuds. Les deux changements de v2 continuent à manquer le premier coup de référence. [Coups, scores et variantes aux trois budgets](interior-changed-moves.json)

Le cache d’évaluation statique (`83b457e`) a aussi été essayé sur la version avec détection rapide du pat. Sur les 77 positions à froid, les deux passages retrouvent exactement coup, score et nœuds, pour une somme de temps médians de 1,566 s avec cache et 1,598 s sans cache. Le gain chronométrique est faible sur cette mesure et son effet en partie reste ouvert. Le candidat de sélection utilise l’évaluation sans ce cache. [Mesures](cache-check.json) · [Protocole des passages alternés](compare_fixed_depth.py)

Sur Game25, le moteur UCI à froid choisit Rf3 à D5 et D6, puis Tg2 à D7, D8 et D9. La base jouait déjà Tg2 à D5 et D6. Si l’on impose Rf3, le candidat trouve …Te8 dès D5 mais évalue encore la position à +442 cp pour les Blancs. Même à D9, …Te8 reste son choix et son évaluation vaut +349 cp. Le retour au bon coup à D7 précède donc une évaluation correcte de la ligne perdante. **Prochaine piste** : examiner les échecs de dame contre le roi exposé après …Te8 et mesurer le coût d’une meilleure reconnaissance de l’attaque. [Rejeu UCI, profondeurs 4 à 9](game25-depths.json)

## Sélection en parties

InteriorTT marque **69,01 %** contre la base 1850 sur les 192 ouvertures de développement (202 V, 126 N, 56 D). L’écart estimé est de **+139,1 Elo**, avec un intervalle bilatéral à 95 % de **+64,7 à +228,2**. Avec les Blancs, bilan de 102 V, 65 N, 25 D. Avec les Noirs, 100 V, 61 N, 31 D. La profondeur médiane déclarée passe de 8 pour la base à 9 pour le candidat. Aucun incident de jeu. Le contrôle de chaque départ et de chaque inversion de couleurs passe. Le binaire sélectionné est celui de `13800c8`, SHA-256 `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`. [PGN de sélection](../../../../matches/elo2200/loop/interior-base.pgn) · [Manifeste](../../../../matches/elo2200/loop/interior-base.pgn.meta.json)
