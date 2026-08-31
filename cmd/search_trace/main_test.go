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
