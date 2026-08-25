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
