package search

import (
	"testing"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/tt"
)

// Anti-blunder regression tests (Phase 2 canary).
//
// Each FEN below is the position EquinoxChess actually faced just before
// playing a tactical-blunder move in match matches/sf1600-run6-1+0.1-30.pgn
// (controle 1+0.1 contre Stockfish limite a 1600 Elo). In every case Equinox
// is the side to move, and the PGN annotation right after the actual move
// played at depth 4 reported `pgnDepth4Score` from Equinox's POV — a value
// that hid the impending tactical disaster.
//
// The canary asserts that with deeper search (depth >= 6), Equinox sees the
// trap and reports a losing score (<= -300 cp from side-to-move POV) or a
// mate-against signal. A passing test means more depth is enough to fix
// the blunder; a failing test means the engine genuinely lacks the horizon
// (ordering, quiescence, or pruning issues — all Phase 2 levers).
//
// The 13 cases are: 7 balanced_blunder (Equinox at equilibrium then crashed),
// 4 short conversion_blunder (Equinox at +250cp peak then crashed within
// ~50 plies of the peak), 1 unclassified-loose (Game03), and 1 horizon-no-crash
// (Game02 — the game ended before Equinox could even annotate the crash).

type tacticalBlunderCase struct {
	name           string
	fen            string
	pgnDepth4Score string
}

var tacticalBlunderCases = []tacticalBlunderCase{
	{"Game02_B_ply40", "r4r1k/1pp3pp/p1b1p3/3pP1Nn/P2P3P/qPP1p3/2Q2P2/1K2RBR1 b - - 3 20", "+1.48"},
	{"Game03_W_ply29", "r1b2rk1/4bppp/p1q2n2/1p2p3/3QP3/P1N5/1PP1BPPP/R3R1K1 w - - 0 15", "-1.54"},
	{"Game06_B_ply46", "2rr4/1ppN1pkp/8/p7/8/2Q1PN2/pB3P1K/3q4 b - - 1 23", "+3.04"},
	{"Game09_W_ply127", "8/8/RP1B4/3b4/8/P3k2p/3r4/6K1 w - - 15 64", "+0.82"},
	{"Game16_B_ply62", "4b2k/1Bp2ppp/P3p3/r7/2p3P1/8/5K1P/3R4 b - - 2 31", "+3.03"},
	{"Game17_W_ply17", "rn1q1rk1/pb2bppp/1p2pn2/2ppN3/3P4/2NBP3/PPPB1PPP/R2Q1RK1 w - - 2 9", "+1.69"},
	{"Game21_W_ply87", "8/8/8/1PN5/5kpp/8/5PP1/3r3K w - - 2 44", "-0.36"},
	{"Game22_B_ply36", "r2q1rk1/p1pb1ppp/1p2p3/2bpP1N1/4nBPP/2P5/PP1N3Q/R3KB1R b KQ - 5 18", "-0.03"},
	{"Game25_W_ply49", "r5k1/p1b3pp/7q/8/1B1Pn1P1/4P3/PPP1K3/R2Q2R1 w - - 5 25", "+4.86"},
	{"Game26_B_ply30", "r1b1kb1r/pp4pp/2p5/1N2p3/6p1/P2BP3/1PQ2PqP/2KR2R1 b kq - 1 15", "+2.48"},
	{"Game27_W_ply29", "r1b2rk1/2p1qpp1/2pp1b1p/p7/2B1PBn1/2N2Q2/PPP2PPP/R3R1K1 w - - 0 15", "+1.63"},
	{"Game28_B_ply48", "r3r1k1/1pp4p/3p1Q2/1qnP1R2/p3p3/PP4PP/2P3P1/4R1K1 b - - 0 24", "+1.64"},
	{"Game30_B_ply26", "r4rk1/pppbbppp/2n1p3/3pP3/3Pq2P/P1P5/2P1BPP1/R1BQ1KNR b - - 0 13", "+1.51"},
}

const tacticalLossThreshold = -300

func TestSearchSeesLossingTacticAtSwingPly(t *testing.T) {
	for _, tc := range tacticalBlunderCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pos, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("FromFEN failed: %v", err)
			}

			s := Search{TT: tt.NewTT(1 << 16)}
			_, score, depth, _ := s.SearchRoot(pos, 6)

			if score > tacticalLossThreshold {
				t.Fatalf("expected score <= %d at depth >= 6 (Equinox should see the tactic; PGN reported %s at depth 4), got score=%d at depth=%d",
					tacticalLossThreshold, tc.pgnDepth4Score, score, depth)
			}
		})
	}
}
