package board

import "math/bits"

var knightAttacks [64]Bitboard
var kingAttacks [64]Bitboard

func init() {
	for sq := 0; sq < 64; sq++ {
		file := sq % 8
		rank := sq / 8

		var n Bitboard
		knightDeltas := [8][2]int{{1, 2}, {2, 1}, {-1, 2}, {-2, 1}, {1, -2}, {2, -1}, {-1, -2}, {-2, -1}}
		for i := 0; i < len(knightDeltas); i++ {
			f := file + knightDeltas[i][0]
			r := rank + knightDeltas[i][1]
			if f >= 0 && f < 8 && r >= 0 && r < 8 {
				n |= Bitboard(1) << (r*8 + f)
			}
		}
		knightAttacks[sq] = n

		var k Bitboard
		kingDeltas := [8][2]int{{1, 1}, {1, 0}, {1, -1}, {0, 1}, {0, -1}, {-1, 1}, {-1, 0}, {-1, -1}}
		for i := 0; i < len(kingDeltas); i++ {
			f := file + kingDeltas[i][0]
			r := rank + kingDeltas[i][1]
			if f >= 0 && f < 8 && r >= 0 && r < 8 {
				k |= Bitboard(1) << (r*8 + f)
			}
		}
		kingAttacks[sq] = k
	}
}

// IsSquareAttacked reports whether square is attacked by the given side.
func (b *Board) IsSquareAttacked(square int, bySide int) bool {
	file := square % 8

	if bySide == White {
		if file > 0 {
			from := square - 9
			if from >= 0 {
				if b.Pieces[WhitePawn]&(Bitboard(1)<<from) != 0 {
					return true
				}
			}
		}
		if file < 7 {
			from := square - 7
			if from >= 0 {
				if b.Pieces[WhitePawn]&(Bitboard(1)<<from) != 0 {
					return true
				}
			}
		}
	} else {
		if file > 0 {
			from := square + 7
			if from < 64 {
				if b.Pieces[BlackPawn]&(Bitboard(1)<<from) != 0 {
					return true
				}
			}
		}
		if file < 7 {
			from := square + 9
			if from < 64 {
				if b.Pieces[BlackPawn]&(Bitboard(1)<<from) != 0 {
					return true
				}
			}
		}
	}

	if bySide == White {
		if knightAttacks[square]&b.Pieces[WhiteKnight] != 0 {
			return true
		}
		if kingAttacks[square]&b.Pieces[WhiteKing] != 0 {
			return true
		}
	} else {
		if knightAttacks[square]&b.Pieces[BlackKnight] != 0 {
			return true
		}
		if kingAttacks[square]&b.Pieces[BlackKing] != 0 {
			return true
		}
	}

	var diagMask Bitboard
	if bySide == White {
		diagMask = b.Pieces[WhiteBishop] | b.Pieces[WhiteQueen]
	} else {
		diagMask = b.Pieces[BlackBishop] | b.Pieces[BlackQueen]
	}
	if diagMask != 0 && BishopTargets(square, b.OccupancyBoth)&diagMask != 0 {
		return true
	}

	var orthoMask Bitboard
	if bySide == White {
		orthoMask = b.Pieces[WhiteRook] | b.Pieces[WhiteQueen]
	} else {
		orthoMask = b.Pieces[BlackRook] | b.Pieces[BlackQueen]
	}
	return orthoMask != 0 && RookTargets(square, b.OccupancyBoth)&orthoMask != 0
}

func kingSquare(bb Bitboard) int {
	if bb == 0 {
		return NoSquare
	}
	return bits.TrailingZeros64(uint64(bb))
}
