# EquinoxChess · point au 24 novembre 2026

## Base active · qualification 2200+

Le 24 novembre, `baseline-2200` franchit le critère fixé contre Stockfish 18 limité à 2200 à 1 s + 0,1 s par coup. **430 victoires, 64 nulles et 274 défaites** sur 384 ouvertures jouées dans les deux couleurs, soit 60,16 % [53,23 % à 67,09 %]. L’écart local est de +71,6 Elo [+22,4 à +123,7]. Les ouvertures ont été figées avant les essais. Aucun incident de jeu. [Rapport, PGN et manifeste](tools/elo/2200/loop/README.md#confirmation-du-palier-2200)

Le moteur sélectionné (`13800c8`) accélère la détection des positions non terminales, adapte l’horloge aux itérations complètes et privilégie le coup de table sous la racine. Le duel contre `baseline-1850` donne 69,01 % sur 384 parties. À D6, les 77 positions tactiques coûtent 19,9 % de nœuds en moins avec les mêmes 32 solutions v1, 13 solutions v2 et 10 fautes historiques évitées sur le run6. Game25 montre la limite du dernier indicateur : le nouveau coup de roi y perd l’avantage. L’essai suivant sur une attaque coordonnée est archivé séparément. [Détails par position et essais écartés](tools/elo/2200/loop/README.md)

Sur les 274 défaites de confirmation, 153 présentent une bascule vérifiée et 123 de ces défenses exigent un coup calme. Sur 12 cas de cette famille, un demi-coup de plus récupère deux défenses pour 4,38 fois plus de nœuds. La prochaine expérience porte sur la vérification sélective d’un coup calme réduit à la racine. [Analyse des défaites](tools/elo/2200/loop/README.md#où-les-274-défaites-basculent)

## Qualification antérieure du moteur 1850+

Le candidat `e5daafe` part de septembre (`819a8d5`), accélère le calcul de mobilité et l’évaluation, puis trie les prises en quiétude. Sur les 64 ouvertures du projet à couleurs inversées, il marque 92,19 % contre septembre. Sur 192 ouvertures supplémentaires figées avant le match, il marque 74,09 % contre Stockfish 18 réglé à 1850, à 1 seconde + 0,1 seconde par coup. L’écart local est estimé à +182,5 Elo, avec un intervalle bilatéral à 95 % de +102,1 à +286,6. Face au réglage 2200, le score est de 56,25 % sur 128 parties, avec un intervalle compatible avec une force inférieure ou supérieure. La base 1850 a ensuite obtenu 51,43 % sur 192 paires indépendantes contre ce réglage. [Protocole, résultats et PGN](tools/elo/1850/README.md)

## Référence de septembre

Le moteur est jouable en UCI et dans le navigateur. Le progrès le mieux établi vient de l’allocation du temps de recherche. Les limites tactiques sont mesurables. À profondeur 6, Equinox évite 10 des 13 erreurs historiques du jeu de régression, mais n’évalue que 2 de ces positions à −300 centipions ou moins. Sur le lot de régression v1, la même révision choisit les 32 coups de référence et reconnaît 15 des 16 positions classées perdantes par Stockfish. Sur le lot v2, qui exige un premier coup calme difficile, elle trouve 13 des 32 solutions.

## Où reprendre le code

| Référence | Ce qu’elle contient |
|---|---|
| `182aae5`, 5 août | Révision du moteur reconstruite pour la calibration contre Stockfish 18 limité à 1800. Marge de 15 ms sur l’allocation avec incrément. |
| `7e56aff`, 31 août | Point de départ des expériences de septembre. Inclut l’analyse hors ligne et le diagnostic ciblé des réponses adverses. |
| **`819a8d5`**, figé le 23 novembre | Référence de septembre pour reprendre le développement. Ce commit réunit le traceur de schéma 11, ses tests, son mode d’emploi et le journal des expériences. Les vérifications ci-dessous portent sur cette version du code. |
| Stash local `786d539` | Ancien essai combinant plusieurs termes de pression autour du roi. Les expériences ultérieures sur le cavalier et la dame sont consignées séparément. |

## Activité et décisions

- **Avril — rendre les résultats interprétables.** Corrections des pats à la frontière de recherche, du suivi du meilleur coup, du repli sur interruption et des scores de mat. Le sixième match de 30 parties contre SF1600 finit à 12 victoires, 0 nulle et 18 défaites. Ses positions servent toujours aux régressions. [PGN du run6](matches/sf1600-run6-1+0.1-30.pgn)
- **Juillet à début août — maîtriser l’horloge.** La recherche temporisée dépasse l’ancien plafond de profondeur 4. Le moteur conserve son état de recherche entre commandes `position` et contrôle le compteur des 50 coups avant de réutiliser une borne de la table de transposition. Deux lots de 64 parties contre l’ancienne allocation donnent 73,05 % au nouveau budget avec incrément. La marge de 15 ms est retenue ensuite, avec zéro forfait dans son A/B de 128 parties. [Protocole des matchs appariés](tools/elo/README.md)
- **Août — mettre les hypothèses d’évaluation à l’épreuve.** Les pénalités de 20 centipions pour la pression du cavalier puis de la dame sur la zone du roi ont obtenu 46,88 % et 48,44 % sur leurs A/B longs. Toutes deux ont été écartées. La pression de la dame reste un indice pour retrouver des positions difficiles dans les PGN. [Analyse des positions](tools/kingzone/README.md)
- **Fin août et septembre — expliquer une erreur de recherche.** Les traces suivent 19 réponses adverses donnant échec sans prise et les 13 positions de régression. Chercher à pleine profondeur certains échecs normalement réduits améliore le bilan des 19 cas de 9 à 11 erreurs évitées, mais réintroduit `d2b4` dans `Game25_ply35`. La règle générale `D5D3D1CheckChain` a été rejetée avant match. La dernière sonde compare des sous-arbres isolés et permet d’observer séparément score, ordre et table de transposition. [Journal détaillé et décisions](docs/archives/journal-avril-septembre-2026.md)

## Calibration antérieure du 5 août

Matchs du 5 août contre Stockfish 18 limité en force, à **1 seconde + 0,1 seconde par coup**. Chaque lot comprend 64 ouvertures jouées avec les deux couleurs. L’intervalle de Hoeffding à 95 % utilise la paire d’ouverture comme unité.

| Adversaire | Victoires – nulles – défaites | Score Equinox | Écart Elo estimé et intervalle | Forfaits au temps |
|---|---:|---:|---:|---|
| SF18 limité à 1600 | 79 – 2 – 47 | 62,50 % | +88,7 [−31,2 ; +235,2] | 0 |
| SF18 limité à 1800 | 59 – 7 – 62 | 48,83 % | −8,1 [−132,1 ; +113,7] | 1 pour Stockfish |

Le moteur a joué près de l’équilibre contre SF18-1800 à cette cadence. L’intervalle reste compatible avec un écart important dans les deux sens. Un classement absolu ou une extrapolation à cadence lente demanderait d’autres mesures. Les chiffres d’août proviennent du journal d’expérience. Le dépôt conserve les six lots historiques d’avril ainsi que les nouveaux matchs appariés de novembre et leurs manifestes. [Entrée du 5 août](docs/archives/journal-avril-septembre-2026.md#update-2026-08-05)

## Vérifications de reprise

Sur macOS arm64, avec Go 1.26.2, les suites Go standard et `searchtrace` passent sans cache de résultat. `go vet ./...`, les tests UCI sous détecteur de courses et les 14 tests Python de l’analyseur passent également. Les six positions anti-pat passent. À profondeur 6, les deux indicateurs tactiques reproduisent exactement 10/13 erreurs évitées et 2/13 pertes reconnues. Le test protège aussi chaque succès antérieur individuellement. [Commandes de validation](docs/developpement.md) · [Test tactique](internal/search/search_anti_blunder_test.go)

Les deux points d’entrée ont aussi été essayés avec un binaire reconstruit. Après 1. e4, le moteur UCI a renvoyé `b8c6` à profondeur 6. Le serveur a servi la page et annoncé 20 coups légaux au départ, puis a accepté `e2e4`, joué `g8f6` à profondeur 4 et rendu le trait aux Blancs.

Le rejeu du PGN run6 retrouve les 30 parties, les 1 530 décisions exploitables et l’empreinte SHA-256 du CSV de septembre (`7fa89590…`). À profondeur 6, la sonde `-probe-d5-reduced-full` reproduit 107 comparaisons sur 23 positions. Sur les 32 cas, la recherche observée rend exactement le même coup, le même score, la même profondeur et le même nombre de nœuds que la recherche sans trace. Sur les 19 réponses primaires, elle évite 9 erreurs et reconnaît 7 pertes. [Rejouer la sonde](cmd/search_trace/README.md)

## Référence tactique indépendante du run6

Le lot v1 a été figé dans `850def5` avant de mesurer `819a8d5`. Ses 32 positions proviennent de 32 nouvelles parties Stockfish et de 32 ouvertures distinctes. Aucune ne se retrouve parmi les 3 108 positions distinctes du run6, même en ignorant les compteurs FEN. À profondeur 6, le moteur retrouve **16/16 coups en position tenable et 16/16 en position perdante**. Il reconnaît **15/16 pertes**, sans fausse alerte sur les 16 positions tenables. Deux passages avec processus neufs donnent exactement les mêmes résultats, pour 1 480 222 nœuds par passage. [Lot, protocole et limites](tools/tactics/README.md) · [Résultats par position](tools/tactics/v1/septembre.json)

## Coups calmes difficiles · lot v2

Le lot v2 a été publié dans `1b97ed9` avant toute mesure du moteur de septembre. Sur 32 problèmes issus de 32 autres parties, sans position commune avec le run6 ni les parties sources de v1, `819a8d5` trouve **13/32 premiers coups** à profondeur 6 (8/16 avec les Blancs, 5/16 avec les Noirs). Deux passages concordent sur chaque position et comptent chacun 3 289 616 nœuds. Le même binaire reproduit à l’octet près son rapport v1. [Sélection et résultats v2](tools/tactics/v2/README.md)

## Analyse critique du 23 novembre et suivi

| Priorité | Constat | Action et critère de décision |
|---|---|---|
| 1 · Comparer les changements aux références figées | La référence de septembre trouve 32/32 coups sur v1 et 13/32 sur les coups calmes de v2. | Rejouer les deux lots à profondeur 6. Sur v2, compter les gains parmi les 19 écarts et les régressions parmi les 13 succès. Sur v1, suivre également les 15/16 pertes reconnues. Comparer le coût en nœuds et garder les canaris du run6 comme contrôle historique. |
| 2 · Conserver les preuves des prochains matchs | Le traceur est versionné. Pour les matchs d’août, le dépôt contient les résultats et les empreintes, mais les PGN et leurs manifestes manquent. | Depuis le 24 novembre, chaque match de qualification a son PGN et son manifeste vérifiés. Le score a été recalculé à partir des PGN archivés. [Parties de novembre](matches/elo1850/) |
| 3 · Promotions sans prise en quiétude, expérience close | Le candidat corrige les finales témoins mais perd une solution sur v2 (13/32 → 12/32) et augmente le coût de 39,29 % sur les 77 positions. | Conserver septembre comme référence. Les mesures et le candidat isolé permettent de reprendre l’analyse. [Bilan par lot et par position](tools/tactics/experiments/quiet-promotions/README.md) |
| 4 · Couvrir les usages réels | Les tests UCI exercent surtout les fonctions de gestion de position et d’horloge. Le serveur HTTP partage un objet de partie mutable entre requêtes et son paquet ne contient aucun test. | Ajouter un scénario UCI complet allant de la position au meilleur coup puis à la partie suivante. Côté HTTP, vérifier la cohérence du plateau lors de requêtes simultanées et sérialiser les accès à la partie partagée. Valider les scénarios sous `-race`. |

La sonde `D5D3D1ScoreOnlyShadow` proposée en septembre reste une expérience d’attribution possible si l’objectif est d’expliquer l’effet de la table de transposition sur cette branche précise.
