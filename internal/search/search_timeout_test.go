package search

import (
	"testing"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/tt"
)

func TestSearchRootStoppedBeforeFirstCompletedDepthReturnsFiniteFallback(t *testing.T) {
	// These positions come from matches/sf1600-run4-1+0.1-30.pgn, where
	// Equinox reported -300.00/4 under the short 1+0.1 time control while the
	// game was still non-terminal.
	fens := []string{
		"r4rk1/1bq2ppp/p1nbpn2/P2pN3/1ppP1N2/4P3/1PPBBPPP/R2Q1RK1 w - - 1 14",
		"r4rk1/1bq2ppp/p2bpn2/P2pn3/1ppP1N2/1P2P3/2PBBPPP/R2Q1RK1 w - - 0 15",
		"r4rk1/1bqn1ppp/p2bpn2/P2p4/1ppP1N2/1PP1P3/3BBPPP/R2Q1RK1 w - - 1 16",
	}

	for _, fen := range fens {
		fen := fen
		t.Run(fen, func(t *testing.T) {
			pos, err := board.FromFEN(fen)
			if err != nil {
				t.Fatalf("FromFEN failed: %v", err)
			}

			s := Search{TT: tt.NewTT(1 << 10)}
			s.Stop.Store(true)

			best, score, depth, _ := s.SearchRoot(pos, 4)
			if best == 0 {
				t.Fatalf("expected timeout fallback to return a legal move")
			}
			if depth != 0 {
				t.Fatalf("expected no completed depth, got depth %d", depth)
			}
			if score <= -Infinity || score >= Infinity {
				t.Fatalf("expected finite fallback score, got %d", score)
			}
		})
	}
}
