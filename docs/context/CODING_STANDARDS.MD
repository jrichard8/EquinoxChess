# Coding Standards

## Principes

- Go idiomatique
- code simple a profiler
- flux de donnees explicite
- aucune abstraction gratuite dans le hot path
- documentation synchronisee avec le comportement reel

## Regles Pour Le Noyau

Concerne surtout :

- `internal/board`
- `internal/movegen`
- `internal/search`
- `internal/tt`

Attendus :

- types concrets
- boucles simples
- tableaux prealloues
- comportement deterministe

Eviter :

- interfaces inutiles
- allocations repetitives
- logique cachee
- heuristiques ajoutees sans protocole de mesure

## Regles De Changement

- Toute heuristique de recherche doit etre accompagnee d'une justification et d'un protocole de verification.
- Toute modification de l'evaluation doit idealement ajouter un test d'invariant ou une position de regression.
- Toute doc de contexte doit decrire l'etat courant, pas seulement la cible.
