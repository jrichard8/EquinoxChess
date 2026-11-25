# Calibration contre Stockfish 18 limité à 1850

Le seuil de travail est un score supérieur à 50 % contre Stockfish 18 réglé à 1850, à 1 seconde + 0,1 seconde par coup, avec une borne basse de l’intervalle bilatéral à 95 % supérieure à 50 %. Le calcul utilise une observation par paire d’ouverture, selon le [protocole des matchs](../README.md). Cette échelle dépend de l’adversaire, de sa configuration, du matériel et de la cadence.

## Résultats du 24 novembre 2026

**Le candidat complet franchit le jalon 1850 sur les 192 ouvertures de confirmation.** Il marque 74,09 % des points, avec un écart estimé de +182,5 Elo par rapport à Stockfish 18 limité à 1850. Sur le repère à 2200, son score de 56,25 % laisse encore une marge d’incertitude de part et d’autre de l’équilibre.

| Match, moteur cité en premier | Paires | V–N–D | Score | Écart Elo [IC 95 %] |
|---|---:|---:|---:|---:|
| Septembre – SF1850 | 64 | 59–4–65 | 47,66 % | −16,3 [−141,6 ; +104,7] |
| Évaluation accélérée – septembre | 64 | 96–19–13 | 82,42 % | +268,4 [+111,0 ; +887,2] |
| Évaluation accélérée – SF1850 | 64 | 90–4–34 | 71,88 % | +163,0 [+34,1 ; +360,6] |
| Candidat complet – septembre | 64 | 111–14–3 | 92,19 % | +428,8 [+192,8 ; +∞] |
| Candidat complet – évaluation accélérée | 64 | 66–38–24 | 66,41 % | +118,4 [−4,0 ; +280,2] |
| Candidat complet – évaluation accélérée, confirmation | 128 | 144–70–42 | 69,92 % | +146,5 [+55,5 ; +262,5] |
| **Candidat complet – SF1850, confirmation** | **192** | **277–15–92** | **74,09 %** | **+182,5 [+102,1 ; +286,6]** |
| Candidat complet – SF2200 | 64 | 66–12–50 | 56,25 % | +43,7 [−75,7 ; +174,8] |

Dans le duel contre septembre, la profondeur médiane du premier coup hors ouverture passe de 4 à 7. Le candidat termine une itération plus profonde sur **64 ouvertures sur 64** à ce premier coup avec les Blancs. Face à SF1850 sur le lot réservé, il marque 139 victoires, 6 nulles et 47 défaites avec les Blancs, et 138 victoires, 9 nulles et 45 défaites avec les Noirs.

Les PGN et leurs manifestes contiennent les ouvertures, les noms, les résultats, la cadence et les empreintes vérifiées. Les révisions moteur sont `f6871de` (septembre), `c9ff718` (mobilité seule, chronométrée), `6b24f99` (évaluation accélérée) et `91aeb4e` (candidat complet). Le commit `2c43fdf` adapte les compteurs de nœuds attendus par les tests du traceur et ajoute un test du tri des captures. Les assertions sur les coups, scores, profondeurs, chemins et non-interférence du traceur passent. [Parties et manifestes](../../../matches/elo1850/) · [PGN de confirmation à 1850](../../../matches/elo1850/capt-sf1850-holdout.pgn)

Les huit matchs totalisent 1 408 parties, sans forfait au temps, coup illégal, déconnexion ni plantage. Les matchs ont été joués sur un Apple M4 Pro, six parties simultanées pour les matchs principaux, trois pour le premier duel entre candidats et le repère 2200, avec un maximum de neuf simultanées entre les matchs. Le moteur Equinox cherche sur un thread. Chaque Stockfish utilise un thread et 16 Mio de table. Les intervalles sont ceux du calcul par paire du projet. La borne supérieure peut être infinie quand la borne de score atteint 100 %.

## Sélection et confirmation

- Comparer à septembre deux configurations successives sur les 64 ouvertures de `openings-v2.epd`, couleurs inversées. La première accélère le calcul de l’évaluation. La seconde ajoute le tri des prises en quiétude.
- Vérifier v1, v2 et les 13 canaris du run6 à profondeur 6. Pour confirmer la configuration complète, exiger un gain direct contre septembre selon l’intervalle du projet, la conservation des coups de référence déjà trouvés et la réussite des tests.
- Tester une seule configuration retenue sur les **192 nouvelles ouvertures** de ce dossier contre Stockfish 1850. Jouer les 384 parties et calculer l’intervalle sur les 192 paires, sans arrêt anticipé. Garder ce résultat distinct des matchs de sélection.

Le duel du tri des prises contre la version déjà accélérée donne 66,41 % sur 64 paires, avec un intervalle de −4 à +280 Elo. Pour isoler sa contribution, un second duel utilise les **128 premières ouvertures** du nouveau lot (256 parties, arrêt fixé avant le premier coup). Sur ce second duel, le tri apporte un gain confirmé de +146,5 Elo à cette cadence (IC 95 % de +55,5 à +262,5). Les deux duels sont analysés séparément. Le repère contre Stockfish 2200 reprend les 64 ouvertures v2.

## Origine des ouvertures

La suite a été figée dans `b2c150f` avant sa première partie avec Equinox. Le générateur part de la position initiale, joue huit demi-coups et choisit parmi cinq variantes de Stockfish à 3 000 nœuds restant à 60 cp du meilleur coup. Une vérification à 16 000 nœuds garde les positions entre −60 et +60 cp. La graine est 18502200. Le générateur rejette les positions finales déjà présentes dans la suite précédente ou dans ce lot, en ignorant les deux compteurs FEN. Plusieurs lignes peuvent appartenir à la même famille d’ouvertures.

Stockfish utilise un thread et 16 Mio de table. L’empreinte SHA-256 du fichier EPD est `590061cbd1ff2801ffac0e97e872835d13afc2a02119e9b9398e81284f6bcfac`. [Générateur](build_holdout.py) · [Ouvertures](holdout.epd)

## Travail à profondeur 6

| Lot | Septembre et évaluation accélérée | Avec le tri des prises | Résultat tactique dans les trois versions |
|---|---:|---:|---|
| v1 | 1 480 222 nœuds | 897 840 | 32/32 coups, 15/16 pertes signalées |
| v2 | 3 289 616 | 1 558 741 | 13/32 coups, 3 fausses alertes |
| run6 | 7 772 789 | 1 245 101 | 10/13 erreurs évitées, 2 pertes signalées |
| **Total** | **12 542 627** | **3 701 682** | **77 coups et 77 scores identiques** |

La première configuration compte directement la mobilité sur les bitboards, réutilise les attaques des pièces à longue portée pour détecter les échecs et évite de recalculer la même évaluation pour chaque prise. La seconde cherche d’abord les captures de plus grande valeur en quiétude hors échec, avec l’attaquant le moins cher en premier en cas d’égalité de valeur de la victime. En échec, elle continue de rechercher toutes les évasions dans l’ordre de septembre. Les promotions sans prise restent filtrées hors échec. Sur Game17, le coût passe de 2 788 051 à 159 809 nœuds, toujours pour `d1f3` à +67 cp. Sur 0JLoe, les trois versions jouent `h5h6` à +38 cp. [Mesures par position et empreintes](tactics.json)

Sur un sous-échantillon systématique de 12 positions de v1/v2, la médiane de quatre passages à profondeur 6 descend de **1,428 s** (septembre) à **0,398 s** (mobilité seule), **0,289 s** (évaluation accélérée complète) et **0,146 s** (avec le tri des prises). Chaque position est mesurée dans un processus UCI neuf, de la fin de `isready` au retour du meilleur coup. L’ordre des quatre moteurs alterne entre les passages. Ces temps ont été relevés sur le M4 Pro pendant un match, et servent à expliquer la performance à cadence fixe. [Mesures chronométrées](timings.json)

## Rejouer un match

Construire chaque exécutable depuis la révision indiquée dans un clone propre avec `go build -trimpath -buildvcs=true -o /chemin/choisi ./cmd/uci`. Pour septembre, le binaire de match a été recompilé depuis `f6871de` avec `-buildvcs=false` et vérifié à l’octet près. Son SHA-256 est `e84d7f9af0e2ae4912fb1f26b2cf9152e76361cf385adf8dbe976afb89d99830`. Les lots tactiques ont utilisé une version avec empreinte Git intégrée. Leurs rapports conservent le SHA-256 de cet exécutable historique.

Depuis la racine du dépôt, cette commande relance la confirmation avec deux exécutables déjà construits. Choisir un nouveau chemin pour le PGN. Le lanceur vérifie les empreintes des binaires et des ouvertures et la structure des paires. Les résultats d’un match chronométré peuvent varier d’une exécution à l’autre.

```sh
go run ./cmd/elo_match -engine-a /chemin/equinox -name-a Captures \
  -engine-b /chemin/stockfish -name-b SF1850 \
  -option-b UCI_LimitStrength=true -option-b UCI_Elo=1850 \
  -option-b Threads=1 -option-b Hash=16 \
  -openings tools/elo/1850/holdout.epd -tc 1+0.1 -concurrency 6 \
  -pgn /tmp/equinox-confirmation.pgn
```

Pour recalculer les statistiques du PGN conservé, utiliser `go run ./cmd/matchstats -pgn matches/elo1850/capt-sf1850-holdout.pgn -engine-a Captures -engine-b SF1850`.
