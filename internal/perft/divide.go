package perft

import (
	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/move"
	"github.com/jrichard8/MarsGopherChess/internal/movegen"
)

type DivideResult struct {
	Move  uint32
	Nodes uint64
}

func PerftDivide(b *board.Board, depth int, out *[movegen.MaxMoves]DivideResult) int {
	var moves [movegen.MaxMoves]uint32
	count := movegen.GenerateLegalMoves(b, &moves)
	outCount := 0
	if depth == 0 {
		return 0
	}
	for i := 0; i < count; i++ {
		m := moves[i]
		b.MakeMove(m)
		var nodes uint64
		if depth == 1 {
			nodes = 1
		} else {
			nodes = Perft(b, depth-1)
		}
		b.UnmakeMove()
		out[outCount] = DivideResult{Move: m, Nodes: nodes}
		outCount++
	}
	return outCount
}

func PromotionSuffix(piece int) byte {
	switch piece {
	case board.WhiteQueen, board.BlackQueen:
		return 'q'
	case board.WhiteRook, board.BlackRook:
		return 'r'
	case board.WhiteBishop, board.BlackBishop:
		return 'b'
	case board.WhiteKnight, board.BlackKnight:
		return 'n'
	default:
		return 0
	}
}

func MoveToUCI(m uint32) [5]byte {
	from := move.From(m)
	to := move.To(m)
	promo := move.PromotionPiece(m)

	var out [5]byte
	out[0] = byte('a' + (from % 8))
	out[1] = byte('1' + (from / 8))
	out[2] = byte('a' + (to % 8))
	out[3] = byte('1' + (to / 8))
	if promo != move.NoPiece {
		out[4] = PromotionSuffix(promo)
	}
	return out
}
