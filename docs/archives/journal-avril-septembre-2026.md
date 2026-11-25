# Experiment log · April to September 2026

Entries are ordered from newest to oldest. Decisions and next steps reflect what was known on each date. To resume development, see the [current project status](../../status.md).

The revisions cited refer to the current history. Temporary filenames and checksums identify the original artifacts. [Git revision map](../revisions-git.tsv)

## Update 2026-09-09 — D5ReducedFullShadow

- Schema 11 adds `-probe-d5-reduced-full`, an observation-only probe without an allowlist. For every
  quiet check eligible for late move reduction (LMR) at remaining depth 5 in the final iteration, it
  copies the same board, transposition table (TT), killers and history. It compares three isolated
  searches: a reduced null-window search at `child 3`, a fresh full-depth search at `child 4`, and
  the normal reduced search followed by a full re-search only if the score beats alpha. It records
  scores, principal variations (PVs), alpha crossings, TT and ordering changes, heuristics and D3→D1
  chains.
- All 32 traces were reproduced with a fresh search and TT. All 107 shadow records matched exactly.
  The 32 main results also matched a third, untraced search bit for bit on move, score, depth and
  nodes. Primary checks held at `9/19` historical errors avoided and `7/19` losses recognized.
  Canaries held at `10/13` and `2/13`. Game25_ply35 retained `d1e2`, `+146`, `155 300` nodes. Game09
  retained `d6c5`, `-128`, `16 772` nodes.
- Across 23 cases the probe found `107` unique D5 paths. Reduced searches beat alpha six times and
  fresh full-depth searches 11 times. Fourteen scores differed, including five that disagreed on
  crossing alpha. The probe found 229 D3 events and 385 D3→D1 chains across 17 cases. Its `74 894`
  nodes were `0.848%` of the `8 831 770` main nodes. Shadow results were discarded.
- Of 107 child positions, 103 already had a TT entry. Thirteen entries cut off the reduced search
  and none cut off the full search. The two searches left different child entries in 107/107 cases
  and differed in 2,068 TT slots overall. The fresh full search wrote more slots (`3 541` versus
  `2 159`). The first ordered child move still matched in 107/107 cases. Killers diverged in 26
  cases and history in 17, placing most of the state differences deeper in the subtrees.
- On the regressing line `Qe2 Qh1+`, alpha was `-146`. The reduced search returned exactly `-146`
  and triggered no re-search. The fresh full search returned `-145`, beating alpha by 1 cp. The
  normal sequence therefore stayed at `-146`. The substitute line `Qe2 Ng3+` returned `-146` at both
  depths. Both lines contained D3→D1 chains (five and eight respectively).
- No clean D5 structural predicate emerged. Five of the six positive prerequisites had identical
  reduced and full scores. Only Game30_ply52 `Rf5 Qh4+` crossed alpha (`-19 -> -11`), resembling the
  regression `Qe2 Qh1+`. Both the positive Game25_ply69 `a4 Qc7+` and the substitute regression
  `Qe2 Ng3+` had unchanged scores. All six positives and both regressions had at least one D3→D1
  chain. Their TT, killer and history changes also overlapped.
- **Attribution improved. D5D3D1 remains rejected.** A score-delta threshold, an alpha crossing
  alone, a D3→D1 chain or a local ordering change all failed to separate the gains from
  Game25_ply35. These observations do not support a new D5 predicate.
- Proposed next experiment: `D5D3D1ScoreOnlyShadow`. Keep general D3→D1, compute each full-depth D5
  on a complete copy of the TT and heuristics, return only its score and discard the copied state.
  This single ablation would distinguish score effects from TT, killer and history effects on the
  six gains and the regression, without an allowlist or changes to evaluation, depth, quiescence or
  predicates.
- SHA-256 checksums: corpus `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`,
  schema 11 binary `53462bf4c8ad14abd42a99f048db5effcc91e577628f4d9d8310a3552e3d8532`, report
  `9bdc9181c7c08a985346aa2b2798db2914a2c703a109f170dbfb570033870dd4`.

## Update 2026-09-09 — D5BranchAblation

- Schema 10 adds `-d5-ablation-case` and `-d5-ablation-line` to remove one D5 exemption from D5D3D1
  on an exact root/reply path. The line must contain two legal moves ending in a quiet check. The
  selected move retains normal LMR (`depth 5 -> child 3`), its null-window search and any full
  re-search. All other D5 exemptions and the D3→D1 coupling stay active. The
  `d5_check_ablation_lmr_preserved` event confirms the exact application.
- The Game25_ply35 candidate trace contained four D5 paths: `Bb4 Qh1+`, `Bb4 Ng3+`, `Bb4 Nd2+` and
  `Qe2 Qh1+`. Each was disabled separately in a full 32-case replay. All four variants retained the
  erroneous `d2b4` at `+135`. The alternative `d1e2` remained only an upper bound of `+135` and the
  historical move's margin remained a zero lower bound. This rules out a single D5 exemption as the
  cause.
- Game25_ply35 took `133 665` nodes with the intact candidate. Ablations: `Qe2 Qh1+` `147 082`
  (`+10.04%`), `Bb4 Qh1+` `131 803` (`-1.39%`), `Bb4 Ng3+` `133 612` (`-0.04%`), and `Bb4 Nd2+`
  `133 642` (`-0.02%`). Each report contained exactly one ablation. All other 31 cases reproduced
  the intact move, score and nodes. Both runs on all 32 cases were deterministic for each variant.
  Primary results stayed at `11/19` errors avoided and `8/19` losses recognized, and canaries at
  `10/13` and `2/13`.
- Disabling all four initial paths together still gave Game25 `d2b4`, `+135`, `150 561` nodes. The
  changed search exposed a fifth D5, `Qe2 Ng3+`, absent from the original candidate trace. Testing
  all 32 subsets of the five paths restored `d1e2` if and only if both alternatives under the
  competing move were removed: `Qe2 Qh1+` and `Qe2 Ng3+`. Removing that pair alone gave `d1e2`,
  `+145`, `145 619` nodes. The D3→D1 control without any D5 gave the same move and score (`148 952`
  nodes).
- **Two interchangeable D5 paths under the competing `Qe2` branch caused the regression.** Removing
  `Qh1+` exposed `Ng3+`. The three exemptions under the erroneous `Bb4` did not determine the
  choice. D5D3D1 remained ineligible for engine promotion or an Elo match. An allowlist derived from
  these ablations would be unsuitable for production.
- Proposed next experiment: `D5ReducedFullShadow`. For each D5 quiet check, record the reduced and
  full-depth scores, alpha crossing, TT and ordering effects and any downstream D3→D1 chain in an
  isolated copy. Compare the six positive prerequisites with the two regressing `Qe2 Qh1+/Ng3+`
  paths before proposing a single structural predicate.
- SHA-256 checksums: corpus `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`,
  schema 10 binary `453a29ee70f7436d58c962dce736dd652091c6e5863cf9645ac0c85c2dd9600f`, intact
  candidate `63890845a50c7ee44bc57862416fc4ea3a3041ab4840d5e4605ce8ff6df92ea4`, `Qe2 Qh1+` ablation
  `5638446c4cd43ad2db02df1dee282d175a460b9887efe4b18d9913dfa2010253`. The `Bb4 Qh1+`, `Bb4 Ng3+` and
  `Bb4 Nd2+` ablations respectively produced
  `d61ea4a2a34402b7bbb9d9984e42e3a85f6ce22228867a8780d6dcae83badd8f`,
  `e2de950cfabd9aeeee7b86d94887ff0430c8bfd2be29cb40b3b6821c5ce3d180` and
  `3eae3e5f22fd27c37002e22f7cd32244fbe151d4447edc3f9c2918feef9a3c94`.

## Update 2026-09-04 — D5D3D1CheckChain without an allowlist

- Schema 9 adds `-general-d5-d3-d1-checks`. This mode rejects an allowlist, a bridge or probe, and
  the previous candidate. At the final iteration, any quiet check normally eligible for LMR at
  remaining depth 5 or 3 gets normal depth (`child 4` or `child 2`). The D1 futility exemption
  applies only to legal quiet checks two plies below a D3 check forced to full depth by the same
  candidate. The candidate changes no ply, TT, quiescence or evaluation term.
- Game25_ply69 passes without a corpus-specific prerequisite: `Qc7+` is forced to full depth at D5,
  `Qc3+` at D3, and every observed D1 exemption has an exact D3 parent. The result stays `c4d3`,
  `-741` at depth 6 (`78 659` nodes). The unit test fixes the expected result and checks the
  iteration, depths and coupling for every D1 event.
- Across 32 cases there were `127` D5 applications (`127` unique paths) in 23 cases, `1 081` D3
  (`1 006` unique) in 28 cases and `2 092` D1 (`1 972` unique) in 25 cases. All 32 candidate replays
  were deterministic with fresh searches and TTs. All 32 traced baselines matched their untraced
  searches exactly.
- Against normal search, primary errors avoided rose from `9/19` to `11/19` and recognized losses
  from `7/19` to `8/19`. New avoidances were Game07_ply177 (`g1h2 -> g6f5`), Game25_ply69
  (`a3a4 -> c4d3`) and Game30_ply52 (`f8f5 -> c5d6`). Game17_ply103 now recognized the loss
  (`-130 -> -582`). Canaries stayed at `10/13` avoidances and `2/13` recognized losses. Game09 held
  `d6c5`, `-128`.
- One primary regression check failed. Game25_ply35 avoided `d2b4` with `d1e2`, `+146` on the
  baseline and `d1e2`, `+145` on the previous candidate. The unrestricted D5 rule also searched
  `Qh1+` after the competing `d1e2`. That move no longer beat `+135`, and the engine returned to the
  historical `d2b4` at `+135`. The new candidate reproduced six of seven targeted prerequisites, but
  fell from `12/19` to `11/19` avoidances against the previous candidate. Loss recognition remained
  `8/19`.
- Cost relative to normal search: `8 831 770 -> 8 991 901` nodes, `+1.813%` overall and a per-case
  median of `+0.751%`. Three cases exceeded `+50%`: Game17_ply103 `+68.41%`, Game25_ply59 `+55.16%`
  and Game25_ply75 `+86.36%`. Against the previous D3→D1 candidate with seven prerequisites, total
  nodes fell by `0.072%`. The similar aggregate cost did not offset the Game25_ply35 decision
  regression.
- **The structural and tactical signal was confirmed. A primary regression ruled out promotion and
  an Elo match.** Proposed next experiment: `D5BranchAblation`. Keep global D3→D1 and disable the
  four exact D5 paths observed in Game25_ply35 individually, starting with `d1e2 Qh1+`, to identify
  which path changes the choice. A new D5 rule requires a structural predicate separating the
  regressing path from the six positives. A new allowlist will not qualify for production.
- SHA-256 checksums: corpus `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`,
  schema 9 binary `e66a1a9cc40b91e73f952f3571b3292b32c0a438ba261e03943a609067c4b639`, normal
  baseline `d0af24eff28e96959cd3c7a7f1e76f0212bade099cec2c48631c0380f2b45d91`, previous candidate
  remeasured with the same binary
  `4400b741b3ce23dad549942e31105ceb30953e736117f32335dca566ae5213ee`, D5D3D1 candidate
  `e985b67e3f6f3a55d851be640a51840c152c7d14f3a85e031e501758a9ab8439`.

## Update 2026-09-04 — D3→D1 structural probe and generalized candidate

- Schema 8 runs a full-depth shadow on copies of the board, TT and heuristics for each quiet check
  eligible for LMR at remaining depth 3 in the final iteration. Shadow nodes, scores, TT entries,
  killers and history never feed back into the main search. Two plies deeper, the probe records only
  legal quiet checks that the depth-1 futility predicate would prune.
- An initial observation-only check found the pattern frequently but missed `Qc3+ -> b3+`. Reaching
  that branch first requires full-depth `Qc7+`. That preliminary set was excluded from candidate
  selection. The final protocol keeps the seven frozen RootMove/ReplyMove LMR pairs at full depth
  and probes all 32 cases.
- The final instrumentation recorded `830` D3 probes (`772` unique paths) in 28 cases and `1 770`
  D3→D1 chains (`1 626` unique) in 25 cases, showing the pattern well beyond Game25. Shadows used
  `22 075` nodes against `8 838 032` main nodes (`0.250%`). The 25 untargeted searches showed no
  interference. The seven targeted searches and shadows matched across two fresh searches and TTs.
- The probe reproduced the control exactly. After `a4 Qc7+ Kb3`, `Qc3+` used baseline LMR at
  `child 1`, and its full-depth shadow at `child 2` returned `+644` in seven nodes. Six D1 quiet
  checks met the futility condition in this shadow, including `b3+` after `Ka2` with
  `staticEval 480`, alpha `644` and the exact line `a4 Qc7+ Kb3 Qc3+ Ka2 b3+`.
- With the generalization condition met, the coupled `D3CheckNoLMR+D1CheckNoFutility` candidate was
  measured at the final iteration. Each D1 exemption stayed limited to two plies below a D3 check
  searched at full depth by that candidate. It applied `870` D3 NoLMR exemptions (`795` unique
  paths) and `1 862` D1 NoFutility exemptions (`1 742` unique), in 28 and 24 cases respectively.
- Primary errors avoided rose from `9/19` to `12/19` and losses recognized from `7/19` to `8/19`.
  Three earlier mistakes disappeared without a new avoidance regression: Game07_ply177
  (`g1h2 -> g6f5`, `-302 -> -326`), Game25_ply69 (`a3a4 -> c4d3`, `-644 -> -741`) and Game30_ply52
  (`f8f5 -> c5d6`, `+11 -> +18`). Game25_ply75 also changed move but already avoided its historical
  error.
- Canaries stayed at `10/13` errors avoided and `2/13` losses recognized. Game09 kept `d6c5`,
  `-128`. Game06 kept `d1d4` and now saw the mate (`-603 -> -31994`). All 32 candidate replays were
  deterministic with independent searches and TTs.
- Node count changed from 8,838,032 to 8,998,390, or `+1.814%` overall, with a per-case median of
  `+0.420%`. Three cases exceeded `+50%`: Game17_ply103 `+68.41%`, Game25_ply59 `+54.92%` and
  Game25_ply75 `+85.10%`. Game25_ply69 rose `+43.14%` and Game07_ply177 fell `12.30%`.
- **The depth-6 check was positive, but engine promotion and an Elo match were deferred.** The
  nested candidate still depended on seven corpus-specific RootMove/ReplyMove exemptions. In normal
  search it would not guarantee reaching `Qc3+` on Game25. The next minimal experiment was
  `D5D3D1CheckChain`: generalize the outer prerequisite to D5 quiet checks eligible for LMR, keep
  the D3→D1 coupling, remove case allowlists and rerun the same checks before considering an Elo
  A/B.
- SHA-256 checksums: corpus `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`,
  targeted probe binary `4c82b16e800e80ae0fd643df3f52e79eb4dff383d5a55f0800a956e264c04daa`, targeted
  probe report `1866096f199f79a3bdb5eb965cbce5e57b632b9a5c719440512f0e4f045cb6fa`, general candidate
  binary `9526e617852ddfa5cde352276a7040fb418d9a49da2d8c7316b558daeeb64d4c`, candidate baseline
  report `6f4b47da6658107b09f7b5b8541d576121d110b01cb2a96457e6f70cda701e65`, candidate report
  `0814f12ccc9718d8a3002a04a2fd740cdc1caeb5adf0d1e1f96328fd62c32c5c`.

## Update 2026-09-03 — TargetNestedCheckNoLMR+BridgeNoFutility

- Schema 7 of `searchtrace` adds a second exact gate. The baseline keeps `Qc3+` exempt from LMR on
  `a4 Qc7+ Kb3`. The candidate adds only a futility exemption for `b3+` after the complete prefix
  `a4 Qc7+ Kb3 Qc3+ Ka2`. The final move must match the validated line, be legal and give a
  non-capturing check. Depths and windows stay the same.
- On the aspiration pass, normal depth-1 futility applies
  (`staticEval 480 + margin 150 <= alpha 644`). The exemption fires once. The engine then verifies
  `b3+` is legal and searches it at child depth 0. In-check quiescence sees `cxb3` and then the
  capture `Qxb3#`. The propagated leaf is `qsearch_checkmate` at `-31992`.
- The new score fails outside the aspiration window and triggers a full-window search. On that pass,
  `Qc3+` again receives the exact LMR exemption. For `b3+`, alpha is `498`, so the futility
  predicate is false and the move is searched normally. Across the two passes of iteration 6 there
  is one effective futility exemption and two NoLMR applications.
- The causal check passes. Game25_ply69 drops the bad `a4` (`-644`) for `d1d3` (`-728`). On the
  decisive pass, the historical `a4` branch returns at most `-754`, putting its margin against
  `d1d3` at most `-26 cp`. The exact tactical line is `a4 Qc7+ Kb3 Qc3+ Ka2 b3+ cxb3 Qxb3#` with no
  extra ply.
- Primary errors avoided rise from `9/19` to `10/19`, with recognized losses steady at `7/19`.
  Canaries hold at `10/13` avoidances and `2/13` recognitions. Game09 retains `d6c5`, `-128`,
  `16 772` nodes. All 31 untargeted cases reproduce the untraced search, and the target repeats
  identically across two fresh TTs.
- Game25 rises from `54 331 -> 91 546` nodes (`+68.50%`), and all 32 cases from
  `8 831 785 -> 8 869 000` (`+0.421%`), with zero median change. The local increase includes the
  full-window search required by the corrected score. One case accounts for too much of the change
  to extrapolate a general cost.
- **The exact LMR/futility interaction was confirmed. Engine promotion was deferred.** A hard-coded
  path cannot justify a general rule or Elo match. The next experiment should instrument quiet
  checks reduced by LMR at depth 3 followed by quiet checks pruned by futility at depth 1 across the
  corpus. Consider a general `D3CheckNoLMR+D1CheckNoFutility` candidate only if the pattern appears
  beyond this position.
- SHA-256 checksums: corpus `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`,
  schema 7 binary `1e56c75614b78065fd210609147e81403dc00c6aaff446797f98afe660644136`, NoLMR baseline
  report `0544c4a197ed4ef4dcbd491201cea05376a92c6bdeecf536b5d3f44a66f9e5cb`, combined candidate
  report `316d5928fbae4aace866cbe4ff6e37c9676fc1cb6ae96fe6725e36cc8675aee1`.

## Update 2026-09-03 — TargetNestedCheckNoLMR

- The `TargetNestedCheckNoLMR` counterfactual was added only to the schema 6 `searchtrace` binary.
  It requires the exact `a4 Qc7+ Kb3 Qc3+ Ka2 b3+` line, the unique intermediate move `Qc3+`,
  confirmation that the move is legal and gives a quiet check, and the requested final iteration.
  The normal engine is unaffected.
- On the decisive depth-6 pass in Game25_ply69, `Qc3+` changes once from LMR
  `depth 3 -> child depth 1` to normal `child depth 2` without an extension. Earlier iterations, the
  other 31 cases, other moves and other paths are unaffected.
- The newly searched ply gets past the old TT cutoff on `Ka2` and exposes the next obstacle. The
  engine generates `b3+` at depth 1 and orders it at index 3, then futility-prunes it before
  legality checking or search. The reported static value is `+480` against alpha `+644`. The `Qc3+`
  score therefore reaches only alpha `+644`.
- The effectiveness check fails. Game25_ply69 keeps `a4` at `-644` with no improvement in root
  margin. Primary results stay at `9/19` errors avoided and `7/19` losses recognized. None of the 32
  best moves or scores changes.
- Canaries hold at `10/13` avoidances and `2/13` recognitions. Game09 stays at `d6c5`, `-128`,
  `16 772` nodes. All 31 untargeted cases match the untraced search, and the target is identical
  across two fresh searches and TTs.
- Cost with no benefit: `8 831 770 -> 8 831 785` nodes across the 32 cases, an increase of `+15`
  (`+0.00017%`). Game25 alone accounts for `54 316 -> 54 331` (`+0.0276%`). The overall median
  change is zero.
- **TargetNestedCheckNoLMR alone was rejected.** It isolates a sequence involving LMR, the TT and
  then futility. The next minimal experiment is `TargetNestedCheckNoLMR+BridgeNoFutility`: keep the
  exact `Qc3+` exemption and exempt only `b3+` from depth-1 futility on the same prefix. This needs
  no extra ply or change to the TT, quiescence or global rules.
- SHA-256 checksums: corpus `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`,
  schema 6 binary `bdc1fdf83f93a310d0cb44d56e71ee6df6a21d27fed5a5eef9afa7e4faff9f94`, baseline
  report `5b6877da2f3c87d9e9d64924df23215a5d631e686b99ae5e6e6c195b4dd312c8`, candidate report
  `20d4191cb2d4e3d74b3ffe9fc921055b9200c200d644e7e99aa98e2f5200bbb2`.

## Update 2026-08-31 — TargetBridgeNoFutility

- The TargetReply+1 set was frozen at `90ea1da` before this experiment. At the time of this entry it
  had not been pushed, and the historical stash was intact.
- Schema 5 of `searchtrace` accepts an exact UCI line. It traces each step after the root/reply pair
  only when the entire preceding prefix matches. The validated Game25_ply69 line is
  `a4 Qc7+ Kb3 Qc3+ Ka2 b3+`. The final `b3+` was verified legal, quiet and checking.
- The `TargetBridgeNoFutility` prerequisite fails. The decisive depth-6 pass never reaches `b3+`, so
  pruning that move cannot explain the result. No exemption was implemented.
- On the decisive pass, `Kb3` is the TT/killer move and is searched first at full depth. The engine
  generates legal `Qc3+`, orders it fifth and applies LMR `depth 3 -> child depth 1` with a
  reduction of one. At `Ka2`, the reduced search finds an `exact` TT entry with matching key and
  halfmove counter, depth `1` and score `-480`, and cuts off immediately. The score for `Qc3+`
  returns `+480`, below alpha `+644`, so there is no full re-search.
- Only at iteration 5 does this line enter non-check quiescence before `b3+`. The engine generates
  the move at index 0, then filters it as a non-capture. This observation does not explain the
  decisive iteration or justify widening quiescence.
- The search matches the targeted full-depth baseline exactly. Game25_ply69 chooses `a4`, score
  `-644`, `54 316` nodes. All 31 untargeted cases match the untraced search. Canaries remain at
  `10/13` avoidances and `2/13` recognized losses. Game09 keeps `d6c5`, `-128`, `16 772` nodes.
- **TargetBridgeNoFutility was rejected before running a counterfactual.** The next mechanism to
  isolate is nested LMR on `Qc3+` coupled with the shallow TT cutoff. Proposed experiment:
  `TargetNestedCheckNoLMR`. Exempt only `Qc3+` from LMR on the exact prefix, retaining the TT,
  quiescence, futility and global depth.
- SHA-256 checksums: observation-only binary
  `a7872e0f2df59c5db998e39ba0617c4f7b0abefc6328814f29c53f89c131f53a`, schema 5 report
  `8e7605a8454f93fb975627c7e4883992821ee8335470cf74b75c34324e194141`. The regenerated corpus matched
  its historical checksum `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`.

## Update 2026-08-28 — TargetReply+1

- Schema 4 of `searchtrace` adds an exact case allowlist for counterfactuals. For every root move on
  the decisive pass it reports the score, alpha-beta window and value type (`exact`, upper bound or
  lower bound). The reported margin compares the historical mistake with its best alternative while
  preserving fail-low bounds.
- `TargetReply+1` was limited to the seven previously identified LMR-only pairs at the final depth-6
  iteration. All seven replies moved from normal child depth 4 to 5. No extension leaked into
  iterations 1 through 5. Both counterfactual runs matched in each of the seven cases. The other 25
  reproduced the untraced search exactly.
- Canaries stayed at `10/13` errors avoided and `2/13` losses recognized. Game09 kept `d6c5`, score
  `-128` and `16 772` nodes. On the 19 primary cases, avoidance rose from `9/19` to `11/19` while
  loss recognition stayed at `7/19`.
- Game25_ply69 showed a local causal effect. The root score of the targeted reply fell from `-644`
  to `-754 cp`. The engine replaced `a3a4` with `d1d3`, leaving the historical move a margin of at
  most `-26 cp`. The explored continuation `a4 Qc7+ Kb3 Qc3+ Ka2 b3+ cxb3 Qxb3#` reached a
  `qsearch_checkmate` leaf. The extra ply put `b3+` and then `cxb3` in main search before quiescence
  saw `Qxb3#`. This establishes horizon sensitivity. The trace could not yet distinguish whether the
  full-depth branch had missed `b3+` because of depth-1 futility, its window, the TT or ordering.
- Game07_ply183 also avoided `h2h1` in favor of `g6f7` with an exact margin of `-4 cp`. The targeted
  reply still scored `-309`. The change came from TT/ordering state and obtaining an exact score for
  the alternative. The other five root choices stayed the same. Game30_ply52 even rose from `+11` at
  full depth to `+19 cp` with the extra ply.
- Across the seven cases, `525 303 -> 599 446` nodes relative to the full-depth counterfactual
  (`+14.11%` total, `+10.29%` median), with two outliers at `+53.44%` and `+76.29%`. Across all 32,
  `8 838 032 -> 8 912 175` (`+0.84%`). The other 25 were bit-for-bit identical apart from the mode
  label.
- **TargetReply+1 confirmed local horizon sensitivity on one position. The evidence did not justify
  a general extension.** No normal-engine change was proposed. Trace the exact branch through `b3+`,
  then exempt only that move from the observed cause (depth-1 futility if confirmed) before
  considering a general predicate. No global ply increase is needed.
- Final SHA-256 checksums: binary
  `ba5c9448705cab98d5eed4d8389350e6fea4722ae2fa1eb8197f40a97dabf60c`, observation report
  `18b781cfde68abd61283220fbd57d4f818dc65eff5b218a7c4d28cb5a2b2dd59`, targeted full-depth report
  `75a590c587881d8c8409510a2dc24b2cb6750c9e7e8de8972d6fee251209293d`, TargetReply+1 report
  `b4100557a659313de4fd7fffea95fbd7e86f0da3a2d254326fa7becf34f89623`.

## Update 2026-08-28

- The `searchtrace` diagnostics now report the best continuation explored on each actual search of
  the targeted reply, its LMR/full mode, child depth and windows, branch score, UCI line, and
  propagated quiescence or terminal leaf. The report moves to schema 3. The normal engine is
  unchanged.
- Only the exact `RootMove/ReplyMove` pair at ply 1 of the final requested iteration can be forced
  to full depth. It must normally be LMR-eligible and actually give check. Iterations 1–5, all other
  checks and all non-checking replies retain normal behavior. Each counterfactual must match on two
  runs with fresh search state and TT.
- All 32 observation-only depth-6 replays reproduced the old best move, score, depth and node count.
  All 32 traced/untraced comparisons and 32 repeated counterfactuals passed.
- In the seven primary LMR-only cases, `7/7` moved from child depth 3 to 4. Every line still reached
  a quiescence stand pat. The counterfactual moved the main horizon by one ply and retained the
  reply. Game25_ply69 even kept the exact line `Kb3 Rf8 g5 Qxc2+ Rxc2`: `Qxc2+` simply moved from
  quiescence into main search.
- Six target reply scores matched exactly (Game07_ply149/177/183, Game17_ply95 and Game25_ply35/69).
  Only Game30_ply52 changed, from `+19` to `+11 cp` from the root perspective. After `Rf5 Qh4+ Kg8`
  the line changed from `Nxd5 exd5` to `Rxc7 Bd4 Be3`. The 8 cp difference was too small to displace
  the historical mistake `Rf5`.
- No best move changed in the 32 cases. Canaries stayed at `10/13` avoidances and `2/13`
  recognitions. Game09 kept `-128`. The seven affected cases cost `+1.21%` extra nodes in total,
  peaking at `+2.69%` on Game30_ply52. Across all 32 the cost was `+0.071%`.
- **LMR still searches the observed quiet reply.** Searching it one ply earlier returned the same
  bound in six of seven cases and only a small adjustment in the seventh. The mixed effect of the
  earlier global `CheckNoLMR` came from other checks in the tree and their TT/ordering interactions.
  No search change was proposed.
- Final SHA-256 checksums: diagnostic binary
  `68985f000037f204862ad671ec16cd76f3aebc79b96e569e364500ada8532cca`, observation report
  `141c491f7402a52aa2876013bf277124ddc14b5fe72fbfba3a8c0dcf389a2fdb`, counterfactual report
  `a2a1a0267e69f91f69b177c911c6c17fb90b97190210bfbbba37211007e637d4`.

## Update 2026-08-27

- The separate canary contract was committed at `c0e6a69` and pushed to `origin/main` with the
  earlier diagnostics commit `dfab943`. The `CheckNoLMR` experiment thus started from a clean,
  synchronized HEAD marked `vcs.modified=false`.
- Isolated experiment: temporarily add one condition to LMR eligibility, `!InCheck(b)` after the
  legal move. Checking moves then get full depth. TT, PV, futility, null move, quiescence,
  evaluation, reductions and LMR thresholds remain unchanged.
- Criteria fixed before testing: all seven primary LMR-only replies must reach full depth. Canaries
  must retain at least `10/13` avoidances and `2/13` recognitions, with no reappearance of an
  avoided historical move. At least one of six affected bad choices must disappear. At depth 6
  across 32 cases, median node increase must be `<=10%`, total `<=15%`, with no individual case
  `>50%`.
- All seven target replies reached full depth. Canaries stayed at `10/13` avoidances and `2/13`
  recognitions. Game09 still avoided `Ra7` and scored `-128`.
- Two affected mistakes disappeared (Game25_ply69 and Game30_ply52), but Game25_ply35 returned to
  its historical `d2b4`. The affected subset went from `6/7` to `5/7` bad choices, including one
  individual regression forbidden by the criteria.
- Median cost was `+1.20%` and total cost `+2.86%`. Game25_ply59 rose `+55.34%` and Game25_ply75
  `+103.81%`. Both exceeded the preset individual limit.
- **Global CheckNoLMR was rejected.** No Elo match was run. The engine patch and candidate test were
  removed after the full suite was validated. The next step was to trace the continuation and
  horizon after the affected quiet checks, keeping the existing general policy for LMR and
  quiescence.
- SHA-256 checksums: baseline trace
  `6ec5507d2ebfbef9e85e45e33858151c336790f2273bdc02ac8f61618e39f35a`, candidate trace
  `f22b37b9448a5ee8a356dd5b4b08ae9e943ab0c0afdd46d7e8868d8cabb5ceb3`, candidate patch
  `49fa11f262be4eda082e3e0653a8e2bf931ce69cd06b6d3065e57cb6993e6eaf`.

## Update 2026-08-26

- The 13 tactical canaries now have two independent metrics, measured in one fresh depth-6 search
  per position. `BadMoveAvoidance` compares the best move with the historical mistake.
  `LossRecognition` keeps the existing `score <= -300` threshold from the side to move's
  perspective.
- Baseline avoidance is `10/13`. Game03, Game17 and Game21 still play the historical move.
  Recognition is `2/13`: only Game06 and Game21 see the loss below the threshold. Game09 avoids
  `Ra7` while scoring `-128`, illustrating why the two measures are distinct.
- Each of the ten existing avoidances and two recognitions is protected individually, as well as the
  aggregate floors of `>=10/13` and `>=2/13`. A new gain on one position cannot mask a regression on
  another.
- The `search_trace` report moves to schema 2 and exposes `bad_move_avoided`,
  `historical_bad_move_selected` and `loss_recognized` separately at each depth. The historical
  `root_move_best` field remains available. Full replay at depth 4 confirms `5/13` avoidances and
  `0/13` recognized losses. At depth 6 the figures are `10/13` and `2/13`. Report
  `/tmp/equinox-search-trace-metrics-depth4-6-20260826.json`, SHA-256
  `f92cc92990456290e431c7e89b29e8e03db75e466c09d264270d04ca73ce5116`. Binary SHA-256
  `3f74a83e903064ee59a4a3deae81eb0df66ba4333b7891e1dc4181e34816460d`. This step changes only
  instrumentation.

## Update 2026-08-25

- The isolated `QueenZone20` candidate was implemented and tested against `2d0a2d8`. It penalizes
  each distinct opposing queen geometrically attacking the king ring by 20 cp, once per queen, only
  when `nonPawnMaterial > 2000`. It adds no bishop or rook term, square multiplicity, synergy, cap
  or search change.
- Checks covered occupancy-aware rays including the first occupied square, ignored pins, a missing
  king, blockers of either color, two queens, friendly queens, multiple hits, color symmetry, sign
  and perspective, and exact integration into `Evaluate`. Targeted tests, the race detector and
  `go vet ./...` passed. Independent review found no functional defect.
- At depth 6, exactly `2/13` canaries still passed (Game06 and Game21). Game09 held `-128` as
  expected: no queens and the phase gate closed at 1,660 non-pawn material. Main changes were Game02
  `-7 -> +17`, Game03 `-157 -> -175`, Game25 `+453 -> +436`, Game26 `-36 -> -56` and Game27
  `+73 -> +88`. No canary changed pass/fail status.
- Before A/B testing, four canaries had an opposing queen attacking the ring, but the two queens'
  static threats canceled in Game02, Game06 and Game28. The net adjustment was `-20` on Game25,
  `+20` on Game30 and zero on the other eleven. Outside the canaries, net queen danger still showed
  an exploratory association with severe swings: `+35.9` percentage points, 95% bootstrap interval
  `[+13.7, +51.4]`.
- The first screening match was interrupted and excluded. A baseline time forfeit in game 4 violated
  the hard criterion. The stop took effect after 30 completed games. The partial file had no valid
  sidecar and was excluded from every result and aggregate.
- Valid screening match, 32 games on 16 paired v2 openings: `12-6-14`, score `46.88%`, estimate
  `-21.7 Elo`, 95% Hoeffding interval `[-331.4, +250.0]`. Neither engine had a time forfeit, crash,
  illegal move, disconnect, stall or unknown termination. The result cleared the preset continuation
  threshold (`>= -40 Elo`).
- Time-only/depth-zero annotations in the screening match: candidate `15/1969` (`0.76%`), baseline
  `19/1970` (`0.96%`). Mean depths were `7.31` and `7.37`, with equal mean move time of `93.74 ms`.
  Binary and input hashes matched the manifest.
- Separate long validation with the same binaries, 128 games on 64 v2 openings: `46-32-50`, score
  `48.44%`, estimate `-10.9 Elo`, 95% Hoeffding interval `[-135.3, +110.7]`. First half `23-15-26`
  (`47.66%`). Second half on the 32 openings absent from screening `23-17-24` (`49.22%`). Both
  segments scored below even.
- Neither engine had an incident in long validation. Depth-zero rates were equal to reported
  precision: candidate `44/9090` (`0.484%`), baseline `44/9098` (`0.484%`). Mean times were `94.28`
  versus `94.25 ms` and mean depths `7.02` versus `7.14`.
- As White the candidate scored `31-15-18` (`60.16%`) and as Black `15-17-32` (`36.72%`). Most of
  the White advantage came from the suite. The baseline scored `63.28%` as White and `39.84%` as
  Black, leaving the candidate `3.12` percentage points behind with each color.
- Frozen SHA-256 checksums: baseline
  `7a7c18bf755f3048c63621070d5268050607cf0e8818eec26815c19cb211532b`, candidate
  `d1887ec08fc371a8575f0207eb600ab955bd53cd113d30a83da8fbe2ad98eed1`, long-match PGN
  `b45010669eca5bb5c92d1f1c525cd6b6296793c8b34a61d23c08febca4bcab58`.
- **QueenZone20 was rejected.** The long-match estimate was negative, with no compensating Elo or
  operational benefit. The engine patch and tests were removed. Immediate interpolation to 10 cp and
  restoring the multi-piece stash were ruled out. The historical correlation identified a tactical
  symptom without validating a causal static term.
- Offline analyzer instrumentation classifies each immediate opponent reply by capture, check, quiet
  check, promotion, whether the side was already in check and eligibility for the current quiescence
  search. Theoretical eligibility includes all evasions when already in check, otherwise captures
  before delta pruning. It does not show that search reached quiescence there or that a
  non-promoting capture would survive pruning.
- Net queen danger was defined in advance as `queen_attackers - placebo_queen_attackers > 0` before
  the move, with the primary analysis restricted to middlegames outside check and excluding
  canaries. The replay criteria required at least 20 severe swings across ten games, a strict
  majority of invisible quiet checks and at least a 15-point excess over the severe control group
  without net queen danger.
- Run6 replay reproduced 30 games, 1,530 rows, 417 severe swings and all 13 canaries. Among severe
  primary cases with net queen danger there were 78 rows across 11 games: nine captures (`11.5%`),
  21 checks (`26.9%`), including 19 invisible quiet checks (`24.4%`), zero promotions and 50 quiet
  non-checks (`64.1%`). Current quiescence would admit 22/78 (`28.2%`): 14 evasions and eight
  captures outside check. It would filter 56/78 (`71.8%`) as non-captures.
- Severe control without net queen danger: 109 rows across 23 games, 25 captures (`22.9%`), 12
  checks (`11.0%`) including nine invisible quiet checks (`8.3%`), zero promotions and 76 invisible
  replies (`69.7%`). Quiet checks were `+16.1` points more frequent with net queen danger but well
  short of a majority. The 19 cases were concentrated in six games.
- Only the central majority criterion failed (`24.4% <= 50%`). **This signal alone did not justify
  adding quiet checks to quiescence or returning to a static term.** In the separate canary sample,
  seven of nine severe canaries received an invisible quiet check (including Game09), but only
  Game25 had positive net queen danger. This selected sample did not overturn the broader result.
- Artifacts: CSV `/tmp/equinox-kingzone-tactical-run6-20260825.csv`, SHA-256
  `7fa89590629317332ad5ebdaf032e38448f74fd9e4946f5c104e380b9348fce7`. JSON
  `/tmp/equinox-kingzone-tactical-run6-20260825.json`, SHA-256
  `fa9c5f489cfeedf0f7443beb12ae270494826cd2d155c0b1390554305dc178b7`. Analyzer SHA-256
  `e4d8535ce6e1ae5131bd3511b0fffb9cf4ac4f932b3ca4d8d938db753320a6ea`. All 14 Python unit tests,
  `compileall` and `git diff --check` passed. No engine or search file had changed at this stage.
- An opt-in tracer of actual search was added in `internal/search/trace.go` and `cmd/search_trace/`.
  By default `Search.Trace == nil`. The evaluation formula, quiescence rules, priorities, cutoffs
  and reductions stay the same. Hooks compile only with the `searchtrace` tag and are eliminated
  from ordinary builds. Events record iterative depth, aspiration/full pass,
  pseudo-legal/ordered/legal indices, TT/PV/killer moves, futility, null move, beta cutoffs,
  LMR/full search, terminal states and quiescence.
- The frozen corpus contains the 19 primary invisible quiet checks and 13 canaries. Each was
  replayed at depths 4 and 6 with a fresh 65,536-entry TT, giving 64 traced searches. Repeating
  every search without tracing gave `64/64` identical best moves, scores, completed depths and node
  counts. The tracer was observation-only across the corpus.
- At iterative depth 1, all 19 primary replies reached a quiescence node. Fourteen were generated
  and filtered as non-captures. Five stopped on stand pat before move traversal. This confirms the
  filter's behavior without establishing its effect at useful search depths.
- On the decisive depth-4 pass, all 19 primary replies were generated, legal, ordered and searched
  in main Negamax. Eight received full depth only, eight LMR only, two separate reduced and full
  visits, and one LMR followed by full re-search. The historical move remained best in 18/19 cases.
- On the decisive depth-6 pass, 7/19 replies received full depth, 7/19 a one-ply LMR reduction
  (`depth 5 -> child depth 3`) with no full re-search, 3/19 nodes cut off on the TT and 2/19 on
  another move before the target. There were no futility prunes or null-move cutoffs on this pass.
  The historical move remained best in 10/19 cases: 6/7 LMR-only, 4/7 full-depth and 0/5 TT/beta
  shortcuts. The selected sample is too small for causal inference.
- The three exclusive primary TT cutoffs at D6 were beta bounds with compatible key, halfmove clock
  and depth (`entry depth >= node depth`). Their TT move differed from the targeted quiet check:
  another opponent move already proved the cutoff. Both remaining beta cutoffs likewise came from a
  different move searched before the target. These cutoffs do not indicate an erroneously discarded
  reply.
- The engine applies PV priority only at the root. Among 19 primary cases, the bad root move had PV
  priority in 15 at D4 and eight at D6. The TT prioritized it in 12 and eight respectively. At the
  reply node at D4, the TT prioritized the target in five cases and a killer in six. At D6 the
  corresponding counts were four and seven. The remaining late quiet checks are candidates for LMR.
- At D6, 10/13 canary replies were searched explicitly and 3/13 served by a valid TT entry. All were
  accounted for. Among the seven severe quiet-check canaries, Game02 was recognized as terminal,
  Game06/Game09/Game21/Game26 received full depth, Game25 triggered LMR followed by full re-search,
  and Game28 was itself the TT move. On Game09, `h2+` was ordered first as a killer and searched at
  full depth. Its score remained `-128`. Neither quiescence nor LMR hid this check.
- The engine avoided the historical error on 10/13 positions at depth 6, yet the existing test
  passed only 2/13 because it required a global score `<= -300`. That test measures loss
  recognition. Game09 avoids `Ra7` but scores `-128` and still fails the test. Track avoidance and
  recognition separately in future experiments.
- **Keep quiet checks out of quiescence on this evidence.** The remaining isolated search hypothesis
  consistent with the traces is to exempt checking moves from LMR while retaining the TT, futility,
  quiescence and evaluation. The seven primary LMR-only cases offer a short counterfactual test. The
  signal is promising but not causal, and the quiet-check canaries do not directly confirm it.
- Search artifacts: diagnostic binary `/tmp/equinox-search-trace-a26755e-20260825`, SHA-256
  `1d1665abb617fdd9289791827220ca6b9610d06891845b619b2e6f06da02db12`. Report
  `/tmp/equinox-search-trace-depth4-6-run6-20260825.json`, SHA-256
  `17d89b7b04e5760fcb23cd2ff2bc37530fbf4318c67be1cf4397860d7779faea`. Targeted tracer tests,
  decisive-pass summary tests and `go vet` passed. The only failures in `go test ./...` were the 11
  historical anti-blunder cases, with exactly their baseline scores. The
  `wip king zone safety experiment` stash was intact.

## Update 2026-08-24

- Offline instrumentation under `tools/kingzone/` replays annotated PGNs from the player's
  perspective, separates finite scores from mates and measures enemy bishops, rooks and queens
  geometrically attacking the king ring before and after Equinox's move. No engine, search or
  evaluation file changed.
- Source `sf1600-run6-1+0.1-30.pgn`, SHA-256
  `1c98886a75d26560326cbfae249742af76ea7dfe813e3d8e3fd66fa74cf0d4dc`. Replay covers 30 games, 3,281
  plies, 1,530 usable Equinox decisions, 417 severe swings and exactly 13/13 canaries.
- The primary variable is Equinox's score from its own perspective plus Stockfish's reply score from
  Stockfish's perspective. A positive Stockfish mate or a finite gap of at least 300 cp is severe.
  The primary analysis never converts mate to centipawns. Intervals use 10,000 game-block bootstrap
  samples with seed `20260824`.
- Within the preset middlegame/out-of-check filter (`nonPawnMaterial > 2000`), an opposing queen
  attacks the ring in 184/921 rows across 24/30 games. Severe swing rates are 51.09% exposed and
  13.30% unexposed, a +37.79-point difference with 95% bootstrap interval `[+20.00, +50.83]` points.
  The median finite swing is 249 versus 109 cp, a +140 cp difference with interval `[+9, +248]`.
- After removing canaries, the excess risk remains +37.14 points, interval `[+18.56, +50.77]`, and
  the median difference +141.5 cp. The symmetric placebo (Equinox queen attacking the opposing king)
  falls to +4.70 points, interval `[-7.67, +20.65]`, and +38 cp in the median.
- Canary coverage: queen `4/13`, bishop `1/13`, rook `2/13`. Bishop and rook missed the preset
  coverage threshold. The rook also had a negative overall correlation. Game09 alone did not justify
  combining all three piece types.
- Before and after `64.Ra7` on the Game09 control, one enemy bishop attacks two ring squares, one
  rook attacks three, and no queen attacks the ring. That is two attackers, five hits and four
  distinct squares, outside check. The pressure predates the move and stays the same. A future queen
  term would not directly fix this canary.
- **The only causal candidate selected for the next experiment was `QueenZone20`.** Penalize each
  distinct opposing queen attacking at least one ring square by 20 cp, once per queen, within the
  existing phase gate. Exclude square multiplicity, bishops, rooks, synergies and the stash's cap.
  The historical signal justified an isolated A/B test. Elo gain remained unmeasured.
- Eight unit tests, Python compilation and `diff --check` passed. Final report SHA-256
  `972560ded3cfa1b7584964f1e37db0adff72a80cd614e64a09116e42606e0151`, final CSV SHA-256
  `77c0dfe103bcd92cf9e9dccf7b7da436a861102ef1c840fb627c4d3109f47dda`.

## Update 2026-08-21

- The `wip king zone safety experiment` stash was audited without popping it. The full patch
  combined knights, bishops, rooks and queens, attacked-square multiplicity, multi-attacker synergy
  and a `250 cp` cap. It was not restored wholesale.
- Within the existing king-safety gate (`nonPawnMaterial > 2000`), test an isolated `20 cp` penalty
  per enemy knight attacking at least one king-ring square. Count each knight once. Sliding attacks,
  occupancy, search and time management stay unchanged.
- Canary checks covered `0/1/2` attackers, multiple hits counted once, opposing color, a missing
  king, color symmetry and sign by side to move. The score matches the baseline below the phase
  threshold and changes by exactly `20 cp` above it. Targeted tests, race detector, `go vet ./...`
  and independent review passed.
- The full suite retained exactly `11/13` historical anti-blunder failures, with `2/13` positions
  still passing. Game02 and Game25 improved by `20 cp`, Game17 worsened by `10 cp`, Game27 improved
  by `2 cp` and Game09 stayed at `-128`.
- Screening match: `32` games on `16` openings at `1+0.1`. Result `14-8-10`, score `56.25%`,
  estimate `+43.7 Elo`, 95% Hoeffding interval `[-216.9, +385.6] Elo`, zero incidents. This met the
  continuation criterion.
- Long validation of the same binary, `128` games on `64` openings: `46-28-54`, score `46.88%`,
  estimate `-21.7 Elo`, 95% Hoeffding interval `[-148.0, +98.8] Elo`. The candidate scored
  `24-14-26` as White and `22-14-28` as Black. Neither engine had a time forfeit, crash, illegal
  move, disconnect, stall or unknown termination.
- Time-only annotations remained rare but were slightly more frequent for the candidate: `32/8424`
  (`0.38%`) versus `25/8427` (`0.30%`). Mean annotated depth was `7.27` versus `7.20` with equal
  mean move time of `93.8 ms`.
- **KnightZone20 was rejected for lack of a long-match gain.** Its point estimate fell below the
  baseline, although the match established neither a statistical nor a practical regression. The
  patch and tests were removed and the original stash kept intact. No immediate interpolation to
  `10 cp` was warranted.
- Next step: instrument the distribution of sliding threats (bishop/rook/queen) against the king
  ring outside search and measure their correlation with canary/PGN swings before choosing a single
  causal evaluation term. Defer the multi-piece stash pending that measurement.

## Update 2026-08-05

- External validation of the same `Overhead15ms` binary against Stockfish 18 limited to `1600 Elo`:
  `128` games on `64` v2 openings at `1+0.1`. Result `79-2-47`, score `62.50%`, estimate
  `+88.7 Elo`, 95% Hoeffding interval `[-31.2, +235.2] Elo`.
- Neither engine had a time forfeit, crash, illegal move, disconnect, stall or unknown termination.
  Colors were balanced: `40-1-23` as White and `39-1-24` as Black.
- The raw score matched the previous external calibration of the `100 ms` policy (`62.50%`,
  `78-4-46`), while incidents fell from `3` Equinox and `2` Stockfish time forfeits to zero.
  Cross-suite results provide calibration. The direct long A/B result of `+19 Elo` supports
  attribution.
- All `11256` plies have annotations. Equinox produced `66/5636` time-only comments (`1.17%`),
  consistent with timing out before completing a depth. This timed-search risk remains a separate
  issue. No UCI incident was observed.
- **The fixed 15 ms overhead was accepted.** There were no forfeits in the `32` game screen, `128`
  game long A/B or `128` game external match, and no detected strength loss. Stop testing the
  `10/20 ms` constants.
- Change frozen at `8611e7a Add UCI move overhead margin`. Clean binary rebuilt from this commit,
  SHA-256 `ae1d10278a8ec18a8fd42d7a5bd8ec5f08cc67fa74024084f215f6f5fd3bc5ba`.
- Calibration against Stockfish 18 limited to `1800 Elo`, `128` games on `64` openings: `59-7-62`,
  score `48.83%`, estimate `-8.1 Elo`, 95% Hoeffding interval `[-132.1, +113.7] Elo`. Equinox had no
  abnormal terminations. Stockfish lost one game on time.
- Excluding that Stockfish forfeit gives `58-7-62` over `127` games, score `48.43%`, estimate
  `-10.9 Elo`, or a nominal calibration around `1789`. Equinox played approximately even with
  SF18-1800 at this time control. The sample neither proves nor rules out `1800+`.
- In timed search, `75/6013` Equinox moves (`1.25%`) had only a time annotation, consistent with a
  `bestDepth=0` return. Mean annotated depth was `5.91`. Equinox had zero time forfeits.
- The root TT fallback hypothesis was tested without changing the accepted binary. Passive
  instrumentation across `32` games on `16` openings at `1+0.1` captured the root TT slot before
  search and logged only `bestDepth=0` returns.
- Coverage was zero. All `16/16` `bestDepth=0` returns were nonterminal with legal moves, but none
  had a matching-key TT entry: `0/16` legal TT moves and `0/16` fallbacks that would change the
  choice. The PV from a completed depth was kept separate and necessarily zero in this `bestDepth=0`
  sample.
- **Rejected before strength testing.** The candidate patch passed targeted tests, race detection,
  vet and independent review, but the measured conditions never exercised the hypothesis. The patch
  was removed. No interpretable Elo A/B could be run, and the engine stayed at the accepted commit.
- Next direction: revisit the king-zone pressure experiment from an audited stash diff, isolating
  the smallest evaluation term and its canary positions before any Elo test.

## Update 2026-08-04

- Long validation of the fixed `15 ms` overhead on the `64` v2 openings: `128` paired games at
  `1+0.1`, result `52-31-45`, score `52.73%`, estimate `+19.0 Elo`, 95% Hoeffding interval
  `[-101.8, +144.8] Elo`.
- The point estimate was comfortably above the preset `-40 Elo` threshold. Neither engine had a time
  forfeit or other abnormal termination. The result met the practical non-inferiority criterion. The
  conservative interval still included zero, leaving an Elo gain unproven.
- All `16394` plies had annotations. Time-only comments, consistent with a timeout before any depth
  finished, were similarly frequent: `41/8199` candidate moves (`0.50%`) versus `39/8195` baseline
  moves (`0.48%`).
- The candidate averaged `93.89 ms` and annotated depth `7.29`, against `99.85 ms` and depth `7.36`
  for the baseline. The candidate scored `57.03%` as White and `48.44%` as Black, without the strong
  asymmetry seen in the `20 ms` test.
- **Proceed to external validation against SF18-1600** with the same frozen candidate binary,
  SHA-256 `9b5d5a983a9dbfc7b644621de0584061e30f1bd1866d1a68ef2127483258253e`.

## Update 2026-08-03

- A fixed UCI overhead of `10 ms` was tested in isolation after the `remaining/10` cap and before
  the `1 ms` floor. Initial budget at `1+0.1`: `100 -> 90 ms`.
- Screening match against the `100 ms` policy, `32` games on `16` pairs: `18-1-13`, score `57.81%`,
  estimate `+54.7 Elo`. Excluding the two games decided on time: `17-1-12`, score `59.17%`.
- Mean candidate search time was `96.2 ms` (`92.9 ms` on the first ten moves, `96.8 ms` thereafter).
  Average depth was `6.58` versus `6.72` for the baseline.
- The safety criterion failed with `1` candidate and `1` baseline time forfeit. The candidate
  flagged in a position it scored around `+7.8` at ply `81`. There were no crashes, illegal moves,
  disconnects or stalls.
- **The 10 ms overhead was rejected.** The patch was removed. The `128` game A/B and SF18-1600
  revalidation were skipped because the candidate failed the zero-forfeit requirement.
- A fixed `20 ms` overhead was tested at the same point. Initial budget at `1+0.1`: `100 -> 80 ms`.
  Screen against the `100 ms` policy, `32` games on `16` pairs: `14-6-12`, score `53.12%`, estimate
  `+21.7 Elo`, conservative 95% Hoeffding interval `[-250.0, +331.4] Elo`.
- Both engines passed the `20 ms` safety check with no forfeits, crashes, illegal moves,
  disconnects, stalls or unknown terminations.
- The candidate averaged `91.7 ms` (`83.9 ms` over its first ten moves, `93.2 ms` thereafter),
  versus `99.9 ms` for the baseline. Moves with depth annotations averaged `7.01` for both. No
  explicit depth annotation was zero.
- In the PGN, `23/2017` candidate moves and `20/2015` baseline moves had only a time annotation. A
  `-debug all` replay confirmed that this format corresponds to `info depth 0`, a timeout before any
  depth finishes. Both engines showed this behavior. Track it over the longer sample.
- Long `20 ms` validation on `64` v2 openings, `128` games: `40-28-60`, score `42.19%`, estimate
  `-54.7 Elo`, 95% Hoeffding interval `[-188.9, +64.4] Elo`. The point estimate missed the preset
  threshold of about `-40 Elo`, although the conservative interval remained inconclusive.
- Neither engine had a time forfeit or other abnormal termination. The candidate had slightly fewer
  `depth 0` returns: `35/8237` moves (`0.42%`) against `37/8251` (`0.45%`) for the baseline.
- Mean candidate cost was `91.45 ms` at depth `7.14`, versus `99.84 ms` at depth `7.34` for the
  baseline. The candidate scored `24-18-22` as White but `16-10-38` as Black. Pairing and `64` games
  per color rule out a simple imbalance in the protocol.
- **The 20 ms overhead was rejected for loss of strength.** The patch was removed and no SF18-1600
  validation was launched. The positive short screen was too noisy to predict the long result.
- A fixed `15 ms` overhead was tested as the sole intermediate value. Initial budget at `1+0.1`:
  `100 -> 85 ms`. Screen on `32` games and `16` pairs: `17-2-13`, score `56.25%`, estimate
  `+43.7 Elo`, 95% Hoeffding interval `[-216.9, +385.6] Elo`.
- Both engines passed the `15 ms` safety check with no forfeits or other abnormal terminations. The
  `depth 0` rate remained low, with `7/2050` (`0.34%`) for the candidate against `9/2047` (`0.44%`)
  for the baseline.
- The candidate averaged `93.8 ms` and annotated depth `6.54`, versus `100.0 ms` and `6.62` for the
  baseline.
- **Proceed conditionally to the long `128` game match with 15 ms overhead.** The short screen met
  the preset criteria. The noise observed with the `20 ms` candidate prevented treating it as Elo
  validation.

## Update 2026-07-31

- Added a paired Elo harness with `32` fixed EPD positions, two games per opening with colors
  reversed, structurally validated PGN, pair-based statistics and a SHA-256 manifest (`tools/elo/`).
- Three isolated search changes were rejected after measurement: strict TT/PV priority, root-only PV
  priority and counting legal moves for LMR.
- The accepted UCI correction retains TT, history and killers between `position` commands until
  `ucinewgame`. A TT guard on `HalfmoveClock` prevents cutoffs from a different fifty-move counter
  while retaining the TT move for ordering.
- Long combined A/B against baseline, `64` games at `1+0.1`: `34-6-24`, score `57.81%`, estimate
  `+54.7 Elo` (95% Hoeffding interval still inconclusive).
- Combined change versus the guard alone, two sets of `64`: total `61-23-44`, score `56.64%`,
  estimate `+46.4 Elo`. Both sets were positive. The conservative interval still included zero, so
  the signal repeated without establishing a definitive Elo gain.
- The TT guard moved Game09 from `-342` back to `-128`. The anti-blunder canary fell from `3/13` to
  `2/13`. Keep that regression visible and look for a way to recover it.
- UCI increment time allocation was corrected and retained. The budget is
  `remaining/40 + 3*increment/4`, capped at `remaining/10` with a floor of `1 ms`. The `movetime`
  and no-increment paths are unchanged. At `1+0.1` the nominal budget rises from `50 ms` to
  `100 ms`.
- Two complete A/B matches against the old budget, `64` games each at `1+0.1`, gave `43-6-15` and
  `42-11-11`. Combined `85-17-26`, score `73.05%`, estimate `+173.2 Elo`, conservative 95% Hoeffding
  interval `[+42.4, +382.1] Elo`. Both repetitions and the pooled interval showed a gain. The change
  was accepted.
- Mean time per move rose from `49.9 ms` to `99.8 ms`. Across both sets, excluding terminal markers,
  mean depth rose from `5.95` to `6.58`, median from `6` to `7`, and p90 from `8` to `9`. There were
  no time forfeits, crashes or illegal moves in the `128` main games or the initial `32` game
  screen.
- Targeted UCI, TT and match-statistics tests, `go test -race ./cmd/uci` and `go vet ./...` passed.
  The only failures in `go test ./...` were the `11/13` known anti-blunder cases.
- The external harness now records per-engine UCI options in the manifest and breaks down abnormal
  terminations by engine. The v2 suite contains `64` unique legal openings (`128` games with colors
  reversed).
- External screen against Stockfish 18 limited to 1600, `32` games at `1+0.1`: `20-1-11`, score
  `64.06%`, indicative `+100.4 Elo`, zero incidents. This screen overlaps the long suite and is
  reported separately from the final result.
- Main external calibration against SF18-1600, `128` games on `64` pairs: `78-4-46`, score `62.50%`,
  estimate `+88.7 +/- 61.8 Elo` (LOS `99.8%`). The more conservative paired Hoeffding interval
  remained inconclusive: `[-31.2, +235.2] Elo`.
- Strong color asymmetry at this very short time control: Equinox as White `47-3-14` (`75.8%`), as
  Black `31-1-32` (`49.2%`). White's overall score was `63.3%`. Pairing controls this bias in the
  total score.
- Five time forfeits in the long set: `3` by Equinox and `2` by Stockfish. No crashes, illegal
  moves, disconnects or stalls. The internal A/B established a strength gain from the new
  allocation, but the safety margin against flagging still needed work.
- Nominal Equinox calibration was around `1689` at `1+0.1` (usual interval approximately
  `[1627, 1751]`), well above the historical snapshot. Opponent, protocol and suite differences
  preclude causal comparison with run6. The `1800` level had not yet been demonstrated.

## Snapshot 2026-04-29

- The engine plays complete games with clean PGN scores, mate notation and no misleading aspiration bounds.
- Against SF1600 at `1+0.1` over `30` games: `12-18-0` (wins–losses–draws), score `40.0%`,
  `-70.4 +/- 133.7` Elo, draw ratio `0.0%`. Run6 PGN: `matches/sf1600-run6-1+0.1-30.pgn`.
- **Phase 1 complete** (scores, PV, timeout and mate fixes). **Phase 2 in progress** (move ordering,
  quiescence and pruning).

## SF1600 match history at `1+0.1` (`30` games each)

| Run | Score | W-D-L | Stalemates | Note |
|-----|-------|-------|------------|------|
| baseline | `31.7%` | `1-17-12` | Several | Starting point, `-133` Elo |
| run3 | `36.7%` | `7-8-15` | `6` | First stalemate fixes in quiescence and terminal Negamax |
| run4 | `38.3%` | `11-1-18` | `0` | After the root PV fail-high fix |
| run5 | `46.7%` | `13-2-15` | `0` | After the timeout fallback fix |
| run6 | `40.0%` | `12-0-18` | `0` | After the mate score fix, with a clean PGN |

Scores plateaued around `40-46%` within the noise (`+/- 133` Elo). Score reporting was now reliable. Playing strength remained the open question.

## Active phase 2 canaries (blunder avoidance)

`TestSearchTacticalCanaryMetrics` in [the tactical regression tests](../../internal/search/search_anti_blunder_test.go).

- `13` FENs from run6: `7` balanced_blunder, `2` unclassified and `4` short conversion_blunder positions.
- Avoidance: the depth-6 best move must differ from the historical mistake. **Status: `10/13`**,
  with each previous success protected.
- Recognition: the depth-6 score must be `<= -300 cp` or a forced mate against the side to move.
  **Status: `2/13`** (Game06 and Game21), both protected. Some missed losses still score positively:
  G16 `+288`, G25 `+453` and G28 `+120`.
- A search change must retain at least `10/13` avoidances and `2/13` recognitions without
  reintroducing an avoided historical move. Clearing those thresholds alone does not establish an
  Elo gain.

## Phase 1 canaries (stalemate avoidance, passing)

`TestSearchAvoidsKnownStalematingMoves` in [the stalemate regression tests](../../internal/search/search_anti_stalemate_test.go).

- `6` FENs from run3. All `6/6` have passed since the root PV fail-high fix.

## Phase 2 fixes tested

| Fix | Canary result | Decision | Notes |
|-----|--------------|----------|------|
| Correct MVV-LVA ordering (king attacker `10000`, promotions) | `2/13 -> 2/13` | Keep | Ordering correctness fix, neutral on canaries |
| Quiescence expanded to checks (`qPly`, second pass) | `2/13 -> 2/13`, time `x1.8`, breaks mate-score test | Revert | Poor trade-off |

## Depth diagnosis · 2026-04-29

Isolated measurements on the `5` priority failing positions, using `SearchRoot` at depths `4/6/8/10/12` without a time limit:

| Position | d4 | d6 | d8 | d10 | d12 | Move stable from |
|----------|------|------|------|------|------|-----------------|
| G02_B    | +16  | -7   | -3   | +9   | -20  | d6 (`g7g6`) |
| G17_W    | +90  | +67  | +50  | +37  | +17  | d4 (`d1f3`) |
| G22_B    | -3   | -32  | -69  | -76  | -94  | d4 (`e4f2`) |
| G26_B    | -3   | -36  | -49  | -53  | -54  | d4 (`g2g1`) |
| G30_B    | +25  | +12  | 0    | -20  | -31  | d6 (`c6e5`) |

No position crossed `-300 cp` even at depth `12`. Scores declined slowly and leveled off (G22 fell fastest to `-94`, G17 stayed at `+17`). Move choices were stable by depth `6`.

**Conclusion at the time:** more effective depth was unlikely to resolve the canaries. Auditing LMR and null move lost priority. The working diagnosis was **evaluation**: the engine failed to recognize fundamentally lost positions. **Phase 3 was brought forward.**

Cumulative wall time was approximately 42 minutes for the `5` positions at depth `12` (G17 alone took `21 min`). This was a one-off diagnosis. The `cmd/depth_diagnostic` tool was removed afterwards.

## Overall plan

| Phase | Objective | Status |
|-------|----------|--------|
| 1. Stabilize the engine | Stop Elo regressions and score bugs | Complete |
| 2. Clean up search | Ordering, quiescence, pruning. Get the anti-blunder canaries passing | Paused (canaries `2/13`, diagnosis indicated a problem beyond depth) |
| 3. Improve evaluation | King safety, threat awareness, tapered evaluation and pawn structure | **In progress** (brought forward after the April 29, 2026 diagnosis) |
| 4. Elo infrastructure | Reproducible gauntlet, SF1400/1600/1800 baselines and rejection criteria | **In progress** (paired A/B available) |
| 5. Target 2000 Elo | Beat SF1400, hold against SF1600, then surpass SF1600 | Planned |

## Phase 3 priorities

The `5` priority failures share a theme: an exposed opposing king or a nearby attack whose risk Equinox misses. Test the following changes one at a time against the anti-blunder canaries:

1. **King safety.** Existing evaluation uses only a `+/-10/-5` pawn shield, gated on `nonPawnMaterial > 2000`. Count attackers, weight the pieces attacking the king zone and consider open files toward the king. Most likely to help G02, G26 and G30.
2. **Threat awareness.** Penalize enemy heavy pieces aimed at undefended targets, such as a queen x-ray or a rook on an open file toward the king.
3. **Tapered evaluation.** Interpolate between middlegame and endgame scores with separate king piece-square tables.
4. **Pawn structure.** Add connected pawns, candidate passers and backward pawns to the existing doubled, isolated and simple passed-pawn terms.
5. **Rook activity.** Score open and semi-open files and the seventh rank.

## Experiment rules

- Test one phase 2 fix per iteration and measure it on the canaries before changing anything else.
- Keep changes that improve the canaries, with an explicit exception for documented correctness fixes.
- Start a new match only after `>= 10/13` canaries pass.
- Prefer reproducible Elo progress to a one-off increase.
- Treat documentation as part of the engine.

## Next step

Test a root fallback using the TT move when no search depth completes.
If the TT contains a nonzero move for the root key and that move is legal,
use it in place of the first legal move. Otherwise retain the existing fallback.
Keep cutoffs, move ordering, pruning and time budgets unchanged.
Add tests for a legal TT move, a stale or illegal move and an absent TT entry.
Measure coverage on returns with
`bestDepth=0` before a v2 A/B. If TT hits are rare or the A/B result is
neutral or negative, remove the patch and resume the work on
king-zone attack pressure.
