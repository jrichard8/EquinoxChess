# EquinoxChess · bilan de reprise du 23 novembre 2026

Le moteur est jouable en UCI et dans le navigateur. Le progrès le mieux établi vient de l’allocation du temps de recherche. Les limites tactiques sont mesurables. À profondeur 6, Equinox évite 10 des 13 erreurs historiques du jeu de régression, mais n’évalue que 2 de ces positions à −300 centipions ou moins. La prochaine itération gagnerait à mesurer une amélioration sur des parties indépendantes du lot qui a servi à mettre ces erreurs au jour.

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

## Dernière calibration de force documentée

Matchs du 5 août contre Stockfish 18 limité en force, à **1 seconde + 0,1 seconde par coup**. Chaque lot comprend 64 ouvertures jouées avec les deux couleurs. L’intervalle de Hoeffding à 95 % utilise la paire d’ouverture comme unité.

| Adversaire | Victoires – nulles – défaites | Score Equinox | Écart Elo estimé et intervalle | Forfaits au temps |
|---|---:|---:|---:|---|
| SF18 limité à 1600 | 79 – 2 – 47 | 62,50 % | +88,7 [−31,2 ; +235,2] | 0 |
| SF18 limité à 1800 | 59 – 7 – 62 | 48,83 % | −8,1 [−132,1 ; +113,7] | 1 pour Stockfish |

Le moteur a joué près de l’équilibre contre SF18-1800 à cette cadence. L’intervalle reste compatible avec un écart important dans les deux sens. Un classement absolu ou une extrapolation à cadence lente demanderait d’autres mesures. Les PGN suivis par Git sont les six lots historiques d’avril. Les chiffres d’août proviennent du journal d’expérience, sans nouveau match en novembre. [Entrée du 5 août](docs/archives/journal-avril-septembre-2026.md#update-2026-08-05)

## Vérifications de reprise

Sur macOS arm64, avec Go 1.26.2, les suites Go standard et `searchtrace` passent sans cache de résultat. `go vet ./...`, les tests UCI sous détecteur de courses et les 14 tests Python de l’analyseur passent également. Les six positions anti-pat passent. À profondeur 6, les deux indicateurs tactiques reproduisent exactement 10/13 erreurs évitées et 2/13 pertes reconnues. Le test protège aussi chaque succès antérieur individuellement. [Commandes de validation](docs/developpement.md) · [Test tactique](internal/search/search_anti_blunder_test.go)

Les deux points d’entrée ont aussi été essayés avec un binaire reconstruit. Après 1. e4, le moteur UCI a renvoyé `b8c6` à profondeur 6. Le serveur a servi la page et annoncé 20 coups légaux au départ, puis a accepté `e2e4`, joué `g8f6` à profondeur 4 et rendu le trait aux Blancs.

Le rejeu du PGN run6 retrouve les 30 parties, les 1 530 décisions exploitables et l’empreinte SHA-256 du CSV de septembre (`7fa89590…`). À profondeur 6, la sonde `-probe-d5-reduced-full` reproduit 107 comparaisons sur 23 positions. Sur les 32 cas, la recherche observée rend exactement le même coup, le même score, la même profondeur et le même nombre de nœuds que la recherche sans trace. Sur les 19 réponses primaires, elle évite 9 erreurs et reconnaît 7 pertes. [Rejouer la sonde](cmd/search_trace/README.md)

## Analyse critique et ordre de reprise proposé

| Priorité | Constat | Action et critère de décision |
|---|---|---|
| 1 · Élargir les positions de validation | Les 19 cas primaires et les 13 canaris viennent du même lot de 30 parties. Les variantes successives ont été ajustées avec ces positions sous les yeux. Éviter le coup historique peut aussi conduire à un autre mauvais coup. | Constituer un lot de parties issu d’autres ouvertures, le séparer par partie avant tout réglage, puis figer coups de référence et critères de réussite. Rapporter séparément choix du coup, reconnaissance de la perte et régressions individuelles sur les deux lots. |
| 2 · Conserver les preuves des prochains matchs | Le traceur est versionné. Pour les matchs d’août, le dépôt contient les résultats et les empreintes, mais les PGN et leurs manifestes manquent. | Pour le prochain match, conserver ensemble la révision, les empreintes des binaires, le PGN, le manifeste et le résumé. Vérifier que le score peut être recalculé à partir de ces fichiers. |
| 3 · Tester une seule correction tactique | Hors échec, la quiétude filtre les coups sans prise avant de tester la promotion. Une promotion sans prise peut donc disparaître à la frontière. C’est une hypothèse plus circonscrite qu’une nouvelle série de règles D5. | Commencer par une finale où la promotion change le résultat à l’horizon et un témoin avec promotion par prise. Si l’effet est reproduit, isoler le changement, mesurer choix, score et nœuds, puis appliquer le protocole de match apparié. [Filtre de quiétude](internal/search/search.go) |
| 4 · Couvrir les usages réels | Les tests UCI exercent surtout les fonctions de gestion de position et d’horloge. Le serveur HTTP partage un objet de partie mutable entre requêtes et son paquet ne contient aucun test. | Ajouter un scénario UCI complet allant de la position au meilleur coup puis à la partie suivante. Côté HTTP, vérifier la cohérence du plateau lors de requêtes simultanées et sérialiser les accès à la partie partagée. Valider les scénarios sous `-race`. |

La sonde `D5D3D1ScoreOnlyShadow` proposée en septembre reste une expérience d’attribution possible si l’objectif est d’expliquer l’effet de la table de transposition sur cette branche précise. Pour progresser en force de jeu, je donnerais d’abord la priorité à un lot indépendant et à une correction dont l’effet peut y être observé.
