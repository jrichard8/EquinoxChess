package uci

import (
	"errors"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/move"
	"github.com/jrichard8/MarsGopherChess/internal/movegen"
)

var errInvalidUCI = errors.New("invalid uci move")
var errMoveNotFound = errors.New("move not found")

func MoveToUCI(m move.Move) string {
	from := move.From(uint32(m))
	to := move.To(uint32(m))
	promo := move.PromotionPiece(uint32(m))

	var out [5]byte
	out[0] = byte('a' + (from % 8))
	out[1] = byte('1' + (from / 8))
	out[2] = byte('a' + (to % 8))
	out[3] = byte('1' + (to / 8))
	if promo != move.NoPiece {
		out[4] = promotionSuffix(promo)
		return string(out[:])
	}
	return string(out[:4])
}

func ParseUCIMove(b *board.Board, s string) (move.Move, error) {
	if len(s) < 4 || len(s) > 5 {
		return 0, errInvalidUCI
	}

	fromFile := int(s[0] - 'a')
	fromRank := int(s[1] - '1')
	toFile := int(s[2] - 'a')
	toRank := int(s[3] - '1')
	if fromFile < 0 || fromFile > 7 || toFile < 0 || toFile > 7 || fromRank < 0 || fromRank > 7 || toRank < 0 || toRank > 7 {
		return 0, errInvalidUCI
	}
	from := fromRank*8 + fromFile
	to := toRank*8 + toFile

	promo := move.NoPiece
	if len(s) == 5 {
		promoChar := s[4]
		switch promoChar {
		case 'q', 'r', 'b', 'n':
			promo = promotionPiece(b.SideToMove, promoChar)
			if promo == move.NoPiece {
				return 0, errInvalidUCI
			}
		default:
			return 0, errInvalidUCI
		}
	}

	var moves [movegen.MaxMoves]uint32
	count := movegen.GeneratePseudoLegalMoves(b, &moves)
	for i := 0; i < count; i++ {
		m := moves[i]
		if move.From(m) != from || move.To(m) != to {
			continue
		}
		mp := move.PromotionPiece(m)
		if promo == move.NoPiece {
			if mp != move.NoPiece {
				continue
			}
			return move.Move(m), nil
		}
		if mp == promo {
			return move.Move(m), nil
		}
	}

	return 0, errMoveNotFound
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

func promotionPiece(side int, promo byte) int {
	if side == board.White {
		switch promo {
		case 'q':
			return board.WhiteQueen
		case 'r':
			return board.WhiteRook
		case 'b':
			return board.WhiteBishop
		case 'n':
			return board.WhiteKnight
		}
		return move.NoPiece
	}
	switch promo {
	case 'q':
		return board.BlackQueen
	case 'r':
		return board.BlackRook
	case 'b':
		return board.BlackBishop
	case 'n':
		return board.BlackKnight
	}
	return move.NoPiece
}
