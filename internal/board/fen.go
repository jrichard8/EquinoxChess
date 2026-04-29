package board

import (
	"fmt"
	"strconv"
	"strings"
)

// FromFEN builds a board from a FEN string.
func FromFEN(fen string) (*Board, error) {
	fields := strings.Fields(fen)
	if len(fields) < 6 {
		return nil, fmt.Errorf("invalid FEN: expected 6 fields, got %d", len(fields))
	}

	b := &Board{EnPassant: NoSquare}

	if err := b.parsePiecePlacement(fields[0]); err != nil {
		return nil, err
	}

	switch fields[1] {
	case "w":
		b.SideToMove = White
	case "b":
		b.SideToMove = Black
	default:
		return nil, fmt.Errorf("invalid side to move: %q", fields[1])
	}

	if err := b.parseCastling(fields[2]); err != nil {
		return nil, err
	}

	if err := b.parseEnPassant(fields[3]); err != nil {
		return nil, err
	}

	halfmove, err := strconv.Atoi(fields[4])
	if err != nil {
		return nil, fmt.Errorf("invalid halfmove clock: %q", fields[4])
	}
	b.HalfmoveClock = halfmove

	fullmove, err := strconv.Atoi(fields[5])
	if err != nil {
		return nil, fmt.Errorf("invalid fullmove number: %q", fields[5])
	}
	b.FullmoveNumber = fullmove

	b.updateOccupancies()
	b.ComputeZobrist()

	return b, nil
}

func (b *Board) parsePiecePlacement(field string) error {
	ranks := strings.Split(field, "/")
	if len(ranks) != 8 {
		return fmt.Errorf("invalid piece placement: %q", field)
	}

	rank := 7
	for _, row := range ranks {
		file := 0
		for i := 0; i < len(row); i++ {
			ch := row[i]
			switch {
			case ch >= '1' && ch <= '8':
				file += int(ch - '0')
				continue
			default:
				sq := rank*8 + file
				if sq < 0 || sq >= 64 {
					return fmt.Errorf("invalid square in placement: %q", field)
				}
				if err := b.setPiece(ch, sq); err != nil {
					return err
				}
				file++
			}
		}
		if file != 8 {
			return fmt.Errorf("invalid piece placement row: %q", row)
		}
		rank--
	}
	return nil
}

func (b *Board) setPiece(ch byte, sq int) error {
	mask := Bitboard(1) << sq
	switch ch {
	case 'P':
		b.Pieces[WhitePawn] |= mask
	case 'N':
		b.Pieces[WhiteKnight] |= mask
	case 'B':
		b.Pieces[WhiteBishop] |= mask
	case 'R':
		b.Pieces[WhiteRook] |= mask
	case 'Q':
		b.Pieces[WhiteQueen] |= mask
	case 'K':
		b.Pieces[WhiteKing] |= mask
	case 'p':
		b.Pieces[BlackPawn] |= mask
	case 'n':
		b.Pieces[BlackKnight] |= mask
	case 'b':
		b.Pieces[BlackBishop] |= mask
	case 'r':
		b.Pieces[BlackRook] |= mask
	case 'q':
		b.Pieces[BlackQueen] |= mask
	case 'k':
		b.Pieces[BlackKing] |= mask
	default:
		return fmt.Errorf("invalid piece: %q", ch)
	}
	return nil
}

func (b *Board) parseCastling(field string) error {
	if field == "-" {
		b.CastlingRights = 0
		return nil
	}

	var rights uint8
	for i := 0; i < len(field); i++ {
		switch field[i] {
		case 'K':
			rights |= CastleWhiteKing
		case 'Q':
			rights |= CastleWhiteQueen
		case 'k':
			rights |= CastleBlackKing
		case 'q':
			rights |= CastleBlackQueen
		default:
			return fmt.Errorf("invalid castling rights: %q", field)
		}
	}
	b.CastlingRights = rights
	return nil
}

func (b *Board) parseEnPassant(field string) error {
	if field == "-" {
		b.EnPassant = NoSquare
		return nil
	}
	if len(field) != 2 {
		return fmt.Errorf("invalid en passant square: %q", field)
	}

	file := int(field[0] - 'a')
	rank := int(field[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return fmt.Errorf("invalid en passant square: %q", field)
	}
	b.EnPassant = rank*8 + file
	return nil
}
