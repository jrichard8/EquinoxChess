package main

import (
	"reflect"
	"testing"

	"EquinoxChess/internal/search"
)

func TestSummarizeUsesDecisiveFullWindowResearchPass(t *testing.T) {
	events := []search.TraceEvent{
		{Iteration: 4, Pass: "aspiration", Role: "reply_move", Node: "negamax", Stage: "lmr_reduced", OrderIndex: 8},
		{Iteration: 4, Pass: "aspiration", Role: "target_line", Node: "quiescence", Stage: "enter_after_reply"},
		{Iteration: 4, Pass: "full_window_research", Role: "reply_move", Node: "negamax", Stage: "full_depth", OrderIndex: 2},
		{Iteration: 4, Pass: "full_window_research", Role: "reply_move", Node: "negamax", Stage: "search_complete", OrderIndex: 2},
	}
	summary := summarize(events, 4)
	if summary.FinalPass != "full_window_research" {
		t.Fatalf("FinalPass = %q", summary.FinalPass)
	}
	if summary.ReplyFinalPass.StageCounts["negamax.lmr_reduced"] != 0 {
		t.Fatal("decisive-pass summary leaked aspiration LMR event")
	}
	if summary.ReplyFinalPass.StageCounts["negamax.full_depth"] != 1 {
		t.Fatal("decisive-pass summary missed full-depth event")
	}
	if summary.TargetLineQEntriesFinal != 1 || summary.TargetLineQEntriesFinalPass != 0 {
		t.Fatalf(
			"qsearch counts final=%d final-pass=%d",
			summary.TargetLineQEntriesFinal,
			summary.TargetLineQEntriesFinalPass,
		)
	}
}

func TestParseDepthsSortsAndDeduplicates(t *testing.T) {
	depths, err := parseDepths("6, 4,6")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(depths, []int{4, 6}) {
		t.Fatalf("depths = %v", depths)
	}
}

func TestBuildAggregatesSeparatesAvoidanceFromLossRecognition(t *testing.T) {
	cases := []corpusCase{
		{
			Cohort: "canary",
			Runs: []run{{
				RequestedDepth: 6,
				RootMoveBest:   false,
				Score:          -128,
			}},
		},
		{
			Cohort: "canary",
			Runs: []run{{
				RequestedDepth: 6,
				RootMoveBest:   true,
				Score:          -31994,
			}},
		},
	}

	result := buildAggregates(cases, []int{6})["6"]["canary"]
	if result.Cases != 2 {
		t.Fatalf("Cases = %d, want 2", result.Cases)
	}
	if result.BadMoveAvoided != 1 || result.HistoricalBadMoveSelected != 1 {
		t.Fatalf("avoidance split = %d avoided/%d selected, want 1/1",
			result.BadMoveAvoided, result.HistoricalBadMoveSelected)
	}
	if result.LossRecognized != 1 {
		t.Fatalf("LossRecognized = %d, want 1", result.LossRecognized)
	}
	if result.RootMoveBest != result.HistoricalBadMoveSelected {
		t.Fatalf("legacy RootMoveBest = %d, selected = %d",
			result.RootMoveBest, result.HistoricalBadMoveSelected)
	}
}

func TestSelectFinalContinuationUsesDecisivePass(t *testing.T) {
	continuations := []continuationOutput{
		{TraceContinuation: search.TraceContinuation{Iteration: 6, Pass: "aspiration", Score: 10}},
		{TraceContinuation: search.TraceContinuation{Iteration: 6, Pass: "full_window_research", Score: 20}},
	}
	selected := selectFinalContinuation(continuations, 6, "full_window_research")
	if selected == nil || selected.Score != 20 {
		t.Fatalf("selected = %#v, want full-window score 20", selected)
	}
}

func TestFormatContinuationConvertsReplyAndLeafPOV(t *testing.T) {
	output := formatContinuation(search.TraceContinuation{
		Score:      125,
		LeafPly:    5,
		LeafScore:  80,
		LeafReason: "qsearch_stand_pat",
	})
	if output.ScoreRootPOV != -125 {
		t.Fatalf("ScoreRootPOV = %d, want -125", output.ScoreRootPOV)
	}
	if output.LeafAfterReplyPly != 3 || output.LeafScoreRootPOV != -80 {
		t.Fatalf("leaf relative ply/root POV = %d/%d, want 3/-80",
			output.LeafAfterReplyPly, output.LeafScoreRootPOV)
	}
}

func TestBuildRootMarginRetainsAlphaBetaBounds(t *testing.T) {
	results := []rootMoveOutput{
		{TraceRootMoveResult: search.TraceRootMoveResult{Score: 25, ScoreBound: "exact"}, MoveUCI: "a1a2"},
		{TraceRootMoveResult: search.TraceRootMoveResult{Score: 17, ScoreBound: "upper"}, MoveUCI: "b1b2"},
		{TraceRootMoveResult: search.TraceRootMoveResult{Score: 12, ScoreBound: "upper"}, MoveUCI: "c1c2"},
	}
	margin := buildRootMargin(results, "a1a2", "a1a2")
	if margin == nil {
		t.Fatal("missing root margin")
	}
	if margin.BestCompetitor.MoveUCI != "b1b2" ||
		margin.HistoricalMinusBestCompetitorCP != 8 || margin.MarginBound != "lower" {
		t.Fatalf("margin = %#v, want best upper-bound competitor and lower-bound margin 8", margin)
	}

	margin = buildRootMargin(results, "b1b2", "a1a2")
	if margin == nil || margin.HistoricalMinusBestCompetitorCP != -8 || margin.MarginBound != "upper" {
		t.Fatalf("losing historical margin = %#v, want upper-bound margin -8", margin)
	}
}

func TestParseCounterfactualCasesRequiresExactKnownNames(t *testing.T) {
	cases := []corpusCase{{Name: "Game07_ply149"}, {Name: "Game30_ply52"}}
	selected, names, err := parseCounterfactualCases(
		"Game30_ply52, Game07_ply149",
		"plus-one",
		cases,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !selected["Game07_ply149"] || !selected["Game30_ply52"] ||
		!reflect.DeepEqual(names, []string{"Game07_ply149", "Game30_ply52"}) {
		t.Fatalf("selected=%v names=%v", selected, names)
	}
	if _, _, err := parseCounterfactualCases("missing", "plus-one", cases); err == nil {
		t.Fatal("unknown case accepted")
	}
	if _, _, err := parseCounterfactualCases("Game07_ply149", "observational", cases); err == nil {
		t.Fatal("observational mode accepted a target case")
	}
}

func TestD5D3D1SelectionRejectsEveryAllowlistMechanism(t *testing.T) {
	if err := validateGeneralD5D3D1Selection(true, "observational", nil, "", false, false); err != nil {
		t.Fatalf("valid selection rejected: %v", err)
	}
	tests := []struct {
		name        string
		mode        string
		cases       []string
		bridge      string
		probe       bool
		generalD3D1 bool
	}{
		{name: "mode", mode: "full-depth"},
		{name: "case allowlist", mode: "observational", cases: []string{"Game25_ply69"}},
		{name: "bridge", mode: "observational", bridge: "Game25_ply69"},
		{name: "probe", mode: "observational", probe: true},
		{name: "old candidate", mode: "observational", generalD3D1: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateGeneralD5D3D1Selection(
				true,
				tc.mode,
				tc.cases,
				tc.bridge,
				tc.probe,
				tc.generalD3D1,
			); err == nil {
				t.Fatal("invalid allowlist-bearing selection accepted")
			}
		})
	}
}

func TestD5AblationSelectionRequiresOneExactCaseAndTwoMoveLine(t *testing.T) {
	cases := []corpusCase{{Name: "Game25_ply35"}}
	valid := []string{"d1e2", "h3h1"}
	if err := validateD5AblationSelection(true, "Game25_ply35", valid, cases); err != nil {
		t.Fatalf("valid ablation rejected: %v", err)
	}
	tests := []struct {
		name     string
		general  bool
		caseName string
		line     []string
	}{
		{name: "candidate disabled", caseName: "Game25_ply35", line: valid},
		{name: "missing case", general: true, line: valid},
		{name: "missing line", general: true, caseName: "Game25_ply35"},
		{name: "short line", general: true, caseName: "Game25_ply35", line: []string{"d1e2"}},
		{name: "unknown case", general: true, caseName: "missing", line: valid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateD5AblationSelection(tc.general, tc.caseName, tc.line, cases); err == nil {
				t.Fatal("invalid D5 ablation accepted")
			}
		})
	}
}

func TestD5ReducedFullSelectionIsObservationOnlyAndExclusive(t *testing.T) {
	if err := validateD5ReducedFullSelection(
		true, "observational", nil, "", false, false, false, "", nil,
	); err != nil {
		t.Fatalf("valid D5 shadow rejected: %v", err)
	}
	tests := []struct {
		name       string
		mode       string
		cases      []string
		bridge     string
		d3Probe    bool
		generalD3  bool
		generalD5  bool
		ablateCase string
		ablateLine []string
	}{
		{name: "counterfactual", mode: "full-depth"},
		{name: "case allowlist", mode: "observational", cases: []string{"Game25_ply35"}},
		{name: "bridge", mode: "observational", bridge: "Game25_ply35"},
		{name: "D3 probe", mode: "observational", d3Probe: true},
		{name: "D3 candidate", mode: "observational", generalD3: true},
		{name: "D5 candidate", mode: "observational", generalD5: true},
		{name: "ablation case", mode: "observational", ablateCase: "Game25_ply35"},
		{name: "ablation line", mode: "observational", ablateLine: []string{"d1e2", "h3h1"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateD5ReducedFullSelection(
				true, tc.mode, tc.cases, tc.bridge, tc.d3Probe, tc.generalD3,
				tc.generalD5, tc.ablateCase, tc.ablateLine,
			); err == nil {
				t.Fatal("invalid mixed D5 shadow accepted")
			}
		})
	}
}

func TestParseAndValidateD5AblationLine(t *testing.T) {
	tc := &corpusCase{
		Name: "Game25_ply35",
		FEN:  "r5k1/p1bn2pp/8/8/3Pn3/4P2q/PPPB1PR1/R2Q1K2 w - - 5 18",
	}
	line, err := parseAndValidateD5AblationLine(tc, []string{"d1e2", "h3h1"})
	if err != nil || len(line) != 2 {
		t.Fatalf("valid line=%v err=%v", line, err)
	}
	if _, err := parseAndValidateD5AblationLine(tc, []string{"d1e2"}); err == nil {
		t.Fatal("one-move ablation line accepted")
	}
	if _, err := parseAndValidateD5AblationLine(tc, []string{"d1e2", "a8b8"}); err == nil {
		t.Fatal("non-checking ablation reply accepted")
	}
}

func TestBuildGeneralD3D1SummaryIncludesD5Stage(t *testing.T) {
	cases := []corpusCase{
		{
			Name: "Game25_ply69",
			Runs: []run{{
				GeneralD3D1Events: []generalD3D1Output{
					{TraceGeneralD3D1Event: search.TraceGeneralD3D1Event{Stage: "d5_check_forced_full_depth"}, LineUCI: []string{"a3a4", "b7c7"}},
					{TraceGeneralD3D1Event: search.TraceGeneralD3D1Event{Stage: "d5_check_ablation_lmr_preserved"}, LineUCI: []string{"d1e2", "h3h1"}},
					{TraceGeneralD3D1Event: search.TraceGeneralD3D1Event{Stage: "d3_check_forced_full_depth"}, LineUCI: []string{"a3a4", "b7c7", "c4b3", "c7c3"}},
					{TraceGeneralD3D1Event: search.TraceGeneralD3D1Event{Stage: "d1_check_forced_no_futility"}, LineUCI: []string{"a3a4", "b7c7", "c4b3", "c7c3", "b3a2", "b4b3"}},
				},
			}},
		},
	}

	result := buildGeneralD3D1Summary(cases, true)
	if !result.Applied || result.D5Applications != 1 || result.D5Ablations != 1 ||
		result.D3Applications != 1 || result.D1Applications != 1 || result.CasesWithD5 != 1 ||
		result.CasesWithD5Ablation != 1 || result.CasesWithD3 != 1 || result.CasesWithD1 != 1 ||
		result.UniqueD5Paths != 1 || result.UniqueD5AblationPaths != 1 || result.UniqueD3Paths != 1 ||
		result.UniqueD1Paths != 1 || !reflect.DeepEqual(result.D5CaseNames, []string{"Game25_ply69"}) ||
		!reflect.DeepEqual(result.D5AblationCaseNames, []string{"Game25_ply69"}) ||
		!reflect.DeepEqual(result.D1CaseNames, []string{"Game25_ply69"}) {
		t.Fatalf("summary = %#v", result)
	}
}

func TestBuildD5ReducedFullSummary(t *testing.T) {
	cases := []corpusCase{{
		Name: "Game25_ply35",
		Runs: []run{{
			D5ReducedFullShadowNodes: 42,
			D5ReducedFullProbes: []d5ReducedFullOutput{
				{
					TraceD5ReducedFullProbe: search.TraceD5ReducedFullProbe{
						ReducedExceedsAlpha:       true,
						AlphaCrossingDisagrees:    true,
						FullMinusReduced:          -20,
						SequentialDidFullResearch: true,
						SequentialMinusFreshFull:  5,
						ParentMoveIsTT:            true,
						ChildTTKeyMatch:           true,
						ChildTTReducedCutoff:      true,
						ChildTTEntryChanged:       true,
						FirstOrderedMoveChanged:   true,
						ReducedKillersChanged:     true,
						ReducedHistoryChanged:     true,
						D3ProbeCount:              2,
						D3D1ChainCount:            1,
					},
					LineUCI: []string{"d1e2", "h3h1"},
				},
			},
		}},
	}}
	result := buildD5ReducedFullSummary(cases, true)
	if !result.Applied || result.Occurrences != 1 || result.UniquePaths != 1 ||
		result.CasesWithProbes != 1 || result.ReducedCrossesAlpha != 1 ||
		result.FreshFullCrossesAlpha != 0 || result.AlphaCrossingDisagreements != 1 ||
		result.FullScoreDifferences != 1 || result.SequentialFullResearches != 1 ||
		result.SequentialFreshDifferences != 1 || result.ParentTTMoves != 1 ||
		result.InitialChildTTHits != 1 || result.InitialReducedTTCutoffs != 1 ||
		result.ChildTTEntriesChanged != 1 || result.FirstOrderedMovesChanged != 1 ||
		result.ReducedKillersChanged != 1 || result.ReducedHistoryChanged != 1 || result.D3Probes != 2 ||
		result.D3D1Chains != 1 || result.CasesWithD3D1Chains != 1 ||
		!reflect.DeepEqual(result.D3D1ChainCaseNames, []string{"Game25_ply35"}) ||
		result.ShadowNodes != 42 {
		t.Fatalf("summary = %#v", result)
	}
}

func TestScoreDifferenceBound(t *testing.T) {
	tests := []struct {
		left  string
		right string
		want  string
	}{
		{left: "exact", right: "exact", want: "exact"},
		{left: "exact", right: "upper", want: "lower"},
		{left: "upper", right: "exact", want: "upper"},
		{left: "lower", right: "upper", want: "lower"},
		{left: "upper", right: "lower", want: "upper"},
		{left: "upper", right: "upper", want: "indeterminate"},
	}
	for _, tc := range tests {
		if got := scoreDifferenceBound(tc.left, tc.right); got != tc.want {
			t.Errorf("scoreDifferenceBound(%q,%q)=%q, want %q", tc.left, tc.right, got, tc.want)
		}
	}
}

func TestBridgeSelectionAndLineValidation(t *testing.T) {
	const fen = "1r4k1/1q4pp/8/3Pb3/1pK1P1P1/P7/2P3R1/3R4 w - - 0 35"
	tc := corpusCase{
		Name:         "Game25_ply69",
		FEN:          fen,
		RootMoveUCI:  "a3a4",
		ReplyMoveUCI: "b7c7",
	}
	line, err := parseBridgeLineText("a3a4, b7c7,c4b3,c7c3,b3a2,b4b3")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateBridgeSelection(
		tc.Name,
		line,
		"full-depth",
		map[string]bool{tc.Name: true},
		[]corpusCase{tc},
	); err != nil {
		t.Fatal(err)
	}
	path, bridgeMove, err := parseAndValidateBridgeLine(&tc, line)
	if err != nil {
		t.Fatal(err)
	}
	if len(path) != 5 || moveUCI(bridgeMove) != "b4b3" {
		t.Fatalf("bridge path/move = %d/%s, want 5/b4b3", len(path), moveUCI(bridgeMove))
	}
	if err := validateBridgeNoLMRSelection("c7c3", tc.Name, line); err != nil {
		t.Fatal(err)
	}
	noLMRPly, err := validateBridgeNoLMRMove(&tc, line, "c7c3")
	if err != nil {
		t.Fatal(err)
	}
	if noLMRPly != 3 {
		t.Fatalf("bridge NoLMR ply=%d, want 3", noLMRPly)
	}
	if err := validateBridgeNoLMRSelection("b4b3", tc.Name, line); err == nil {
		t.Fatal("bridge NoLMR selection accepted the final bridge move")
	}
	if _, err := validateBridgeNoLMRMove(&tc, line, "c4b3"); err == nil {
		t.Fatal("bridge NoLMR validation accepted a non-checking move")
	}
	if err := validateBridgeNoFutilitySelection("b4b3", "c7c3", tc.Name, line); err != nil {
		t.Fatal(err)
	}
	if err := validateBridgeNoFutilitySelection("b4b3", "", tc.Name, line); err == nil {
		t.Fatal("bridge NoFutility selection accepted a missing NoLMR dependency")
	}
	if err := validateBridgeNoFutilitySelection("c7c3", "c7c3", tc.Name, line); err == nil {
		t.Fatal("bridge NoFutility selection accepted a non-final move")
	}
	if err := validateBridgeSelection(
		tc.Name,
		line,
		"plus-one",
		map[string]bool{tc.Name: true},
		[]corpusCase{tc},
	); err == nil {
		t.Fatal("bridge selection accepted plus-one mode")
	}
}
