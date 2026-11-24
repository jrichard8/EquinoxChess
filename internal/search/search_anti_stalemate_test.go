package search

import (
	"testing"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/move"
	"github.com/jrichard8/MarsGopherChess/internal/movegen"
	"github.com/jrichard8/MarsGopherChess/internal/tt"
)

// Anti-stalemate regression tests.
//
// Each FEN below is the position EquinoxChess actually faced just before
// playing a stalemating move in match matches/sf1600-run3-1+0.1-30.pgn
// (controle 1+0.1 contre Stockfish limite a 1600 Elo). In every case the
// side to move is Equinox, the position is overwhelmingly winning, and the
// "stalemateUCI" field is the exact pat-inflicting move played in the PGN.
//
// The tests assert that SearchRoot does NOT pick a move that gives stalemate
// to the opponent. They are deliberately strict: any non-stalemating legal
// move is acceptable. The depth chosen (4) matches the depth annotation seen
// in the match PGN, so a passing test means the bug is reproducibly fixed at
// the depth where the regression actually fired.

type antiStalemateCase struct {
	name         string
	fen          string
	stalemateUCI string
}

var antiStalemateCases = []antiStalemateCase{
	{
		name:         "Game3_W_39_Nxd5",
		fen:          "3k4/4R3/8/3p2B1/P1R5/2N5/2P2PPP/6K1 w - - 0 39",
		stalemateUCI: "c3d5",
	},
	{
		name:         "Game7_W_72_c3",
		fen:          "1k6/7R/P1Q5/3QNQ2/8/8/2P3P1/6K1 w - - 3 72",
		stalemateUCI: "c2c3",
	},
	{
		name:         "Game16_B_54_a5",
		fen:          "6r1/2k5/pp1p4/4n3/4q3/1q6/3K4/5r2 b - - 11 54",
		stalemateUCI: "a6a5",
	},
	{
		name:         "Game19_W_64_c3",
		fen:          "1B4Q1/7Q/2k2B1P/Q7/2B5/5R2/2P5/6K1 w - - 5 64",
		stalemateUCI: "c2c3",
	},
	{
		name:         "Game20_B_56_b6",
		fen:          "r6k/1p5p/5q2/8/3q4/1K1br3/6q1/8 b - - 15 56",
		stalemateUCI: "b7b6",
	},
	{
		name:         "Game21_W_57_g4",
		fen:          "R7/2k5/1R6/8/2PQ4/5QP1/8/7K w - - 19 57",
		stalemateUCI: "g3g4",
	},
}

func uciOf(m uint32) string {
	from := int(m & 0x3F)
	to := int((m >> 6) & 0x3F)
	promo := move.PromotionPiece(m)
	buf := [5]byte{}
	buf[0] = byte('a' + (from % 8))
	buf[1] = byte('1' + (from / 8))
	buf[2] = byte('a' + (to % 8))
	buf[3] = byte('1' + (to / 8))
	if promo != move.NoPiece {
		switch promo {
		case board.WhiteQueen, board.BlackQueen:
			buf[4] = 'q'
		case board.WhiteRook, board.BlackRook:
			buf[4] = 'r'
		case board.WhiteBishop, board.BlackBishop:
			buf[4] = 'b'
		case board.WhiteKnight, board.BlackKnight:
			buf[4] = 'n'
		}
		return string(buf[:5])
	}
	return string(buf[:4])
}

// moveLeadsToStalemate replays the move on b and reports whether the
// resulting child position is a stalemate (no legal move and not in check).
func moveLeadsToStalemate(b *board.Board, m uint32) bool {
	b.MakeMove(m)
	defer b.UnmakeMove()

	var moves [movegen.MaxMoves]uint32
	count := movegen.GeneratePseudoLegalMoves(b, &moves)
	if hasAnyLegalMove(b, &moves, count) {
		return false
	}
	return !InCheck(b)
}

func TestSearchAvoidsKnownStalematingMoves(t *testing.T) {
	for _, tc := range antiStalemateCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pos, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("FromFEN failed: %v", err)
			}

			s := Search{TT: tt.NewTT(1 << 16)}
			best, score, depth, _ := s.SearchRoot(pos, 4)

			if best == 0 {
				t.Fatalf("SearchRoot returned no move; expected a non-stalemating move (depth=%d, score=%d)", depth, score)
			}

			chosen := uciOf(best)
			if chosen == tc.stalemateUCI {
				t.Fatalf("SearchRoot picked the known pat-inflicting move %s (score=%d, depth=%d)", chosen, score, depth)
			}

			// Robustness: even if the chosen UCI string differs (e.g. the
			// engine reorders generation), the chosen move must not deliver
			// stalemate to the opponent.
			fresh, err := board.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("FromFEN failed (re-parse): %v", err)
			}
			if moveLeadsToStalemate(fresh, best) {
				t.Fatalf("SearchRoot picked a stalemating move %s (score=%d, depth=%d)", chosen, score, depth)
			}
		})
	}
}
