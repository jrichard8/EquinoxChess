package api

import (
	"errors"
	"strconv"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/search"
	"EquinoxChess/internal/tt"
	"EquinoxChess/internal/uci"
)

type Game struct {
	Board  *board.Board
	Search search.Search
}

func NewGame() *Game {
	g := &Game{}
	g.NewGame()
	return g
}

func (g *Game) NewGame() {
	g.Board = board.NewBoard()
	g.Search = search.Search{TT: tt.NewTT(1 << 20)}
}

func (g *Game) PlayMove(uciStr string) error {
	if g.Board == nil {
		return errors.New("game not initialized")
	}
	m, err := uci.ParseUCIMove(g.Board, uciStr)
	if err != nil {
		return err
	}
	g.Board.MakeMove(m)
	if search.IsMovedSideInCheck(g.Board) {
		g.Board.UnmakeMove()
		return errors.New("illegal move")
	}
	return nil
}

func (g *Game) EngineMove(depth int) (string, error) {
	if g.Board == nil {
		return "", errors.New("game not initialized")
	}
	best, _, _, _ := g.Search.SearchRoot(g.Board, depth)
	if best == 0 {
		return "", errors.New("no move found")
	}
	g.Board.MakeMove(best)
	return uci.MoveToUCI(best), nil
}

func (g *Game) FEN() string {
	if g.Board == nil {
		return ""
	}
	b := g.Board

	buf := make([]byte, 0, 96)

	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			sq := rank*8 + file
			piece := pieceAt(b, sq)
			if piece == 0 {
				empty++
				continue
			}
			if empty > 0 {
				buf = append(buf, byte('0'+empty))
				empty = 0
			}
			buf = append(buf, piece)
		}
		if empty > 0 {
			buf = append(buf, byte('0'+empty))
		}
		if rank > 0 {
			buf = append(buf, '/')
		}
	}

	buf = append(buf, ' ')
	if b.SideToMove == board.White {
		buf = append(buf, 'w')
	} else {
		buf = append(buf, 'b')
	}

	buf = append(buf, ' ')
	castling := 0
	if b.CastlingRights&board.CastleWhiteKing != 0 {
		buf = append(buf, 'K')
		castling++
	}
	if b.CastlingRights&board.CastleWhiteQueen != 0 {
		buf = append(buf, 'Q')
		castling++
	}
	if b.CastlingRights&board.CastleBlackKing != 0 {
		buf = append(buf, 'k')
		castling++
	}
	if b.CastlingRights&board.CastleBlackQueen != 0 {
		buf = append(buf, 'q')
		castling++
	}
	if castling == 0 {
		buf = append(buf, '-')
	}

	buf = append(buf, ' ')
	if b.EnPassant == board.NoSquare {
		buf = append(buf, '-')
	} else {
		file := b.EnPassant % 8
		rank := b.EnPassant / 8
		buf = append(buf, byte('a'+file))
		buf = append(buf, byte('1'+rank))
	}

	buf = append(buf, ' ')
	buf = strconv.AppendInt(buf, int64(b.HalfmoveClock), 10)
	buf = append(buf, ' ')
	fullmove := b.FullmoveNumber
	if fullmove == 0 {
		fullmove = 1
	}
	buf = strconv.AppendInt(buf, int64(fullmove), 10)

	return string(buf)
}

func pieceAt(b *board.Board, square int) byte {
	mask := board.Bitboard(1) << square
	if b.Pieces[board.WhitePawn]&mask != 0 {
		return 'P'
	}
	if b.Pieces[board.WhiteKnight]&mask != 0 {
		return 'N'
	}
	if b.Pieces[board.WhiteBishop]&mask != 0 {
		return 'B'
	}
	if b.Pieces[board.WhiteRook]&mask != 0 {
		return 'R'
	}
	if b.Pieces[board.WhiteQueen]&mask != 0 {
		return 'Q'
	}
	if b.Pieces[board.WhiteKing]&mask != 0 {
		return 'K'
	}
	if b.Pieces[board.BlackPawn]&mask != 0 {
		return 'p'
	}
	if b.Pieces[board.BlackKnight]&mask != 0 {
		return 'n'
	}
	if b.Pieces[board.BlackBishop]&mask != 0 {
		return 'b'
	}
	if b.Pieces[board.BlackRook]&mask != 0 {
		return 'r'
	}
	if b.Pieces[board.BlackQueen]&mask != 0 {
		return 'q'
	}
	if b.Pieces[board.BlackKing]&mask != 0 {
		return 'k'
	}
	return 0
}
