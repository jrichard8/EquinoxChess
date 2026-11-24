//go:build quietverify && searchtrace && !trace1850

package search

import (
	"testing"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/tt"
)

func TestQuietDefenseRechecksBestMeasuredRejectionOnce(t *testing.T) {
	position := board.NewBoard()
	trace := &SearchTrace{}
	worker := Search{TT: tt.NewTT(1 << 16), Trace: trace}
	best, score, depth, nodes := worker.SearchRoot(position, 6)
	plain := Search{TT: tt.NewTT(1 << 16)}
	plainBest, plainScore, plainDepth, plainNodes := plain.SearchRoot(position, 6)
	if best != plainBest || score != plainScore || depth != plainDepth || nodes != plainNodes {
		t.Fatal("observing root moves changed the search")
	}

	type pass struct {
		depth int
		name  string
	}
	rejected := make(map[pass]quietDefense)
	checked := make(map[pass]bool)
	for _, row := range trace.RootMoveResults {
		key := pass{row.Iteration, row.Pass}
		if row.SearchMode == "lmr_reduced" && row.ChildBeta == -row.Alpha+quietDefenseMargin &&
			row.Score <= row.Alpha && row.Score > row.Alpha-quietDefenseMargin {
			d := rejected[key]
			d.consider(row.Move, row.Score)
			rejected[key] = d
		}
		if row.SearchMode != "root_quiet_verification" {
			continue
		}
		d := rejected[key]
		if checked[key] || row.Iteration < 5 || row.Move != d.move || !d.withinWindow(row.Alpha) ||
			row.ChildDepth != row.ParentDepth-1 || row.Reduction != 0 {
			t.Fatalf("invalid full-depth verification: %+v, challenger %+v", row, d)
		}
		checked[key] = true
	}
	if len(checked) == 0 {
		t.Fatal("fixture did not exercise verification")
	}
}
