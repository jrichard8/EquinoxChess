//go:build searchtrace

package search

import (
	"reflect"
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

func TestSearchTraceBridgeLocatesNestedLMRTTCutoff(t *testing.T) {
	const fen = "1r4k1/1q4pp/8/3Pb3/1pK1P1P1/P7/2P3R1/3R4 w - - 0 35"
	lineUCI := []string{"a3a4", "b7c7", "c4b3", "c7c3", "b3a2", "b4b3"}
	linePosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(line): %v", err)
	}
	line := make([]uint32, 0, len(lineUCI))
	for _, moveText := range lineUCI {
		m, err := uci.ParseUCIMove(linePosition, moveText)
		if err != nil {
			t.Fatalf("ParseUCIMove(%s): %v", moveText, err)
		}
		line = append(line, m)
		linePosition.MakeMove(m)
		if IsMovedSideInCheck(linePosition) {
			t.Fatalf("fixture move %s is illegal", moveText)
		}
		if moveText == "c7c3" && !InCheck(linePosition) {
			t.Fatal("fixture nested NoLMR move must give check")
		}
	}
	if !InCheck(linePosition) {
		t.Fatal("fixture bridge move must give check")
	}

	position, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(search): %v", err)
	}
	trace := &SearchTrace{
		RootMove:                  line[0],
		ReplyMove:                 line[1],
		ForceTargetReplyFullDepth: true,
		ForceTargetReplyIteration: 6,
		BridgePath:                append([]uint32(nil), line[:len(line)-1]...),
		BridgeMove:                line[len(line)-1],
		BridgeLine:                append([]uint32(nil), line...),
	}
	worker := Search{TT: tt.NewTT(1 << 16), Trace: trace}
	best, score, depth, _ := worker.SearchRoot(position, 6)
	if best != line[0] || score != -644 || depth != 6 {
		t.Fatalf("full-depth result=(%s,%d,%d), want historical a4/-644/6",
			uci.MoveToUCI(best), score, depth)
	}

	qc3Reduced := false
	ka2TTCutoff := false
	bridgeAtFinal := false
	bridgeFilteredEarlier := false
	for _, event := range trace.Events {
		if event.Iteration == 6 && event.Pass == "aspiration" && event.Role == "bridge_step" &&
			event.TargetMove == line[3] && event.Stage == "lmr_reduced" &&
			event.Depth == 3 && event.Reduction == 1 && event.ChildDepth == 1 {
			qc3Reduced = true
		}
		if event.Iteration == 6 && event.Pass == "aspiration" && event.Role == "bridge_step" &&
			event.TargetMove == line[4] && event.Stage == "tt_cutoff" && event.Depth == 1 &&
			event.TTKeyMatch && event.TTHalfmoveMatch && event.TTDepthEnough &&
			event.TTEntryDepth == 1 && event.TTStoredScore == -480 && event.TTFlag == "exact" {
			ka2TTCutoff = true
		}
		if event.Iteration == 6 && event.Role == "bridge_move" {
			bridgeAtFinal = true
		}
		if event.Iteration == 5 && event.Role == "bridge_move" &&
			event.TargetMove == line[5] && event.Node == "quiescence" && event.Stage == "capture_filtered" {
			bridgeFilteredEarlier = true
		}
	}
	if !qc3Reduced || !ka2TTCutoff || bridgeAtFinal || !bridgeFilteredEarlier {
		t.Fatalf("bridge trace reduced=%t ttCutoff=%t bridgeFinal=%t qFilteredEarlier=%t",
			qc3Reduced, ka2TTCutoff, bridgeAtFinal, bridgeFilteredEarlier)
	}

	candidatePosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(candidate): %v", err)
	}
	candidateTrace := &SearchTrace{
		RootMove:                  line[0],
		ReplyMove:                 line[1],
		ForceTargetReplyFullDepth: true,
		ForceTargetReplyIteration: 6,
		ForceBridgeStepNoLMR:      true,
		BridgeNoLMRPly:            3,
		ForceBridgeNoLMRIteration: 6,
		BridgePath:                append([]uint32(nil), line[:len(line)-1]...),
		BridgeMove:                line[len(line)-1],
		BridgeLine:                append([]uint32(nil), line...),
	}
	candidateWorker := Search{TT: tt.NewTT(1 << 16), Trace: candidateTrace}
	candidateBest, candidateScore, candidateDepth, candidateNodes :=
		candidateWorker.SearchRoot(candidatePosition, 6)
	if candidateBest != line[0] || candidateScore != -644 || candidateDepth != 6 || candidateNodes != 54331 {
		t.Fatalf("nested NoLMR result=(%s,%d,%d,%d), want historical a4/-644/6/54331",
			uci.MoveToUCI(candidateBest), candidateScore, candidateDepth, candidateNodes)
	}
	forcedAtTarget := false
	forcedLeak := false
	bridgeFutility := false
	qc3ReturnedAtAlpha := false
	for _, event := range candidateTrace.Events {
		if event.Stage == "bridge_step_forced_full_depth" {
			if event.Iteration == 6 && event.Role == "bridge_step" && event.TargetMove == line[3] &&
				event.CurrentMove == line[3] && event.Depth == 3 && event.ChildDepth == 2 &&
				event.Reduction == 0 && event.Extension == 0 {
				forcedAtTarget = true
			} else {
				forcedLeak = true
			}
		}
		if event.Iteration == 6 && event.Pass == "aspiration" && event.Role == "bridge_move" &&
			event.TargetMove == line[5] && event.Stage == "futility_pruned" && event.Depth == 1 &&
			event.OrderIndex == 3 && event.Alpha == 644 && event.Score == 480 {
			bridgeFutility = true
		}
		if event.Iteration == 6 && event.Pass == "aspiration" && event.Role == "bridge_step" &&
			event.TargetMove == line[3] && event.Stage == "search_complete" && event.Score == 644 {
			qc3ReturnedAtAlpha = true
		}
	}
	if !forcedAtTarget || forcedLeak || !bridgeFutility || !qc3ReturnedAtAlpha {
		t.Fatalf("nested NoLMR forced=%t leak=%t bridgeFutility=%t qc3AtAlpha=%t",
			forcedAtTarget, forcedLeak, bridgeFutility, qc3ReturnedAtAlpha)
	}

	combinedPosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(combined): %v", err)
	}
	combinedTrace := &SearchTrace{
		RootMove:                  line[0],
		ReplyMove:                 line[1],
		ForceTargetReplyFullDepth: true,
		ForceTargetReplyIteration: 6,
		ForceBridgeStepNoLMR:      true,
		BridgeNoLMRPly:            3,
		ForceBridgeNoLMRIteration: 6,
		ForceBridgeNoFutility:     true,
		ForceBridgeNoFutilityIter: 6,
		BridgePath:                append([]uint32(nil), line[:len(line)-1]...),
		BridgeMove:                line[len(line)-1],
		BridgeLine:                append([]uint32(nil), line...),
	}
	combinedWorker := Search{TT: tt.NewTT(1 << 16), Trace: combinedTrace}
	combinedBest, combinedScore, combinedDepth, combinedNodes := combinedWorker.SearchRoot(combinedPosition, 6)
	if uci.MoveToUCI(combinedBest) != "d1d3" || combinedScore != -728 || combinedDepth != 6 ||
		combinedNodes != 91546 {
		t.Fatalf("combined result=(%s,%d,%d,%d), want d1d3/-728/6/91546",
			uci.MoveToUCI(combinedBest), combinedScore, combinedDepth, combinedNodes)
	}
	forcedNoFutility := false
	forcedNoFutilityLeak := false
	bridgeLegal := false
	bridgePruned := false
	for _, event := range combinedTrace.Events {
		if event.Stage == "bridge_move_forced_no_futility" {
			if event.Iteration == 6 && event.Pass == "aspiration" && event.Role == "bridge_move" &&
				event.TargetMove == line[5] && event.CurrentMove == line[5] && event.Ply == 5 &&
				event.Depth == 1 && event.OrderIndex == 3 && event.Alpha == 644 && event.Score == 480 {
				forcedNoFutility = true
			} else {
				forcedNoFutilityLeak = true
			}
		}
		if event.Iteration == 6 && event.Pass == "aspiration" && event.Role == "bridge_move" &&
			event.TargetMove == line[5] && event.Stage == "legal" {
			bridgeLegal = true
		}
		if event.Iteration == 6 && event.Pass == "aspiration" && event.Role == "bridge_move" &&
			event.TargetMove == line[5] && event.Stage == "futility_pruned" {
			bridgePruned = true
		}
	}
	if !forcedNoFutility || forcedNoFutilityLeak || !bridgeLegal || bridgePruned {
		t.Fatalf("combined NoFutility forced=%t leak=%t bridgeLegal=%t bridgePruned=%t",
			forcedNoFutility, forcedNoFutilityLeak, bridgeLegal, bridgePruned)
	}
	combinedContinuation := finalTraceContinuation(t, combinedTrace, 6)
	wantContinuation := []string{
		"a3a4", "b7c7", "c4b3", "c7c3", "b3a2", "b4b3", "c2b3", "c3b3",
	}
	if combinedContinuation.Pass != "full_window_research" || combinedContinuation.Score != 754 ||
		combinedContinuation.LeafReason != "qsearch_checkmate" || combinedContinuation.LeafScore != -31992 ||
		len(combinedContinuation.Line) != len(wantContinuation) {
		t.Fatalf("combined continuation pass=%s score=%d leaf=%s/%d moves=%d",
			combinedContinuation.Pass, combinedContinuation.Score, combinedContinuation.LeafReason,
			combinedContinuation.LeafScore, len(combinedContinuation.Line))
	}
	for index, moveText := range wantContinuation {
		if got := uci.MoveToUCI(combinedContinuation.Line[index]); got != moveText {
			t.Fatalf("combined continuation move %d=%s, want %s", index, got, moveText)
		}
	}

	probePosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(probe): %v", err)
	}
	probeTrace := &SearchTrace{
		RootMove:                  line[0],
		ReplyMove:                 line[1],
		ForceTargetReplyFullDepth: true,
		ForceTargetReplyIteration: 6,
		ProbeD3CheckChains:        true,
		ProbeD3CheckIteration:     6,
	}
	probeWorker := Search{TT: tt.NewTT(1 << 16), Trace: probeTrace}
	probeBest, probeScore, probeDepth, probeNodes := probeWorker.SearchRoot(probePosition, 6)
	if probeBest != line[0] || probeScore != -644 || probeDepth != 6 || probeNodes != 54316 {
		t.Fatalf("shadow probe changed main result=(%s,%d,%d,%d)",
			uci.MoveToUCI(probeBest), probeScore, probeDepth, probeNodes)
	}
	probeFound := false
	chainFound := false
	for _, probe := range probeTrace.D3CheckProbes {
		if probe.Iteration == 6 && probe.Pass == "aspiration" && probe.Ply == 3 &&
			probe.Move == line[3] && probe.ReducedChildDepth == 1 && probe.FullChildDepth == 2 &&
			probe.FullDepthScore == 644 && probe.ShadowNodes == 7 && probe.FutilityChecks == 6 {
			probeFound = true
		}
	}
	for _, chain := range probeTrace.D3CheckFutilityChains {
		if chain.Iteration == 6 && chain.Pass == "aspiration" && chain.D3Ply == 3 &&
			chain.D3Move == line[3] && chain.D1Ply == 5 && chain.D1Move == line[5] &&
			chain.StaticEval == 480 && chain.Alpha == 644 {
			chainFound = true
		}
	}
	if !probeFound || !chainFound {
		t.Fatalf("structural shadow probe foundD3=%t foundChain=%t", probeFound, chainFound)
	}

	generalPosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(general): %v", err)
	}
	generalTrace := &SearchTrace{
		RootMove:                  line[0],
		ReplyMove:                 line[1],
		ForceTargetReplyFullDepth: true,
		ForceTargetReplyIteration: 6,
		ForceGeneralD3D1Checks:    true,
		ForceGeneralD3D1Iteration: 6,
	}
	generalWorker := Search{TT: tt.NewTT(1 << 16), Trace: generalTrace}
	generalBest, generalScore, generalDepth, generalNodes := generalWorker.SearchRoot(generalPosition, 6)
	if uci.MoveToUCI(generalBest) != "c4d3" || generalScore != -741 || generalDepth != 6 ||
		generalNodes != 77748 {
		t.Fatalf("general result=(%s,%d,%d,%d), want c4d3/-741/6/77748",
			uci.MoveToUCI(generalBest), generalScore, generalDepth, generalNodes)
	}
	generalD3Count := 0
	generalD1Count := 0
	generalLeak := false
	for _, event := range generalTrace.GeneralD3D1Events {
		switch event.Stage {
		case "d3_check_forced_full_depth":
			generalD3Count++
			generalLeak = generalLeak || event.Iteration != 6 || event.Depth != 3 || event.ChildDepth != 2
		case "d1_check_forced_no_futility":
			generalD1Count++
			generalLeak = generalLeak || event.Iteration != 6 || event.Depth != 1 || event.ChildDepth != 0
		default:
			generalLeak = true
		}
	}
	for _, d1Event := range generalTrace.GeneralD3D1Events {
		if d1Event.Stage != "d1_check_forced_no_futility" {
			continue
		}
		coupled := false
		for _, d3Event := range generalTrace.GeneralD3D1Events {
			if d3Event.Stage != "d3_check_forced_full_depth" ||
				d3Event.Iteration != d1Event.Iteration || d3Event.Pass != d1Event.Pass ||
				d3Event.Ply != d1Event.Ply-2 || len(d3Event.Line) > len(d1Event.Line) {
				continue
			}
			prefixMatches := true
			for index := range d3Event.Line {
				if d3Event.Line[index] != d1Event.Line[index] {
					prefixMatches = false
					break
				}
			}
			if prefixMatches {
				coupled = true
				break
			}
		}
		if !coupled {
			generalLeak = true
			break
		}
	}
	if generalD3Count == 0 || generalD1Count == 0 || generalLeak {
		t.Fatalf("general candidate d3=%d d1=%d leak=%t",
			generalD3Count, generalD1Count, generalLeak)
	}

	d5GeneralPosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(D5D3D1): %v", err)
	}
	d5GeneralTrace := &SearchTrace{
		RootMove:                  line[0],
		ReplyMove:                 line[1],
		ForceGeneralD5D3D1Checks:  true,
		ForceGeneralD3D1Iteration: 6,
	}
	d5GeneralWorker := Search{TT: tt.NewTT(1 << 16), Trace: d5GeneralTrace}
	d5GeneralBest, d5GeneralScore, d5GeneralDepth, d5GeneralNodes :=
		d5GeneralWorker.SearchRoot(d5GeneralPosition, 6)
	d5Count := 0
	d5D3Count := 0
	d5D1Count := 0
	d5Leak := false
	knownD5 := false
	knownD3 := false
	for _, event := range d5GeneralTrace.GeneralD3D1Events {
		switch event.Stage {
		case "d5_check_forced_full_depth":
			d5Count++
			d5Leak = d5Leak || event.Iteration != 6 || event.Depth != 5 || event.ChildDepth != 4
			knownD5 = knownD5 || event.Ply == 1 && event.Move == line[1]
		case "d3_check_forced_full_depth":
			d5D3Count++
			d5Leak = d5Leak || event.Iteration != 6 || event.Depth != 3 || event.ChildDepth != 2
			knownD3 = knownD3 || event.Ply == 3 && event.Move == line[3]
		case "d1_check_forced_no_futility":
			d5D1Count++
			d5Leak = d5Leak || event.Iteration != 6 || event.Depth != 1 || event.ChildDepth != 0
		default:
			d5Leak = true
		}
	}
	for _, d1Event := range d5GeneralTrace.GeneralD3D1Events {
		if d1Event.Stage != "d1_check_forced_no_futility" {
			continue
		}
		coupled := false
		for _, d3Event := range d5GeneralTrace.GeneralD3D1Events {
			if d3Event.Stage != "d3_check_forced_full_depth" ||
				d3Event.Iteration != d1Event.Iteration || d3Event.Pass != d1Event.Pass ||
				d3Event.Ply != d1Event.Ply-2 || len(d3Event.Line) > len(d1Event.Line) {
				continue
			}
			prefixMatches := true
			for index := range d3Event.Line {
				if d3Event.Line[index] != d1Event.Line[index] {
					prefixMatches = false
					break
				}
			}
			if prefixMatches {
				coupled = true
				break
			}
		}
		if !coupled {
			d5Leak = true
			break
		}
	}
	if uci.MoveToUCI(d5GeneralBest) != "c4d3" || d5GeneralScore != -741 || d5GeneralDepth != 6 ||
		d5GeneralNodes != 78659 || d5Count == 0 || d5D3Count == 0 || d5D1Count == 0 || d5Leak ||
		!knownD5 || !knownD3 {
		t.Fatalf("D5D3D1 result=(%s,%d,%d,%d) d5=%d d3=%d d1=%d leak=%t known=%t/%t",
			uci.MoveToUCI(d5GeneralBest), d5GeneralScore, d5GeneralDepth, d5GeneralNodes,
			d5Count, d5D3Count, d5D1Count, d5Leak, knownD5, knownD3)
	}
}

func TestTraceBridgeStepNoLMRRequiresExactGate(t *testing.T) {
	line := []uint32{11, 22, 33, 44, 55, 66}
	trace := &SearchTrace{
		ForceBridgeStepNoLMR:      true,
		BridgeNoLMRPly:            3,
		ForceBridgeNoLMRIteration: 6,
		BridgeLine:                line,
		iteration:                 6,
	}
	copy(trace.path[:], line[:4])
	worker := Search{Trace: trace}
	if !worker.traceBridgeStepNoLMR(3, line[3]) {
		t.Fatal("exact nested NoLMR gate did not match")
	}
	trace.iteration = 5
	if worker.traceBridgeStepNoLMR(3, line[3]) {
		t.Fatal("nested NoLMR leaked into an earlier iteration")
	}
	trace.iteration = 6
	trace.path[2] = 99
	if worker.traceBridgeStepNoLMR(3, line[3]) {
		t.Fatal("nested NoLMR accepted a different path prefix")
	}
	trace.path[2] = line[2]
	if worker.traceBridgeStepNoLMR(3, 99) || worker.traceBridgeStepNoLMR(2, line[2]) ||
		worker.traceBridgeStepNoLMR(5, line[5]) {
		t.Fatal("nested NoLMR accepted a different move, ply, or final bridge move")
	}
}

func TestSearchTraceD5BranchAblationsStayOnExactPath(t *testing.T) {
	const fen = "r5k1/p1bn2pp/8/8/3Pn3/4P2q/PPPB1PR1/R2Q1K2 w - - 5 18"
	tests := []struct {
		name      string
		line      []string
		wantNodes uint64
	}{
		{name: "competitor_qh1", line: []string{"d1e2", "h3h1"}, wantNodes: 147082},
		{name: "historical_qh1", line: []string{"d2b4", "h3h1"}, wantNodes: 131803},
		{name: "historical_ng3", line: []string{"d2b4", "e4g3"}, wantNodes: 133612},
		{name: "historical_nd2", line: []string{"d2b4", "e4d2"}, wantNodes: 133642},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			linePosition, err := board.FromFEN(fen)
			if err != nil {
				t.Fatalf("FromFEN(line): %v", err)
			}
			line := make([]uint32, 0, len(tc.line))
			for _, moveText := range tc.line {
				m, err := uci.ParseUCIMove(linePosition, moveText)
				if err != nil {
					t.Fatalf("ParseUCIMove(%s): %v", moveText, err)
				}
				line = append(line, m)
				linePosition.MakeMove(m)
				if IsMovedSideInCheck(linePosition) {
					t.Fatalf("fixture move %s is illegal", moveText)
				}
			}
			if !InCheck(linePosition) {
				t.Fatal("fixture ablation reply must give check")
			}

			position, err := board.FromFEN(fen)
			if err != nil {
				t.Fatalf("FromFEN(search): %v", err)
			}
			trace := &SearchTrace{
				ForceGeneralD5D3D1Checks:  true,
				ForceGeneralD3D1Iteration: 6,
				GeneralD5AblationPaths:    [][]uint32{line},
			}
			worker := Search{TT: tt.NewTT(1 << 16), Trace: trace}
			best, score, completedDepth, nodes := worker.SearchRoot(position, 6)

			ablationEvents := 0
			forcedLeak := false
			for _, event := range trace.GeneralD3D1Events {
				exactLine := len(event.Line) == len(line)
				if exactLine {
					for index := range line {
						exactLine = exactLine && event.Line[index] == line[index]
					}
				}
				if event.Stage == "d5_check_ablation_lmr_preserved" {
					if exactLine && event.Iteration == 6 && event.Depth == 5 &&
						event.ChildDepth == 3 && event.Reduction == 1 {
						ablationEvents++
					} else {
						forcedLeak = true
					}
				}
				if event.Stage == "d5_check_forced_full_depth" && exactLine {
					forcedLeak = true
				}
			}
			if uci.MoveToUCI(best) != "d2b4" || score != 135 || completedDepth != 6 ||
				nodes != tc.wantNodes || ablationEvents != 1 || forcedLeak {
				t.Fatalf("result=(%s,%d,%d,%d) ablations=%d leak=%t",
					uci.MoveToUCI(best), score, completedDepth, nodes, ablationEvents, forcedLeak)
			}
		})
	}

	withoutD5Position, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(without D5): %v", err)
	}
	withoutD5Trace := &SearchTrace{
		ForceGeneralD3D1Checks:    true,
		ForceGeneralD3D1Iteration: 6,
	}
	withoutD5Worker := Search{TT: tt.NewTT(1 << 16), Trace: withoutD5Trace}
	withoutD5Best, withoutD5Score, withoutD5Depth, withoutD5Nodes :=
		withoutD5Worker.SearchRoot(withoutD5Position, 6)
	if uci.MoveToUCI(withoutD5Best) != "d1e2" || withoutD5Score != 145 || withoutD5Depth != 6 ||
		withoutD5Nodes != 148952 {
		t.Fatalf("without D5 result=(%s,%d,%d,%d), want d1e2/145/6/148952",
			uci.MoveToUCI(withoutD5Best), withoutD5Score, withoutD5Depth, withoutD5Nodes)
	}

	encodedLines := make([][]uint32, 0, len(tests))
	for _, tc := range tests {
		linePosition, err := board.FromFEN(fen)
		if err != nil {
			t.Fatalf("FromFEN(subset line): %v", err)
		}
		line := make([]uint32, 0, len(tc.line))
		for _, moveText := range tc.line {
			m, err := uci.ParseUCIMove(linePosition, moveText)
			if err != nil {
				t.Fatalf("ParseUCIMove(subset %s): %v", moveText, err)
			}
			line = append(line, m)
			linePosition.MakeMove(m)
		}
		encodedLines = append(encodedLines, line)
	}
	cascadeLinePosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(cascade line): %v", err)
	}
	cascadeLine := []uint32{}
	for _, moveText := range []string{"d1e2", "e4g3"} {
		m, err := uci.ParseUCIMove(cascadeLinePosition, moveText)
		if err != nil {
			t.Fatalf("ParseUCIMove(cascade %s): %v", moveText, err)
		}
		cascadeLine = append(cascadeLine, m)
		cascadeLinePosition.MakeMove(m)
	}
	allOriginalPosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(all original ablations): %v", err)
	}
	allOriginalTrace := &SearchTrace{
		ForceGeneralD5D3D1Checks:  true,
		ForceGeneralD3D1Iteration: 6,
		GeneralD5AblationPaths:    encodedLines,
	}
	allOriginalWorker := Search{TT: tt.NewTT(1 << 16), Trace: allOriginalTrace}
	allOriginalBest, allOriginalScore, allOriginalDepth, allOriginalNodes :=
		allOriginalWorker.SearchRoot(allOriginalPosition, 6)
	cascadeForced := false
	for _, event := range allOriginalTrace.GeneralD3D1Events {
		if event.Stage != "d5_check_forced_full_depth" || len(event.Line) != len(cascadeLine) {
			continue
		}
		matches := true
		for index := range cascadeLine {
			matches = matches && event.Line[index] == cascadeLine[index]
		}
		cascadeForced = cascadeForced || matches
	}
	if uci.MoveToUCI(allOriginalBest) != "d2b4" || allOriginalScore != 135 ||
		allOriginalDepth != 6 || allOriginalNodes != 150561 || !cascadeForced {
		t.Fatalf("all original ablations result=(%s,%d,%d,%d) cascade=%t",
			uci.MoveToUCI(allOriginalBest), allOriginalScore, allOriginalDepth,
			allOriginalNodes, cascadeForced)
	}

	pairedPosition, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN(paired competitor ablation): %v", err)
	}
	pairedTrace := &SearchTrace{
		ForceGeneralD5D3D1Checks:  true,
		ForceGeneralD3D1Iteration: 6,
		GeneralD5AblationPaths:    [][]uint32{encodedLines[0], cascadeLine},
	}
	pairedWorker := Search{TT: tt.NewTT(1 << 16), Trace: pairedTrace}
	pairedBest, pairedScore, pairedDepth, pairedNodes := pairedWorker.SearchRoot(pairedPosition, 6)
	pairedAblations := 0
	for _, event := range pairedTrace.GeneralD3D1Events {
		if event.Stage == "d5_check_ablation_lmr_preserved" {
			pairedAblations++
		}
	}
	if uci.MoveToUCI(pairedBest) != "d1e2" || pairedScore != 145 || pairedDepth != 6 ||
		pairedNodes != 145619 || pairedAblations != 2 {
		t.Fatalf("paired competitor ablation result=(%s,%d,%d,%d) ablations=%d",
			uci.MoveToUCI(pairedBest), pairedScore, pairedDepth, pairedNodes, pairedAblations)
	}
}

func TestSearchTraceD5ReducedFullShadowIsDeterministicAndNonInterfering(t *testing.T) {
	const fen = "r5k1/p1bn2pp/8/8/3Pn3/4P2q/PPPB1PR1/R2Q1K2 w - - 5 18"
	run := func(traced bool) (uint32, int, int, uint64, *SearchTrace) {
		position, err := board.FromFEN(fen)
		if err != nil {
			t.Fatalf("FromFEN: %v", err)
		}
		var trace *SearchTrace
		if traced {
			trace = &SearchTrace{ProbeD5ReducedFull: true, ProbeD5Iteration: 6}
		}
		worker := Search{TT: tt.NewTT(1 << 16), Trace: trace}
		best, score, depth, nodes := worker.SearchRoot(position, 6)
		return best, score, depth, nodes, trace
	}

	best, score, depth, nodes, first := run(true)
	plainBest, plainScore, plainDepth, plainNodes, _ := run(false)
	secondBest, secondScore, secondDepth, secondNodes, second := run(true)
	if uci.MoveToUCI(best) != "d1e2" || score != 146 || depth != 6 || nodes != 155300 {
		t.Fatalf("baseline result=(%s,%d,%d,%d), want d1e2/146/6/155300",
			uci.MoveToUCI(best), score, depth, nodes)
	}
	if best != plainBest || score != plainScore || depth != plainDepth || nodes != plainNodes {
		t.Fatalf("trace interference: traced=(%s,%d,%d,%d) plain=(%s,%d,%d,%d)",
			uci.MoveToUCI(best), score, depth, nodes,
			uci.MoveToUCI(plainBest), plainScore, plainDepth, plainNodes)
	}
	if best != secondBest || score != secondScore || depth != secondDepth || nodes != secondNodes ||
		!reflect.DeepEqual(first.D5ReducedFullProbes, second.D5ReducedFullProbes) {
		t.Fatalf("shadow nondeterminism: first=(%s,%d,%d,%d,%d probes) second=(%s,%d,%d,%d,%d probes)",
			uci.MoveToUCI(best), score, depth, nodes, len(first.D5ReducedFullProbes),
			uci.MoveToUCI(secondBest), secondScore, secondDepth, secondNodes,
			len(second.D5ReducedFullProbes))
	}
	if len(first.D5ReducedFullProbes) == 0 {
		t.Fatal("no D5 reduced/full probes recorded")
	}

	wantLines := map[string]bool{
		"d1e2,h3h1": false,
		"d1e2,e4g3": false,
		"d2b4,h3h1": false,
	}
	for _, probe := range first.D5ReducedFullProbes {
		if probe.Iteration != 6 || probe.Ply != 1 || probe.Reduction != 1 ||
			probe.ReducedChildDepth != 3 || probe.FullChildDepth != 4 || len(probe.Line) != 2 {
			t.Fatalf("invalid D5 probe gate: %#v", probe)
		}
		line := uci.MoveToUCI(probe.Line[0]) + "," + uci.MoveToUCI(probe.Line[1])
		if _, ok := wantLines[line]; ok {
			wantLines[line] = true
			t.Logf("%s reduced=%d fresh=%d sequential=%d alpha=%d crossings=%t/%t d3=%d chains=%d",
				line, probe.ReducedScore, probe.FreshFullScore, probe.SequentialScore,
				probe.Alpha, probe.ReducedExceedsAlpha, probe.FreshFullExceedsAlpha,
				probe.D3ProbeCount, probe.D3D1ChainCount)
		}
		if line == "d1e2,h3h1" && (probe.Alpha != -146 || probe.ReducedScore != -146 ||
			probe.FreshFullScore != -145 || probe.SequentialScore != -146 ||
			probe.ReducedExceedsAlpha || !probe.FreshFullExceedsAlpha ||
			!probe.AlphaCrossingDisagrees || probe.FullMinusReduced != 1 ||
			probe.SequentialMinusFreshFull != -1 || probe.D3D1ChainCount != 5) {
			t.Errorf("Qe2 Qh1+ shadow mismatch: %#v", probe)
		}
		if line == "d1e2,e4g3" && (probe.ReducedScore != -146 ||
			probe.FreshFullScore != -146 || probe.AlphaCrossingDisagrees ||
			probe.D3D1ChainCount != 8) {
			t.Errorf("Qe2 Ng3+ shadow mismatch: %#v", probe)
		}
	}
	for line, found := range wantLines {
		if !found {
			t.Errorf("missing expected D5 path %s", line)
		}
	}
}

func TestTraceBridgeNoFutilityRequiresExactGate(t *testing.T) {
	line := []uint32{11, 22, 33, 44, 55, 66}
	trace := &SearchTrace{
		ForceBridgeNoFutility:     true,
		ForceBridgeNoFutilityIter: 6,
		BridgePath:                append([]uint32(nil), line[:5]...),
		BridgeMove:                line[5],
		iteration:                 6,
	}
	copy(trace.path[:], line[:5])
	worker := Search{Trace: trace}
	if !worker.traceBridgeNoFutility(5, line[5]) {
		t.Fatal("exact bridge NoFutility gate did not match")
	}
	trace.iteration = 5
	if worker.traceBridgeNoFutility(5, line[5]) {
		t.Fatal("bridge NoFutility leaked into an earlier iteration")
	}
	trace.iteration = 6
	trace.path[4] = 99
	if worker.traceBridgeNoFutility(5, line[5]) {
		t.Fatal("bridge NoFutility accepted a different path prefix")
	}
	trace.path[4] = line[4]
	if worker.traceBridgeNoFutility(5, 99) || worker.traceBridgeNoFutility(4, line[5]) {
		t.Fatal("bridge NoFutility accepted a different move or ply")
	}
}
