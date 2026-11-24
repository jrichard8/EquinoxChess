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

func assertCheckingReplyTrace(t *testing.T, fen, rootUCI, replyUCI string) {
	position, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN: %v", err)
	}
	rootMove, err := uci.ParseUCIMove(position, rootUCI)
	if err != nil {
		t.Fatalf("ParseUCIMove(root): %v", err)
	}
	position.MakeMove(rootMove)
	replyMove, err := uci.ParseUCIMove(position, replyUCI)
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
	plainPosition, err := board.FromFEN(fen)
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

	counterfactualPosition, err := board.FromFEN(fen)
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

	plusOnePosition, err := board.FromFEN(fen)
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

func assertNestedLMRBridge(t *testing.T) (string, []uint32) {
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

	return fen, line
}

type ablationCase struct {
	name       string
	line       []string
	wantNodes  uint64
	wantEvents int
}

func assertBranchAblations(t *testing.T, fen string, tests []ablationCase) {
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
				nodes != tc.wantNodes || ablationEvents != tc.wantEvents || forcedLeak {
				t.Fatalf("result=(%s,%d,%d,%d) ablations=%d leak=%t",
					uci.MoveToUCI(best), score, completedDepth, nodes, ablationEvents, forcedLeak)
			}
		})
	}

}

func assertLiveD5Shadows(t *testing.T, wantMove string, wantScore int, wantNodes uint64) []TraceD5ReducedFullProbe {
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
	if uci.MoveToUCI(best) != wantMove || score != wantScore || depth != 6 || nodes != wantNodes {
		t.Fatalf("current result=(%s,%d,%d,%d)", uci.MoveToUCI(best), score, depth, nodes)
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

	return first.D5ReducedFullProbes
}
