# Journal des expériences · avril à septembre 2026

Les entrées sont classées de la plus récente à la plus ancienne. Chaque décision et chaque prochaine étape reflètent la date de leur entrée. Pour reprendre le projet, lire le [bilan actuel](../../status.md).

Les révisions citées renvoient à l’historique actuel. Les noms de fichiers temporaires et les empreintes identifient les artefacts de l’époque. [Correspondance des révisions](../revisions-git.tsv)

## Update 2026-09-09 — D5ReducedFullShadow

- Le traceur passe en schema 11 et ajoute `-probe-d5-reduced-full`, une sonde
  strictement observationnelle et sans allowlist. Pour chaque echec calme
  reellement eligible LMR a profondeur restante 5 dans l'iteration finale,
  elle part du meme plateau, de la meme TT et des memes killers/historiques,
  puis compare trois vues isolees : recherche reduite `child 3` a fenetre
  nulle, recherche full-depth fraiche `child 4`, et sequence normale reduite
  puis re-recherche pleine seulement si le score depasse alpha. Elle publie
  scores/PV, franchissements d'alpha, TT/ordre/heuristiques et chaines D3→D1.
- Securite validee : les 32 traces ont ete reproduites avec recherche et TT
  fraiches, avec egalite exacte de leurs 107 enregistrements shadow. Les 32
  resultats principaux egalent aussi bit-a-bit une troisieme recherche non
  tracee (coup, score, profondeur et noeuds). Les gates restent donc primaires
  `9/19` evitements et `7/19` pertes reconnues, canaris `10/13` et `2/13`.
  Game25_ply35 reste sain avec `d1e2`, `+146`, `155 300` noeuds; Game09 reste
  `d6c5`, `-128`, `16 772` noeuds.
- Portee : `107` D5 uniques sur 23 cas. Les recherches reduites depassent alpha
  6 fois, les full fraiches 11 fois; 14 scores different et 5 franchissements
  d'alpha divergent. La sonde trouve 229 D3 et 385 chaines D3→D1 sur 17 cas.
  Ses `74 894` noeuds representent `0.848%` des `8 831 770` noeuds principaux
  et ne sont jamais reinjectes dans la recherche reelle.
- TT/ordre : 103/107 positions enfant avaient deja une entree TT, dont 13
  coupaient a profondeur reduite mais aucune a profondeur pleine. Reduced et
  full laissent une entree TT enfant differente dans 107/107 cas et different
  sur 2 068 slots TT au total; la full fraiche ecrit davantage (`3 541` slots
  contre `2 159`). Pourtant leur premier coup ordonne reste identique dans
  107/107 cas. Les killers divergent dans 26 cas et l'historique dans 17 :
  l'effet d'etat existe donc surtout plus bas dans les sous-arbres, pas sur le
  premier coup enfant.
- La branche regressive est maintenant expliquee localement. Sur
  `Qe2 Qh1+`, alpha vaut `-146`; le reduced rend exactement `-146` et ne
  declenche donc pas de re-recherche, tandis que le full frais rend `-145` et
  franchit alpha de 1 cp. La sequence normale reste a `-146`. La voie de
  substitution `Qe2 Ng3+` vaut `-146` dans les deux profondeurs. Les deux ont
  des chaines D3→D1 (respectivement 5 et 8).
- Aucun predicat structurel D5 propre n'emerge. Parmi les six pre-requis
  positifs, cinq ont des scores reduced/full identiques; seul
  Game30_ply52 `Rf5 Qh4+` franchit alpha (`-19 -> -11`). Ce dernier ressemble
  donc a la regression `Qe2 Qh1+`, tandis que le positif Game25_ply69
  `a4 Qc7+` et la regression substituable `Qe2 Ng3+` ont tous deux un score
  inchange. Les six positifs comme les deux regressifs possedent au moins une
  chaine D3→D1; les deltas TT, killers et historique se recouvrent aussi.
- Verdict : **progres d'attribution, mais D5D3D1 reste rejete**. La sonde
  falsifie l'idee qu'un seuil de delta, le seul franchissement d'alpha, la
  presence d'une chaine D3→D1 ou un changement d'ordre local puisse separer
  les gains de Game25_ply35. Aucun nouveau predicat D5 ne doit etre formule a
  partir de ces observations.
- Prochaine experience minimale recommandee : `D5D3D1ScoreOnlyShadow`. Garder
  D3→D1 general, calculer chaque D5 full-depth dans une copie complete de la
  TT et des heuristiques, reinjecter uniquement son score puis jeter tout son
  etat. Cette unique ablation dira si les six gains et la regression passent
  par la valeur retournee ou par la contamination TT/killers/history, sans
  allowlist et sans modifier evaluation, profondeur, qsearch ou predicats.
- Artefacts : corpus SHA-256
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`;
  binaire schema 11
  `53462bf4c8ad14abd42a99f048db5effcc91e577628f4d9d8310a3552e3d8532`;
  rapport
  `9bdc9181c7c08a985346aa2b2798db2914a2c703a109f170dbfb570033870dd4`.

## Update 2026-09-09 — D5BranchAblation

- Le traceur passe en schema 10. `-d5-ablation-case` et
  `-d5-ablation-line` retirent du candidat D5D3D1 une seule exemption D5 sur
  un chemin exact racine/reponse. La ligne doit contenir deux coups legaux et
  finir par un echec calme. Le coup ablate conserve le LMR normal
  `depth 5 -> child 3`, sa recherche nulle et son eventuelle re-recherche
  pleine; tous les autres D5 et tout le couplage D3→D1 restent actifs. Un
  evenement `d5_check_ablation_lmr_preserved` prouve l'application exacte.
- La trace candidate de Game25_ply35 contenait quatre D5 : `Bb4 Qh1+`,
  `Bb4 Ng3+`, `Bb4 Nd2+` et `Qe2 Qh1+`. Chacun a ete ablate seul sur un rejeu
  complet des 32 cas. Les quatre variantes conservent pourtant le mauvais
  `d2b4`, score `+135`; `d1e2` reste seulement une borne haute `+135` et la
  marge du coup historique reste une borne basse nulle. L'hypothese d'une
  exemption D5 unique est donc falsifiee.
- Details Game25_ply35 : candidat intact `133 665` noeuds; ablation
  `Qe2 Qh1+` `147 082` (`+10.04%`), `Bb4 Qh1+` `131 803` (`-1.39%`),
  `Bb4 Ng3+` `133 612` (`-0.04%`) et `Bb4 Nd2+` `133 642` (`-0.02%`). Chaque
  rapport contient exactement une ablation. Les 31 autres cas reproduisent
  exactement coup, score et noeuds du candidat intact; les 32 doubles rejeux
  de chaque variante sont deterministes. Les gates restent donc primaires
  `11/19` evitements et `8/19` pertes reconnues, canaris `10/13` et `2/13`.
- Le controle factoriel localise l'interaction. Ablater les quatre chemins
  initiaux ensemble ne suffit pas : Game25 conserve `d2b4`, `+135`,
  `150 561` noeuds, car la recherche modifiee expose alors un cinquieme D5,
  `Qe2 Ng3+`, absent de la trace candidate initiale. L'exploration des 32
  sous-ensembles des cinq chemins restaure `d1e2` si et seulement si les deux
  alternatives sous le concurrent sont retirees : `Qe2 Qh1+` et
  `Qe2 Ng3+`. Cette paire seule donne `d1e2`, `+145`, `145 619` noeuds; le
  controle D3→D1 sans aucun D5 donne le meme coup/score (`148 952` noeuds).
- Verdict : **la regression n'est causee par aucun des quatre D5 pris
  isolement, mais par deux voies D5 substituables sous la branche concurrente
  `Qe2`**. Retirer `Qh1+` fait apparaitre `Ng3+`; les trois exemptions sous le
  mauvais `Bb4` ne determinent pas le choix. D5D3D1 reste rejete pour promotion
  moteur et match Elo, et aucune allowlist issue de cette ablation ne doit
  devenir une regle de production.
- Prochaine experience minimale recommandee : `D5ReducedFullShadow`. Pour
  chaque echec calme D5 du corpus, enregistrer dans une copie isolee le score
  reduit, le score full-depth, le franchissement d'alpha, les effets TT/ordre
  et la presence d'une chaine D3→D1 en aval. Comparer ensuite les six
  pre-requis positifs aux deux voies regressives `Qe2 Qh1+/Ng3+` avant de
  formuler un unique predicat structurel D5.
- Artefacts : corpus SHA-256
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`;
  binaire schema 10
  `453a29ee70f7436d58c962dce736dd652091c6e5863cf9645ac0c85c2dd9600f`;
  candidat intact
  `63890845a50c7ee44bc57862416fc4ea3a3041ab4840d5e4605ce8ff6df92ea4`;
  ablation `Qe2 Qh1+`
  `5638446c4cd43ad2db02df1dee282d175a460b9887efe4b18d9913dfa2010253`;
  ablations `Bb4 Qh1+`, `Bb4 Ng3+`, `Bb4 Nd2+`
  `d61ea4a2a34402b7bbb9d9984e42e3a85f6ce22228867a8780d6dcae83badd8f`,
  `e2de950cfabd9aeeee7b86d94887ff0430c8bfd2be29cb40b3b6821c5ce3d180`,
  `3eae3e5f22fd27c37002e22f7cd32244fbe151d4447edc3f9c2918feef9a3c94`.

## Update 2026-09-04 — D5D3D1CheckChain sans allowlist

- Le traceur passe en schema 9 et ajoute `-general-d5-d3-d1-checks`. Le mode
  refuse explicitement toute allowlist, tout bridge/probe et l'ancien candidat.
  A l'iteration finale, tout echec calme normalement eligible LMR a profondeur
  restante 5 ou 3 est recherche a sa profondeur normale (`child 4` ou
  `child 2`). L'exemption futility D1 reste strictement limitee aux echecs
  calmes legaux situes deux plies sous un D3 effectivement force par le meme
  candidat. Aucun ply, TT, qsearch ou terme d'evaluation n'est modifie.
- Le cas causal Game25_ply69 passe sans pre-requis code par le corpus : `Qc7+`
  est force full-depth a D5, `Qc3+` l'est a D3 et toutes les exemptions D1
  observees ont un parent D3 exact. Le resultat reste `c4d3`, `-741`, a
  profondeur 6 (`78 659` noeuds). Le test unitaire fige le resultat et verifie
  iteration, profondeurs et couplage de chaque evenement D1.
- Portee mesuree sur les 32 cas : `127` applications D5 (`127` chemins uniques)
  sur 23 cas, `1 081` D3 (`1 006` uniques) sur 28 cas et `2 092` D1 (`1 972`
  uniques) sur 25 cas. Les 32 rejeux candidats sont strictement deterministes
  avec recherches et TT fraiches; les 32 baselines tracees egalent exactement
  leurs recherches non tracees.
- Face a la recherche normale, les primaires progressent de `9/19` a `11/19`
  mauvais coups evites et de `7/19` a `8/19` pertes reconnues. Les gains
  d'evitement sont Game07_ply177 (`g1h2 -> g6f5`), Game25_ply69
  (`a3a4 -> c4d3`) et Game30_ply52 (`f8f5 -> c5d6`); Game17_ply103 reconnait
  maintenant la perte (`-130 -> -582`). Les canaris restent exactement a
  `10/13` evitements et `2/13` reconnaissances. Game09 reste `d6c5`, `-128`.
- Un gate de non-regression primaire echoue : Game25_ply35 evitait `d2b4` avec
  `d1e2`, `+146`, dans la baseline et `d1e2`, `+145`, avec le precedent
  candidat. La regle D5 sans allowlist recherche aussi `Qh1+` apres le
  concurrent `d1e2`; celui-ci ne depasse plus `+135` et le moteur reselectionne
  le coup historique `d2b4`, `+135`. Le nouveau candidat reproduit donc 6 des
  7 anciens pre-requis cibles, mais passe de `12/19` a `11/19` evitements face
  au candidat precedent; la reconnaissance reste `8/19`.
- Cout face a la recherche normale : `8 831 770 -> 8 991 901` noeuds,
  `+1.813%`; mediane par cas `+0.751%`. Trois queues depassent `+50%` :
  Game17_ply103 `+68.41%`, Game25_ply59 `+55.16%` et Game25_ply75 `+86.36%`.
  Face au precedent candidat D3→D1 avec 7 pre-requis, le total baisse de
  `0.072%`, mais cette quasi-egalite agregee ne compense pas la regression de
  choix sur Game25_ply35.
- Verdict : **signal structurel et tactique confirme, mais candidat rejete pour
  promotion moteur et match Elo a cause d'une regression primaire**. La
  prochaine experience minimale recommandee est `D5BranchAblation` : conserver
  D3→D1 global, puis desactiver un par un les quatre D5 exacts observes dans
  Game25_ply35, en commencant par `d1e2 Qh1+`, afin de prouver quel chemin fait
  basculer le choix. Une nouvelle regle D5 ne devra etre formulee que si un
  predicat structurel separe ce chemin regressif des six pre-requis positifs;
  aucune nouvelle allowlist ne sera candidate a la production.
- Artefacts : corpus SHA-256
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`;
  binaire schema 9
  `e66a1a9cc40b91e73f952f3571b3292b32c0a438ba261e03943a609067c4b639`;
  baseline normale
  `d0af24eff28e96959cd3c7a7f1e76f0212bade099cec2c48631c0380f2b45d91`;
  precedent candidat remesure avec le meme binaire
  `4400b741b3ce23dad549942e31105ceb30953e736117f32335dca566ae5213ee`;
  candidat D5D3D1
  `e985b67e3f6f3a55d851be640a51840c152c7d14f3a85e031e501758a9ab8439`.

## Update 2026-09-04 — D3→D1 structural probe and generalized candidate

- Le traceur passe en schema 8. Pour chaque echec calme reellement eligible LMR
  a profondeur restante 3 dans l'iteration finale, une recherche full-depth
  est executee sur des copies du plateau, de la TT et des heuristiques. Les
  noeuds, scores, TT, killers et historiques shadow ne reviennent jamais dans
  la recherche principale. Deux plies plus bas, la sonde enregistre uniquement
  les echecs calmes legaux que le predicat futility profondeur 1 elimine.
- Un premier controle purement observationnel a revele un motif tres frequent,
  mais ne retrouvait pas `Qc3+ -> b3+` : cette branche exige d'abord le
  pre-requis causal `Qc7+` full-depth. Ce lot preliminaire a donc ete exclu du
  choix du candidat. Le protocole final conserve les 7 paires RootMove/ReplyMove
  LMR deja figees en full-depth et sonde ensuite les 32 cas.
- Instrumentation finale : `830` probes D3 (`772` chemins uniques) sur 28 cas;
  `1 770` chaines D3→D1 (`1 626` uniques) sur 25 cas. Le motif existe donc tres
  largement au-dela de Game25. Les shadows totalisent `22 075` noeuds contre
  `8 838 032` noeuds principaux (`0.250%`). Les 25 recherches non ciblees sont
  strictement non interferentes; les 7 recherches ciblees et leurs shadows
  sont identiques sur deux recherches/TT fraiches.
- Le cas temoin est retrouve exactement. Apres `a4 Qc7+ Kb3`, `Qc3+` est
  baseline-LMR `child 1` et son shadow full-depth `child 2` remonte `+644` en
  7 noeuds. Six echecs calmes D1 satisfont la futility dans ce shadow, dont
  `b3+` apres `Ka2`, avec `staticEval 480`, alpha `644` et la ligne exacte
  `a4 Qc7+ Kb3 Qc3+ Ka2 b3+`.
- La condition de generalisation etant satisfaite, le candidat couple
  `D3CheckNoLMR+D1CheckNoFutility` a ete mesure a l'iteration finale. Toute
  exemption D1 reste limitee a deux plies sous un echec D3 que le meme candidat
  vient de passer full-depth. Il applique `870` NoLMR D3 (`795` chemins uniques)
  et `1 862` NoFutility D1 (`1 742` uniques), sur respectivement 28 et 24 cas.
- Resultat fonctionnel positif : les primaires passent de `9/19` a `12/19`
  mauvais coups evites et de `7/19` a `8/19` pertes reconnues. Trois mauvais
  choix disparaissent sans regression d'evitement : Game07_ply177
  (`g1h2 -> g6f5`, `-302 -> -326`), Game25_ply69
  (`a3a4 -> c4d3`, `-644 -> -741`) et Game30_ply52
  (`f8f5 -> c5d6`, `+11 -> +18`). Game25_ply75 change aussi de coup, mais
  evitait deja son coup historique.
- Les canaris restent exactement `10/13` evitements et `2/13` reconnaissances.
  Game09 conserve `d6c5`, `-128`; Game06 conserve `d1d4` mais voit maintenant
  le mat (`-603 -> -31994`). Les 32 rejeux candidats sont deterministes avec
  recherches et TT independantes.
- Cout : `8 838 032 -> 8 998 390` noeuds, soit `+1.814%`; mediane par cas
  `+0.420%`. Trois queues depassent `+50%` : Game17_ply103 `+68.41%`,
  Game25_ply59 `+54.92%` et Game25_ply75 `+85.10%`. Game25_ply69 vaut
  `+43.14%`, tandis que Game07_ply177 baisse de `12.30%`.
- Verdict : **signal general positif au gate profondeur 6, mais pas encore de
  promotion moteur ni de match Elo**. Le candidat imbrique reste precede de 7
  exemptions RootMove/ReplyMove codees par le corpus; applique seul dans le
  moteur normal, il ne garantirait donc pas l'acces a `Qc3+` sur Game25. La
  prochaine experience minimale est `D5D3D1CheckChain` : generaliser aussi le
  pre-requis externe aux echecs calmes LMR a depth 5, conserver le couplage D3
  puis D1, supprimer toute allowlist de cas et remesurer les memes gates avant
  d'envisager un A/B Elo.
- Artefacts : corpus SHA-256
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`;
  binaire sonde ciblee
  `4c82b16e800e80ae0fd643df3f52e79eb4dff383d5a55f0800a956e264c04daa`;
  rapport sonde ciblee
  `1866096f199f79a3bdb5eb965cbce5e57b632b9a5c719440512f0e4f045cb6fa`;
  binaire candidat general
  `9526e617852ddfa5cde352276a7040fb418d9a49da2d8c7316b558daeeb64d4c`;
  rapport baseline candidat
  `6f4b47da6658107b09f7b5b8541d576121d110b01cb2a96457e6f70cda701e65`;
  rapport candidat
  `0814f12ccc9718d8a3002a04a2fd740cdc1caeb5adf0d1e1f96328fd62c32c5c`.

## Update 2026-09-03 — TargetNestedCheckNoLMR+BridgeNoFutility

- Le schema `searchtrace` passe a 7 avec un second gate exact. La baseline
  conserve `Qc3+` sans LMR sur `a4 Qc7+ Kb3`; le candidat ajoute uniquement
  l'exemption de futility de `b3+` apres le prefixe complet
  `a4 Qc7+ Kb3 Qc3+ Ka2`. Le coup final doit etre celui de la ligne validee,
  legal, calme et donnant echec; aucune profondeur ou fenetre n'est changee.
- Dans la passe aspiration, le predicat normal de futility est bien actif a
  profondeur 1 (`staticEval 480 + marge 150 <= alpha 644`). Le gate s'applique
  une fois, puis `b3+` devient legal et est recherche a profondeur enfant 0.
  La qsearch en echec voit `cxb3`, puis la capture `Qxb3#`; la feuille propagee
  est `qsearch_checkmate` a `-31992`.
- Ce nouveau score fait echouer la fenetre aspiration et declenche la recherche
  pleine. Dans cette seconde passe, `Qc3+` est encore exactement exempte de LMR
  mais `b3+` ne requiert plus d'exception : avec alpha `498`, le predicat de
  futility est faux et le coup est recherche normalement. Il n'y a donc qu'une
  exemption de futility effective, et deux applications NoLMR sur les deux
  passes de la seule iteration 6.
- Le gate causal passe : Game25_ply69 abandonne le mauvais `a4` (`-644`) pour
  `d1d3` (`-728`). Dans la passe decisive, la branche historique `a4` remonte
  au plus `-754`; sa marge face a `d1d3` est donc au plus `-26 cp`. La ligne
  tactique obtenue est exactement
  `a4 Qc7+ Kb3 Qc3+ Ka2 b3+ cxb3 Qxb3#`, sans ply supplementaire.
- Les primaires progressent de `9/19` a `10/19` mauvais coups evites; les pertes
  reconnues restent `7/19`. Les canaris restent strictement `10/13` evitements
  et `2/13` reconnaissances; Game09 reste `d6c5`, `-128`, `16 772` noeuds.
  Les 31 cas non cibles reproduisent exactement leur recherche non tracee et le
  cas cible est deterministe sur deux recherches avec TT fraiches.
- Cout mesure : `54 331 -> 91 546` noeuds sur Game25 (`+68.50%`) et
  `8 831 785 -> 8 869 000` sur les 32 cas (`+0.421%`), mediane globale nulle.
  La hausse locale comprend la recherche pleine rendue necessaire par la
  correction du score, mais reste trop concentree pour extrapoler un cout
  global depuis ce seul cas.
- Verdict : **validation causale positive du couple exact LMR/futility, sans
  promotion moteur**. Le chemin code en dur n'est pas une regle generalisable
  et aucun match Elo n'est justifie. La prochaine etape recommandee est
  d'instrumenter, sur tout le corpus, les paires structurelles « echec calme
  reduit par LMR a depth 3 » puis « echec calme elague par futility a depth 1 ».
  Un candidat general `D3CheckNoLMR+D1CheckNoFutility` ne devra etre formule que
  si le motif existe au-dela de cette position unique.
- Artefacts : corpus SHA-256
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`;
  binaire schema 7
  `1e56c75614b78065fd210609147e81403dc00c6aaff446797f98afe660644136`;
  rapport baseline NoLMR
  `0544c4a197ed4ef4dcbd491201cea05376a92c6bdeecf536b5d3f44a66f9e5cb`;
  rapport candidat combine
  `316d5928fbae4aace866cbe4ff6e37c9676fc1cb6ae96fe6725e36cc8675aee1`.

## Update 2026-09-03 — TargetNestedCheckNoLMR

- Le contre-factuel `TargetNestedCheckNoLMR` a ete ajoute uniquement au
  binaire `searchtrace` schema 6. Il exige la ligne exacte
  `a4 Qc7+ Kb3 Qc3+ Ka2 b3+`, le coup intermediaire unique `Qc3+`, sa legalite
  et son caractere calme/echec, ainsi que l'iteration finale demandee. Aucun
  changement n'atteint le moteur normal.
- Mecanisme valide sans fuite : dans Game25_ply69, passe decisive profondeur 6,
  `Qc3+` passe une seule fois de LMR `depth 3 -> child depth 1` a la profondeur
  normale `child depth 2`, sans extension. Les iterations precedentes, les 31
  autres cas, les autres coups et les autres chemins restent inchanges.
- Le nouveau ply traverse l'ancienne coupure TT sur `Ka2`, mais revele le
  verrou suivant : `b3+` est maintenant genere et ordonne en index 3 a
  profondeur 1, puis elague par futility avant legalite/recherche. La valeur
  statique publiee est `+480`, sous alpha `+644`; `Qc3+` remonte donc seulement
  a alpha `+644`.
- Le gate d'efficacite echoue : Game25_ply69 conserve `a4`, score `-644`, et
  sa marge racine ne s'ameliore pas. Les primaires restent a `9/19` mauvais
  coups evites et `7/19` pertes reconnues; aucun des 32 meilleurs coups ou
  scores ne change.
- Securite exacte : les canaris restent `10/13` evitements et `2/13`
  reconnaissances. Game09 reste `d6c5`, `-128`, `16 772` noeuds. Les 31 cas
  non cibles egalent leur recherche non tracee et le cas cible est identique
  sur deux recherches/TT fraiches.
- Cout negligeable mais sans benefice : `8 831 770 -> 8 831 785` noeuds sur les
  32 cas, soit `+15` (`+0.00017%`); Game25 seul fait
  `54 316 -> 54 331` (`+0.0276%`), mediane globale nulle.
- Verdict : **rejet de TargetNestedCheckNoLMR seul**. Il isole neanmoins une
  chaine causale sequentielle LMR/TT puis futility. La prochaine experience
  minimale est `TargetNestedCheckNoLMR+BridgeNoFutility` : conserver seulement
  l'exemption exacte de `Qc3+`, puis exempter uniquement `b3+` de la futility
  profondeur 1 sur le meme prefixe, sans ply supplementaire ni changement TT,
  qsearch ou regle globale.
- Artefacts : corpus SHA-256
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`;
  binaire schema 6
  `bdc1fdf83f93a310d0cb44d56e71ee6df6a21d27fed5a5eef9afa7e4faff9f94`;
  rapport baseline
  `5b6877da2f3c87d9e9d64924df23215a5d631e686b99ae5e6e6c195b4dd312c8`;
  rapport candidat
  `20d4191cb2d4e3d74b3ffe9fc921055b9200c200d644e7e99aa98e2f5200bbb2`.

## Update 2026-08-31 — TargetBridgeNoFutility

- Le lot TargetReply+1 a ete fige dans le commit `90ea1da` avant cette nouvelle
  experience. Aucun push n'a ete effectue et le stash historique reste intact.
- Le schema `searchtrace` passe a 5 et accepte une ligne UCI exacte. Chaque
  etape apres la paire racine/reponse est tracee seulement si tout le prefixe
  precedent correspond. La ligne Game25_ply69 validee est
  `a4 Qc7+ Kb3 Qc3+ Ka2 b3+`; `b3+` est verifie legal, calme et donnant echec.
- Le pre-gate `TargetBridgeNoFutility` echoue : dans la passe decisive
  profondeur 6, la branche n'atteint jamais `b3+`, donc aucune futility sur ce
  coup ne peut etre la cause et aucune exemption n'a ete implementee.
- Chemin exact dans la passe decisive : `Kb3` est TT/killer, premier coup et
  full-depth. `Qc3+` est genere et legal mais seulement cinquieme coup ordonne;
  il subit LMR `depth 3 -> child depth 1` avec reduction 1. Au noeud `Ka2`, la
  recherche reduite rencontre une entree TT `exact`, cle/halfmove valides,
  profondeur `1`, score `-480`, puis coupe immediatement. `Qc3+` remonte a
  `+480`, sous alpha `+644`, et ne declenche donc aucune recherche pleine.
- A l'iteration 5 seulement, la meme ligne atteint la qsearch hors echec avant
  `b3+`; le coup est genere en index 0 puis filtre comme non-capture. Cette
  observation est reelle mais non decisive et ne justifie pas d'elargir la
  qsearch.
- La recherche reste exactement la baseline full-depth cible : Game25_ply69
  choisit `a4`, score `-644`, `54 316` noeuds. Les 31 cas non cibles egalent
  leur recherche non tracee; les canaris restent `10/13` evitements et `2/13`
  reconnaissances. Game09 reste `d6c5`, `-128`, `16 772` noeuds.
- Verdict : **rejet de TargetBridgeNoFutility avant contre-factuel**. Le point
  causal suivant est le LMR imbrique de `Qc3+` combine a la coupure TT peu
  profonde. La prochaine experience minimale est `TargetNestedCheckNoLMR` :
  exempter uniquement `Qc3+` de LMR sur ce prefixe exact, sans modifier TT,
  qsearch, futility ni profondeur globale.
- Artefacts : binaire observationnel SHA-256
  `a7872e0f2df59c5db998e39ba0617c4f7b0abefc6328814f29c53f89c131f53a`;
  rapport schema 5 SHA-256
  `8e7605a8454f93fb975627c7e4883992821ee8335470cf74b75c34324e194141`;
  corpus regenere identique au SHA-256 historique
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`.

## Update 2026-08-28 — TargetReply+1

- Le diagnostic `searchtrace` passe en schema 4. Les contre-factuels utilisent
  desormais une allowlist de cas exacte et publient, pour tous les coups racine
  de la passe decisive, score, fenetre alpha-beta et nature de la valeur
  (`exact`, borne haute ou borne basse). La marge publiee compare ainsi le
  mauvais coup historique a son meilleur concurrent sans convertir les
  fail-low en faux scores exacts.
- `TargetReply+1` a ete limite aux 7 paires precedemment identifiees LMR-only,
  a l'iteration finale de profondeur 6. Le mecanisme passe bien les 7/7
  reponses de profondeur enfant normale 4 a 5; aucune extension ne fuit dans
  les iterations 1 a 5. Les 7 doubles rejeux contre-factuels sont identiques et
  les 25 autres cas reproduisent exactement la recherche non tracee.
- Les canaris restent strictement inchanges : `10/13` mauvais coups evites et
  `2/13` pertes reconnues. Game09 garde `d6c5`, score `-128` et `16 772`
  noeuds. Sur les 19 cas primaires, l'evitement progresse de `9/19` a `11/19`;
  la reconnaissance de perte reste `7/19`.
- Signal causal fort sur Game25_ply69 : le score racine de la reponse cible
  passe de `-644` a `-754 cp`; `a3a4` est remplace par `d1d3` et sa marge
  devient au plus `-26 cp` contre ce concurrent. La continuation exploree est
  `a4 Qc7+ Kb3 Qc3+ Ka2 b3+ cxb3 Qxb3#` et atteint une feuille
  `qsearch_checkmate`. Le ply ajoute place `b3+` puis `cxb3` dans la recherche
  principale avant que la qsearch voie `Qxb3#`. La sensibilite a l'horizon est
  donc reelle, mais la trace actuelle ne dit pas encore si la branche
  full-depth perdait `b3+` par futility de profondeur 1, fenetre, TT ou ordre.
- Game07_ply183 evite aussi `h2h1` au profit de `g6f7`, avec une marge exacte
  de `-4 cp`, mais le score de la reponse cible reste `-309`. Ce second effet
  vient donc de l'etat TT/ordering et de l'exactification du concurrent, pas
  d'une degradation de la branche cible. Les cinq autres choix racine ne
  changent pas; Game30_ply52 revient meme de `+11` full-depth a `+19 cp` avec
  le ply supplementaire.
- Cout diagnostique : sur les 7 cas, `525 303 -> 599 446` noeuds par rapport au
  contre-factuel full-depth (`+14.11%`, mediane `+10.29%`), avec deux queues a
  `+53.44%` et `+76.29%`. Sur les 32 cas, `8 838 032 -> 8 912 175`
  (`+0.84%`), les 25 cas non cibles etant bit-a-bit identiques hors etiquette
  de mode.
- Verdict : **TargetReply+1 confirme une sensibilite causale locale a l'horizon
  sur un cas, mais ne justifie pas une extension globale**. Aucun changement du
  moteur normal n'est propose. La prochaine experience devra tracer la branche
  exacte jusqu'a `b3+`, puis exempter seulement ce coup de la cause observee
  (notamment futility profondeur 1 si elle est confirmee), sans ajouter de ply
  global, avant de chercher un predicat generalisable.
- Artefacts finaux : binaire SHA-256
  `ba5c9448705cab98d5eed4d8389350e6fea4722ae2fa1eb8197f40a97dabf60c`;
  rapport observationnel
  `18b781cfde68abd61283220fbd57d4f818dc65eff5b218a7c4d28cb5a2b2dd59`;
  rapport full-depth cible
  `75a590c587881d8c8409510a2dc24b2cb6750c9e7e8de8972d6fee251209293d`;
  rapport TargetReply+1
  `b4100557a659313de4fd7fffea95fbd7e86f0da3a2d254326fa7becf34f89623`.

## Update 2026-08-28

- Instrumentation `searchtrace` etendue sans changement du moteur normal :
  chaque recherche effective de la reponse cible publie maintenant sa meilleure
  continuation exploree, mode LMR/full, profondeur et fenetres enfant, score de
  branche, ligne UCI et feuille qsearch/terminale propagee. Le rapport
  passe en schema 3.
- Contre-factuel strict : seule la paire exacte `RootMove/ReplyMove`, au ply 1
  de l'iteration finale demandee, peut etre forcee full-depth; elle doit etre
  normalement eligible LMR et donner reellement echec. Les iterations 1 a 5,
  tous les autres echecs et toutes les reponses non-check restent inchanges.
  Chaque rejeu contre-factuel est execute deux fois avec recherche et TT
  fraiches et doit etre identique.
- Non-interference validee : le rejeu observationnel profondeur 6 retrouve sur
  les 32/32 cas exactement les anciens meilleur coup, score, profondeur et
  nombre de noeuds. Les 32 verifications traced/untraced et les 32 repetitions
  contre-factuelles passent.
- Sur les 7 cas primaires LMR-only, le mecanisme cible passe `7/7` de profondeur
  enfant 3 a 4. Toutes les lignes atteignent encore un stand-pat de qsearch :
  le contre-factuel decale donc bien l'horizon principal d'un ply, sans perdre
  la reponse. Game25_ply69 conserve meme exactement `Kb3 Rf8 g5 Qxc2+ Rxc2` :
  `Qxc2+` migre simplement de la qsearch vers la recherche principale.
- Resultat causal : six scores de la reponse cible restent strictement
  identiques (Game07_ply149/177/183, Game17_ply95, Game25_ply35/69). Seul
  Game30_ply52 passe de `+19` a `+11 cp` depuis le POV racine; sa ligne diverge
  apres `Rf5 Qh4+ Kg8`, de `Nxd5 exd5` vers `Rxc7 Bd4 Be3`. Les 8 cp ne
  suffisent pas : le mauvais coup historique `Rf5` reste choisi.
- Aucun meilleur coup ne change sur les 32 cas. Les canaris restent a
  `10/13` evitements et `2/13` reconnaissances; Game09 reste a `-128`. Le cout
  des 7 cas affectes est `+1.21%` de noeuds au total, maximum `+2.69%` sur
  Game30_ply52; sur les 32 cas, `+0.071%`.
- Verdict : **la reponse calme observee ne disparait pas dans LMR**. La chercher
  un ply plus tot produit la meme borne dans 6/7 cas et un ajustement mineur
  dans le septieme. Les effets mixtes du precedent `CheckNoLMR` global venaient
  donc d'autres echecs de l'arbre et de leurs interactions TT/ordering, pas de
  la reponse cible elle-meme. Aucun changement de recherche n'est propose a ce
  stade.
- Artefacts finaux : binaire diagnostique SHA-256
  `68985f000037f204862ad671ec16cd76f3aebc79b96e569e364500ada8532cca`;
  rapport observationnel SHA-256
  `141c491f7402a52aa2876013bf277124ddc14b5fe72fbfba3a8c0dcf389a2fdb`;
  rapport contre-factuel SHA-256
  `a2a1a0267e69f91f69b177c911c6c17fb90b97190210bfbbba37211007e637d4`.

## Update 2026-08-27

- Le contrat canari separe a ete committe sous `c0e6a69` puis pousse sur
  `origin/main`, avec le commit diagnostique precedent `dfab943`. La baseline
  de l'experience `CheckNoLMR` est donc un HEAD propre, synchronise et marque
  `vcs.modified=false`.
- Experience isolee : ajout temporaire d'une seule condition a l'eligibilite
  LMR, `!InCheck(b)` apres le coup legal. Les coups donnant echec sont ainsi
  recherches a profondeur pleine; TT, PV, futility, null move, qsearch,
  evaluation, reductions et seuils LMR restent inchanges.
- Gate fixe avant resultat : les 7 reponses primaires LMR-only doivent devenir
  full-depth; canaris au moins `10/13` evitements et `2/13` reconnaissances,
  sans nouveau coup historique; au moins un des 6 mauvais choix affectes doit
  disparaitre; cout profondeur 6 sur 32 cas, median `<=10%`, total `<=15%` et
  aucun cas individuel `>50%`.
- Mecanisme valide : les 7/7 reponses ciblees ne sont plus reduites et passent
  full-depth. Les canaris restent exactement a `10/13` evitements et `2/13`
  reconnaissances; Game09 evite toujours `Ra7` et reste a `-128`.
- Efficacite mixte : deux mauvais choix affectes disparaissent
  (Game25_ply69 et Game30_ply52), mais Game25_ply35 recommence a selectionner
  son coup historique `d2b4`. Le sous-ensemble affecte passe donc seulement de
  `6/7` a `5/7` mauvais choix, avec une regression individuelle interdite par
  le gate.
- Cout global acceptable mais queue excessive : median `+1.20%`, total
  `+2.86%`; Game25_ply59 monte de `+55.34%` et Game25_ply75 de `+103.81%`.
  Les deux caps individuels depassent le seuil prefixe.
- Verdict : **rejet de CheckNoLMR global**. Aucun match Elo n'a ete lance. Le
  patch moteur et son test candidat ont ete retires apres validation de la
  suite complete. La prochaine etape doit instrumenter la continuation et
  l'horizon apres les echecs calmes affectes, plutot que desactiver LMR pour
  tous les echecs ou elargir la qsearch.
- Artefacts : baseline trace SHA-256
  `6ec5507d2ebfbef9e85e45e33858151c336790f2273bdc02ac8f61618e39f35a`;
  candidat trace SHA-256
  `f22b37b9448a5ee8a356dd5b4b08ae9e943ab0c0afdd46d7e8868d8cabb5ceb3`;
  patch candidat SHA-256
  `49fa11f262be4eda082e3e0653a8e2bf931ce69cd06b6d3065e57cb6993e6eaf`.

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
  baseline `2d0a2d8` : penalite de 20 cp par dame adverse distincte touchant
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
- Changement fige dans le commit `8611e7a Add UCI move overhead margin`.
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

`TestSearchTacticalCanaryMetrics` dans [internal/search/search_anti_blunder_test.go](../../internal/search/search_anti_blunder_test.go).

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

`TestSearchAvoidsKnownStalematingMoves` dans [internal/search/search_anti_stalemate_test.go](../../internal/search/search_anti_stalemate_test.go).

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
