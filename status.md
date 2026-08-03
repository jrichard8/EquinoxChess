# EquinoxChess — Status

## Update 2026-07-31

- Harness Elo apparie ajoute : `32` EPD fixes, deux parties par ouverture avec
  couleurs inversees, PGN valide structurellement, statistiques par paire et
  manifeste SHA-256 (`tools/elo/`).
- Trois changements de recherche ont ete rejetes apres mesure isolee : priorite
  TT/PV dure, priorite PV racine seule et comptage des coups legaux pour le LMR.
- Correction UCI retenue techniquement : `position` conserve maintenant TT,
  history et killers jusqu'a `ucinewgame`. Un garde-fou TT sur
  `HalfmoveClock` interdit les cutoffs issus d'un autre compteur 50-coups tout
  en conservant le coup TT pour l'ordering.
- A/B long combine vs baseline, `64` parties a `1+0.1` : `34-6-24`, score
  `57.81%`, estimation `+54.7 Elo` (Hoeffding 95% encore inconclusif).
- A/B causal combine vs garde-fou seul, deux lots de `64` : total
  `61-23-44`, score `56.64%`, estimation `+46.4 Elo`. Les deux lots sont
  positifs, mais l'intervalle conservateur contient encore zero : signal
  reproductible, pas preuve Elo definitive.
- Cout connu : le garde-fou TT fait repasser Game09 de `-342` a `-128`; le
  canari anti-blunder revient donc de `3/13` a `2/13`. Ce recul reste visible et
  devra etre compense, pas masque.
- Gestion du temps UCI avec increment corrigee et retenue : budget
  `remaining/40 + 3*increment/4`, plafonne a `remaining/10` et borne a `1 ms`.
  `movetime` et le chemin sans increment restent inchanges. A `1+0.1`, le
  budget nominal passe de `50 ms` a `100 ms`.
- Deux A/B complets contre l'ancien budget, `64` parties chacun a `1+0.1` :
  `43-6-15` puis `42-11-11`. Agregat principal `85-17-26`, score `73.05%`,
  estimation `+173.2 Elo`; intervalle Hoeffding 95% conservateur
  `[+42.4, +382.1] Elo`. Le gain est positif dans les deux repetitions et dans
  l'intervalle agrege : changement accepte.
- Le temps moyen par coup passe bien de `49.9 ms` a `99.8 ms`; sur les deux
  lots, la profondeur moyenne hors marqueurs terminaux passe de `5.95` a `6.58`,
  la mediane de `6` a `7` et le p90 de `8` a `9`. Aucun forfait au temps, crash
  ou coup illegal sur les `128` parties principales (ni sur le gate initial de
  `32` parties).
- Validation : tests UCI/TT/matchstats cibles, `go test -race ./cmd/uci` et
  `go vet ./...` verts. `go test ./...` reste rouge uniquement sur les `11/13`
  cas deja connus du canari anti-blunder.
- Infrastructure externe durcie : options UCI propres a chaque moteur tracees
  dans le manifeste, terminaisons anormales ventilees par moteur, et suite v2
  de `64` ouvertures legales/uniques (`128` parties avec couleurs inversees).
- Gate externe vs Stockfish 18 limite a 1600, `32` parties a `1+0.1` :
  `20-1-11`, score `64.06%`, `+100.4 Elo` indicatifs, zero incident. Ce gate
  chevauche la suite longue et n'est pas additionne au resultat final.
- Calibration externe principale vs SF18-1600, `128` parties / `64` paires :
  `78-4-46`, score `62.50%`, estimation `+88.7 +/- 61.8 Elo` (LOS `99.8%`).
  L'intervalle Hoeffding apparie, plus conservateur, reste inconclusif :
  `[-31.2, +235.2] Elo`.
- Forte asymetrie couleur au TC ultra-court : Equinox blanc `47-3-14`
  (`75.8%`), Equinox noir `31-1-32` (`49.2%`); score global des Blancs
  `63.3%`. Le pairage controle ce biais dans le score total.
- Cinq forfaits au temps dans le lot long : `3` pour Equinox, `2` pour
  Stockfish. Aucun crash, coup illegal, disconnect ou stall. Le gain de force
  de la nouvelle allocation reste etabli par l'A/B interne, mais sa marge
  anti-flag est insuffisante pour declarer la gestion du temps terminee.
- Bilan niveau : calibration nominale Equinox autour de `1689` a `1+0.1`
  (intervalle usuel environ `[1627, 1751]`), nettement meilleure que le snapshot
  historique mais non comparable causalement a run6 (SF/protocole/suite
  differents). Le niveau `1800` n'est pas encore demontre.

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
- **Etat : `2/13` verts** (Game06, Game21). Sur certains rouges, le score reste meme positif a depth `6` : G16 `+288`, G25 `+453`, G28 `+120`
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
| 4. Infrastructure Elo | Gauntlet reproductible, baselines SF1400/1600/1800, seuils de rejet | **En cours** (A/B apparie disponible) |
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

Corriger de facon isolee la marge anti-flag de la gestion du temps, sans toucher
a la recherche ni a l'evaluation. Premiere candidate recommandee : soustraire un
overhead fixe de `10 ms` au budget calcule (apres le plafond, avant le plancher
de `1 ms`), ce qui donne `90 ms` au lieu de `100 ms` a `1+0.1`. L'accepter
seulement si un A/B apparie sur la suite v2 conserve l'essentiel du gain Elo et
si une validation externe longue ne produit plus aucun forfait Equinox.
