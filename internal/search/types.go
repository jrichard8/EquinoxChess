package search

import (
	"math/bits"
	"sync/atomic"
	"time"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/tt"
)

const (
	Infinity      = 50000
	MateScore     = 32000
	MateThreshold = 31000
	DrawScore     = 0
	MaxDepth      = 256
)

type Search struct {
	Nodes       uint64
	Ply         int
	MaxDepth    int
	KillerMoves [MaxDepth][2]uint32
	History     [12][64]int
	TT          *tt.TranspositionTable
	Stop        atomic.Bool
	PVMove      [MaxDepth]uint32
	StartTime   time.Time
	TimeLimit   time.Duration
	Trace       *SearchTrace
}

func InCheck(b *board.Board) bool {
	kingSq := board.NoSquare
	if b.SideToMove == board.White {
		bb := b.Pieces[board.WhiteKing]
		if bb != 0 {
			kingSq = bits.TrailingZeros64(uint64(bb))
		}
		if kingSq == board.NoSquare {
			return false
		}
		return b.IsSquareAttacked(kingSq, board.Black)
	}
	bb := b.Pieces[board.BlackKing]
	if bb != 0 {
		kingSq = bits.TrailingZeros64(uint64(bb))
	}
	if kingSq == board.NoSquare {
		return false
	}
	return b.IsSquareAttacked(kingSq, board.White)
}

func IsMovedSideInCheck(b *board.Board) bool {
	movedSide := b.SideToMove ^ 1
	kingSq := board.NoSquare
	if movedSide == board.White {
		bb := b.Pieces[board.WhiteKing]
		if bb != 0 {
			kingSq = bits.TrailingZeros64(uint64(bb))
		}
	} else {
		bb := b.Pieces[board.BlackKing]
		if bb != 0 {
			kingSq = bits.TrailingZeros64(uint64(bb))
		}
	}
	if kingSq == board.NoSquare {
		return false
	}
	return b.IsSquareAttacked(kingSq, b.SideToMove)
}

func hasNonPawnMaterial(b *board.Board, side int) bool {
	if side == board.White {
		return (b.Pieces[board.WhiteKnight] |
			b.Pieces[board.WhiteBishop] |
			b.Pieces[board.WhiteRook] |
			b.Pieces[board.WhiteQueen]) != 0
	}
	return (b.Pieces[board.BlackKnight] |
		b.Pieces[board.BlackBishop] |
		b.Pieces[board.BlackRook] |
		b.Pieces[board.BlackQueen]) != 0
}

func totalNonPawnMaterial(b *board.Board) int {
	count := 0
	count += bits.OnesCount64(uint64(b.Pieces[board.WhiteKnight]))
	count += bits.OnesCount64(uint64(b.Pieces[board.WhiteBishop]))
	count += bits.OnesCount64(uint64(b.Pieces[board.WhiteRook]))
	count += bits.OnesCount64(uint64(b.Pieces[board.WhiteQueen]))
	count += bits.OnesCount64(uint64(b.Pieces[board.BlackKnight]))
	count += bits.OnesCount64(uint64(b.Pieces[board.BlackBishop]))
	count += bits.OnesCount64(uint64(b.Pieces[board.BlackRook]))
	count += bits.OnesCount64(uint64(b.Pieces[board.BlackQueen]))
	return count
}
