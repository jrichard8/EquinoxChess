# Recherche du palier 2200+

Référence au départ de la boucle : `baseline-1850` (binaire UCI `91aeb4e`, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`). Son repère indépendant à 2200 est de 51,43 % sur 384 parties. [Mesure et diagnostic précédents](../README.md)

## Confirmation du palier 2200+

**60,16 % contre Stockfish 18 limité à 2200** sur les 384 ouvertures figées de `confirm-01.epd` à couleurs inversées (768 parties). Le score est de **430 V, 64 N, 274 D**. L’intervalle bilatéral de Hoeffding à 95 % par paire est de **53,23 % à 67,09 %**, soit un écart de **+71,6 Elo [+22,4 à +123,7]** par rapport au réglage adverse. Le critère annoncé avant les parties est franchi. Cette calibration concerne la cadence 1 s + 0,1 s, le matériel utilisé et cette distribution d’ouvertures. Le tag de qualification est `baseline-2200`.

| Couleur du candidat | Victoires | Nulles | Défaites | Score |
|---|---:|---:|---:|---:|
| Blancs | 218 | 34 | 132 | 61,20 % |
| Noirs | 212 | 30 | 142 | 59,11 % |

Toutes les paires et tous les départs concordent avec l’EPD. Le candidat gagne les deux parties sur 116 ouvertures et perd les deux sur 45. Les deux moteurs terminent sans forfait, coup illégal, déconnexion ni plantage. Le binaire est identique à celui sélectionné face à la base 1850. L’essai utilise un Apple M4 Pro, cutechess-cli 1.4.0 et huit parties simultanées pour la confirmation. Au plus deux autres parties de développement tournaient en parallèle. Stockfish utilise un thread et 16 Mio de table. [768 parties](../../../../matches/elo2200/loop/interior-sf2200-confirm.pgn) · [Manifeste avec commandes et empreintes](../../../../matches/elo2200/loop/interior-sf2200-confirm.pgn.meta.json) · [Audit des paires](confirmation-audit.json)

Les tests Go passent pour le moteur normal, le traceur de la version actuelle et le rejeu des mesures historiques avec les tags `searchtrace trace1850`. Le contrôle de concurrence UCI et `go vet` passent. Les deux constructions normales `a43a8a3` et `8b75863` (intégration des tests historiques) reproduisent coup, score et nœuds sur les 77 positions en UCI.

## Reconstruire la base 2200+

Le tag `baseline-2200` pointe sur le rapport de qualification (`18ce418`). Le moteur **InteriorTT** qui a joué les 768 parties se reconstruit à la révision `a43a8a3`. Avec Go 1.26.2 sur macOS arm64 (`CGO_ENABLED=1`, `GOARM64=v8.0`), lancer :

```sh
lab=$(mktemp -d)
git clone --quiet https://github.com/jrichard8/MarsGopherChess.git "$lab/source"
git -C "$lab/source" checkout --quiet --detach a43a8a34313c4c1fe717b7ccfbc5083f10ee44ed
(cd "$lab/source" && go build -trimpath -buildvcs=true -o "$lab/Base2200" ./cmd/uci)
go version -m "$lab/Base2200"
echo "ff4749fe461ab76b7ff383a6180e34d48bb2e15b00c24673b923ef876446648c  $lab/Base2200" | shasum -a 256 -c -
```

Go inscrit la révision dans le binaire. Deux clones indépendants ont produit cette empreinte avec `vcs.modified=false`. Le manifeste des parties conserve l’empreinte du binaire utilisé avant la réécriture, `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`. À D6, la reconstruction et ce binaire ont donné les mêmes coups, scores et nœuds sur les 77 positions de v1, v2 et run6, avec deux recherches à froid par position et par exécutable. Sur les 192 ouvertures de la campagne 2400 à D8, le rejeu retrouve chaque premier coup et le nombre de nœuds de chaque position, soit 100 004 748 nœuds pour la base. [Relevé de référence](../../2400/window/opening-nodes.json)

Les anciens identifiants des manifestes et des annotations de tags se retrouvent dans la [correspondance des révisions Git](../../../../docs/revisions-git.tsv). [Script de comparaison UCI](compare_fixed_depth.py)

## Où les 274 défaites basculent

Après la clôture du match, SF18 à pleine force a rejugé les décisions perdantes. Une bascule stricte exige une alternative à ≥ −100 cp, un coup joué à ≤ −200 cp et un écart d’au moins 150 cp, du point de vue du joueur au trait. Le filtre à 5 000 nœuds en trouve 111 après vérification à 80 000 nœuds. Le second passage, directement à 80 000 nœuds sur les défaites encore inexpliquées, en ajoute 42. Sur les 121 autres défaites, 109 présentent un franchissement de −200 cp de moindre ampleur et 12 atteignent une position inférieure sans franchissement attribué selon ce critère. [274 défaites, positions et variantes](confirmation-losses.json) · [Analyseur](../analyze_losses.py)

Parmi les **153 bascules strictes** (75 avec les Blancs, 78 avec les Noirs), **123 défenses de référence sont des coups calmes**, contre 19 prises, 10 échecs sans prise et une promotion. Le coup joué est calme dans 105 cas. La bascule médiane arrive 40 demi-coups après la sortie d’ouverture, avec 22 pièces et une profondeur de jeu de 7. À la paire 3 avec les Noirs, par exemple, le roi en d7 vaut −263 cp alors que la tour en d8 maintient −58 cp. Ces fréquences décrivent les défaites classées.

Sur 12 de ces 123 cas, choisis à intervalles réguliers dans l’ordre des paires, le rejeu à froid reproduit le coup de la partie 9 fois sur 12. Passer de D à D+1 récupère deux positions au sens des seuils et en fait reculer une de −233 à −355 cp. Une défense exacte de l’arbitre est trouvée à D+1, contre zéro à D. Le travail monte de 3 651 347 à 16 004 032 nœuds (×4,38). **Prochaine expérience retenue** : comparer le meilleur coup calme écarté à la racine à profondeur pleine quand la première recherche l’a réduit. Mesurer les défenses récupérées, les défenses perdues et le surcoût avant un nouveau duel. Le lot de confirmation 01 sert désormais au diagnostic. Toute nouvelle qualification demandera d’autres ouvertures réservées. [12 rejeux](confirmation-quiet-depth.json) · [Sélection et arbitrage](check_quiet_depth.py)

## Décision

Les matchs de développement utilisent `1+0,1` et chaque ouverture dans les deux couleurs. Premier essai : priorité au coup de table ou à la variante principale à la racine, sur les 128 premières paires de la suite diagnostique. Deuxième piste indépendante : améliorer l’évaluation positionnelle à partir des parties archivées avant cette boucle. Les révisions, empreintes des binaires et PGN de développement seront conservés pour les essais favorables comme défavorables.

La promotion demande un gain en match direct contre la base avec une borne basse positive (intervalle bilatéral de Hoeffding à 95 % par paire). Les lots tactiques v1, v2 et run6 mesurent séparément ce que le candidat gagne et perd. Pour le palier 2200+, utiliser **384 autres ouvertures** à couleurs inversées, sans arrêt anticipé ni choix de version sur ce lot. À cette cadence, la borne basse du score contre Stockfish 18 réglé à 2200 doit dépasser 50 %. Stockfish utilise un thread et 16 Mio de table.

## Premiers essais de développement

| Candidat | Résultat contre la base | v1 | v2 | Erreurs run6 évitées | Nœuds sur les 77 positions |
|---|---|---:|---:|---:|---:|
| Base 1850 | référence | 32/32 | 13/32 | 10/13 | 3 701 682 |
| Priorité au coup mémorisé, `df9de88` | 119 V, 65 N, 72 D sur 128 paires. 59,18 %, écart +64,5 Elo [−19,6 ; +157,1] | 32/32 | 13/32 | 9/13 | 2 845 694 |
| Tables par phase, `794fcc7` | match séparé | 32/32 | 12/32 | 10/13 | 3 916 379 |
| Combinaison, `abfb9fe` | matchs séparés | 32/32 | 10/32 | 10/13 | 3 097 958 |

Avec le coup mémorisé en premier, Game27_W_ply29 rejoue `a2a3`. Les tables par phase réduisent l’erreur absolue moyenne face à l’évaluation statique de SF18 de 238,2 à 168,4 cp sur 4 551 positions de validation, après ajustement sur 28 781 positions d’autres ouvertures issues des PGN 1850. La reconnaissance des pertes v1 passe toutefois de 15/16 à 10/16. L’ajustement modifie aussi les valeurs matérielles implicites. En pleine phase d’ouverture, la contribution individuelle d’un pion en a2 passe de +105 à −1 cp et celle du pion h2 de +105 à −15 cp, avant les termes de structure et de mobilité. Cette dérive fournit une piste à contrôler pour tout nouvel ajustement. La génération des tables utilise Python, python-chess 1.999 et NumPy 2.2.6. Pour reproduire, copier le répertoire `tools/elo/2200/loop/static_eval` du commit `794fcc7` dans un checkout de `baseline-1850`, y compiler la commande Go, puis passer le binaire ainsi obtenu à `fit_tables.py STOCKFISH BASE_STATIC SORTIE.go` depuis le checkout du candidat. [Ajustement](tapered-fit.json) · [Mesure du premier match](../../../../matches/elo2200/loop/hash-base.pgn.meta.json)

L’optimisation de recherche `0f5cf3a` reconnaît une position non terminale lorsqu’un pion peut avancer légalement. Sur les 77 positions à profondeur 6 en UCI, deux passages par position et par moteur ont reproduit coup, score et nœuds. La somme des temps médians par position passe de 1,621 à 1,275 s (−21,4 %). Les passages alternent l’ordre des moteurs, chacun dans un processus neuf, sur la machine utilisée pour les matchs. [Mesures appariées](witness-parity.json)

La combinaison réunit les trois changements. Les analyses par position sont archivées pour [la priorité au coup mémorisé](hash-tactics.json), [les tables par phase](tapered-tactics.json) et [la combinaison](combined-tactics.json).

## Choix du candidat de sélection

Les tables apprises seules terminent à **38,28 %** (84 V, 28 N, 144 D), soit −83,0 Elo [−179,2 ; +2,0] contre la base. Aucun incident de jeu. Les scores statiques se rapprochent de ceux de Stockfish sur les positions de validation, mais le match conduit à écarter cet ajustement global. [Parties](../../../../matches/elo2200/loop/tapered-base.pgn) · [Manifeste](../../../../matches/elo2200/loop/tapered-base.pgn.meta.json)

Deux essais isolent la régression de l’ordre des coups sur Game27_W_ply29. Avec la même évaluation historique et les mêmes optimisations de calcul, donner la priorité à la variante principale à la racine suffit à réintroduire `a2a3` ; donner la priorité au coup de table uniquement sous la racine évite de nouveau cette erreur avec `c4b3` (la base jouait `e1e2`). Les deux versions ont le même budget de temps en partie. La comparaison tactique utilise une profondeur fixe.

| Changement | v1 | v2 | run6 évités | Pertes run6 reconnues | Nœuds v1 + v2 + run6 |
|---|---:|---:|---:|---:|---:|
| [Vérification du coup nul du côté au trait](verified-tactics.json), `2a9ff6f` | 32/32 | 13/32 | 9/13 | 2/13 | 2 782 306 |
| [Recherche à fenêtre nulle sur les coups suivants](scout-tactics.json), `c56ad71` | 32/32 | 13/32 | 9/13 | 3/13 | 2 897 200 |
| [Position du roi selon la phase](king-tactics.json), `36e2a0a` | 32/32 | 12/32 | 10/13 | 1/13 | 3 677 481 |
| [Priorité à la racine seule](RootPV-tactics.json), `6298632` | 32/32 | 13/32 | 9/13 | 2/13 | 3 426 891 |
| [Priorité au coup de table sous la racine](InteriorTT-tactics.json), `a43a8a3` | 32/32 | 13/32 | 10/13 | 2/13 | 2 964 949 |

Le candidat **InteriorTT** (`a43a8a3`) garde les 15 alertes correctes de v1, les 3 fausses alertes de v2 et les deux pertes identifiées du run6. Il réunit la preuve rapide d’un coup de pion légal, le calcul direct des occupations, la gestion de l’horloge et la priorité au coup mémorisé sous la racine. Avec une horloge UCI, il peut finir une itération au-delà du budget nominal, dans la limite du double. Deux itérations stables abaissent le budget de démarrage de la suivante. Le délai explicite `movetime` conserve son plafond. Les matchs opposent des configurations complètes ; leurs résultats ne quantifient pas séparément le gain de chaque changement.

Le match direct d’InteriorTT porte sur les 192 paires de `baseline.epd`, sans arrêt anticipé. S’il passe le critère de sélection, le même exécutable joue les 384 paires de `confirm-01.epd` contre Stockfish 18 réglé à 2200. Le match de développement de la combinaison avec les tables utilise uniquement les 64 ouvertures v2 historiques. Le contrôle des couleurs, des empreintes, de la légalité et de l’ordre des ouvertures se rejoue avec `python3 tools/elo/2200/loop/audit_match.py MATCH.pgn LOT.epd`.

La combinaison avec les tables termine à 51,17 % contre SF2200 sur les ouvertures v2 (59 V, 13 N, 56 D, écart +8,1 Elo [−113,7 ; +132,1]). Les 64 paires sont complètes, sans incident. [PGN](../../../../matches/elo2200/loop/combined-sf-dev.pgn) · [Manifeste](../../../../matches/elo2200/loop/combined-sf-dev.pgn.meta.json)

Les 384 départs de confirmation sont tous distincts. Aucun ne figure parmi les 219 695 positions des parties, EPD et lots tactiques du dépôt au commit `314d236`, après normalisation du trait, des droits de roque et de la prise en passant et retrait des compteurs FEN. Le générateur utilise la graine 220020261125. [Audit du lot figé](confirm-01-audit.json) · [Vérification reproductible](audit_openings.py)

## Ce que change InteriorTT à profondeur fixe

Sur les 77 positions, le travail descend de 3 701 682 à **2 964 949 nœuds** (−19,9 %). Les 32 coups de v1 et les 13 solutions trouvées sur v2 sont identiques à ceux de la base. Deux coups changent parmi les échecs de v2. Sur le run6, les dix mêmes erreurs sont évitées et les deux mêmes pertes reconnues. `0JLoe` conserve `h5h6` avec 3 885 nœuds contre 4 466. Game17_W_ply17 reste un échec tactique (`d1f3`) avec un surcoût de 11 701 nœuds. [Détail par position](InteriorTT-tactics.json)

Sur les **92 bascules de défaites** déjà identifiées contre SF2200, le rejeu de l’historique de chaque position à la profondeur de la partie donne 4 défenses récupérées et 3 perdues par rapport à la base rejouée à froid. Le seuil de sauvetage est le passage de ≤ −200 à ≥ −100 cp chez l’arbitre SF18 (80 000 nœuds). La régression est le passage inverse. Le travail descend de 13 922 455 à 12 227 917 nœuds (−12,2 %). Dans la paire 112 avec les Noirs, InteriorTT roque au lieu de prendre en d4 et passe de −548 à +1 cp. Dans la paire 57 avec les Blancs, il pousse en c5 au lieu de prendre en c4 et passe de −79 à −429 cp. Cette sélection porte uniquement sur d’anciennes défaites. [92 décisions](interior-defenses.json) · [Rejeu et arbitrage](replay_defenses.py)

Les quatre changements de coup masqués par les totaux v2 et run6 ont été rejugés avec SF18 à pleine force. **Game25_W_ply49 est une régression** malgré l’évitement de l’ancienne faute `d4d5`. À un million de nœuds par recherche, la tour en g2 de la base vaut +153 cp. Le roi en f3 du candidat vaut −208 cp après la réponse …Te8, suivie dans la variante de …Df6+ et …De6. À 80 000 et 320 000 nœuds, l’arbitre préfère déjà la tour. Sur Game27, `c4b3` vaut −38 cp contre −71 cp pour `e1e2` à un million de nœuds. Les deux changements de v2 continuent à manquer le premier coup de référence. [Coups, scores et variantes aux trois budgets](interior-changed-moves.json)

Le cache d’évaluation statique (`d79da5d`) a aussi été essayé sur la version avec détection rapide du pat. Sur les 77 positions à froid, les deux passages retrouvent exactement coup, score et nœuds, pour une somme de temps médians de 1,566 s avec cache et 1,598 s sans cache. Le gain chronométrique est faible sur cette mesure et son effet en partie reste ouvert. Le candidat de sélection utilise l’évaluation sans ce cache. [Mesures](cache-check.json) · [Protocole des passages alternés](compare_fixed_depth.py)

Sur Game25, le moteur UCI à froid choisit Rf3 à D5 et D6, puis Tg2 à D7, D8 et D9. La base jouait déjà Tg2 à D5 et D6. Si l’on impose Rf3, le candidat trouve …Te8 dès D5 mais évalue encore la position à +442 cp pour les Blancs. Même à D9, …Te8 reste son choix et son évaluation vaut +349 cp. Le retour au bon coup à D7 précède donc une évaluation correcte de la ligne perdante. Ce diagnostic motive le test de pression coordonnée décrit plus bas. [Rejeu UCI, profondeurs 4 à 9](game25-depths.json)

## Sélection en parties

InteriorTT marque **69,01 %** contre la base 1850 sur les 192 ouvertures de développement (202 V, 126 N, 56 D). L’écart estimé est de **+139,1 Elo**, avec un intervalle bilatéral à 95 % de **+64,7 à +228,2**. Avec les Blancs, bilan de 102 V, 65 N, 25 D. Avec les Noirs, 100 V, 61 N, 31 D. La profondeur médiane déclarée passe de 8 pour la base à 9 pour le candidat. Aucun incident de jeu. Le contrôle de chaque départ et de chaque inversion de couleurs passe. Le binaire sélectionné pour ces parties portait le SHA-256 `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`. [PGN de sélection](../../../../matches/elo2200/loop/interior-base.pgn) · [Manifeste](../../../../matches/elo2200/loop/interior-base.pgn.meta.json)

Le contrôle contre Clock isole l’ordre intérieur des coups. Sur 64 paires, InteriorTT obtient 47 V, 46 N et 35 D (54,69 %, écart +32,7 Elo [−87,2 ; +161,2]). Ce contrôle laisse ouverte l’ampleur de l’apport propre du tri. Le gain de +139,1 face à la base porte sur la configuration complète. [Duel isolé](../../../../matches/elo2200/loop/interior-clock.pgn) · [Manifeste](../../../../matches/elo2200/loop/interior-clock.pgn.meta.json)

## Essai suivant sur Game25

La variante `54cccef` (KingAttack) part d’InteriorTT. Elle valorise les attaques où une dame et une deuxième pièce contrôlent des cases voisines du roi. Le bonus croît avec le nombre et le poids des contacts. L’hypothèse et les mesures tactiques utilisent exclusivement les anciennes positions du projet. InteriorTT reste l’unique candidat testé sur le lot de confirmation 01.

À D6, Game25 passe de Rf3 à De1. SF18 évalue De1 à +100, +129 puis +132 cp à 80 000, 320 000 et un million de nœuds. À ce dernier budget, Rf3 vaut −208 cp et Tg2 +153 cp. Les mêmes 32 et 13 solutions de v1 et de v2 sont retrouvées. Une fausse alerte de perte disparaît sur v2. Le run6 garde 10 évitements et 2 pertes reconnues. Le coût total sur les 77 positions passe de 2 964 949 à 3 073 432 nœuds (+3,7 %). L’effet en partie est mesuré par un duel de développement fixé à 64 ouvertures v2, couleurs inversées. [Relevé tactique](attack-tactics.json) · [Arbitrage de De1](attack-game25.json)

Sur les 92 anciennes bascules, KingAttack récupère 5 défenses et en perd 3 par rapport à InteriorTT, avec 12 177 692 nœuds contre 12 227 917. Par rapport à la base 1850, le bilan sur ce même lot est de 6 récupérations et 1 régression. [Décisions et coûts](attack-defenses.json)

Le duel KingAttack contre la base qualifiée se termine à **48,83 %** sur 128 parties (37 V, 51 N, 40 D), soit −8,1 Elo [−132,1 à +113,7]. Aucune terminaison anormale. Malgré le sauvetage de Game25, ce duel laisse le gain en partie indémontré. La variante reste sur `codex/2200-coordinated-attack`. **Décision de fin de boucle** : conserver InteriorTT comme base 2200+. Les sept matchs de cette boucle totalisent 2 048 parties. [Duel KingAttack](../../../../matches/elo2200/loop/attack-interior.pgn) · [Manifeste](../../../../matches/elo2200/loop/attack-interior.pgn.meta.json)
