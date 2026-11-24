//go:build searchtrace && !trace1850

package search

import (
	"testing"

	"EquinoxChess/internal/uci"
)

func TestSearchTraceTargetedReplyCounterfactualRecordsContinuation(t *testing.T) {
	assertCheckingReplyTrace(t, "r5k1/p1bn2pp/8/8/3Pn3/4P2q/PPPB1PR1/R2Q1K2 w - - 5 18", "d2b4", "e4g3")
}

func TestSearchTraceBridgeLocatesNestedLMRTTCutoff(t *testing.T) {
	assertNestedLMRBridge(t)
}

func TestSearchTraceD5BranchAblationsStayOnExactPath(t *testing.T) {
	const fen = "r5k1/p1bn2pp/8/8/3Pn3/4P2q/PPPB1PR1/R2Q1K2 w - - 5 18"
	assertBranchAblations(t, fen, []ablationCase{
		{name: "competitor_qh1", line: []string{"d1e2", "h3h1"}, wantNodes: 406164, wantEvents: 1},
		{name: "historical_qh1", line: []string{"d2b4", "h3h1"}, wantNodes: 400573, wantEvents: 0},
		{name: "historical_ng3", line: []string{"d2b4", "e4g3"}, wantNodes: 400513, wantEvents: 1},
		{name: "historical_nd2", line: []string{"d2b4", "e4d2"}, wantNodes: 400540, wantEvents: 1},
	})
}

func TestSearchTraceD5ReducedFullShadowIsDeterministicAndNonInterfering(t *testing.T) {
	probes := assertLiveD5Shadows(t, "d2b4", 143, 325323)
	wantChains := map[string]int{"d2b4,h3h1": 8, "d2b4,e4g3": 6, "d2b4,e4d2": 2}
	for _, probe := range probes {
		if probe.Iteration != 6 || probe.Ply != 1 || probe.Reduction != 1 ||
			probe.ReducedChildDepth != 3 || probe.FullChildDepth != 4 || len(probe.Line) != 2 {
			t.Fatalf("invalid D5 probe gate: %#v", probe)
		}
		line := uci.MoveToUCI(probe.Line[0]) + "," + uci.MoveToUCI(probe.Line[1])
		if chains, ok := wantChains[line]; ok {
			if probe.D3D1ChainCount != chains || probe.Alpha != -143 || probe.ReducedScore != -143 ||
				probe.FreshFullScore != -143 || probe.SequentialScore != -143 || probe.AlphaCrossingDisagrees {
				t.Fatalf("changed D5 shadow on %s: %#v", line, probe)
			}
			delete(wantChains, line)
		}
	}
	if len(wantChains) != 0 {
		t.Fatalf("missing D5 paths: %v", wantChains)
	}
}
