# EquinoxChess — Status

## Update 2026-08-05

- Validation externe du meme binaire `Overhead15ms` contre Stockfish 18 limite
  a `1600 Elo`, `128` parties / `64` ouvertures v2 a `1+0.1` : `79-2-47`,
  score `62.50%`, estimation `+88.7 Elo`, intervalle Hoeffding 95%
  `[-31.2, +235.2] Elo`.
- Securite externe validee : zero forfait, crash, coup illegal, disconnect,
  stall ou terminaison inconnue pour Equinox et Stockfish. Les couleurs sont
  equilibrees : `40-1-23` avec les Blancs, `39-1-24` avec les Noirs.
- Le score brut est identique a la calibration externe precedente de la
  politique `100 ms` (`62.50%`, `78-4-46`), tandis que les incidents passent
  de `3` forfaits Equinox et `2` Stockfish a zero. La comparaison inter-lots
  reste une calibration; le long A/B direct `+19 Elo` fournit l'attribution
  causale.
- Les `11256` plies sont annotees. Equinox produit `66/5636` commentaires
  temps-seul (`1.17%`), compatibles avec un timeout avant profondeur complete;
  ce risque de recherche temporisee reste a traiter separement, sans incident
  UCI observe.
- Decision : **overhead fixe de 15 ms accepte**. Il totalise zero forfait sur
  le gate `32`, le long A/B `128` et l'externe `128`, sans perte de force
  detectee. Ne plus tester les constantes `10/20 ms`.

## Update 2026-08-04

- Validation longue de l'overhead fixe `15 ms` sur les `64` ouvertures v2,
  `128` parties appariees a `1+0.1` : `52-31-45`, score `52.73%`, estimation
  `+19.0 Elo`, intervalle Hoeffding 95% `[-101.8, +144.8] Elo`.
- Les criteres prefixes passent : point estime largement superieur a
  `-40 Elo`, zero forfait et zero autre terminaison anormale pour les deux
  moteurs. Le resultat etablit une non-inferiorite pratique, pas un gain Elo
  prouve, car l'intervalle conservateur contient encore zero.
- Les annotations couvrent les `16394` plies. Les commentaires temps-seul,
  compatibles avec un timeout avant profondeur complete, restent symetriques :
  `41/8199` coups candidat (`0.50%`) contre `39/8195` baseline (`0.48%`).
- Cout moyen limite : candidat `93.89 ms`, profondeur annotee `7.29`; baseline
  `99.85 ms`, profondeur `7.36`. Le candidat marque `57.03%` avec les Blancs
  et `48.44%` avec les Noirs, sans l'asymetrie alarmante du test `20 ms`.
- Decision : **GO vers la validation externe SF18-1600**, en reutilisant
  strictement le binaire candidat fige SHA-256
  `9b5d5a983a9dbfc7b644621de0584061e30f1bd1866d1a68ef2127483258253e`.

## Update 2026-08-03

- Overhead UCI fixe de `10 ms` teste isolement apres le cap `remaining/10` et
  avant le plancher `1 ms`; budget initial a `1+0.1` : `100 -> 90 ms`.
- Gate causal vs politique `100 ms`, `32` parties / `16` paires : `18-1-13`,
  score `57.81%`, estimation `+54.7 Elo`. Hors les deux parties finies au
  temps : `17-1-12`, score `59.17%`.
- Cout de recherche faible : temps candidat moyen `96.2 ms` (`92.9 ms` sur ses
  dix premiers coups, `96.8 ms` ensuite), profondeur moyenne `6.58` contre
  `6.72` pour la baseline.
- Securite non validee : `1` forfait candidat et `1` forfait baseline. Le
  candidat abandonne au temps une position qu'il evalue autour de `+7.8` au
  ply `81`; aucun crash, coup illegal, disconnect ou stall.
- Decision : **rejet de l'overhead 10 ms**. Le patch moteur a ete retire. Les
  A/B `128` et la revalidation SF18-1600 n'ont volontairement pas ete lances,
  car le gate ne satisfait pas le critere dur de zero forfait candidat.
- Overhead UCI fixe de `20 ms` teste au meme emplacement; budget initial a
  `1+0.1` : `100 -> 80 ms`. Gate causal vs politique `100 ms`, `32` parties /
  `16` paires : `14-6-12`, score `53.12%`, estimation `+21.7 Elo`, intervalle
  Hoeffding 95% conservateur `[-250.0, +331.4] Elo`.
- Securite du gate `20 ms` validee : zero forfait, crash, coup illegal,
  disconnect, stall ou terminaison inconnue pour les deux moteurs.
- Temps candidat moyen `91.7 ms` (`83.9 ms` sur ses dix premiers coups,
  `93.2 ms` ensuite), contre `99.9 ms` pour la baseline. Sur les coups avec
  profondeur annotee, moyenne identique de `7.01`; aucune profondeur explicite
  egale a zero.
- Point de vigilance : `23/2017` coups candidat et `20/2015` coups baseline
  n'ont qu'une annotation temps dans le PGN. Un replay `-debug all` confirme
  que ce format correspond a `info depth 0`, donc a un timeout avant toute
  profondeur complete. Le phenomene n'est pas specifique au candidat, mais
  doit etre suivi sur l'echantillon long.
- Validation longue `20 ms` sur les `64` ouvertures v2, `128` parties :
  `40-28-60`, score `42.19%`, estimation `-54.7 Elo`, intervalle Hoeffding 95%
  `[-188.9, +64.4] Elo`. Le point estime echoue le seuil prefixe d'environ
  `-40 Elo`, meme si l'intervalle conservateur reste inconclusif.
- Securite longue confirmee : zero forfait et zero autre terminaison anormale.
  Les `depth 0` sont meme legerement moins frequents pour le candidat :
  `35/8237` coups (`0.42%`) contre `37/8251` (`0.45%`) pour la baseline.
- Cout moyen mesure : `91.45 ms` et profondeur `7.14` pour le candidat, contre
  `99.84 ms` et profondeur `7.34` pour la baseline. Le candidat fait
  `24-18-22` avec les Blancs mais `16-10-38` avec les Noirs; l'appariement et
  les `64` parties par couleur excluent un simple desequilibre de protocole.
- Decision : **rejet de l'overhead 20 ms pour perte de force**. Le patch moteur
  a ete retire et la validation SF18-1600 n'a pas ete lancee. Le gate court
  positif etait trop bruite pour predire le resultat long.
- Overhead fixe de `15 ms` teste comme unique interpolation; budget initial a
  `1+0.1` : `100 -> 85 ms`. Gate causal `32` parties / `16` paires :
  `17-2-13`, score `56.25%`, estimation `+43.7 Elo`, intervalle Hoeffding 95%
  `[-216.9, +385.6] Elo`.
- Securite du gate `15 ms` validee : zero forfait et zero autre terminaison
  anormale pour les deux moteurs. Les `depth 0` restent rares et ne montent
  pas contre la baseline : `7/2050` (`0.34%`) contre `9/2047` (`0.44%`).
- Cout du gate : temps moyen `93.8 ms` et profondeur annotee moyenne `6.54`
  pour le candidat, contre `100.0 ms` et `6.62` pour la baseline.
- Decision : **GO conditionnel pour le long `128` de l'overhead 15 ms**. Le
  gate court satisfait les criteres prefixes, mais ne constitue pas une
  validation Elo au vu du bruit observe avec le candidat `20 ms`.

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

Committer l'overhead `15 ms`, ses tests et ce bilan afin de rendre le binaire
reproductible. Construire ensuite une nouvelle baseline immuable depuis ce
commit et mesurer `128` parties contre Stockfish 18 limite a `1800 Elo`, avec
les memes ouvertures, le meme controle et les memes options techniques. Cette
calibration doit estimer l'ecart restant vers l'objectif `1800+`; elle ne doit
pas etre presentee comme un Elo absolu. Conserver comme piste separee la
reduction des `depth 0` en recherche temporisee.
