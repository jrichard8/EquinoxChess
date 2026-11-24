# Objectif 2400 · première mesure et défenses calmes

## Protocole fixé avant les parties

La base est le moteur qualifié `baseline-2200`, binaire de match du commit `13800c8` (SHA-256 `bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95`). Le nouveau lot compte 192 ouvertures, graine 240020261124, avec les mêmes critères d’équilibre que les lots précédents. Le générateur exclut les EPD et les positions des PGN archivés. Un deuxième contrôle vérifie aussi les positions tactiques v1 et v2 contre l’archive `ddc5924`.

1. Figer le lot dans Git. Le jouer intégralement, dans les deux couleurs, contre Stockfish 18 avec UCI_Elo=2400, limitation activée, un thread et 16 Mio de table. Cadence 1 s + 0,1 s par coup. Calculer l’intervalle bilatéral de Hoeffding à 95 % par paire, sans arrêt anticipé.
2. Tester un seul candidat qui vérifie à profondeur normale une défense calme écartée par la réduction des coups tardifs à la racine. Dès D5, utiliser une fenêtre réduite élargie de 80 cp pour classer les coups calmes rejetés. Après les autres coups, revérifier au plus un candidat par passage racine, à condition qu’il reste à moins de 80 cp du meilleur score. Exclure prises, promotions et coups donnant échec.
3. Rejouer à froid les 123 défenses calmes de la qualification 2200 à leur profondeur de partie. Arbitrer les coups modifiés avec SF18 à pleine force et 80 000 nœuds. Compter les passages de ≤ −200 à ≥ −100 cp, les passages inverses et les nœuds. Contrôler aussi v1, v2 et le run6 à D6.
4. Opposer directement le candidat à la base sur les 128 premières ouvertures de `tools/elo/2200/baseline.epd`, puis mesurer sa performance face à SF2400 sur les mêmes 192 ouvertures que la mesure initiale. Ces deux résultats ont des rôles distincts. Garder la base qualifiée en l’absence de gain direct établi. Toute sélection sur ce lot demanderait une confirmation 2400 sur de nouvelles ouvertures avant de qualifier une version.

Les matchs emploient cutechess-cli. Le PGN et le manifeste de chaque match conservent les ouvertures, les couleurs et les empreintes des exécutables.

## Coût du moteur Go

Sur l’Apple M4 Pro, Go 1.26.2, la recherche de la base à D7 parcourt 432 453 nœuds de milieu de jeu en 97,7 ms et 33 693 nœuds de finale en 7,28 ms. Cela donne respectivement 4,42 et 4,63 millions de nœuds/s. Le benchmark monothread réutilise une table UCI d’un million d’entrées et la vide avant chaque répétition. Les deux mesures sans profileur rapportent 0 octet et 0 allocation par opération. Les matchs tournaient sur la même machine. [Sortie des benchmarks](go-bench.txt)

Dans le profil CPU du milieu de jeu, 66,75 % des échantillons se trouvent en quiétude en incluant ses appels, et 18,58 % dans l’évaluation en incluant ses appels. Ces pourcentages peuvent se recouvrir. La génération des coups glissants et le tri des prises de quiétude prennent respectivement 8,60 % et 7,71 % du temps propre. Cela donne des cibles d’optimisation en Go. Ces deux positions ne mesurent ni tout le processus UCI ni l’écart qu’aurait une implémentation native du même algorithme. [Profil CPU](go-cpu.txt)

```sh
GOMAXPROCS=1 go test ./internal/search -run '^$' -bench '^BenchmarkColdRoot$' -benchmem -benchtime=4s
GOMAXPROCS=1 go test ./internal/search -run '^$' -bench '^BenchmarkColdRoot$/middlegame$' -benchtime=8s -cpuprofile /tmp/equinox-root.pprof -o /tmp/equinox-root.test
go tool pprof -top /tmp/equinox-root.test /tmp/equinox-root.pprof
```

## Rejouer les matchs

Depuis le dépôt, ces commandes reconstruisent les deux exécutables dans des arbres propres et écrivent les parties dans un répertoire temporaire. Git doit connaître la branche de recherche. Les PGN et leurs manifestes permettent de contrôler les empreintes avec la même chaîne de compilation. À la pendule, les résultats individuels peuvent varier.

```sh
git fetch origin main codex/2400-quiet-defense
lab=$(mktemp -d)
git clone --quiet --no-local . "$lab/base"
git -C "$lab/base" checkout --quiet 13800c8
(cd "$lab/base" && go build -trimpath -buildvcs=true -o "$lab/Base2200" ./cmd/uci)
git clone --quiet --no-local . "$lab/recherche"
git -C "$lab/recherche" fetch --quiet "$(git remote get-url origin)" codex/2400-quiet-defense
git -C "$lab/recherche" checkout --quiet ff3ca1c
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
