# Développer EquinoxChess

## Suivre un coup dans le code

1. Les commandes UCI et les requêtes HTTP utilisent `api.Game`, qui réunit un plateau et une recherche. En UCI, `position` remplace le plateau et conserve la table de transposition et les heuristiques. `ucinewgame` réinitialise la partie. [Façade de jeu](../internal/api/game.go) · [Boucle UCI](../cmd/uci/main.go)
2. Le plateau comporte douze bitboards de pièces, les occupations, le trait, les droits de roque, la case en passant, les compteurs et la clé Zobrist. La pile d’annulation permet d’explorer puis de retirer un coup. Les coups sont encodés sur 32 bits. Le générateur parcourt les rayons des pièces glissantes jusqu’au premier obstacle. [Plateau](../internal/board) · [Encodage](../internal/move) · [Génération](../internal/movegen/movegen.go)
3. La recherche monothread approfondit par itérations avec negamax alpha-bêta et fenêtres d’aspiration. Sous la racine, le coup de table est essayé en premier. Le tri des autres coups utilise les prises, les coups tueurs et l’historique. Le coup nul, la réduction des coups tardifs (LMR) et l’élagage de futilité limitent les branches explorées. À la frontière, la recherche de quiétude examine les prises hors échec et les évasions en échec. [Recherche](../internal/search/search.go) · [Table de transposition](../internal/tt)
4. L’évaluation statique combine matériel, tables pièce-case, paire de fous, structure de pions, pions passés, mobilité et bouclier du roi. Les diagnostics de pression sur le roi et les variantes de réduction se trouvent dans les outils expérimentaux. [Évaluation](../internal/search/eval.go) · [Traceur](../cmd/search_trace/README.md)

## Modifier et valider

Garder le cœur de recherche déterministe, avec des types concrets, des tableaux de coups préalloués et une table de transposition de taille fixe. Garder les allocations, les logs et les goroutines hors de la boucle de recherche. Pour une optimisation, mesurer les allocations, le temps et les nœuds à profondeur fixe, puis la qualité du jeu à temps égal.

| Changement | Contrôle depuis la racine |
|---|---|
| Plateau ou génération | `go test -count=1 ./internal/board ./internal/movegen ./internal/perft`. Perft compte les feuilles légales, dont 4 865 609 depuis la position initiale à profondeur 5 et 4 085 603 depuis Kiwipete à profondeur 4. |
| Recherche ou évaluation | `go test -count=1 ./...`. Pour voir les résultats position par position, lancer `go test -count=1 -v ./internal/search -run 'TestSearchTacticalCanaryMetrics\|TestSearchAvoidsKnownStalematingMoves'`. Ajouter une position de régression ou un invariant qui exerce le mécanisme modifié. |
| Instrumentation | `go test -count=1 -tags searchtrace ./...`. Le rejeu des parcours historiques utilise `go test -count=1 -tags 'searchtrace trace1850' ./...`. En mode observationnel, le traceur compare coup, score, profondeur et nœuds avec une recherche sans trace. |
| UCI et contrôles statiques | `go test -race -count=1 ./cmd/uci` et `go vet ./...`. |
| Coût de recherche | `go test -run '^$' -bench BenchmarkNegamax -benchmem ./internal/search` sur la même machine pour les deux versions. |
| Analyse Python | Avec les dépendances de [l’analyseur](../tools/kingzone/README.md), lancer `python3 -m unittest discover -s tools/kingzone -p 'test_*.py'`. |

Une hypothèse de gain de force se teste sur un seul changement à la fois. Figer les deux exécutables, jouer les mêmes ouvertures avec couleurs inversées, conserver le PGN et son manifeste, puis lire le score par paire, son intervalle et les terminaisons anormales. Les tests de positions permettent de repérer quel comportement a changé. Le match mesure son effet en partie. [Protocole et commandes](../tools/elo/README.md)
