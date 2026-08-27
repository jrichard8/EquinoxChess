# EquinoxChess — Status

## Update 2026-08-26

- Le contrat des 13 canaris tactiques est maintenant separe en deux metriques
  independantes, calculees sur une seule recherche fraiche a profondeur 6 par
  position. `BadMoveAvoidance` compare le meilleur coup au coup historique
  fautif; `LossRecognition` conserve sans changement le seuil `score <= -300`
  depuis le POV du camp au trait.
- Baseline formalisee : evitement `10/13`; Game03, Game17 et Game21 rejouent
  encore le coup historique. Reconnaissance `2/13`; seuls Game06 et Game21
  voient la perte sous le seuil. Game09 evite bien `Ra7` tout en restant a
  `-128`, ce qui illustre pourquoi les deux verdicts ne sont pas
  interchangeables.
- Garde-fous : les dix evitements et les deux reconnaissances deja acquis sont
  proteges individuellement, en plus des planchers agreges `>=10/13` et
  `>=2/13`. Une amelioration sur un cas ne peut donc pas masquer une regression
  sur un autre.
- Le rapport `search_trace` passe en schema 2 et publie separement
  `bad_move_avoided`, `historical_bad_move_selected` et `loss_recognized` a
  chaque profondeur. Le champ historique `root_move_best` reste disponible.
  Le rejeu complet confirme profondeur 4 : `5/13` evitements et `0/13`
  reconnaissances; profondeur 6 : `10/13` et `2/13`. Rapport
  `/tmp/equinox-search-trace-metrics-depth4-6-20260826.json` SHA-256
  `f92cc92990456290e431c7e89b29e8e03db75e466c09d264270d04ca73ce5116`;
  binaire SHA-256
  `3f74a83e903064ee59a4a3deae81eb0df66ba4333b7891e1dc4181e34816460d`.
  Aucun comportement du moteur n'est modifie par cette etape.

## Update 2026-08-25

- Candidat causal `QueenZone20` implemente puis teste isolement depuis la
  baseline `a26755e` : penalite de 20 cp par dame adverse distincte touchant
  geometriquement l'anneau du roi, une fois par dame, sous le gate strict
  `nonPawnMaterial > 2000`. Aucun fou, tour, multiplicite de cases, synergie,
  cap ou changement de recherche n'a ete ajoute.
- Semantique validee : rayons occupancy-aware, premiere case occupee incluse
  puis arret, pins ignores, roi absent, blockers des deux couleurs, deux dames,
  dame propre, multi-hit, symetrie des couleurs, signe/POV et integration
  exacte dans `Evaluate`. Tests cibles et race detector verts, `go vet ./...`
  vert, revue independante sans defaut fonctionnel.
- Canaris profondeur 6 : toujours exactement `2/13` passages (Game06 et
  Game21). Game09 reste exactement a `-128`, comme attendu : aucune dame et
  gate ferme a 1660 de materiel non-pion. Variations principales : Game02
  `-7 -> +17`, Game03 `-157 -> -175`, Game25 `+453 -> +436`, Game26
  `-36 -> -56`, Game27 `+73 -> +88`; aucun nouveau passage ni echec.
- Nuance mesuree avant A/B : quatre canaris ont une dame adverse sur l'anneau,
  mais les menaces des deux dames s'annulent statiquement sur Game02, Game06 et
  Game28. L'ajustement net vaut seulement `-20` sur Game25, `+20` sur Game30 et
  zero sur les onze autres. Sur le corpus large hors canaris, le danger dame
  net conserve toutefois un signal exploratoire positif : risque severe
  `+35.9` points, intervalle bootstrap 95% `[+13.7, +51.4]`.
- Premier gate interrompu et exclu : un forfait baseline a la partie 4 avait
  invalide le critere dur; l'interruption n'a pris effet qu'apres 30 parties
  completes. Ce fichier partiel, sans sidecar valide, n'est utilise dans aucun
  resultat ni agregat.
- Gate valide, 32 parties / 16 ouvertures appariees v2 : `12-6-14`, score
  `46.88%`, estimation `-21.7 Elo`, Hoeffding 95%
  `[-331.4, +250.0]`. Zero forfait, crash, coup illegal, disconnect, stall ou
  terminaison inconnue pour les deux moteurs. Le seuil prefixe de poursuite
  (`>= -40 Elo`) est franchi.
- Au gate, commentaires temps-seul/depth0 : candidat `15/1969` (`0.76%`),
  baseline `19/1970` (`0.96%`). Profondeur moyenne `7.31` contre `7.37`, temps
  moyen identique `93.74 ms`. Les hashes des binaires et entrees correspondent
  au manifeste.
- Validation longue autonome, memes binaires, 128 parties / 64 ouvertures v2 :
  `46-32-50`, score `48.44%`, estimation `-10.9 Elo`, Hoeffding 95%
  `[-135.3, +110.7]`. Premiere moitie `23-15-26` (`47.66%`), seconde moitie
  sur les 32 ouvertures absentes du gate `23-17-24` (`49.22%`) : les deux
  segments restent de signe negatif.
- Securite longue parfaite pour les deux moteurs. Les depth0 sont exactement
  symetriques : `44/9090` (`0.484%`) candidat contre `44/9098` (`0.484%`)
  baseline. Temps moyen `94.28` contre `94.25 ms`; profondeur moyenne `7.02`
  contre `7.14`.
- Couleurs : candidat blanc `31-15-18` (`60.16%`), candidat noir `15-17-32`
  (`36.72%`). La forte prime aux Blancs vient surtout du lot : la baseline
  marque respectivement `63.28%` avec les Blancs et `39.84%` avec les Noirs;
  la perte candidat est donc identique de `3.12` points dans chaque couleur.
- Binaires figes : baseline SHA-256
  `7a7c18bf755f3048c63621070d5268050607cf0e8818eec26815c19cb211532b`,
  candidat SHA-256
  `d1887ec08fc371a8575f0207eb600ab955bd53cd113d30a83da8fbe2ad98eed1`.
  PGN long SHA-256
  `b45010669eca5bb5c92d1f1c525cd6b6296793c8b34a61d23c08febca4bcab58`.
- Decision : **rejet de QueenZone20**, car le point long est negatif et le
  candidat n'apporte aucun benefice Elo ou operationnel compensatoire. Le patch
  moteur et ses tests ont ete retires; ne pas interpoler immediatement a 10 cp
  ni restaurer la combinaison multi-pieces du stash. La correlation historique
  identifie un symptome tactique, pas un terme statique causal valide.
- Instrumentation tactique ajoutee uniquement a l'analyseur offline : chaque
  reponse adverse immediate est classee capture, echec, echec calme,
  promotion, position deja en echec et eligibilite par la quiescence actuelle.
  Cette eligibilite est theorique : toutes les evasions si le camp adverse est
  deja en echec, sinon captures admises avant delta pruning; elle ne prouve pas
  que la recherche a atteint cette position a sa frontiere de quiescence, ni
  qu'une capture non promotionnelle n'y serait pas ensuite elaguee.
- Definition prefixee du danger dame net : `queen_attackers -
  placebo_queen_attackers > 0` en PRE, avec analyse primaire mi-partie,
  hors echec et hors canaris. Gate prefixe avant rejeu : au moins 20 swings
  severes / 10 parties, majorite stricte d'echecs calmes invisibles et exces
  d'au moins 15 points sur le controle severe sans danger dame net.
- Rejeu run6 inchange : 30 parties, 1530 lignes, 417 severes et 13/13 canaris.
  Dans le sous-ensemble primaire severe a danger dame net : 78 lignes / 11
  parties; 9 captures (`11.5%`), 21 echecs (`26.9%`) dont 19 echecs calmes
  invisibles (`24.4%`), zero promotion et 50 coups calmes sans echec (`64.1%`).
  La quiescence actuelle serait eligible pour 22/78 (`28.2%`) : 14 evasions
  et 8 captures hors echec; 56/78 (`71.8%`) sont filtres comme non-captures.
- Controle severe sans danger dame net : 109 lignes / 23 parties, 25 captures
  (`22.9%`), 12 echecs (`11.0%`) dont 9 echecs calmes invisibles (`8.3%`),
  zero promotion et 76 reponses invisibles (`69.7%`). Les echecs calmes sont
  enrichis de `+16.1` points avec danger dame net, mais restent loin d'une
  majorite; les 19 cas sont concentres dans 6 parties.
- Le gate tactique echoue donc uniquement sur le critere causal central de
  majorite (`24.4% <= 50%`) : **ne pas lancer une extension de quiescence aux
  echecs calmes sur ce seul signal**, et ne pas revenir a un terme statique.
  Nuance canaris, volontairement separee : 7 des 9 canaris severes recoivent un
  echec calme invisible (dont Game09), mais un seul a un danger dame net
  positif (Game25); cet echantillon selectionne ne renverse pas le gate large.
- Artefacts : CSV `/tmp/equinox-kingzone-tactical-run6-20260825.csv` SHA-256
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`;
  JSON `/tmp/equinox-kingzone-tactical-run6-20260825.json` SHA-256
  `fa9c5f489cfeedf0f7443beb12ae270494826cd2d155c0b1390554305dc178b7`;
  analyseur SHA-256
  `e4d8535ce6e1ae5131bd3511b0fffb9cf4ac4f932b3ca4d8d938db753320a6ea`.
  Les 14 tests unitaires Python, `compileall` et `git diff --check` passent;
  a cette etape, aucun fichier moteur ou recherche n'etait modifie.
- Traceur de recherche reel ajoute en mode diagnostique opt-in sous
  `internal/search/trace.go` et `cmd/search_trace/`. `Search.Trace == nil` par
  defaut; aucune formule d'evaluation, regle de quiescence, priorite, coupure ou
  reduction n'est changee. Les hooks sont compiles seulement avec le tag
  `searchtrace` et elimines des builds moteur ordinaires. Les evenements
  publient profondeur iterative, passe
  aspiration/recherche pleine, indices pseudo-legal/ordonne/legal, TT/PV/killer,
  futility, null move, beta cutoff, LMR/recherche pleine, terminal et qsearch.
- Corpus fige : les 19 echecs calmes invisibles primaires plus les 13 canaris,
  chacun rejoue a profondeur 4 et 6 avec une TT fraiche de 65536 entrees, soit
  64 recherches tracees. Chaque recherche est repetee sans trace : `64/64`
  donnent exactement le meme meilleur coup, score, profondeur terminee et
  nombre de noeuds. Le traceur est donc observationnel sur tout le corpus.
- Verification qsearch a profondeur iterative 1 : les 19 reponses primaires
  atteignent toutes le noeud de quiescence; 14 sont generees puis filtrees
  comme non-captures et 5 sont arretees avant parcours par stand-pat. Cette
  observation confirme la semantique du filtre, mais pas sa causalite aux
  profondeurs utiles.
- Passe decisive profondeur 4, primaire : les 19/19 reponses sont generees,
  legales, ordonnees et recherchees dans le Negamax principal. Ventilation
  exclusive : 8 recherchees seulement a profondeur pleine, 8 seulement avec
  LMR, 2 avec une visite LMR et une visite pleine separee, 1 avec LMR puis
  re-recherche pleine. Le coup historique reste meilleur dans 18/19 cas.
- Passe decisive profondeur 6, primaire : 7/19 reponses sont recherchees a
  profondeur pleine, 7/19 subissent une reduction LMR de 1 ply (`depth 5 ->
  child depth 3`) sans aucune re-recherche pleine, 3/19 noeuds coupent sur TT
  et 2/19 sur un autre coup avant la reponse cible. Zero futility prune et zero
  null-move cutoff dans la passe decisive. Le coup historique reste meilleur
  dans 10/19 cas : 6/7 lorsque la reponse est LMR-only, contre 4/7 lorsqu'elle
  est pleine et 0/5 dans les raccourcis TT/beta; echantillon trop petit et
  selectionne pour conclure causalement.
- Les trois TT cutoffs primaires exclusifs a profondeur 6 sont des bornes beta
  avec cle, halfmove clock et profondeur compatibles (`entry depth >= node
  depth`). Leur coup TT est different de l'echec calme cible : un autre coup
  adverse suffit deja a prouver la coupure. Les deux beta cutoffs restants sont
  egalement produits par un autre coup recherche avant la cible. TT et beta ne
  constituent donc pas une disparition fautive de la reponse.
- Priorites : la PV n'est appliquee qu'a la racine dans le moteur actuel. Sur
  les 19 cas primaires, le mauvais coup racine est prioritaire PV dans 15 cas a
  profondeur 4 puis 8 a profondeur 6; la TT le priorise respectivement dans 12
  puis 8 cas. Au noeud de reponse, la TT priorise la cible dans 5 cas et un
  killer dans 6 cas a profondeur 4, puis respectivement 4 et 7 cas a profondeur
  6; les echecs calmes tardifs restants sont les candidats a LMR.
- Canaris profondeur 6 : 10/13 reponses sont explicitement recherchees et 3/13
  servies par une TT valide; aucune n'est perdue. Pour les 7 canaris severes a
  echec calme : Game02 est reconnu comme terminal, Game06/Game09/Game21/Game26
  sont recherches a profondeur pleine, Game25 declenche LMR puis re-recherche
  pleine, et Game28 est le coup TT lui-meme. Game09 : `h2+` est ordonne en
  premier comme killer et recherche a profondeur pleine; son score reste
  `-128`, donc son echec n'est masque ni par qsearch ni par LMR.
- Distinction canari importante : le coup historique fautif n'est plus choisi
  dans 10/13 positions a profondeur 6, alors que le test existant n'en declare
  que 2/13 verts parce qu'il exige un score global `<= -300`. Il mesure donc la
  reconnaissance d'une position perdante, pas directement l'evitement du coup
  fautif. Game09 illustre l'ecart : `Ra7` est evite, mais le score `-128` laisse
  le test rouge. Conserver les deux indicateurs separement dans les prochains
  gates.
- Decision : **ne pas etendre la qsearch aux echecs calmes** sur ce diagnostic.
  La seule hypothese de recherche isolee encore compatible avec les traces est
  d'exempter de LMR les coups qui donnent echec, sans toucher TT, futility,
  qsearch ou evaluation. Les 7 cas primaires LMR-only fournissent un gate
  counterfactuel court; le signal est prometteur mais non causal et les canaris
  a echec calme ne le confirment pas directement.
- Artefacts recherche : binaire diagnostique
  `/tmp/equinox-search-trace-a26755e-20260825` SHA-256
  `1d1665abb617fdd9289791827220ca6b9610d06891845b619b2e6f06da02db12`;
  rapport `/tmp/equinox-search-trace-depth4-6-run6-20260825.json` SHA-256
  `17d89b7b04e5760fcb23cd2ff2bc37530fbf4318c67be1cf4397860d7779faea`.
  Tests cibles du traceur, tests du resume de passe decisive et `go vet` verts.
  `go test ./...` ne conserve que les 11 echecs anti-blunder historiques avec
  exactement les scores baseline; aucun nouvel echec. Le stash
  `wip king zone safety experiment` reste intact.

## Update 2026-08-24

- Instrumentation hors recherche ajoutee sous `tools/kingzone/`. Elle rejoue le
  PGN annote avec le POV du joueur, separe les scores finis des mats et mesure,
  avant/apres le coup Equinox, les fous, tours et dames ennemis touchant
  geometriquement l'anneau du roi. Aucun fichier moteur, recherche ou
  evaluation n'a ete modifie.
- Source analysee : `sf1600-run6-1+0.1-30.pgn`, SHA-256
  `1c98886a75d26560326cbfae249742af76ea7dfe813e3d8e3fd66fa74cf0d4dc`.
  Le replay couvre 30 parties, 3281 plies, 1530 decisions Equinox exploitables,
  417 swings severes et retrouve exactement les 13/13 canaris.
- Variable primaire : score Equinox au POV Equinox plus score de la reponse
  Stockfish au POV Stockfish. Un mat positif Stockfish ou un ecart fini d'au
  moins 300 cp est severe; aucun mat n'est transforme en cp dans l'analyse
  primaire. Les intervalles reposent sur 10000 bootstraps par blocs de parties,
  seed `20260824`.
- Sous le gate prefixe milieu de partie/hors echec (`nonPawnMaterial > 2000`),
  la dame adverse touche l'anneau sur 184/921 lignes dans 24/30 parties. Le
  taux severe vaut 51.09% expose contre 13.30% non expose, soit +37.79 points;
  l'intervalle bootstrap 95% est `[+20.00, +50.83]` points. Le swing fini
  median vaut 249 contre 109 cp, difference +140 cp, intervalle `[+9, +248]`.
- Le signal dame ne depend pas des cibles historiques : apres retrait des
  canaris, le risque reste a +37.14 points, intervalle `[+18.56, +50.77]`, et
  la difference mediane a +141.5 cp. Le placebo symetrique (dame Equinox vers
  le roi adverse) tombe a +4.70 points, intervalle `[-7.67, +20.65]`, et
  +38 cp de mediane.
- Couverture canaris : dame `4/13`, fou `1/13`, tour `2/13`. Le fou et la tour
  echouent donc le seuil prefixe de couverture; la tour a aussi une correlation
  globale de signe negatif. Ne pas combiner les trois types malgre le cas
  Game09.
- Golden Game09 confirme avant et apres `64.Ra7` : un fou adverse touche deux
  cases, une tour trois, aucune dame; total deux attaquants, cinq hits et quatre
  cases distinctes, hors echec. La pression est preexistante et inchangee par
  le coup : un futur terme dame ne corrigera pas directement ce canari.
- Decision : **seul candidat causal retenu pour l'experience suivante :
  `QueenZone20`**, penalite de 20 cp par dame adverse distincte touchant au
  moins une case de l'anneau, comptee une seule fois et sous le gate de phase
  existant. Les multiplicites de cases, fous, tours, synergies et cap du stash
  restent exclus. Le signal historique justifie un A/B isole, pas un gain Elo.
- Validation outil : 8 tests unitaires, compilation Python et `diff --check`
  verts. Rapport final SHA-256
  `972560ded3cfa1b7584964f1e37db0adff72a80cd614e64a09116e42606e0151`;
  CSV final SHA-256
  `77c0dfe103bcd92cf9e9dccf7b7da436a861102ef1c840fb627c4d3109f47dda`.

## Update 2026-08-21

- Le stash `wip king zone safety experiment` a ete audite sans etre depile. Son
  patch complet melangeait cavaliers, fous, tours et dames, multiplicite des
  cases attaquees, synergie multi-attaquants et cap `250 cp`; il n'a pas ete
  restaure tel quel.
- Hypothese minimale testee isolement : dans le gate de securite du roi deja
  existant (`nonPawnMaterial > 2000`), penalite de `20 cp` par cavalier ennemi
  touchant au moins une case de l'anneau du roi. Chaque cavalier ne compte
  qu'une fois; aucune attaque glissante, occupation, recherche ou gestion du
  temps n'a ete modifiee.
- Canaris valides : `0/1/2` attaquants, multi-hit compte une fois, couleur
  adverse, roi absent, symetrie des couleurs et signe selon le trait. Le score
  reste celui de la baseline sous le gate de phase et varie exactement de
  `20 cp` au-dessus. Tests cibles, race detector, `go vet ./...` et revue
  independante verts.
- La suite complete conserve exactement les `11/13` echecs anti-blunder
  historiques, donc `2/13` positions passent toujours. Les variations sont
  mixtes : Game02 et Game25 s'ameliorent de `20 cp`, Game17 recule de `10 cp`,
  Game27 s'ameliore de `2 cp`; Game09 reste a `-128`.
- Gate `32` parties / `16` ouvertures a `1+0.1` : `14-8-10`, score `56.25%`,
  estimation `+43.7 Elo`, intervalle Hoeffding 95% `[-216.9, +385.6] Elo`,
  zero incident. Ce signal court a satisfait le critere de poursuite.
- Validation longue du meme binaire, `128` parties / `64` ouvertures :
  `46-28-54`, score `46.88%`, estimation `-21.7 Elo`, intervalle Hoeffding 95%
  `[-148.0, +98.8] Elo`. Le candidat fait `24-14-26` avec les Blancs et
  `22-14-28` avec les Noirs; zero forfait, crash, coup illegal, disconnect,
  stall ou terminaison inconnue pour les deux moteurs.
- Les commentaires temps-seul restent rares mais un peu plus frequents pour le
  candidat : `32/8424` (`0.38%`) contre `25/8427` (`0.30%`). Profondeur moyenne
  annotee `7.27` contre `7.20`, temps moyen identique `93.8 ms`.
- Decision : **rejet de KnightZone20 pour absence de gain long**. Le point
  estime ne depasse pas la baseline, meme si aucune regression statistique ou
  pratique n'est demontree. Le patch moteur et ses tests ont ete retires; le
  stash original reste intact. Ne pas interpoler immediatement a `10 cp`.
- Prochaine etape recommandee : instrumenter hors recherche la distribution
  des menaces glissantes (fou/tour/dame) sur l'anneau du roi et leur correlation
  avec les swings des canaris/PGN, avant de choisir un unique terme d'evaluation
  causal. Ne pas restaurer le terme multi-pieces du stash sans cette mesure.

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
- Changement fige dans le commit `182aae5 Add UCI move overhead margin`.
  Binaire propre reconstruit depuis ce commit, SHA-256
  `ae1d10278a8ec18a8fd42d7a5bd8ec5f08cc67fa74024084f215f6f5fd3bc5ba`.
- Calibration contre Stockfish 18 limite a `1800 Elo`, `128` parties /
  `64` ouvertures : `59-7-62`, score `48.83%`, estimation `-8.1 Elo`,
  intervalle Hoeffding 95% `[-132.1, +113.7] Elo`. Equinox ne subit aucune
  anomalie; Stockfish abandonne une partie au temps.
- Sensibilite sans ce forfait Stockfish : `58-7-62` sur `127`, score `48.43%`,
  estimation `-10.9 Elo`, soit une calibration nominale autour de `1789`.
  Equinox tient donc approximativement SF18-1800 a ce controle, mais
  l'objectif `1800+` n'est ni prouve ni exclu par cet echantillon.
- Recherche temporisee : `75/6013` coups Equinox (`1.25%`) n'ont qu'une
  annotation temps, compatibles avec un retour `bestDepth=0`. La profondeur
  moyenne annotee est `5.91`; zero forfait Equinox.
- Hypothese de fallback racine TT testee sans modifier le binaire accepte :
  instrumentation passive sur `32` parties / `16` ouvertures a `1+0.1`, avec
  capture de la case TT racine avant la recherche puis journalisation des seuls
  retours `bestDepth=0`.
- Couverture nulle : `16/16` retours `bestDepth=0` etaient non terminaux avec
  des coups legaux, mais aucun n'avait une entree TT de meme cle, donc
  `0/16` coup TT legal et `0/16` fallback qui aurait change. Le PV d'une
  profondeur completee est reste separe du diagnostic et vaut necessairement
  zero sur cet echantillon `bestDepth=0`.
- Decision : **rejet avant gate de force**. Le patch candidat, pourtant valide
  par tests cibles, race detector, vet et revue independante, a ete retire car
  l'hypothese n'est pas exercee dans le regime mesure. Aucun A/B Elo n'est
  interpretable ni lance; le code moteur reste identique au commit accepte.
- Prochaine piste : reprendre l'experience de pression de zone du roi depuis
  un diff audite du stash, en isolant d'abord le plus petit terme d'evaluation
  et ses positions canaris avant tout gate Elo.

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

`TestSearchTacticalCanaryMetrics` dans [internal/search/search_anti_blunder_test.go](internal/search/search_anti_blunder_test.go).

- `13` FEN extraites du run6 : `7` balanced_blunder + `2` unclassified + `4` conversion_blunder courtes
- Metrique evitement : le meilleur coup profondeur 6 doit differer du coup
  historique fautif. **Etat : `10/13`**, avec chaque succes protege.
- Metrique reconnaissance : le score profondeur 6 doit rester `<= -300 cp` ou
  mate-against. **Etat : `2/13`** (Game06, Game21), avec les deux succes
  proteges. Sur certains cas non reconnus, le score reste positif : G16 `+288`,
  G25 `+453`, G28 `+120`.
- Un gate de recherche doit conserver au minimum `10/13` evitements et `2/13`
  reconnaissances, sans nouvelle selection d'un coup historique fautif. Ces
  planchers de securite ne constituent pas a eux seuls une preuve de gain Elo.

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

Tester isolement un fallback racine base sur le coup TT lorsqu'aucune
profondeur n'est completee. Si la TT contient pour la cle racine un coup non
nul present dans la liste legale, l'utiliser a la place du premier coup legal;
sinon conserver exactement le fallback actuel. Ne modifier ni cutoff, ni move
ordering, ni pruning, ni budget temps. Ajouter des tests pour un coup TT legal,
un coup stale/illegal et l'absence de TT, puis mesurer la couverture sur les
cas `bestDepth=0` avant un A/B v2. Si le hit TT est rare ou si l'A/B est
neutre/negatif, retirer le patch et reprendre la piste king-zone attack
pressure.
