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
