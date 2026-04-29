# Engine Status

## Identite

- Nom UCI : `EquinoxChessV1.0 (Go)`
- Langage : Go
- Style : moteur classique alpha-beta, sans NNUE
- Execution : single-thread

## Vision Produit

Le projet vise toujours un moteur educatif, lisible et capable d'atteindre environ 2000 Elo, mais le travail prioritaire n'est plus d'ajouter les briques de base.

Le besoin actuel est de transformer un moteur deja jouable en moteur mesurable, stable et progressif en Elo.

## Ce Qui Est Deja Fait

- UCI jouable
- Serveur web local
- Perft de reference
- Hash Zobrist incremental
- TT
- Quiescence
- Null move
- LMR
- Killer/history

## Ce Qui Manque Pour Une Version "Force"

- protocole Elo reproductible
- suite de regression tactique et strategique
- evaluation plus robuste en finales et conversions
- meilleure gestion des nulles et repetitions
- instrumentation plus fine des causes de defaite

## Definition De Succes Pour Les Prochaines Iterations

- Chaque changement de force doit etre mesure.
- Toute regression sur la suite tactique ou les matches de reference doit bloquer le merge.
- La roadmap doit privilegier les gains robustes avant les optimisations micro-perf.
