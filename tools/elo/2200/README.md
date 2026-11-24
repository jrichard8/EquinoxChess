# Référence 1850+ contre Stockfish 2200

Le tag `baseline-1850` pointe sur `536e375`, avec le code de jeu de `e5daafe`. La mesure emploie l’exécutable qualifié contre SF1850, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`. Septembre reste disponible à `819a8d5` pour les comparaisons historiques. [Qualification 1850+](../1850/README.md)

## Protocole fixé avant la première partie

- **192 nouvelles ouvertures**, huit demi-coups depuis la position initiale, deux couleurs par ouverture, 384 parties sans arrêt anticipé. Le générateur de 1850 conserve sa distribution et utilise une nouvelle graine, `220020261124`. Il écarte les positions des anciennes suites EPD et de tous les PGN de matchs du projet à `536e375`, ainsi que du PGN source de v1. La comparaison conserve trait, roques et prise en passant légale et ignore les compteurs. Des positions distinctes peuvent appartenir à la même famille d’ouverture.
- **1 s + 0,1 s par coup**, six parties simultanées sur Apple M4 Pro, Stockfish 18 limité par UCI à 2200, un thread et 16 Mio de table. Le PGN et son manifeste sont produits par `cmd/elo_match`. Lire l’intervalle bilatéral de Hoeffding à 95 % par paire. Le jalon 2200 demande une borne basse du score supérieure à 50 %.
- **Analyse des défaites** avec Stockfish à pleine force, un thread, budgets en nœuds. Pour chaque défaite, rechercher le premier coup d’Equinox qui fait passer une position encore défendable (meilleur coup ≥ −100 cp) à une position nettement inférieure (coup joué ≤ −200 cp), avec une perte d’au moins 150 cp. Vérifier le meilleur coup et le coup joué depuis la même position à budget accru. Conserver les défaites sans bascule confirmée et les résultats par couleur. Chercher ensuite un mécanisme de recherche vérifiable dans les positions confirmées.

Pour régénérer le lot, utiliser le Stockfish du manifeste et les entrées du tag de référence, avec le générateur de ce dossier parent. La liste des PGN ci-dessous est celle du dépôt à `536e375`.

```sh
python3 tools/elo/1850/build_holdout.py /chemin/stockfish /tmp/2200.epd \
  --seed 220020261124 --exclude-epd tools/elo/1850/holdout.epd \
  --exclude-pgn matches/*.pgn matches/elo1850/*.pgn tools/tactics/v1/source.pgn
```

## Méthode de diagnostic

L’arbitre réinitialise son état à chaque recherche. Pour chaque coup d’Equinox dans une défaite, il cherche librement à 5 000 nœuds puis impose le coup joué si celui-ci diffère de sa préférence. Le filtre large (meilleur ≥ −180 cp, joué ≤ −125 cp, écart ≥ 100 cp) déclenche deux recherches à 80 000 nœuds. Seule la première bascule satisfaisant les seuils définitifs est retenue par partie. Les scores sont toujours du point de vue d’Equinox. Le fichier conserve la distance au mat séparément et code les mats par ±32 000 pour les comparaisons de seuils. Le filtre rapide et l’arbitrage à budget fini peuvent manquer une bascule.

Sur ces positions, un rejeu sans horloge conserve l’historique de la position et repart avec une table vide de 1 048 576 entrées. La profondeur D reprend celle du PGN, plafonnée à 8. La sonde compare quatre recherches.

| Recherche | Question |
|---|---|
| Référence à D | Le moteur rejoue-t-il le coup de la partie à froid ? Le binaire qualifié doit retrouver le coup, la profondeur et les nœuds du rejeu. |
| Référence à D+1 | Un demi-coup de recherche supplémentaire suffit-il à défendre ? |
| Profondeur entière à la racine | Que donne la suppression des réductions de coups tardifs uniquement à la racine, aux mêmes profondeurs nominales qu’à D ? |
| Priorité au coup mémorisé | Que donne le fait de chercher d’abord le coup de table de transposition, ou le coup de la variante principale à la racine, que septembre et la base 1850+ placent en tête avant de retrier les coups ? |

Les deux contre-épreuves s’exécutent uniquement dans une copie du moteur. Stockfish réarbitre à 80 000 nœuds les nouveaux coups qu’elles font apparaître. Une récupération fait passer le coup du rejeu de référence de ≤ −200 cp à ≥ −100 cp. Le mouvement inverse compte comme une régression. Ces décisions sont sélectionnées dans des défaites du moteur : leur fréquence renseigne sur les erreurs étudiées. L’adoption d’une modification demanderait un essai en parties contre la base 1850+.

### Reproduction

Le PGN et son manifeste occupent `matches/elo2200/`. Utiliser le même Stockfish que dans le manifeste et un exécutable UCI portant l’empreinte de la base. Les fichiers JSON servent à retrouver chaque décision par numéro de paire et couleur d’Equinox.

```sh
set -eu
diag=$(mktemp -d)
repo="$PWD"
sf=$(command -v stockfish)
uci="$repo/equinoxchess-uci"
cd "$repo"
python3 tools/elo/2200/analyze_losses.py matches/elo2200/baseline-sf2200.pgn "$sf" "$diag/losses.json"
mkdir "$diag/src"
git archive baseline-1850 | tar -x -C "$diag/src"
(
  cd "$diag/src"
  git apply "$repo/tools/elo/2200/replay.patch"
  LOSS_INPUT="$diag/losses.json" LOSS_OUTPUT="$diag/search.json" \
    go test -count=1 -run '^TestLossReplay$' ./internal/search
)
python3 tools/elo/2200/judge_replay.py "$diag/losses.json" "$diag/search.json" \
  "$uci" "$sf" "$diag/grades.json"
```
