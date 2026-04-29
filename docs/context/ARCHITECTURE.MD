# Architecture Snapshot

## Vue D'ensemble

Le projet est organise autour de deux executables et d'un noyau moteur dans `internal/`.

- `cmd/uci`
  Interface UCI pour cutechess, GUIs et benchs.
- `cmd/server`
  Petit serveur HTTP avec UI web pour jouer localement.
- `cmd/list_moves`
  Outil de debug pour lister des coups.
- `cmd/search_test`
  Binaire utilitaire de test manuel de recherche.

## Noyau Moteur

- `internal/board`
  Representation du plateau, FEN, hash Zobrist, repetition, insuffisance materielle, make/unmake et null move.
- `internal/move`
  Encodage compact des coups en `uint32`.
- `internal/movegen`
  Generation pseudo-legale et legale, y compris promotions, en passant et roques.
- `internal/search`
  Negamax alpha-beta, iterative deepening, aspiration windows, quiescence, TT probe/store, killer/history, null-move pruning, LMR et evaluation statique.
- `internal/tt`
  Table de transposition a taille fixe et adressage par masque.
- `internal/uci`
  Conversion coups <-> UCI.
- `internal/api`
  Facade `Game` utilisee par l'UCI et le serveur web.
- `internal/perft`
  Verification de correction de generation des coups.
- `internal/zobrist`
  Clefs de hash pre-generees.

## Representation Du Plateau

Le plateau utilise :

- 12 bitboards de pieces.
- 3 bitboards d'occupation.
- `SideToMove`, droits de roque, case en passant.
- `HalfmoveClock`, `FullmoveNumber`.
- `ZobristKey`.
- Une pile `undoStack` de taille fixe indexee par ply.

## Recherche Actuelle

La recherche est single-thread et basee sur un `Search` mutable :

- iterative deepening depuis `SearchRoot`
- aspiration windows a partir de la profondeur 2
- negamax alpha-beta
- quiescence search
- transposition table
- null-move pruning avec verification
- late move reductions
- killer moves et history heuristic
- stop flag atomique et limite de temps simple

## Evaluation Actuelle

L'evaluation reside dans `internal/search/eval.go` et combine :

- materiel
- piece-square tables
- paire de fous
- pions doubles et isoles
- mobilite tres simple
- pions passes
- bouclier de pions devant le roi en milieu de jeu

Elle ne contient pas encore :

- tapered eval opening/endgame
- structure de pions detaillee
- evaluation de finales
- SEE
- evaluation de menace et initiative

## Dette Technique Importante

- La doc historique parlait d'une architecture cible, pas de l'etat present.
- L'evaluation et le protocole de mesure sont en retard par rapport a la sophistication de la recherche.
- Certains comportements de draw, conversion et gestion tactique meritent une revue prioritaire avant toute nouvelle optimisation.
