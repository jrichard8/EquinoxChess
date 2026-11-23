# Promotions sans prise à la frontière de recherche · 23 novembre 2026

**Le candidat est écarté en l’état.** La recherche de quiétude voit désormais une promotion et une sous-promotion gagnantes sur les positions témoins. Sur les trois lots figés, elle perd une solution de v2, n’en gagne aucune et augmente le nombre de nœuds de 39,29 % au total. Le moteur de septembre reste la référence de jeu.

## Changement isolé et vérification directe

Le candidat `81fdfe3` admet en quiétude, hors échec, les quatre promotions sans prise. La génération de coups, les évasions, le stand-pat et la règle existante qui exempte les promotions de l’élagage delta sont communs aux deux moteurs. Le candidat se trouve sur la branche `codex/qsearch-quiet-promotions`.

| Position témoin, recherche de quiétude seule | Score septembre → candidat (cp au trait) | Nœuds |
|---|---:|---:|
| Promotion blanche en dame, `a7a8q` | +220 → +881 | 1 → 5 |
| Position miroir pour les Noirs, `a2a1q` | +220 → +881 | 1 → 5 |
| Sous-promotion `c7c8r`, la dame ferait pat | +250 → +526 | 1 → 6 |
| Témoin avec prise, `a7b8q` | +613 → +613 | 5 → 5 |

## Comparaison à profondeur 6

| Lot | Indicateurs septembre → candidat | Gains / régressions de choix | Nœuds septembre | Nœuds candidat | Écart |
|---|---|---:|---:|---:|---:|
| v1, 32 positions | Coups de référence 32 → 32. Pertes reconnues 15/16 → 15/16, fausses alertes 0 → 0. | 0 / 0 | 1 480 222 | 1 574 941 | +94 719 (+6,40 %) |
| v2, 32 coups calmes | Coups de référence 13 → 12. Blancs 8/16 → 7/16, Noirs 5/16 → 5/16. Trois fausses alertes dans les deux versions, sur les mêmes positions. | 0 / 1 | 3 289 616 | 3 400 898 | +111 282 (+3,38 %) |
| run6, 13 canaris | Coups fautifs historiques évités 10 → 10. Pertes reconnues 2 → 2. | 0 / 0 | 7 772 789 | 12 495 002 | +4 722 213 (+60,75 %) |
| **Total, 77 positions** | | **0 / 1** | **12 542 627** | **17 470 841** | **+4 928 214 (+39,29 %)** |

La régression de v2 se situe sur **`0JLoe`**. Septembre trouvait `h5h6` (+38 cp). Le candidat choisit `f5f6` (+26 cp). À profondeur 18, Stockfish donnait +522 cp à `h5h6` et +60 cp au deuxième coup, `e5d4`. Sur **`0Np2W`**, l’autre changement de choix, les deux versions manquent `g2g5`. Septembre joue `g2g3`, le candidat `f3f2`. Les 75 autres premiers coups sont identiques.

Le surcoût est concentré. Sur **`Game17_W_ply17`**, les deux moteurs jouent l’erreur historique `d1f3` et annoncent +67 cp. Le candidat explore 7 041 277 nœuds contre 2 788 051, soit **90,1 % du surcoût du run6** à lui seul. Dans v1, `G18-p26` ajoute 81 075 nœuds sur les 94 719 supplémentaires du lot. Les nombres de nœuds mesurent le travail à profondeur fixe. Aucun match Elo n’a été joué pour cette expérience.

Le diagnostic des deux positions retrace les promotions noires responsables du surcoût sur Game17 et la coupure à profondeur réduite sur 0JLoe. [Traces, contre-épreuve et reproduction](diagnostic.md)

## Provenance et reproduction

- Les révisions UCI `819a8d5` et `81fdfe3` ont été compilées depuis des arbres propres sous Go 1.26.2, macOS arm64. Les empreintes des exécutables, des entrées et les 77 paires de résultats sont conservées dans [paired.json](paired.json). Les champs `cp` et `mate` sont au point de vue du camp au trait.
- Pour v1 et v2, utiliser [measure.py](../../measure.py) avec chaque binaire et les deux fichiers `positions.json` figés. Deux processus UCI neufs par position doivent rendre le même coup, le même score, la même profondeur et le même nombre de nœuds. Les deux rapports de septembre ont reproduit à l’octet près les rapports versionnés. La table UCI compte 1 048 576 entrées.
- Pour run6, reconstruire le [CSV tactique de septembre](../../../kingzone/README.md) depuis le PGN (SHA-256 du CSV complet `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`). Garder les 13 lignes dont `canary_name` est renseigné. Utiliser chaque [traceur](../../../../cmd/search_trace/README.md) compilé avec `-tags searchtrace`, en mode observationnel, avec `-depths 6 -tt-size 65536 -expected-primary 0`. Chaque résultat tracé a été confronté à une recherche fraîche sans trace. Les nœuds du tableau comptent un seul passage principal par position. La reconnaissance de perte inclut les scores ≤ −300 cp et les mats subis.
- La suite Go standard et `go vet -tags searchtrace ./...` passent sur le candidat. Dans la suite avec `-tags searchtrace`, trois tests historiques échouent sur leurs comptes de nœuds figés (`TestSearchTraceBridgeLocatesNestedLMRTTCutoff`, `TestSearchTraceD5BranchAblationsStayOnExactPath` et `TestSearchTraceD5ReducedFullShadowIsDeterministicAndNonInterfering`). Ces trois tests repassent sur septembre. Leurs valeurs attendues ont été conservées.
