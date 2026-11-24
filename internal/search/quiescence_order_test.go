package search

import (
	"slices"
	"testing"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/move"
)

func TestQuiescenceCaptureOrder(t *testing.T) {
	ep := move.NewMove(board.E5, board.D6, board.WhitePawn, board.BlackPawn, move.NoPiece, move.FlagEnPassant)
	equal := move.NewMove(board.E5, board.F6, board.WhitePawn, board.BlackPawn, move.NoPiece, 0)
	promo := move.NewMove(board.A7, board.B8, board.WhitePawn, board.BlackRook, board.WhiteQueen, 0)
	quiet := move.NewMove(board.A7, board.A8, board.WhitePawn, move.NoPiece, board.WhiteQueen, 0)
	push := move.NewMove(board.H2, board.H3, board.WhitePawn, move.NoPiece, move.NoPiece, 0)
	moves := []uint32{quiet, ep, push, equal, promo}
	n := orderQuiescenceCaptures(&Search{}, moves)
	if want := []uint32{promo, ep, equal}; !slices.Equal(moves[:n], want) {
		t.Fatalf("captures %v, want %v", moves[:n], want)
	}
	if !slices.Contains(moves[n:], quiet) || !slices.Contains(moves[n:], push) {
		t.Fatal("filtered moves missing from the trace tail")
	}
}
