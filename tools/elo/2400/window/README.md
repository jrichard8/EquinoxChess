# Fenêtre élargie seule à la racine

## Bilan des 768 parties

**MarsGopherChess conserve la base 2200+.** La fenêtre élargie seule manque le critère de sélection en duel direct. Sur les mêmes ouvertures face à SF2400, elle fait mieux que la base et moins bien que la variante avec revérification. Les deux nouveaux matchs se terminent sans incident ni forfait.

| Duel | Victoires / nulles / défaites de la fenêtre seule | Score | Écart Elo à 1+0,1 s, IC 95 % |
|---|---:|---:|---:|
| Contre la base 2200+ | 105 / 145 / 134 | 46,22 % | −26,3 [−96,8 à +42,1] |
| Contre Stockfish 18 réglé à 2400 | 152 / 26 / 206 | 42,97 % | −49,2 [−121,7 à +19,3] |

Contre SF2400, le score historique de la base est de 38,67 % et celui de la variante avec revérification de 45,05 %. La fenêtre seule gagne **4,30 points de pourcentage** face au résultat de la base (77 paires meilleures, 58 égales et 57 moins bonnes) et cède **2,08 points** face à la variante avec revérification (70 meilleures, 51 égales, 71 moins bonnes). Le premier écart vaut +4,58 points si l’on retire aussi la paire touchée par le forfait de Stockfish lors du match historique de la base. La profondeur médiane de la fenêtre seule est de 9 dans les deux matchs. [PGN du duel direct](../../../../matches/elo2400/window-base.pgn) · [PGN contre SF2400](../../../../matches/elo2400/window-sf2400.pgn) · [Audit direct](base-audit.json) · [Audit SF2400](sf2400-audit.json) · [Paires comparées à la base](paired-base.json) · [Paires comparées à la revérification](paired-quiet.json)

Le gain observé sur les défenses calmes se généralise mal au coût de recherche des ouvertures. La fenêtre consomme 2,13 % de nœuds en moins sur les 96 défenses SF2400 et **30,40 % de plus** en repartant des 192 ouvertures à D8. La trace du run6 confirme 345 recherches réduites dans la fenêtre élargie sur 11 des 13 dernières itérations et aucun passage de revérification ajouté.

## Protocole du 24 novembre 2026

Tester le binaire du rejeu à froid précédent, construit au commit `04efa19` avec [le patch figé](../ranking-only.patch) et le tag `quietverify`. À partir de D5, la recherche réduite des coups calmes tardifs à la racine élargit sa fenêtre de 80 centipions. La boucle se termine après le dernier coup légal. La recherche normale à pleine profondeur d’un coup qui dépasse alpha reste celle de la base.

Les deux duels utilisent **les mêmes 192 ouvertures**, dans le même ordre et les deux couleurs, soit 384 parties chacun. Le premier oppose RankingOnly à la base qualifiée 2200+, le second à Stockfish 18 réglé à 2400 avec limitation activée, un thread et 16 Mio de table. Cadence 1 s + 0,1 s par coup et huit parties parallèles. Jouer les deux matchs intégralement. Le lot est celui des deux mesures SF2400 précédentes. Le duel direct antérieur du candidat avec revérification portait sur un autre lot.

Utiliser le score de match et l’intervalle bilatéral de Hoeffding à 95 % par paire. Le critère de sélection en duel direct est une borne basse supérieure à 50 %. Le score SF2400 sur ce lot de développement permet de comparer les trois moteurs. Une qualification 2400 demanderait de nouvelles ouvertures.

Compter les nœuds à profondeur fixée sur v1, v2, les canaris du run6 et les deux corpus de défenses calmes. Rapporter les solutions gagnées et perdues ainsi que les incidents de match. [Ouvertures](../baseline.epd) · [Expérience précédente](../README.md)

## Coût à profondeur fixée

| Positions | Profondeur | Nœuds base | Nœuds fenêtre seule | Variation | Qualité observée |
|---|---:|---:|---:|---:|---|
| 192 ouvertures des matchs | 8 | 100 004 748 | 130 408 283 | +30,40 % | 36 premiers coups changés |
| v1, 32 positions | 6 | 816 043 | 834 306 | +2,24 % | 32 → 32 solutions, 15 → 15 pertes reconnues |
| v2, 32 positions | 6 | 1 229 318 | 1 402 340 | +14,07 % | 13 → 12 solutions, 3 → 3 fausses alertes de perte |
| run6, 13 canaris | 6 | 919 588 | 1 066 327 | +15,96 % | 10 → 10 erreurs historiques évitées, 2 → 2 pertes reconnues |
| 123 défenses du corpus SF2200 | profondeur de partie | 36 730 886 | 39 867 227 | +8,54 % | 7 sauvetages, 5 régressions |
| 96 défenses du corpus SF2400 | profondeur de partie | 30 091 189 | 29 449 704 | −2,13 % | 6 sauvetages, 2 régressions |

Sur les 77 positions v1, v2 et run6 réunies, la fenêtre coûte **338 024 nœuds supplémentaires (+11,40 %)**. Sur les ouvertures, D8 correspond à la profondeur médiane de la base dans sa mesure initiale contre SF2400. Chaque recherche commence avec une table vide et la FEN initiale du PGN. Le nombre de nœuds augmente sur 158 ouvertures, baisse sur 32 et reste identique sur 2. Pour les 36 changements de premier coup, SF18 à pleine force donne un avantage d’au moins 50 cp au candidat dans **2 cas** et à la base dans **3 cas** aux deux budgets de 80 000 et 320 000 nœuds par coup forcé. Les 31 autres changements ne franchissent pas ce seuil dans le même sens aux deux budgets. [Nœuds et coups par ouverture](opening-nodes.json) · [Arbitrages](opening-grades.json)

Les deux dernières lignes reprennent les rejeux archivés, dont les empreintes des deux exécutables correspondent à celles des matchs. Une position est sauvée lorsque le coup de la base vaut au plus −200 cp et celui du candidat au moins −100 cp. La régression est le franchissement inverse. Les chiffres du tableau emploient SF18 à 80 000 nœuds. Sur les 96 défenses SF2400, le contrôle à 320 000 nœuds donne 4 sauvetages et 2 régressions. [Rejeu SF2200](../ranking-2200.json) · [Rejeu SF2400](../ranking-2400.json) · [Contrôle à 320 000 nœuds](../quiet-2400-deeper.json)

## Gains et régressions tactiques

Cinq coups changent sur les 77 positions. À 320 000 nœuds par coup forcé, sur `Game25_W_ply49`, `g1g2` remplace `e2f3` et passe de −242 à +143 cp. Sur `Game30_B_ply26`, `h7h5` remplace `c6e5` et descend de −171 à −439 cp. Les compteurs historiques du run6 ne distinguent pas ces changements. La perte d’une solution v2 vient de `0eK7q`, où `g8g7` vaut −75 cp contre +568 pour `g8h8`. Sur `Game16_B_ply62`, l’évaluation passe de −289 à −216 cp et sur `0Pfb0`, de +1 à −35. Les deux budgets d’arbitrage donnent le même sens aux cinq écarts. [Résultats des trois suites](tactics.json) · [Coups, scores et variantes](tactic-changes.json)

## Reproduction

Le manifeste de match enregistre le SHA-256 historique de `RankingOnly`, `1323dd9cc0c1c8eafa0b311f53e3b69314acd91004db3648a8c5f0a7e6ff93e7`. Avec Go 1.26.2 sur macOS arm64, la reconstruction ci-dessous donne `f2a2806e4901c28513755ade3580624ac00c3f89f0c2cd6d7bd008141adf58b3`. Le marquage Go `vcs.modified=true` correspond au [patch de la fenêtre seule](../ranking-only.patch).

Depuis la racine du dépôt, exécuter d’abord les [commandes de reconstruction de Base2200](../../2200/loop/README.md#reconstruire-la-base-2200), puis dans le même shell :

```sh
git -C "$lab/source" worktree add --quiet --detach "$lab/ranking" 04efa19
git -C "$lab/ranking" apply "$PWD/tools/elo/2400/ranking-only.patch"
(cd "$lab/ranking" && go build -trimpath -buildvcs=true -tags quietverify -o "$lab/RankingOnly" ./cmd/uci)
python3 tools/elo/2400/window/measure_openings.py "$lab/Base2200" "$lab/RankingOnly" "$lab/opening-nodes.json"
python3 tools/elo/2400/window/grade_openings.py "$lab/opening-nodes.json" "$(command -v stockfish)" "$lab/opening-grades.json"
```
