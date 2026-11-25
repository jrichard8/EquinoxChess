# Objectif 2400 · première mesure et défenses calmes

## Résultat et décision

**La base active reste `baseline-2200`.** La vérification d’une défense calme marque 54,10 % en duel direct et améliore le score observé contre SF2400 de 6,38 points sur les mêmes ouvertures. Le duel direct reste compatible avec un gain comme avec une perte. Le palier 2400 reste à atteindre.

| Moteur étudié | Adversaire | Victoires / nulles / défaites | Score | Écart Elo local, IC 95 % par paire |
|---|---|---:|---:|---:|
| Base 2200+ | SF2400 | 131 / 35 / 218 | 38,67 % | −80,1 [−156,6 à −10,6] |
| Candidat | Base 2200+ | 98 / 81 / 77 | 54,10 % | +28,6 [−55,4 à +116,0] |
| Candidat | SF2400 | 159 / 28 / 197 | 45,05 % | −34,5 [−105,6 à +33,8] |

Les 1 024 parties prévues ont été jouées. Face à SF2400, sur 192 ouvertures, le candidat marque davantage que la base sur 71 paires, autant sur 67 et moins sur 54, soit 24,5 points de partie supplémentaires sur 384. Chaque match contre SF2400 contient un forfait au temps de Stockfish. En retirant la paire concernée de chaque match, les scores sont de 38,35 % et 45,03 %. Le duel direct se termine sans incident. Huit parties tournaient en parallèle sur le M4 Pro. Des profils et des analyses hors ligne ont aussi utilisé cette machine. [Duel direct](../../../matches/elo2400/quiet-base.pgn) · [Audit du duel](direct-audit.json) · [Candidat contre SF2400](../../../matches/elo2400/quiet-sf2400.pgn) · [Manifeste](../../../matches/elo2400/quiet-sf2400.pgn.meta.json) · [Audit du candidat](candidate-audit.json)

Le moteur des parties est le commit `04efa19`, compilé avec le tag `quietverify`. La branche `codex/2400-quiet-defense` conserve le candidat, le test de la règle et le traceur de défenses. Les tests Go du candidat et les contrôles des traces courantes et historiques passent. Le duel suivant a testé la fenêtre élargie seule. [Résultats des deux matchs](window/README.md) Dans les défaites à 2400, la revérification supplémentaire coûte 1,69 million de nœuds et fait davantage régresser que sauver de positions aux deux budgets d’arbitrage.

## Protocole fixé avant les parties

La base est le moteur qualifié `baseline-2200`, sources à `a43a8a3` (SHA-256 du binaire historique de match `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`). Le nouveau lot compte 192 ouvertures, graine 240020261124, avec les mêmes critères d’équilibre que les lots précédents. Le générateur exclut les EPD et les positions des PGN archivés. Un deuxième contrôle vérifie aussi les positions tactiques v1 et v2 contre l’archive `023a467`.

1. Figer le lot dans Git. Le jouer intégralement, dans les deux couleurs, contre Stockfish 18 avec UCI_Elo=2400, limitation activée, un thread et 16 Mio de table. Cadence 1 s + 0,1 s par coup. Calculer l’intervalle bilatéral de Hoeffding à 95 % par paire, sans arrêt anticipé.
2. Tester un seul candidat qui vérifie à profondeur normale une défense calme écartée par la réduction des coups tardifs à la racine. Dès D5, utiliser une fenêtre réduite élargie de 80 cp pour classer les coups calmes rejetés. Après les autres coups, revérifier au plus un candidat par passage racine, à condition qu’il reste à moins de 80 cp du meilleur score. Exclure prises, promotions et coups donnant échec.
3. Rejouer à froid les 123 défenses calmes de la qualification 2200 à leur profondeur de partie. Arbitrer les coups modifiés avec SF18 à pleine force et 80 000 nœuds. Compter les passages de ≤ −200 à ≥ −100 cp, les passages inverses et les nœuds. Contrôler aussi v1, v2 et le run6 à D6.
4. Opposer directement le candidat à la base sur les 128 premières ouvertures de `tools/elo/2200/baseline.epd`, puis mesurer sa performance face à SF2400 sur les mêmes 192 ouvertures que la mesure initiale. Ces deux résultats ont des rôles distincts. Garder la base qualifiée en l’absence de gain direct établi. Toute sélection sur ce lot demanderait une confirmation 2400 sur de nouvelles ouvertures avant de qualifier une version.

Les matchs emploient cutechess-cli. Le PGN et le manifeste de chaque match conservent les ouvertures, les couleurs et les empreintes des exécutables.

## Première mesure de la base

Sur les 192 nouvelles ouvertures jouées dans les deux couleurs, la base obtient **131 victoires, 35 nulles, 218 défaites**, soit **38,67 %** face à SF2400. L’écart local est de **−80,1 Elo** (IC 95 % de −156,6 à −10,6). L’unique forfait au temps vient de Stockfish, à la ronde 78. Exclure la paire complète donne 38,35 %. Les profondeurs médianes annoncées sont 8 pour Equinox et 14 pour Stockfish. [Parties](../../../matches/elo2400/base-sf2400.pgn) · [Manifeste](../../../matches/elo2400/base-sf2400.pgn.meta.json) · [Contrôle des ouvertures, couleurs et incidents](base-audit.json)

## Régressions tactiques à profondeur 6

| Lot | Indicateur base → candidat | Nœuds base → candidat |
|---|---|---:|
| v1 | 32 → 32 solutions, 15 → 15 pertes reconnues | 816 043 → 842 517 |
| v2 | 13 → 13 solutions, 3 → 3 fausses alertes de perte | 1 229 318 → 1 505 872 |
| run6 | 10 → 10 erreurs historiques évitées, 2 → 2 pertes reconnues | 919 588 → 1 077 714 |

Sur les 77 positions, le coût monte de **2 964 949 à 3 426 103 nœuds (+15,55 %)**. Le total v2 masque un échange important. Sur `0ImFe`, le candidat trouve `g3e5` (+252 cp contre −201 pour `g5e7`). Sur `0eK7q`, il abandonne `g8h8` (+568 cp) et joue `g8e8` (0 cp). Les deux autres changements sont proches de l’équilibre sur `0Pfb0` ou améliorent une position toujours défavorable sur `Game16_B_ply62`. SF18 à pleine force a arbitré les quatre changements à 80 000 puis 320 000 nœuds par coup forcé. Les scores cités viennent du second budget, du point de vue du joueur au trait. [Comparaison par position](quiet-tactics.json) · [Arbitrages, profondeurs et variantes](tactic-changes.json)

## Vérification du mécanisme sur les défaites à 2200

Sur les 123 défenses calmes historiques rejouées à froid, le choix change sur 20 positions. Le candidat fait passer six positions d’au plus −200 cp à au moins −100 cp et en fait basculer quatre dans l’autre sens. Chaque moteur retrouve huit coups de référence. Le coût passe de 36 730 886 à 42 391 554 nœuds (+15,41 %). Les scores viennent de SF18 à pleine force, à 80 000 nœuds. Le rejeu reconstruit l’historique des coups et vide la table, alors qu’en partie la table et les données d’ordonnancement survivent d’un coup au suivant. [Rejeu des deux exécutables](quiet-2200.json)

Sur les 123 dernières itérations, le candidat déclenche la revérification dans 100 positions et choisit la défense de référence comme coup à revérifier dans 13 positions. La défense de référence reste exclusivement cherchée à profondeur réduite dans 73 positions. Le traceur reproduit le choix et les nœuds du binaire UCI pour chaque position. [Traces par position](trace-2200.json)

Pour isoler l’effet de la fenêtre élargie et celui de la revérification, un contrôle ajouté après le premier rejeu supprime uniquement le passage supplémentaire en fin de racine. Il ne participe à aucun match. Sur ces 123 positions, la fenêtre seule coûte 3 136 341 nœuds supplémentaires. Rétablir la revérification en ajoute encore 2 524 327, change 11 décisions et produit 3 sauvetages pour 3 régressions par rapport à la fenêtre seule. Ces nombres incluent les effets d’une itération sur les suivantes. [Modification sur 04efa19](ranking-only.patch) · [Rejeu sans vérification finale](ranking-2200.json) · [Calcul apparié](compare_replays.py)

Pour reproduire le contrôle, appliquer le patch dans une copie propre de `04efa19` et reconstruire l’UCI avec les mêmes options que `QuietVerify`. Passer à `replay_quiet.py` le rapport de défaites, le binaire de la base, celui à tester, Stockfish et un nouveau fichier de sortie. Puis comparer avec `python3 tools/elo/2400/compare_replays.py tools/elo/2400/quiet-2200.json tools/elo/2400/ranking-2200.json`. La commande `cmd/quiet_trace` de la branche de recherche se compile avec les tags `searchtrace quietverify` et prend le rapport de défaites puis le rejeu complet.

## Défenses dans les défaites contre SF2400

Sur les 218 défaites, SF18 à pleine force confirme dans 120 parties une occasion où le meilleur coup vaut au moins −100 cp, le coup joué au plus −200 cp et l’écart au moins 150 cp. Le filtrage à 5 000 nœuds suivi d’une vérification à 80 000 trouve 90 cas. Le contrôle de chaque décision des 128 autres défaites à 80 000 nœuds en ajoute 30. Parmi les 120 coups de référence, 96 sont calmes, 22 sont des prises et 2 des échecs sans prise. Sur les 98 défaites restantes, 95 franchissent le seuil de −200 cp sans remplir les trois critères de bascule. [Analyse des 218 défaites, variantes et positions](base-losses.json)

Sur les 96 défenses calmes à leur profondeur de partie, le candidat change 12 décisions, sauve 3 positions et en dégrade 1 selon les seuils de rejeu. Les deux moteurs retrouvent 7 coups de référence. Le coût passe de 30 091 189 à 31 141 813 nœuds (+3,49 %). À la dernière profondeur, la vérification se déclenche dans 79 positions et sélectionne le coup de référence dans 19. Sur 48 des 96 positions, la défense de référence reste exclusivement cherchée à profondeur réduite. Dans 20 autres, elle est déjà cherchée à profondeur normale par le mécanisme ordinaire et le coup finalement choisi reste évalué à −200 cp ou moins par SF18. [Rejeux à froid](quiet-2400.json) · [Traces par position](trace-2400.json)

| Comparaison à froid sur les 96 positions | Sauvetages / régressions à 80k | Sauvetages / régressions à 320k | Écart de nœuds Equinox |
|---|---:|---:|---:|
| Fenêtre seule contre base | 6 / 2 | 4 / 2 | −641 485 |
| Candidat complet contre base | 3 / 1 | 2 / 1 | +1 050 624 |
| Ajout de la revérification à la fenêtre | 2 / 4 | 2 / 3 | +1 692 109 |

Le second arbitrage reprend les 17 positions où au moins deux versions jouent des coups différents, avec 320 000 nœuds par coup forcé et le même historique. Il retrouve le même sens du bilan pour chaque comparaison. Les franchissements de seuil ne s’additionnent pas entre lignes. La fenêtre seule est une piste pour un duel à la pendule. [Contrôle sans vérification finale](ranking-2400.json) · [Second arbitrage et variantes](quiet-2400-deeper.json)

## Coût du moteur Go

Sur l’Apple M4 Pro, Go 1.26.2, la recherche de la base à D7 parcourt 432 453 nœuds de milieu de jeu en 97,7 ms et 33 693 nœuds de finale en 7,28 ms. Cela donne respectivement 4,42 et 4,63 millions de nœuds/s. Le benchmark monothread réutilise une table UCI d’un million d’entrées et la vide avant chaque répétition. Les deux mesures sans profileur rapportent 0 octet et 0 allocation par opération. Les matchs tournaient sur la même machine. [Sortie des benchmarks](go-bench.txt)

Dans le profil CPU du milieu de jeu, 66,75 % des échantillons se trouvent en quiétude en incluant ses appels, et 18,58 % dans l’évaluation en incluant ses appels. Ces pourcentages peuvent se recouvrir. La génération des coups glissants et le tri des prises de quiétude prennent respectivement 8,60 % et 7,71 % du temps propre. Cela donne des cibles d’optimisation en Go. Ces deux positions ne mesurent ni tout le processus UCI ni l’écart qu’aurait une implémentation native du même algorithme. [Profil CPU](go-cpu.txt)

```sh
GOMAXPROCS=1 go test ./internal/search -run '^$' -bench '^BenchmarkColdRoot$' -benchmem -benchtime=4s
GOMAXPROCS=1 go test ./internal/search -run '^$' -bench '^BenchmarkColdRoot$/middlegame$' -benchtime=8s -cpuprofile /tmp/equinox-root.pprof -o /tmp/equinox-root.test
go tool pprof -top /tmp/equinox-root.test /tmp/equinox-root.pprof
```

## Rejouer les matchs

Depuis le dépôt, ces commandes reconstruisent les deux exécutables dans des arbres propres et écrivent les parties dans un répertoire temporaire. Les clones récupèrent les branches publiées sur GitHub. Les manifestes conservent les empreintes des binaires de match. Le [contrôle de reconstruction](../2200/loop/README.md#reconstruire-la-base-2200) donne celle de la base dans l’historique actuel. À la pendule, les résultats individuels peuvent varier.

```sh
remote=$(git remote get-url origin)
lab=$(mktemp -d)
git clone --quiet "$remote" "$lab/base"
git -C "$lab/base" checkout --quiet --detach a43a8a3
(cd "$lab/base" && go build -trimpath -buildvcs=true -o "$lab/Base2200" ./cmd/uci)
git clone --quiet "$remote" "$lab/recherche"
git -C "$lab/recherche" checkout --quiet --detach 04efa19
(cd "$lab/recherche" && go build -trimpath -buildvcs=true -tags quietverify -o "$lab/QuietVerify" ./cmd/uci)

jouer_sf2400() {
  go run ./cmd/elo_match -engine-a "$lab/$1" -name-a "$1" \
    -engine-b "$(command -v stockfish)" -name-b SF2400 \
    -option-b UCI_LimitStrength=true -option-b UCI_Elo=2400 \
    -option-b Threads=1 -option-b Hash=16 \
    -openings tools/elo/2400/baseline.epd -openings-count 192 \
    -tc 1+0.1 -concurrency 8 -pgn "$lab/$1-sf2400.pgn"
}
jouer_sf2400 Base2200
go run ./cmd/elo_match -engine-a "$lab/QuietVerify" -name-a QuietVerify \
  -engine-b "$lab/Base2200" -name-b Base2200 -openings tools/elo/2200/baseline.epd \
  -openings-count 128 -tc 1+0.1 -concurrency 8 -pgn "$lab/quiet-base.pgn"
jouer_sf2400 QuietVerify
```
