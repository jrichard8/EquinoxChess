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
