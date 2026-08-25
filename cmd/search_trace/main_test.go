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
