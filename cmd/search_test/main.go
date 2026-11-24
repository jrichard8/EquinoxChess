package main

import (
	"fmt"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/move"
	"github.com/jrichard8/MarsGopherChess/internal/search"
)

func moveToUCI(m uint32) string {
	from := int(m & 0x3F)
	to := int((m >> 6) & 0x3F)
	promo := move.PromotionPiece(m)
	buf := [5]byte{}
	buf[0] = byte('a' + (from % 8))
	buf[1] = byte('1' + (from / 8))
	buf[2] = byte('a' + (to % 8))
	buf[3] = byte('1' + (to / 8))
	if promo != move.NoPiece {
		buf[4] = promotionSuffix(promo)
		return string(buf[:])
	}
	return string(buf[:4])
}

func promotionSuffix(piece int) byte {
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

func main() {
	b, err := board.FromFEN(board.StartFEN)
	if err != nil {
		panic(err)
	}
	var s search.Search
	best, _, _, _ := s.SearchRoot(b, 4)
	s.Nodes = 0
	s.Ply = 0
	score := s.Negamax(b, 4, -search.Infinity, search.Infinity)
	fmt.Printf("bestmove %s\n", moveToUCI(best))
	fmt.Printf("score %d\n", score)
	fmt.Printf("nodes %d\n", s.Nodes)
}
