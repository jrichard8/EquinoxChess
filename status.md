# EquinoxChess — Status

## Snapshot 2026-04-29

- Moteur fonctionnel, parties completes, scores PGN propres (mate notation, plus de bornes d'aspiration trompeuses)
- Plateau Elo vs SF1600 au TC `1+0.1` sur `30` parties : `12-18-0`, `40.0%`, `-70.4 +/- 133.7` Elo, DrawRatio `0.0%`. PGN run6 : `matches/sf1600-run6-1+0.1-30.pgn`
- **Phase 1 close** (nettoyage scores / PV / timeout / mate). **Phase 2 en cours** (recherche : ordering, quiescence, pruning)

## Trajectoire matchs SF1600 `1+0.1` (`30` parties chacun)

| Run | Score | W-D-L | Stalemates | Note |
|-----|-------|-------|------------|------|
| baseline | `31.7%` | `1-17-12` | plusieurs | point de depart, `-133` Elo |
| run3 | `36.7%` | `7-8-15` | `6` | apres premiers anti-pat (quiescence + Negamax terminal) |
| run4 | `38.3%` | `11-1-18` | `0` | apres fix PV-root fail-high |
| run5 | `46.7%` | `13-2-15` | `0` | apres fix fallback timeout |
| run6 | `40.0%` | `12-0-18` | `0` | apres fix scores de mat ; PGN sain |

Plateau `40-46%` dans le bruit (`+/- 133` Elo). La phase comptable est terminee, reste la force de jeu reelle.

## Canari Phase 2 actif (anti-blunder)

`TestSearchSeesLossingTacticAtSwingPly` dans [internal/search/search_anti_blunder_test.go](internal/search/search_anti_blunder_test.go).

- `13` FEN extraites du run6 : `7` balanced_blunder + `2` unclassified + `4` conversion_blunder courtes
- Chaque sous-test : `SearchRoot` a profondeur `6` doit retourner score `<= -300cp` ou mate-against
- **Etat : `2/13` verts** (Game06, Game09, Game21). Sur certains rouges, le score reste meme positif a depth `6` : G16 `+288`, G25 `+442`, G28 `+120`
- Definition of done : `>= 10/13` verts avant de relancer un match

## Canari Phase 1 (anti-stalemate, vert)

`TestSearchAvoidsKnownStalematingMoves` dans [internal/search/search_anti_stalemate_test.go](internal/search/search_anti_stalemate_test.go).

- `6` FEN issues du run3, `6/6` verts depuis le fix PV-root fail-high

## Fixes Phase 2 essayes

| Fix | Effet canari | Decision | Note |
|-----|--------------|----------|------|
| MVV-LVA propre (king-attaquant `10000`, promotions) | `2/13 -> 2/13` | Keep | Correctness fix d'ordering, neutre sur le canari |
| Quiescence elargie aux echecs (`qPly`, second pass) | `2/13 -> 2/13`, +temps `x1.8`, casse mate-score test | Revert | Mauvais trade-off |

Enseignement : les heuristiques generiques d'ordering ou d'extension de qsearch ne suffisent pas individuellement. Le canari ne se debloquera que par soit une augmentation de profondeur effective, soit une comprehension fine des motifs tactiques manques.

## Leviers Phase 2 a explorer

1. **Diagnostic profondeur** sur `4-5` rouges (G02, G17, G22, G26, G30) : `SearchRoot` a depth `8/10/12`, observer a quelle profondeur le score bascule `<= -300cp`. Si jamais : probleme d'eval, pas de profondeur
2. SEE-lite pour ecarter les captures perdantes des killers et de la PV
3. Audit null-move pruning : defenses raccourcies en finale ?
4. Audit LMR : defenses raccourcies en milieu de partie ?
5. Revoir futility profondeur `1` sur menace tactique

Le diagnostic `1` conditionne le choix entre `2-5` et un retour anticipe Phase 3 (eval).

## Plan global

| Phase | Objectif | Statut |
|-------|----------|--------|
| 1. Remettre sur des rails | Stopper les regressions Elo et les bugs de score | Close |
| 2. Nettoyer la recherche | Ordering, quiescence, pruning. Canari anti-blunder au vert | En cours (`2/13`) |
| 3. Faire grandir l'evaluation | Tapered eval, structure de pions, tours, roi de finale | A venir |
| 4. Infrastructure Elo | Gauntlet reproductible, baselines SF1400/1600/1800, seuils de rejet | A venir |
| 5. Viser 2000 Elo | Battre SF1400 -> tenir SF1600 -> depasser SF1600 | A venir |

## Regles de pilotage

- Un fix Phase 2 par iteration, mesure isolee sur le canari avant tout autre changement
- Conserver uniquement les changements qui montent le canari, sauf correctness fix explicitement documente
- Ne lancer un nouveau match qu'apres canari `>= 10/13` verts
- Preferer une progression Elo stable a une hausse ponctuelle non reproductible
- Documentation = partie du moteur

## Prochaine etape

Diagnostic profondeur sur G02, G17, G22, G26, G30 : score et coup choisi a depths `8`, `10`, `12`. Decider ensuite :

- score bascule `<= -300cp` a depth raisonnable -> probleme de profondeur effective -> audit LMR / null-move
- score reste positif meme a depth profonde -> probleme d'eval -> retour anticipe Phase 3
