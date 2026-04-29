# Performance Rules

## Principes

Le moteur cherche deja a rester sobre en allocations sur les chemins chauds. Cette contrainte reste valable, mais elle ne doit plus masquer les problemes de force.

Ordre de priorite actuel :

1. Correction
2. Force Elo
3. Determinisme
4. Performance

## Ce Qui Doit Rester Strict

- pas d'allocations inutiles dans la boucle de recherche
- tableaux fixes pour les coups
- TT de taille fixe
- pas de maps ni reflection dans le hot path
- pas de goroutines dans le coeur de recherche

## Ce Qui Peut Etre Assoupli Hors Hot Path

- instrumentation de bench
- logs de regression
- scripts de match
- outils d'analyse de positions

## Discipline De Mesure

Avant une optimisation :

- mesurer `go test -bench`
- mesurer profondeur atteinte et nodes
- mesurer impact Elo ou au moins impact sur suite tactique

Ne pas accepter une optimisation qui :

- casse le determinisme
- augmente les bugs de conversion
- fait baisser la qualite des choix a temps egal

## Point D'Attention Actuel

Le projet semble davantage limite par la qualite de l'evaluation, la gestion des nulles et la calibration des heuristiques que par la performance brute.
