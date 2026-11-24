//go:build searchtrace && trace1850

package search

import (
	"testing"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/tt"
	"EquinoxChess/internal/uci"
)

func TestSearchTraceTargetedReplyCounterfactualRecordsContinuation(t *testing.T) {
	assertCheckingReplyTrace(t, "8/8/2pk4/3p2q1/3P1N2/8/1P2Nn1K/3B4 w - - 0 75", "d1c2", "f2g4")
}

func TestSearchTraceBridgeLocatesNestedLMRTTCutoff(t *testing.T) {
	fen, line := assertNestedLMRBridge(t)
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
	if candidateBest != line[0] || candidateScore != -644 || candidateDepth != 6 || candidateNodes != 52680 {
		t.Fatalf("nested NoLMR result=(%s,%d,%d,%d), want a4/-644/6/52680",
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
		combinedNodes != 88968 {
		t.Fatalf("combined result=(%s,%d,%d,%d), want d1d3/-728/6/88968",
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
	if probeBest != line[0] || probeScore != -644 || probeDepth != 6 || probeNodes != 52665 {
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
		generalNodes != 74918 {
		t.Fatalf("general result=(%s,%d,%d,%d), want c4d3/-741/6/74918",
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
		d5GeneralNodes != 75812 || d5Count == 0 || d5D3Count == 0 || d5D1Count == 0 || d5Leak ||
		!knownD5 || !knownD3 {
		t.Fatalf("D5D3D1 result=(%s,%d,%d,%d) d5=%d d3=%d d1=%d leak=%t known=%t/%t",
			uci.MoveToUCI(d5GeneralBest), d5GeneralScore, d5GeneralDepth, d5GeneralNodes,
			d5Count, d5D3Count, d5D1Count, d5Leak, knownD5, knownD3)
	}
}

func TestSearchTraceD5BranchAblationsStayOnExactPath(t *testing.T) {
	const fen = "r5k1/p1bn2pp/8/8/3Pn3/4P2q/PPPB1PR1/R2Q1K2 w - - 5 18"
	tests := []ablationCase{
		{name: "competitor_qh1", line: []string{"d1e2", "h3h1"}, wantNodes: 128066, wantEvents: 1},
		{name: "historical_qh1", line: []string{"d2b4", "h3h1"}, wantNodes: 114329, wantEvents: 1},
		{name: "historical_ng3", line: []string{"d2b4", "e4g3"}, wantNodes: 116645, wantEvents: 1},
		{name: "historical_nd2", line: []string{"d2b4", "e4d2"}, wantNodes: 116675, wantEvents: 1},
	}
	assertBranchAblations(t, fen, tests)
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
		withoutD5Nodes != 129186 {
		t.Fatalf("without D5 result=(%s,%d,%d,%d), want d1e2/145/6/129186",
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
		allOriginalDepth != 6 || allOriginalNodes != 130818 || !cascadeForced {
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
		pairedNodes != 126514 || pairedAblations != 2 {
		t.Fatalf("paired competitor ablation result=(%s,%d,%d,%d) ablations=%d",
			uci.MoveToUCI(pairedBest), pairedScore, pairedDepth, pairedNodes, pairedAblations)
	}
}

func TestSearchTraceD5ReducedFullShadowIsDeterministicAndNonInterfering(t *testing.T) {
	probes := assertLiveD5Shadows(t, "d1e2", 146, 134048)
	wantLines := map[string]bool{
		"d1e2,h3h1": false,
		"d1e2,e4g3": false,
		"d2b4,h3h1": false,
	}
	for _, probe := range probes {
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
