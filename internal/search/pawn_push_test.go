package search

import (
	"testing"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/move"
	"EquinoxChess/internal/movegen"
)

func TestPawnPushWitness(t *testing.T) {
	for _, tc := range []struct {
		fen  string
		want bool
	}{
		{board.StartFEN, true},
		{"4r2k/8/8/8/8/8/4P3/4K3 w - - 0 1", true}, // file pin preserves the shield
		{"7k/8/8/8/7b/8/5P2/4K3 w - - 0 1", false}, // diagonal pin
		{"7k/8/8/8/8/8/KP5r/8 w - - 0 1", false},   // horizontal pin
		{"7k/8/8/8/8/P7/P7/K7 w - - 0 1", true},
		{"7k/8/8/8/8/4n3/4P3/4K3 w - - 0 1", false},
		{"4k3/6p1/8/8/8/8/8/K7 b - - 0 1", true},
		{"7k/P7/8/8/8/8/8/4K3 w - - 0 1", true},
	} {
		b, err := board.FromFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		got := hasUnpinnedPawnPush(b)
		if got != tc.want {
			t.Errorf("%s: witness %v, want %v", tc.fen, got, tc.want)
		}
		if got && !InCheck(b) {
			var legal [movegen.MaxMoves]uint32
			n := movegen.GenerateLegalMoves(b, &legal)
			found := false
			for _, m := range legal[:n] {
				if (move.MovingPiece(m) == board.WhitePawn || move.MovingPiece(m) == board.BlackPawn) && !move.IsCapture(m) {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s has no legal pawn push", tc.fen)
			}
		}
	}
}
