package movegen

import (
	"math/bits"
	"testing"

	"github.com/jrichard8/MarsGopherChess/internal/board"
)

func kingSquareForSide(b *board.Board, side int) int {
	var bb board.Bitboard
	if side == board.White {
		bb = b.Pieces[board.WhiteKing]
	} else {
		bb = b.Pieces[board.BlackKing]
	}
	if bb == 0 {
		return board.NoSquare
	}
	return bits.TrailingZeros64(uint64(bb))
}

func TestPseudoLegalVsLegalCounts(t *testing.T) {
	fen := "4r3/8/8/8/8/8/4R3/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	var pseudo [MaxMoves]uint32
	var legal [MaxMoves]uint32
	pseudoCount := GeneratePseudoLegalMoves(b, &pseudo)
	legalCount := GenerateLegalMoves(b, &legal)

	if pseudoCount <= legalCount {
		t.Fatalf("expected pseudo-legal count > legal count, got %d vs %d", pseudoCount, legalCount)
	}
}

func TestLegalMovesDoNotLeaveKingInCheck(t *testing.T) {
	fen := "4r3/8/8/8/8/8/4R3/4K3 w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	var legal [MaxMoves]uint32
	count := GenerateLegalMoves(b, &legal)
	movingSide := b.SideToMove
	oppSide := board.Black
	if movingSide == board.Black {
		oppSide = board.White
	}

	for i := 0; i < count; i++ {
		m := legal[i]
		b.MakeMove(m)
		kingSq := kingSquareForSide(b, movingSide)
		if kingSq == board.NoSquare {
			b.UnmakeMove()
			t.Fatalf("king square missing after move")
		}
		if b.IsSquareAttacked(kingSq, oppSide) {
			b.UnmakeMove()
			t.Fatalf("legal move leaves king in check")
		}
		b.UnmakeMove()
	}
}
