package board

import "math/bits"

// Each mask is a complete line through the square, with the origin removed.
type lines struct{ rank, file, diagonal, antidiagonal uint64 }

var squareLines [64]lines

func init() {
	for from := range squareLines {
		for to := 0; to < 64; to++ {
			if from == to {
				continue
			}
			mask := uint64(1) << to
			fileDelta, rankDelta := to%8-from%8, to/8-from/8
			switch {
			case rankDelta == 0:
				squareLines[from].rank |= mask
			case fileDelta == 0:
				squareLines[from].file |= mask
			case fileDelta == rankDelta:
				squareLines[from].diagonal |= mask
			case fileDelta == -rankDelta:
				squareLines[from].antidiagonal |= mask
			}
		}
	}
}

// Subtraction propagates up to the first blocker, inclusive. Reverse the bits
// to do the same in the opposite direction (Hyperbola Quintessence).
func lineTargets(occupancy, origin, mask uint64) uint64 {
	blockers := occupancy & mask
	forward := blockers - origin
	backward := bits.Reverse64(blockers) - bits.Reverse64(origin)
	return (forward ^ bits.Reverse64(backward)) & mask
}

func KnightTargets(square int) Bitboard { return knightAttacks[square] }

func BishopTargets(square int, occupancy Bitboard) Bitboard {
	l := &squareLines[square]
	origin := uint64(1) << square
	return Bitboard(lineTargets(uint64(occupancy), origin, l.diagonal) | lineTargets(uint64(occupancy), origin, l.antidiagonal))
}

func RookTargets(square int, occupancy Bitboard) Bitboard {
	l := &squareLines[square]
	origin := uint64(1) << square
	return Bitboard(lineTargets(uint64(occupancy), origin, l.file) | lineTargets(uint64(occupancy), origin, l.rank))
}
