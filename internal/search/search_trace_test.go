//go:build searchtrace

package search

import (
	"testing"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/tt"
	"EquinoxChess/internal/uci"
)

func TestSearchTraceIsObservationalAndRecordsQuietReplyFilter(t *testing.T) {
	position, err := board.FromFEN(board.StartFEN)
	if err != nil {
		t.Fatalf("FromFEN: %v", err)
	}
	rootMove, err := uci.ParseUCIMove(position, "e2e4")
	if err != nil {
		t.Fatalf("ParseUCIMove(root): %v", err)
	}
	position.MakeMove(rootMove)
	replyMove, err := uci.ParseUCIMove(position, "e7e5")
	position.UnmakeMove()
	if err != nil {
		t.Fatalf("ParseUCIMove(reply): %v", err)
	}

	plain := Search{TT: tt.NewTT(1 << 12)}
	plainBest, plainScore, plainDepth, plainNodes := plain.SearchRoot(position, 2)

	trace := &SearchTrace{RootMove: rootMove, ReplyMove: replyMove}
	traced := Search{TT: tt.NewTT(1 << 12), Trace: trace}
	best, score, depth, nodes := traced.SearchRoot(position, 2)
	if best != plainBest || score != plainScore || depth != plainDepth || nodes != plainNodes {
		t.Fatalf(
			"trace changed search: got (%#x,%d,%d,%d), plain (%#x,%d,%d,%d)",
			best, score, depth, nodes, plainBest, plainScore, plainDepth, plainNodes,
		)
	}

	wantStages := map[string]bool{
		"root_generated":         false,
		"root_legal":             false,
		"reply_q_generated":      false,
		"reply_capture_filtered": false,
	}
	for _, event := range trace.Events {
		switch {
		case event.Iteration == 1 && event.Role == "root_move" && event.Node == "negamax" && event.Stage == "generated":
			wantStages["root_generated"] = true
		case event.Iteration == 1 && event.Role == "root_move" && event.Node == "negamax" && event.Stage == "legal":
			wantStages["root_legal"] = true
		case event.Iteration == 1 && event.Role == "reply_move" && event.Node == "quiescence" && event.Stage == "generated":
			wantStages["reply_q_generated"] = true
		case event.Iteration == 1 && event.Role == "reply_move" && event.Node == "quiescence" && event.Stage == "capture_filtered":
			wantStages["reply_capture_filtered"] = true
		}
	}
	for stage, found := range wantStages {
		if !found {
			t.Errorf("missing trace stage %s", stage)
		}
	}
}

func TestSearchTraceTargetedReplyCounterfactualRecordsContinuation(t *testing.T) {
	position, err := board.FromFEN("8/8/2pk4/3p2q1/3P1N2/8/1P2Nn1K/3B4 w - - 0 75")
	if err != nil {
		t.Fatalf("FromFEN: %v", err)
	}
	rootMove, err := uci.ParseUCIMove(position, "d1c2")
	if err != nil {
		t.Fatalf("ParseUCIMove(root): %v", err)
	}
	position.MakeMove(rootMove)
	replyMove, err := uci.ParseUCIMove(position, "f2g4")
	if err != nil {
		t.Fatalf("ParseUCIMove(reply): %v", err)
	}
	position.MakeMove(replyMove)
	if !InCheck(position) {
		t.Fatal("fixture reply must give check")
	}
	position.UnmakeMove()
	position.UnmakeMove()

	observedTrace := &SearchTrace{RootMove: rootMove, ReplyMove: replyMove}
	observed := Search{TT: tt.NewTT(1 << 16), Trace: observedTrace}
	observedBest, observedScore, observedDepth, observedNodes := observed.SearchRoot(position, 6)
	plainPosition, err := board.FromFEN("8/8/2pk4/3p2q1/3P1N2/8/1P2Nn1K/3B4 w - - 0 75")
	if err != nil {
		t.Fatalf("plain FromFEN: %v", err)
	}
	plain := Search{TT: tt.NewTT(1 << 16)}
	plainBest, plainScore, plainDepth, plainNodes := plain.SearchRoot(plainPosition, 6)
	if observedBest != plainBest || observedScore != plainScore || observedDepth != plainDepth ||
		observedNodes != plainNodes {
		t.Fatal("observational continuation trace changed the search")
	}

	observedContinuation := finalTraceContinuation(t, observedTrace, 6)
	if observedContinuation.SearchMode != "lmr_reduced" || observedContinuation.ChildDepth != 3 ||
		observedContinuation.Reduction != 1 {
		t.Fatalf("observed mode=%s childDepth=%d reduction=%d, want lmr_reduced/3/1",
			observedContinuation.SearchMode,
			observedContinuation.ChildDepth,
			observedContinuation.Reduction,
		)
	}
	assertTraceContinuation(t, observedContinuation, rootMove, replyMove)

	counterfactualPosition, err := board.FromFEN("8/8/2pk4/3p2q1/3P1N2/8/1P2Nn1K/3B4 w - - 0 75")
	if err != nil {
		t.Fatalf("counterfactual FromFEN: %v", err)
	}
	counterfactualTrace := &SearchTrace{
		RootMove:                  rootMove,
		ReplyMove:                 replyMove,
		ForceTargetReplyFullDepth: true,
		ForceTargetReplyIteration: 6,
	}
	counterfactual := Search{TT: tt.NewTT(1 << 16), Trace: counterfactualTrace}
	_, _, completedDepth, _ := counterfactual.SearchRoot(counterfactualPosition, 6)
	if completedDepth != 6 {
		t.Fatalf("counterfactual completed depth=%d, want 6", completedDepth)
	}
	for _, continuation := range counterfactualTrace.Continuations {
		if continuation.Iteration < 6 && continuation.SearchMode == "target_reply_forced_full_depth" {
			t.Fatalf("counterfactual leaked into iteration %d", continuation.Iteration)
		}
	}
	counterfactualContinuation := finalTraceContinuation(t, counterfactualTrace, 6)
	if counterfactualContinuation.SearchMode != "target_reply_forced_full_depth" ||
		counterfactualContinuation.ChildDepth != 4 || counterfactualContinuation.Reduction != 0 ||
		counterfactualContinuation.Extension != 0 {
		t.Fatalf("counterfactual mode=%s childDepth=%d reduction=%d extension=%d, want forced/4/0/0",
			counterfactualContinuation.SearchMode,
			counterfactualContinuation.ChildDepth,
			counterfactualContinuation.Reduction,
			counterfactualContinuation.Extension,
		)
	}
	assertTraceContinuation(t, counterfactualContinuation, rootMove, replyMove)

	plusOnePosition, err := board.FromFEN("8/8/2pk4/3p2q1/3P1N2/8/1P2Nn1K/3B4 w - - 0 75")
	if err != nil {
		t.Fatalf("plus-one FromFEN: %v", err)
	}
	plusOneTrace := &SearchTrace{
		RootMove:                  rootMove,
		ReplyMove:                 replyMove,
		ForceTargetReplyFullDepth: true,
		TargetReplyExtraPlies:     1,
		ForceTargetReplyIteration: 6,
	}
	plusOne := Search{TT: tt.NewTT(1 << 16), Trace: plusOneTrace}
	_, _, completedDepth, _ = plusOne.SearchRoot(plusOnePosition, 6)
	if completedDepth != 6 {
		t.Fatalf("plus-one completed depth=%d, want 6", completedDepth)
	}
	for _, continuation := range plusOneTrace.Continuations {
		if continuation.Iteration < 6 && continuation.Extension != 0 {
			t.Fatalf("plus-one leaked into iteration %d", continuation.Iteration)
		}
	}
	plusOneContinuation := finalTraceContinuation(t, plusOneTrace, 6)
	if plusOneContinuation.SearchMode != "target_reply_plus_one" ||
		plusOneContinuation.ChildDepth != 5 || plusOneContinuation.Reduction != 0 ||
		plusOneContinuation.Extension != 1 {
		t.Fatalf("plus-one mode=%s childDepth=%d reduction=%d extension=%d, want plus-one/5/0/1",
			plusOneContinuation.SearchMode,
			plusOneContinuation.ChildDepth,
			plusOneContinuation.Reduction,
			plusOneContinuation.Extension,
		)
	}
	assertTraceContinuation(t, plusOneContinuation, rootMove, replyMove)
}

func finalTraceContinuation(t *testing.T, trace *SearchTrace, iteration int) TraceContinuation {
	t.Helper()
	for index := len(trace.Continuations) - 1; index >= 0; index-- {
		if trace.Continuations[index].Iteration == iteration {
			return trace.Continuations[index]
		}
	}
	t.Fatalf("no continuation at iteration %d", iteration)
	return TraceContinuation{}
}

func assertTraceContinuation(t *testing.T, continuation TraceContinuation, rootMove, replyMove uint32) {
	t.Helper()
	if len(continuation.Line) < 3 {
		t.Fatalf("continuation line has %d moves, want at least 3", len(continuation.Line))
	}
	if continuation.Line[0] != rootMove || continuation.Line[1] != replyMove {
		t.Fatal("continuation line does not start with exact root/reply pair")
	}
	if continuation.LeafPly < 2 || continuation.LeafReason == "" {
		t.Fatalf("invalid propagated leaf ply=%d reason=%q",
			continuation.LeafPly, continuation.LeafReason)
	}
}

func TestSearchTraceCounterfactualDoesNotExemptNonCheckingReply(t *testing.T) {
	position, err := board.FromFEN("r1b2rk1/4bppp/p1q2n2/1p2p3/3QP3/P1N5/1PP1BPPP/R3R1K1 w - - 0 15")
	if err != nil {
		t.Fatalf("FromFEN: %v", err)
	}
	rootMove, err := uci.ParseUCIMove(position, "d4e5")
	if err != nil {
		t.Fatalf("ParseUCIMove(root): %v", err)
	}
	position.MakeMove(rootMove)
	replyMove, err := uci.ParseUCIMove(position, "c6e6")
	if err != nil {
		t.Fatalf("ParseUCIMove(reply): %v", err)
	}
	position.MakeMove(replyMove)
	if InCheck(position) {
		t.Fatal("fixture reply must not give check")
	}
	position.UnmakeMove()
	position.UnmakeMove()

	trace := &SearchTrace{
		RootMove:                  rootMove,
		ReplyMove:                 replyMove,
		ForceTargetReplyFullDepth: true,
		TargetReplyExtraPlies:     1,
		ForceTargetReplyIteration: 6,
	}
	s := Search{TT: tt.NewTT(1 << 16), Trace: trace}
	_, _, completedDepth, _ := s.SearchRoot(position, 6)
	if completedDepth != 6 {
		t.Fatalf("completed depth=%d, want 6", completedDepth)
	}
	continuation := finalTraceContinuation(t, trace, 6)
	if continuation.SearchMode != "lmr_reduced" {
		t.Fatalf("non-checking plus-one reply mode=%s, want lmr_reduced", continuation.SearchMode)
	}
}
