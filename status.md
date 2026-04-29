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

## Diagnostic profondeur 2026-04-29

Mesure isolee sur les `5` positions rouges prioritaires, `SearchRoot` a depths `4/6/8/10/12`, sans contrainte de temps :

| Position | d4 | d6 | d8 | d10 | d12 | Coup stable des |
|----------|------|------|------|------|------|-----------------|
| G02_B    | +16  | -7   | -3   | +9   | -20  | d6 (`g7g6`) |
| G17_W    | +90  | +67  | +50  | +37  | +17  | d4 (`d1f3`) |
| G22_B    | -3   | -32  | -69  | -76  | -94  | d4 (`e4f2`) |
| G26_B    | -3   | -36  | -49  | -53  | -54  | d4 (`g2g1`) |
| G30_B    | +25  | +12  | 0    | -20  | -31  | d6 (`c6e5`) |

Aucune position ne franchit `-300 cp` meme a depth `12`. Le score decroit lentement et plateau (G22 le plus rapide a `-94`, G17 reste a `+17`). Les coups choisis sont stables des depth `6`.

**Verdict** : le canari ne se debloquera *pas* par plus de profondeur effective. Audit LMR / null-move n'est plus la voie prioritaire. C'est un probleme d'**evaluation** : le moteur ne voit pas que ces positions sont fondamentalement perdues. **Retour anticipe Phase 3** decide.

Cout : `~42 min` wall-time cumulees pour les `5` positions a depth `12` (G17 seul = `21 min`). Diagnostic one-shot, l'outil `cmd/depth_diagnostic` a ete supprime apres la mesure.

## Plan global

| Phase | Objectif | Statut |
|-------|----------|--------|
| 1. Remettre sur des rails | Stopper les regressions Elo et les bugs de score | Close |
| 2. Nettoyer la recherche | Ordering, quiescence, pruning. Canari anti-blunder au vert | Pause (canari `2/13`, diagnostic dit "pas un pb de profondeur") |
| 3. Faire grandir l'evaluation | King safety, threat awareness, tapered eval, structure de pions | **En cours** (anticipee suite diagnostic 2026-04-29) |
| 4. Infrastructure Elo | Gauntlet reproductible, baselines SF1400/1600/1800, seuils de rejet | A venir |
| 5. Viser 2000 Elo | Battre SF1400 -> tenir SF1600 -> depasser SF1600 | A venir |

## Cibles Phase 3 prioritaires

Les `5` positions rouges du diagnostic partagent un theme dominant : roi adverse expose ou attaque proche, mais Equinox ne percoit pas le risque. Leviers a tester un par un sur le canari (anti-blunder reste la metrique) :

1. **King safety renforcee** — l'eval actuelle n'a qu'un bouclier de pions `+/-10/-5`, conditionne a `nonPawnMaterial > 2000`. Manquent : attacker count, valeur des pieces attaquant la zone du roi, files ouvertes vers le roi. Probable plus gros gain sur G02, G26, G30
2. **Threat awareness** — penaliser les pieces lourdes adverses pointant vers des cibles non defendues (queen X-ray, rook on open file vers king)
3. **Tapered eval** : interpolation middlegame/endgame, PSQT roi distinctes
4. **Structure de pions etendue** : connected, candidate passed, backward (deja : doubled, isolated, simple passed)
5. **Activite des tours** : files ouvertes / semi-ouvertes, `7e` rangee

## Regles de pilotage

- Un fix Phase 2 par iteration, mesure isolee sur le canari avant tout autre changement
- Conserver uniquement les changements qui montent le canari, sauf correctness fix explicitement documente
- Ne lancer un nouveau match qu'apres canari `>= 10/13` verts
- Preferer une progression Elo stable a une hausse ponctuelle non reproductible
- Documentation = partie du moteur

## Prochaine etape

Implementer **king safety renforcee** isolee, mesurer son effet sur le canari anti-blunder. Si gain (>= `+2` verts), garder ; sinon revert et passer au levier suivant (threat awareness ou tapered eval). Meme protocole que les fixes Phase 2 : un levier a la fois, mesure isolee avant tout autre changement.

Pas de match SF1600 tant que le canari n'atteint pas `>= 10/13` verts.
