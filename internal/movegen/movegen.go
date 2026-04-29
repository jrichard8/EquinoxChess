package movegen

import (
	"math/bits"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/move"
)

const MaxMoves = 256

var bishopDirs = [4][2]int{{1, 1}, {-1, 1}, {1, -1}, {-1, -1}}
var rookDirs = [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

func GeneratePseudoLegalMoves(b *board.Board, moves *[MaxMoves]uint32) int {
	count := 0
	side := b.SideToMove
	ownOcc := b.OccupancyWhite
	oppOcc := b.OccupancyBlack
	if side == board.Black {
		ownOcc, oppOcc = oppOcc, ownOcc
	}

	if side == board.White {
		count = genWhitePawns(b, moves, count, ownOcc, oppOcc)
		count = genPieces(b, moves, count, b.Pieces[board.WhiteKnight], ownOcc, oppOcc, knightAttacks, board.WhiteKnight)
		count = genSliding(b, moves, count, b.Pieces[board.WhiteBishop], ownOcc, oppOcc, bishopDirs, board.WhiteBishop)
		count = genSliding(b, moves, count, b.Pieces[board.WhiteRook], ownOcc, oppOcc, rookDirs, board.WhiteRook)
		count = genSliding(b, moves, count, b.Pieces[board.WhiteQueen], ownOcc, oppOcc, bishopDirs, board.WhiteQueen)
		count = genSliding(b, moves, count, b.Pieces[board.WhiteQueen], ownOcc, oppOcc, rookDirs, board.WhiteQueen)
		count = genPieces(b, moves, count, b.Pieces[board.WhiteKing], ownOcc, oppOcc, kingAttacks, board.WhiteKing)
		count = genWhiteCastling(b, moves, count)
	} else {
		count = genBlackPawns(b, moves, count, ownOcc, oppOcc)
		count = genPieces(b, moves, count, b.Pieces[board.BlackKnight], ownOcc, oppOcc, knightAttacks, board.BlackKnight)
		count = genSliding(b, moves, count, b.Pieces[board.BlackBishop], ownOcc, oppOcc, bishopDirs, board.BlackBishop)
		count = genSliding(b, moves, count, b.Pieces[board.BlackRook], ownOcc, oppOcc, rookDirs, board.BlackRook)
		count = genSliding(b, moves, count, b.Pieces[board.BlackQueen], ownOcc, oppOcc, bishopDirs, board.BlackQueen)
		count = genSliding(b, moves, count, b.Pieces[board.BlackQueen], ownOcc, oppOcc, rookDirs, board.BlackQueen)
		count = genPieces(b, moves, count, b.Pieces[board.BlackKing], ownOcc, oppOcc, kingAttacks, board.BlackKing)
		count = genBlackCastling(b, moves, count)
	}

	return count
}

func GenerateLegalMoves(b *board.Board, moves *[MaxMoves]uint32) int {
	var pseudo [MaxMoves]uint32
	count := GeneratePseudoLegalMoves(b, &pseudo)
	out := 0
	movingSide := b.SideToMove
	oppSide := board.Black
	if movingSide == board.Black {
		oppSide = board.White
	}

	for i := 0; i < count; i++ {
		m := pseudo[i]
		b.MakeMove(m)
		kingSq := kingSquare(b, movingSide)
		if kingSq != board.NoSquare && !b.IsSquareAttacked(kingSq, oppSide) {
			moves[out] = m
			out++
		}
		b.UnmakeMove()
	}

	return out
}

func isOpponentKing(captured int, side int) bool {
	if side == board.White {
		return captured == board.BlackKing
	}
	return captured == board.WhiteKing
}

func genWhitePawns(b *board.Board, moves *[MaxMoves]uint32, count int, ownOcc, oppOcc board.Bitboard) int {
	pawns := b.Pieces[board.WhitePawn]
	occ := b.OccupancyBoth
	for pawns != 0 {
		from := popLSB(&pawns)
		fromRank := from / 8
		fromFile := from % 8

		single := from + 8
		if single < 64 {
			if occ&(board.Bitboard(1)<<single) == 0 {
				if single/8 == 7 {
					count = addPromotionMoves(moves, count, from, single, board.WhitePawn, move.NoPiece, board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight)
				} else {
					moves[count] = move.NewMove(from, single, board.WhitePawn, move.NoPiece, move.NoPiece, 0)
					count++
					if fromRank == 1 {
						double := from + 16
						if occ&(board.Bitboard(1)<<double) == 0 {
							moves[count] = move.NewMove(from, double, board.WhitePawn, move.NoPiece, move.NoPiece, move.FlagDoublePawnPush)
							count++
						}
					}
				}
			}
		}

		if fromFile > 0 {
			to := from + 7
			mask := board.Bitboard(1) << to
			if oppOcc&mask != 0 {
				captured := pieceAt(b, to)
				if captured == board.BlackKing {
					// never generate king captures
				} else if to/8 == 7 {
					count = addPromotionMoves(moves, count, from, to, board.WhitePawn, captured, board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight)
				} else {
					moves[count] = move.NewMove(from, to, board.WhitePawn, captured, move.NoPiece, 0)
					count++
				}
			} else if b.EnPassant == to {
				moves[count] = move.NewMove(from, to, board.WhitePawn, board.BlackPawn, move.NoPiece, move.FlagEnPassant)
				count++
			}
		}

		if fromFile < 7 {
			to := from + 9
			mask := board.Bitboard(1) << to
			if oppOcc&mask != 0 {
				captured := pieceAt(b, to)
				if captured == board.BlackKing {
					// never generate king captures
				} else if to/8 == 7 {
					count = addPromotionMoves(moves, count, from, to, board.WhitePawn, captured, board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight)
				} else {
					moves[count] = move.NewMove(from, to, board.WhitePawn, captured, move.NoPiece, 0)
					count++
				}
			} else if b.EnPassant == to {
				moves[count] = move.NewMove(from, to, board.WhitePawn, board.BlackPawn, move.NoPiece, move.FlagEnPassant)
				count++
			}
		}
	}

	return count
}

func genBlackPawns(b *board.Board, moves *[MaxMoves]uint32, count int, ownOcc, oppOcc board.Bitboard) int {
	pawns := b.Pieces[board.BlackPawn]
	occ := b.OccupancyBoth
	for pawns != 0 {
		from := popLSB(&pawns)
		fromRank := from / 8
		fromFile := from % 8

		single := from - 8
		if single >= 0 {
			if occ&(board.Bitboard(1)<<single) == 0 {
				if single/8 == 0 {
					count = addPromotionMoves(moves, count, from, single, board.BlackPawn, move.NoPiece, board.BlackQueen, board.BlackRook, board.BlackBishop, board.BlackKnight)
				} else {
					moves[count] = move.NewMove(from, single, board.BlackPawn, move.NoPiece, move.NoPiece, 0)
					count++
					if fromRank == 6 {
						double := from - 16
						if occ&(board.Bitboard(1)<<double) == 0 {
							moves[count] = move.NewMove(from, double, board.BlackPawn, move.NoPiece, move.NoPiece, move.FlagDoublePawnPush)
							count++
						}
					}
				}
			}
		}

		if fromFile > 0 {
			to := from - 9
			mask := board.Bitboard(1) << to
			if oppOcc&mask != 0 {
				captured := pieceAt(b, to)
				if captured == board.WhiteKing {
					// never generate king captures
				} else if to/8 == 0 {
					count = addPromotionMoves(moves, count, from, to, board.BlackPawn, captured, board.BlackQueen, board.BlackRook, board.BlackBishop, board.BlackKnight)
				} else {
					moves[count] = move.NewMove(from, to, board.BlackPawn, captured, move.NoPiece, 0)
					count++
				}
			} else if b.EnPassant == to {
				moves[count] = move.NewMove(from, to, board.BlackPawn, board.WhitePawn, move.NoPiece, move.FlagEnPassant)
				count++
			}
		}

		if fromFile < 7 {
			to := from - 7
			mask := board.Bitboard(1) << to
			if oppOcc&mask != 0 {
				captured := pieceAt(b, to)
				if captured == board.WhiteKing {
					// never generate king captures
				} else if to/8 == 0 {
					count = addPromotionMoves(moves, count, from, to, board.BlackPawn, captured, board.BlackQueen, board.BlackRook, board.BlackBishop, board.BlackKnight)
				} else {
					moves[count] = move.NewMove(from, to, board.BlackPawn, captured, move.NoPiece, 0)
					count++
				}
			} else if b.EnPassant == to {
				moves[count] = move.NewMove(from, to, board.BlackPawn, board.WhitePawn, move.NoPiece, move.FlagEnPassant)
				count++
			}
		}
	}

	return count
}

func genPieces(b *board.Board, moves *[MaxMoves]uint32, count int, pieces, ownOcc, oppOcc board.Bitboard, attacks [64]board.Bitboard, piece int) int {
	bb := pieces
	for bb != 0 {
		from := popLSB(&bb)
		atk := attacks[from] & ^ownOcc
		for atk != 0 {
			to := popLSB(&atk)
			captured := move.NoPiece
			if oppOcc&(board.Bitboard(1)<<to) != 0 {
				captured = pieceAt(b, to)
				if isOpponentKing(captured, b.SideToMove) {
					continue
				}
			}
			moves[count] = move.NewMove(from, to, piece, captured, move.NoPiece, 0)
			count++
		}
	}

	return count
}

func genSliding(b *board.Board, moves *[MaxMoves]uint32, count int, pieces, ownOcc, oppOcc board.Bitboard, dirs [4][2]int, piece int) int {
	bb := pieces
	occ := b.OccupancyBoth
	for bb != 0 {
		from := popLSB(&bb)
		fromFile := from % 8
		fromRank := from / 8
		for i := 0; i < len(dirs); i++ {
			f := fromFile + dirs[i][0]
			r := fromRank + dirs[i][1]
			for f >= 0 && f < 8 && r >= 0 && r < 8 {
				to := r*8 + f
				mask := board.Bitboard(1) << to
				if occ&mask != 0 {
					if oppOcc&mask != 0 {
						captured := pieceAt(b, to)
						if isOpponentKing(captured, b.SideToMove) {
							break
						}
						moves[count] = move.NewMove(from, to, piece, captured, move.NoPiece, 0)
						count++
					}
					break
				}
				moves[count] = move.NewMove(from, to, piece, move.NoPiece, move.NoPiece, 0)
				count++
				f += dirs[i][0]
				r += dirs[i][1]
			}
		}
	}
	return count
}

func genWhiteCastling(b *board.Board, moves *[MaxMoves]uint32, count int) int {
	if b.CastlingRights&board.CastleWhiteKing != 0 {
		mask := (board.Bitboard(1) << board.F1) | (board.Bitboard(1) << board.G1)
		if b.OccupancyBoth&mask == 0 {
			if !b.IsSquareAttacked(board.E1, board.Black) && !b.IsSquareAttacked(board.F1, board.Black) && !b.IsSquareAttacked(board.G1, board.Black) {
				if b.Pieces[board.WhiteKing]&(board.Bitboard(1)<<board.E1) != 0 && b.Pieces[board.WhiteRook]&(board.Bitboard(1)<<board.H1) != 0 {
					moves[count] = move.NewMove(board.E1, board.G1, board.WhiteKing, move.NoPiece, move.NoPiece, move.FlagCastling)
					count++
				}
			}
		}
	}
	if b.CastlingRights&board.CastleWhiteQueen != 0 {
		mask := (board.Bitboard(1) << board.B1) | (board.Bitboard(1) << board.C1) | (board.Bitboard(1) << board.D1)
		if b.OccupancyBoth&mask == 0 {
			if !b.IsSquareAttacked(board.E1, board.Black) && !b.IsSquareAttacked(board.D1, board.Black) && !b.IsSquareAttacked(board.C1, board.Black) {
				if b.Pieces[board.WhiteKing]&(board.Bitboard(1)<<board.E1) != 0 && b.Pieces[board.WhiteRook]&(board.Bitboard(1)<<board.A1) != 0 {
					moves[count] = move.NewMove(board.E1, board.C1, board.WhiteKing, move.NoPiece, move.NoPiece, move.FlagCastling)
					count++
				}
			}
		}
	}
	return count
}

func genBlackCastling(b *board.Board, moves *[MaxMoves]uint32, count int) int {
	if b.CastlingRights&board.CastleBlackKing != 0 {
		mask := (board.Bitboard(1) << board.F8) | (board.Bitboard(1) << board.G8)
		if b.OccupancyBoth&mask == 0 {
			if !b.IsSquareAttacked(board.E8, board.White) && !b.IsSquareAttacked(board.F8, board.White) && !b.IsSquareAttacked(board.G8, board.White) {
				if b.Pieces[board.BlackKing]&(board.Bitboard(1)<<board.E8) != 0 && b.Pieces[board.BlackRook]&(board.Bitboard(1)<<board.H8) != 0 {
					moves[count] = move.NewMove(board.E8, board.G8, board.BlackKing, move.NoPiece, move.NoPiece, move.FlagCastling)
					count++
				}
			}
		}
	}
	if b.CastlingRights&board.CastleBlackQueen != 0 {
		mask := (board.Bitboard(1) << board.B8) | (board.Bitboard(1) << board.C8) | (board.Bitboard(1) << board.D8)
		if b.OccupancyBoth&mask == 0 {
			if !b.IsSquareAttacked(board.E8, board.White) && !b.IsSquareAttacked(board.D8, board.White) && !b.IsSquareAttacked(board.C8, board.White) {
				if b.Pieces[board.BlackKing]&(board.Bitboard(1)<<board.E8) != 0 && b.Pieces[board.BlackRook]&(board.Bitboard(1)<<board.A8) != 0 {
					moves[count] = move.NewMove(board.E8, board.C8, board.BlackKing, move.NoPiece, move.NoPiece, move.FlagCastling)
					count++
				}
			}
		}
	}
	return count
}

func addPromotionMoves(moves *[MaxMoves]uint32, count, from, to, moving, captured, q, r, b, n int) int {
	moves[count] = move.NewMove(from, to, moving, captured, q, 0)
	count++
	moves[count] = move.NewMove(from, to, moving, captured, r, 0)
	count++
	moves[count] = move.NewMove(from, to, moving, captured, b, 0)
	count++
	moves[count] = move.NewMove(from, to, moving, captured, n, 0)
	count++
	return count
}

func pieceAt(b *board.Board, square int) int {
	mask := board.Bitboard(1) << square
	for i := 0; i < len(b.Pieces); i++ {
		if b.Pieces[i]&mask != 0 {
			return i
		}
	}
	return move.NoPiece
}

func popLSB(bb *board.Bitboard) int {
	lsb := uint64(*bb) & (^(uint64(*bb)) + 1)
	sq := bits.TrailingZeros64(lsb)
	*bb &= *bb - 1
	return sq
}

func kingSquare(b *board.Board, side int) int {
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
