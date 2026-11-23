# Validation tactique indépendante · v1

Le lot contient une décision par partie et par ouverture. Le fichier à conserver pour comparer deux versions du moteur est [positions.json](v1/positions.json). Chaque entrée donne la FEN, le numéro de partie et le demi-coup dans le [PGN source](v1/source.pgn), le coup joué en partie et les deux analyses de référence.

## Référence de septembre à profondeur 6

Le lot a été figé dans le commit `850def5`. Le binaire UCI provient d’un clone propre de `819a8d5` (Go 1.26.2, macOS arm64). Une régénération complète a reproduit à l’octet près le PGN des 128 parties et le fichier des 32 positions. Pour les 32 positions, les deux recherches concordent sur le coup, le score, la profondeur et les nœuds. [Rapport par position et empreintes](v1/septembre.json)

| Groupe | Coup de référence trouvé | Évaluation Equinox ≤ −300 cp | Nœuds sur un passage |
|---|---:|---:|---:|
| 16 positions tenables ou favorables | 16/16 | 0/16 fausse alerte | 993 345 |
| 16 pertes selon Stockfish | 16/16 | 15/16 pertes reconnues | 486 877 |
| Ensemble | 32/32 | 15 pertes signalées | 1 480 222 |

Sur **G57-p59**, Equinox trouve `Kxe1` (`f2e1`) et évalue la position à −277 cp. Stockfish choisit ce même coup et donne −429 puis −416 cp. L’écart porte donc sur l’évaluation de la position après le meilleur choix.

L’accord sur les coups atteint déjà le maximum. Les 32 coups de référence avaient tous été joués dans les parties sources et 25 sont des prises immédiates. Ce lot est sensible à une régression sur ces décisions et permet de suivre le coût de recherche et la détection des pertes. Une mesure de progression sur des choix tactiques plus difficiles demande un autre échantillon aveugle. Le run6 sélectionnait des erreurs historiques d’Equinox ; les deux taux portent sur des populations différentes.

## Constitution du lot

Les parties sont produites par Stockfish 18 depuis les 64 positions à huit demi-coups de la suite [openings-v2.epd](../elo/openings-v2.epd). Une partie oppose deux budgets de 180 et 5 000 nœuds par décision. Les couleurs sont inversées pour la seconde partie de chaque ouverture. La génération s’arrête après 80 demi-coups supplémentaires ou en fin de partie. Le PGN peut donc contenir des parties inachevées. Les annotations servent à repérer des positions candidates.

Le constructeur écarte une ouverture présente dans le run6, puis écarte tout candidat présent **à un demi-coup quelconque** du run6 ou dans plusieurs parties du nouveau PGN. La comparaison conserve les pièces, le trait, les droits de roque et la prise en passant légale, et ignore les deux compteurs FEN. La sélection commence au demi-coup 17. Dans chaque partie, un ordre fixé par SHA-256 et la graine `equinox-tactics-v1` départage les positions. Dès qu’une position est retenue, les deux parties de son ouverture sont closes pour la sélection.

Stockfish examine les deux meilleurs coups (MultiPV 2) avec un thread, 16 Mio de table et un état neuf à **24 000 puis 120 000 nœuds**. Les deux recherches doivent désigner le même premier coup, lui donner un avantage d’au moins 150 centipions sur le deuxième et montrer une prise, un échec ou une promotion dans les trois premiers demi-coups de la variante. À chaque budget, le constructeur lit la dernière paire complète de variantes à profondeur commune avec des scores numériques sans indication de borne. Le score est toujours vu du camp au trait.

| Groupe | Positions | Score du meilleur coup aux deux budgets |
|---|---:|---|
| `tactic` · position tenable ou favorable | 16 | −100 à +600 cp |
| `loss` · perte selon Stockfish | 16 | −900 à −300 cp |

Les seuils et les étiquettes viennent de Stockfish. Un premier passage sur 64 parties a produit seulement 13 pertes admissibles. Le second passage a ajouté les parties à couleurs inversées et exclu les positions partagées entre les deux passages. Le lot a été figé avant la première recherche d’Equinox. Ces parties synthétiques utilisent des ouvertures déjà connues du protocole Elo du projet. Les résultats caractérisent ces deux groupes sélectionnés, à profondeur fixe.

## Relire la mesure ou comparer un candidat

Depuis la racine du dépôt, avec Go et l’environnement Python de [l’analyseur de parties](../kingzone/README.md#setup-and-run), compiler la révision de référence dans un clone propre puis lancer le banc depuis le dépôt courant.

```sh
essai=$(mktemp -d)
git clone --quiet . "$essai/reference"
git -C "$essai/reference" checkout --quiet --detach 819a8d5b12995e2daf68d374d8e667ec06c33fcf
(cd "$essai/reference" && go build -trimpath -buildvcs=true -o "$essai/equinox" ./cmd/uci)
python3 tools/tactics/measure.py "$essai/equinox" "$essai/septembre.json"
```

Pour mesurer un candidat, passer son exécutable UCI et un nouveau chemin de sortie à la dernière commande. Le banc vérifie les empreintes des entrées, retrouve chaque FEN dans son PGN et contrôle son indépendance et ses références. Il démarre un processus UCI neuf par FEN et par répétition, à profondeur 6 sans horloge. Les deux passages doivent donner un coup légal, le même coup, le même score et le même nombre de nœuds à la profondeur demandée. Le rapport enregistre la révision Git inscrite par Go, l’empreinte du binaire, celle du lot et le résultat de chaque position.

Comparer les rapports par identifiant de position. Relever les coups de référence gagnés et perdus dans chaque groupe, les pertes reconnues (score Equinox ≤ −300 cp ou mat subi), les fausses alertes dans le groupe tenable et le coût en nœuds. Les marges de Stockfish sont des estimations à budget fini. Le coup joué dans le PGN sert à retrouver la décision et peut lui-même être le meilleur coup.

## Reconstituer les références

```sh
sortie=$(mktemp -d)
python3 tools/tactics/build_corpus.py "$(command -v stockfish)" "$sortie"
```

Pour réarbitrer les parties conservées, copier `v1/source.pgn` dans ce répertoire avant la commande. La simple mesure d’un candidat utilise les références figées et se passe de Stockfish. Les versions et les empreintes du binaire d’arbitrage, des ouvertures et des PGN figurent dans le lot. Utiliser la même version de python-chess et le même binaire Stockfish pour comparer les empreintes de régénération.
