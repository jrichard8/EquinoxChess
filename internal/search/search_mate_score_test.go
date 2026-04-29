package search

import (
	"testing"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/tt"
)

func TestSearchRootReturnsMateScoreBeyondAlphaBetaInfinity(t *testing.T) {
	// These are the two run5 positions that still produced -300.00/4 after the
	// timeout fallback fix. They are complete depth-4 searches in forced-mate
	// sequences, so the result must stay in the mate-score band instead of being
	// clipped to the alpha-beta infinity bound.
	fens := []string{
		"r4r1k/p1p4p/7p/q1p2N2/4N3/P6P/6P1/Q1R2B1K b - - 1 31",
		"Q5R1/8/8/7k/3p4/1P1B2p1/8/5K2 b - - 0 57",
	}

	for _, fen := range fens {
		fen := fen
		t.Run(fen, func(t *testing.T) {
			pos, err := board.FromFEN(fen)
			if err != nil {
				t.Fatalf("FromFEN failed: %v", err)
			}

			s := Search{TT: tt.NewTT(1 << 16)}
			best, score, depth, _ := s.SearchRoot(pos, 4)
			if best == 0 {
				t.Fatalf("expected a legal move")
			}
			if depth != 4 {
				t.Fatalf("expected completed depth 4, got %d", depth)
			}
			if score >= -MateThreshold {
				t.Fatalf("expected losing mate score below -MateThreshold, got %d", score)
			}
			if score == -30000 {
				t.Fatalf("mate score was clipped to old infinity bound")
			}
		})
	}
}
