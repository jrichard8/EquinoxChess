# Référence 1850+ contre Stockfish 2200

Le tag `baseline-1850` pointe sur `0839ea3`, avec le code de jeu de `91aeb4e`. La mesure emploie l’exécutable qualifié contre SF1850, SHA-256 `1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab`. Septembre reste disponible à `f6871de` pour les comparaisons historiques. [Qualification 1850+](../1850/README.md)

La boucle suivante aboutit à la qualification de `baseline-2200`. [Essais, défaites et match indépendant](loop/README.md#confirmation-du-palier-2200)

## Mesure du 24 novembre 2026

**51,43 % contre Stockfish 2200**, soit 182 victoires, 31 nulles et 171 défaites. L’écart est de **+10,0 Elo**, avec un intervalle bilatéral à 95 % de **−58,7 à +79,4**. La référence joue près de l’équilibre à cette cadence. La borne basse laisse le jalon 2200 ouvert. Le précédent repère sur les 64 ouvertures v2 donnait 56,25 %. Les deux lots sont lus séparément.

| Couleur d’Equinox | Victoires | Nulles | Défaites | Score |
|---|---:|---:|---:|---:|
| Blancs | 94 | 13 | 85 | 52,34 % |
| Noirs | 88 | 18 | 86 | 50,52 % |

Les 192 paires sont complètes. Equinox gagne ses deux parties sur 46 ouvertures et perd les deux sur 46 autres. Un forfait au temps de Stockfish est inclus (paire 99). Retirer toute cette paire donne 51,44 %. Equinox termine sans incident. [PGN](../../../matches/elo2200/baseline-sf2200.pgn) · [Manifeste](../../../matches/elo2200/baseline-sf2200.pgn.meta.json)

Le lot a été figé dans `6632daa`. Son empreinte SHA-256 est `09a333e7b3ed17533558dde6f4a804180d9a1e22c083c717763adac80327435e`. Aucune des 192 positions d’ouverture ne figure parmi les 178 552 positions distinctes des anciens PGN audités, dans les lots tactiques v1/v2 ou dans les suites EPD précédentes. Les positions de départ du PGN ont été comparées à l’EPD pour chaque paire. Cutechess repart au coup 1 et remet les compteurs à zéro. Pour 78 ouvertures, le compteur des 50 coups diffère donc de celui du fichier EPD. Les deux couleurs et tous les diagnostics utilisent les compteurs effectivement joués.

## Où les 171 défaites basculent

L’arbitrage trouve **92 défaites** avec un coup qui fait passer d’une position défendable à une position nettement inférieure. Le filtre rapide en retrouve 74. L’examen des 97 autres défaites, coup par coup à 80 000 nœuds, en ajoute 18. Ces 92 cas comprennent 51 défaites avec les Blancs et 41 avec les Noirs. Le demi-coup médian est le 36ᵉ après la sortie d’ouverture, avec 24 pièces encore présentes en médiane.

- Dans **76 des 92** positions, le meilleur coup de l’arbitre est un coup calme sans prise ni échec. Dans 14 cas il prend, dans 2 il donne échec sans prise. Le coup joué par Equinox était lui-même calme dans 57 cas. Cela motive l’étude de l’ordre et de la profondeur des coups calmes. Cette population sélectionnée ne mesure pas le taux d’erreur de tous les coups calmes.
- Dans **71 autres défaites**, on observe un franchissement du seuil −200 cp de moindre ampleur, sans bascule stricte. À ce premier passage, l’alternative vaut déjà moins de −100 cp dans 66 cas. En médiane, l’arbitre donne −152 cp au meilleur coup et −239 cp au coup joué.
- Pour les **8 dernières**, l’arbitre évalue déjà la position à −200 cp ou moins lors d’au moins une décision, sans identifier un coup d’Equinox franchissant le seuil dans les conditions fixées. L’attribution à un coup unique reste ouverte. [Analyse par défaite, positions et variantes](losses.json)

## Hypothèse retenue après cette mesure

**Essayer en match la priorité effective au coup mémorisé.** À la racine, la variante principale de l’itération précédente prime lorsqu’elle est disponible. Aux autres nœuds, le coup de la table de transposition prime lorsqu’il est disponible. Le changement consiste à essayer ce coup avant d’appliquer le tri habituel aux autres coups. Aujourd’hui le moteur le déplace en première case puis `selectBestMove` peut immédiatement le déloger. Dans 7 des 8 cas où la table de la racine contenait la défense de référence au dernier passage observé, cette défense a été essayée après un autre coup. [Ordre actuel dans Negamax](../../../internal/search/search.go)

| Rejeu sur les 92 positions | Défenses récupérées | Défenses perdues | Nœuds | Écart |
|---|---:|---:|---:|---:|
| Base 1850+, profondeur D | — | — | 13 922 455 | — |
| Un demi-coup supplémentaire partout | 20 | 6 | 55 339 496 | +297,5 % |
| Profondeur entière pour les coups tardifs à la racine | 4 | 3 | 33 189 423 | +138,4 % |
| **Priorité effective au coup mémorisé** | **4** | **2** | **11 512 571** | **−17,3 %** |

Les colonnes mesurent les passages entre les seuils −200 et −100 cp, par rapport au coup de la base rejouée à froid. Le recul comme le gain est compté. Le moteur de référence rejoue le coup de la partie à sa profondeur enregistrée dans **68 cas sur 92**. Dans ces 68 cas, la priorité au coup mémorisé récupère 4 défenses et D+1 en récupère 19. Sur les 92 cas, 59 défenses avaient reçu seulement une recherche réduite à la racine. Leur approfondissement systématique coûte cher pour un faible gain net. Prioriser le coup mémorisé économise des nœuds dans **76 cas sur 92**. C’est le motif principal du choix. Les relectures à froid et la sélection de défaites ne permettent pas d’en déduire un gain Elo. [Résultats des recherches](search.json) · [Arbitrage des coups obtenus](grades.json)

**Deux exemples vérifiés à 320 000 nœuds**, scores du point de vue des Noirs, numérotation du PGN. Dans la paire **112**, Equinox joue `11...Bxd4`. Après `12.Re1+ Kd7 13.Nxd4`, le roi noir a perdu son roque et le score du coup est −562 cp. Le petit roque vaut −8 cp. En rejouant à la profondeur 7 de la partie, la priorité au coup mémorisé retrouve le roque. Dans la paire **45**, le rejeu normal à froid trouve `18...d5` (+110 cp), alors que la priorité au coup mémorisé revient au coup perdant de la partie, `18...e5` (−237 cp). L’effet de l’ordre se propage aussi dans les réponses et les fenêtres de recherche. [Variantes de contrôle](examples.json)

Cet essai a ensuite été joué contre `baseline-1850` à `1+0,1`, avec ouvertures appariées, intervalle par paire et contrôle de v1, v2 et run6. Les 192 ouvertures ci-dessus ont servi au développement. La confirmation du candidat retenu a utilisé 384 autres ouvertures figées. [Boucle de recherche et résultat](loop/README.md#confirmation-du-palier-2200)

## Protocole fixé avant la première partie

- **192 nouvelles ouvertures**, huit demi-coups depuis la position initiale, deux couleurs par ouverture, 384 parties sans arrêt anticipé. Le générateur de 1850 conserve sa distribution et utilise une nouvelle graine, `220020261124`. Il écarte les positions des anciennes suites EPD et de tous les PGN de matchs du projet à `0839ea3`, ainsi que du PGN source de v1. La comparaison conserve trait, roques et prise en passant légale et ignore les compteurs. Des positions distinctes peuvent appartenir à la même famille d’ouverture.
- **1 s + 0,1 s par coup**, six parties simultanées sur Apple M4 Pro, Stockfish 18 limité par UCI à 2200, un thread et 16 Mio de table. Le PGN et son manifeste sont produits par `cmd/elo_match`. Lire l’intervalle bilatéral de Hoeffding à 95 % par paire. Le jalon 2200 demande une borne basse du score supérieure à 50 %.
- **Analyse des défaites** avec Stockfish à pleine force, un thread, budgets en nœuds. Pour chaque défaite, rechercher le premier coup d’Equinox qui fait passer une position encore défendable (meilleur coup ≥ −100 cp) à une position nettement inférieure (coup joué ≤ −200 cp), avec une perte d’au moins 150 cp. Vérifier le meilleur coup et le coup joué depuis la même position à budget accru. Conserver les défaites sans bascule confirmée et les résultats par couleur. Chercher ensuite un mécanisme de recherche vérifiable dans les positions confirmées.

Pour régénérer le lot, utiliser le Stockfish du manifeste et les entrées du tag de référence, avec le générateur de ce dossier parent. La liste des PGN ci-dessous est celle du dépôt à `0839ea3`.

```sh
python3 tools/elo/1850/build_holdout.py /chemin/stockfish /tmp/2200.epd \
  --seed 220020261124 --exclude-epd tools/elo/1850/holdout.epd \
  --exclude-pgn matches/*.pgn matches/elo1850/*.pgn tools/tactics/v1/source.pgn
```

## Méthode de diagnostic

L’arbitre réinitialise son état à chaque recherche et conserve la dernière variante avec score sans indication de borne. Pour chaque coup d’Equinox dans une défaite, il cherche librement à 5 000 nœuds puis impose le coup joué si celui-ci diffère de sa préférence. Le filtre large (meilleur ≥ −180 cp, joué ≤ −125 cp, écart ≥ 100 cp) déclenche deux recherches à 80 000 nœuds. Seule la première bascule satisfaisant les seuils définitifs est retenue par partie. Les scores sont toujours du point de vue d’Equinox. Le fichier conserve la distance au mat séparément et code les mats par ±32 000 pour les comparaisons de seuils. Pour chaque défaite sans bascule repérée par le filtre, une seconde passe examine directement chaque décision à 80 000 nœuds, jusqu’à une bascule stricte ou la fin. Elle relève aussi le premier passage du meilleur coup au-dessus de −200 cp à un coup joué à −200 cp ou moins. L’arbitrage reste une estimation à budget fini.

Sur ces positions, un rejeu sans horloge conserve l’historique de la position et repart avec une table vide de 1 048 576 entrées. La profondeur D reprend celle du PGN, plafonnée à 8. La sonde compare quatre recherches.

| Recherche | Question |
|---|---|
| Référence à D | Le moteur rejoue-t-il le coup de la partie à froid ? Le binaire qualifié doit retrouver le coup, la profondeur et les nœuds du rejeu. |
| Référence à D+1 | Un demi-coup de recherche supplémentaire suffit-il à défendre ? |
| Profondeur entière à la racine | Que donne la suppression des réductions de coups tardifs uniquement à la racine, aux mêmes profondeurs nominales qu’à D ? |
| Priorité au coup mémorisé | Que donne le fait de chercher d’abord le coup de table de transposition, ou le coup de la variante principale à la racine, que septembre et la base 1850+ placent en tête avant de retrier les coups ? |

Les deux contre-épreuves s’exécutent uniquement dans une copie du moteur. Stockfish réarbitre à 80 000 nœuds les nouveaux coups qu’elles font apparaître. Une récupération fait passer le coup du rejeu de référence de ≤ −200 cp à ≥ −100 cp. Le mouvement inverse compte comme une régression. Ces décisions sont sélectionnées dans des défaites du moteur : leur fréquence renseigne sur les erreurs étudiées. L’adoption d’une modification demanderait un essai en parties contre la base 1850+.

### Reproduction

Le PGN et son manifeste occupent `matches/elo2200/`. Utiliser le Stockfish du manifeste. L’arbitrage accepte la base 1850 d’origine et sa reconstruction vérifiée avec Go 1.26.2 sur macOS arm64. Les fichiers JSON servent à retrouver chaque décision par numéro de paire et couleur d’Equinox.

```sh
set -eu
diag=$(mktemp -d)
repo="$PWD"
sf=$(command -v stockfish)
uci="$diag/Base1850"
git clone --quiet "$(git remote get-url origin)" "$diag/moteur"
git -C "$diag/moteur" checkout --quiet --detach 91aeb4e
(cd "$diag/moteur" && go build -trimpath -buildvcs=true -o "$uci" ./cmd/uci)
python3 tools/elo/2200/analyze_losses.py matches/elo2200/baseline-sf2200.pgn "$sf" "$diag/screen.json"
python3 tools/elo/2200/analyze_losses.py matches/elo2200/baseline-sf2200.pgn "$sf" "$diag/losses.json" \
  --exhaustive-misses "$diag/screen.json"
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
