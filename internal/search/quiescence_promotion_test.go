package search

import (
	"testing"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/uci"
)

func TestQuiescenceSeesPromotionsAtHorizon(t *testing.T) {
	// In these positions the expected promotion leaves a quiet child with
	// no legal captures. Its static score is the exact quiescence value.
	cases := []struct {
		name string
		fen  string
		move string
	}{
		{"white_push", "8/P6k/8/8/8/8/8/4K3 w - - 0 1", "a7a8q"},
		{"black_push", "4k3/8/8/8/8/8/p6K/8 b - - 0 1", "a2a1q"},
		// c8=Q stalemates. c8=R leaves a winning rook ending.
		{"rook_underpromotion", "8/k1P5/2K5/8/8/8/8/8 w - - 0 1", "c7c8r"},
		// a8 is blocked. The promotion by capture was already searched.
		{"capture_control", "nr6/P6k/8/8/8/8/8/4K3 w - - 0 1", "a7b8q"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pos, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			m, err := uci.ParseUCIMove(pos, tc.move)
			if err != nil {
				t.Fatal(err)
			}
			standPat := Evaluate(pos)
			pos.MakeMove(m)
			want := -Evaluate(pos)
			pos.UnmakeMove()
			if want <= standPat {
				t.Fatalf("promotion score %d must improve on stand-pat %d", want, standPat)
			}
			var s Search
			got := s.Quiescence(pos, -Infinity, Infinity)
			t.Logf("%s stand-pat=%d expected=%d score=%d nodes=%d", tc.move, standPat, want, got, s.Nodes)
			if got != want {
				t.Errorf("quiescence score %d, want %d after %s", got, want, tc.move)
			}
		})
	}
}
