package board

import (
	"EquinoxChess/internal/move"
	"EquinoxChess/internal/zobrist"
)

// MakeMove updates the board state for a pseudo-legal move.
func (b *Board) MakeMove(m uint32) {
	from := move.From(m)
	to := move.To(m)
	moving := move.MovingPiece(m)
	captured := move.CapturedPiece(m)
	promotion := move.PromotionPiece(m)
	isCapture := move.IsCapture(m)
	isEnPassant := move.IsEnPassant(m)
	isCastling := move.IsCastling(m)
	isDoublePush := move.IsDoublePawnPush(m)

	b.undoStack[b.ply] = Undo{
		Move:           m,
		CastlingRights: b.CastlingRights,
		EnPassant:      b.EnPassant,
		HalfmoveClock:  b.HalfmoveClock,
		Zobrist:        b.ZobristKey,
	}
	b.ply++

	if b.EnPassant != NoSquare {
		b.ZobristKey ^= zobrist.EnPassant[b.EnPassant]
	}

	if moving == WhitePawn || moving == BlackPawn || isCapture {
		b.HalfmoveClock = 0
	} else {
		b.HalfmoveClock++
	}

	oldRights := b.CastlingRights
	newRights := oldRights
	switch moving {
	case WhiteKing:
		newRights &^= CastleWhiteKing | CastleWhiteQueen
	case BlackKing:
		newRights &^= CastleBlackKing | CastleBlackQueen
	case WhiteRook:
		if from == A1 {
			newRights &^= CastleWhiteQueen
		} else if from == H1 {
			newRights &^= CastleWhiteKing
		}
	case BlackRook:
		if from == A8 {
			newRights &^= CastleBlackQueen
		} else if from == H8 {
			newRights &^= CastleBlackKing
		}
	}

	if isCapture {
		switch captured {
		case WhiteRook:
			if to == A1 {
				newRights &^= CastleWhiteQueen
			} else if to == H1 {
				newRights &^= CastleWhiteKing
			}
		case BlackRook:
			if to == A8 {
				newRights &^= CastleBlackQueen
			} else if to == H8 {
				newRights &^= CastleBlackKing
			}
		}
	}

	if newRights != oldRights {
		b.ZobristKey ^= zobrist.Castling[oldRights]
		b.ZobristKey ^= zobrist.Castling[newRights]
		b.CastlingRights = newRights
	}

	fromMask := Bitboard(1) << from
	toMask := Bitboard(1) << to

	if promotion != move.NoPiece {
		b.Pieces[moving] &^= fromMask
		b.ZobristKey ^= zobrist.Piece[moving][from]

		if isCapture {
			b.Pieces[captured] &^= toMask
			b.ZobristKey ^= zobrist.Piece[captured][to]
		}

		b.Pieces[promotion] |= toMask
		b.ZobristKey ^= zobrist.Piece[promotion][to]
	} else {
		b.Pieces[moving] &^= fromMask
		b.ZobristKey ^= zobrist.Piece[moving][from]

		if isEnPassant {
			var capSq int
			var capPiece int
			if b.SideToMove == White {
				capSq = to - 8
				capPiece = BlackPawn
			} else {
				capSq = to + 8
				capPiece = WhitePawn
			}
			capMask := Bitboard(1) << capSq
			b.Pieces[capPiece] &^= capMask
			b.ZobristKey ^= zobrist.Piece[capPiece][capSq]
		} else if isCapture {
			b.Pieces[captured] &^= toMask
			b.ZobristKey ^= zobrist.Piece[captured][to]
		}

		b.Pieces[moving] |= toMask
		b.ZobristKey ^= zobrist.Piece[moving][to]
	}

	if isCastling {
		var rookFrom, rookTo int
		var rookPiece int
		castleRook := false
		if moving == WhiteKing {
			rookPiece = WhiteRook
			if to == G1 {
				rookFrom, rookTo = H1, F1
				castleRook = true
			} else if to == C1 {
				rookFrom, rookTo = A1, D1
				castleRook = true
			}
		} else if moving == BlackKing {
			rookPiece = BlackRook
			if to == G8 {
				rookFrom, rookTo = H8, F8
				castleRook = true
			} else if to == C8 {
				rookFrom, rookTo = A8, D8
				castleRook = true
			}
		}

		if castleRook {
			rookFromMask := Bitboard(1) << rookFrom
			rookToMask := Bitboard(1) << rookTo
			b.Pieces[rookPiece] &^= rookFromMask
			b.Pieces[rookPiece] |= rookToMask
			b.ZobristKey ^= zobrist.Piece[rookPiece][rookFrom]
			b.ZobristKey ^= zobrist.Piece[rookPiece][rookTo]
		}
	}

	if isDoublePush {
		if b.SideToMove == White {
			b.EnPassant = from + 8
		} else {
			b.EnPassant = from - 8
		}
	} else {
		b.EnPassant = NoSquare
	}

	if b.EnPassant != NoSquare {
		b.ZobristKey ^= zobrist.EnPassant[b.EnPassant]
	}

	b.SideToMove ^= 1
	b.ZobristKey ^= zobrist.Side

	b.updateOccupancies()
}

// UnmakeMove restores the board to the state before the last MakeMove.
func (b *Board) UnmakeMove() {
	b.ply--
	undo := b.undoStack[b.ply]
	m := undo.Move

	from := move.From(m)
	to := move.To(m)
	moving := move.MovingPiece(m)
	captured := move.CapturedPiece(m)
	promotion := move.PromotionPiece(m)
	isCapture := move.IsCapture(m)
	isEnPassant := move.IsEnPassant(m)
	isCastling := move.IsCastling(m)

	b.SideToMove ^= 1

	fromMask := Bitboard(1) << from
	toMask := Bitboard(1) << to

	if promotion != move.NoPiece {
		b.Pieces[promotion] &^= toMask
		b.Pieces[moving] |= fromMask
		if isCapture {
			b.Pieces[captured] |= toMask
		}
	} else {
		b.Pieces[moving] &^= toMask
		b.Pieces[moving] |= fromMask
		if isEnPassant {
			var capSq int
			var capPiece int
			if b.SideToMove == White {
				capSq = to - 8
				capPiece = BlackPawn
			} else {
				capSq = to + 8
				capPiece = WhitePawn
			}
			capMask := Bitboard(1) << capSq
			b.Pieces[capPiece] |= capMask
		} else if isCapture {
			b.Pieces[captured] |= toMask
		}
	}

	if isCastling {
		var rookFrom, rookTo int
		var rookPiece int
		castleRook := false
		if moving == WhiteKing {
			rookPiece = WhiteRook
			if to == G1 {
				rookFrom, rookTo = F1, H1
				castleRook = true
			} else if to == C1 {
				rookFrom, rookTo = D1, A1
				castleRook = true
			}
		} else if moving == BlackKing {
			rookPiece = BlackRook
			if to == G8 {
				rookFrom, rookTo = F8, H8
				castleRook = true
			} else if to == C8 {
				rookFrom, rookTo = D8, A8
				castleRook = true
			}
		}

		if castleRook {
			rookFromMask := Bitboard(1) << rookFrom
			rookToMask := Bitboard(1) << rookTo
			b.Pieces[rookPiece] &^= rookFromMask
			b.Pieces[rookPiece] |= rookToMask
		}
	}

	b.CastlingRights = undo.CastlingRights
	b.EnPassant = undo.EnPassant
	b.HalfmoveClock = undo.HalfmoveClock
	b.ZobristKey = undo.Zobrist

	b.updateOccupancies()
}

// MakeNullMove switches side to move without moving any piece.
func (b *Board) MakeNullMove() {
	b.undoStack[b.ply] = Undo{
		Move:           0,
		CastlingRights: b.CastlingRights,
		EnPassant:      b.EnPassant,
		HalfmoveClock:  b.HalfmoveClock,
		Zobrist:        b.ZobristKey,
	}
	b.ply++

	if b.EnPassant != NoSquare {
		b.ZobristKey ^= zobrist.EnPassant[b.EnPassant]
		b.EnPassant = NoSquare
	}

	b.HalfmoveClock++
	b.SideToMove ^= 1
	b.ZobristKey ^= zobrist.Side
}

// UnmakeNullMove restores the board to the state before the last MakeNullMove.
func (b *Board) UnmakeNullMove() {
	b.ply--
	undo := b.undoStack[b.ply]

	b.SideToMove ^= 1
	b.CastlingRights = undo.CastlingRights
	b.EnPassant = undo.EnPassant
	b.HalfmoveClock = undo.HalfmoveClock
	b.ZobristKey = undo.Zobrist
}
