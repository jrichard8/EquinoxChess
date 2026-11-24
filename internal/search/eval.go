package search

import (
	"math/bits"

	"EquinoxChess/internal/board"
)

const (
	PawnValue   = 100
	KnightValue = 320
	BishopValue = 330
	RookValue   = 500
	QueenValue  = 900
	KingValue   = 0
)

var pieceValues = [12]int{
	PawnValue,
	KnightValue,
	BishopValue,
	RookValue,
	QueenValue,
	KingValue,
	PawnValue,
	KnightValue,
	BishopValue,
	RookValue,
	QueenValue,
	KingValue,
}

var pawnPSQT = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, -20, -20, 10, 10, 5,
	5, -5, -10, 0, 0, -10, -5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, 5, 10, 25, 25, 10, 5, 5,
	10, 10, 20, 30, 30, 20, 10, 10,
	50, 50, 50, 50, 50, 50, 50, 50,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightPSQT = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopPSQT = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookPSQT = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var queenPSQT = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingPSQT = [64]int{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

var passedPawnBonus = [8]int{
	0,   // rank 0
	0,   // rank 1 (pion initial blanc)
	10,  // rank 2
	20,  // rank 3
	35,  // rank 4
	60,  // rank 5
	100, // rank 6 (7ème rangée)
	0,   // rank 7 (promotion)
}

const fullPhase = 24

// Combine the original square values with the fitted correction at each phase.
// Precomputing keeps interpolation out of the evaluation loop.
var pieceSquareByPhase [fullPhase + 1][6][64]int16

func init() {
	base := [...]*[64]int{&pawnPSQT, &knightPSQT, &bishopPSQT, &rookPSQT, &queenPSQT, &kingPSQT}
	for phase := 0; phase <= fullPhase; phase++ {
		for kind, table := range base {
			for sq, value := range table {
				correction := (openingCorrection[kind][sq]*phase + endingCorrection[kind][sq]*(fullPhase-phase)) / fullPhase
				pieceSquareByPhase[phase][kind][sq] = int16(pieceValues[kind] + value + correction)
			}
		}
	}
}

func materialPhase(b *board.Board) int {
	minors := b.Pieces[board.WhiteKnight] | b.Pieces[board.BlackKnight] | b.Pieces[board.WhiteBishop] | b.Pieces[board.BlackBishop]
	rooks := b.Pieces[board.WhiteRook] | b.Pieces[board.BlackRook]
	queens := b.Pieces[board.WhiteQueen] | b.Pieces[board.BlackQueen]
	return min(fullPhase, bits.OnesCount64(uint64(minors))+2*bits.OnesCount64(uint64(rooks))+4*bits.OnesCount64(uint64(queens)))
}

func Evaluate(b *board.Board) int {
	score := 0

	phase := materialPhase(b)
	table := &pieceSquareByPhase[phase]
	for kind := board.WhitePawn; kind <= board.WhiteKing; kind++ {
		for bb := b.Pieces[kind]; bb != 0; {
			score += int(table[kind][popLSB(&bb)])
		}
		for bb := b.Pieces[board.BlackPawn+kind]; bb != 0; {
			score -= int(table[kind][popLSB(&bb)^56])
		}
	}

	whiteBishops := bits.OnesCount64(uint64(b.Pieces[board.WhiteBishop]))
	blackBishops := bits.OnesCount64(uint64(b.Pieces[board.BlackBishop]))
	if whiteBishops >= 2 {
		score += 30
	}
	if blackBishops >= 2 {
		score -= 30
	}

	whitePawns := b.Pieces[board.WhitePawn]
	blackPawns := b.Pieces[board.BlackPawn]
	const fileFillMask uint64 = 0x0101010101010101

	whiteFileOcc := uint64(whitePawns)
	whiteFileOcc |= whiteFileOcc >> 8
	whiteFileOcc |= whiteFileOcc >> 16
	whiteFileOcc |= whiteFileOcc >> 32
	whiteFileBits := uint8(whiteFileOcc & 0xFF)

	blackFileOcc := uint64(blackPawns)
	blackFileOcc |= blackFileOcc >> 8
	blackFileOcc |= blackFileOcc >> 16
	blackFileOcc |= blackFileOcc >> 32
	blackFileBits := uint8(blackFileOcc & 0xFF)

	whiteDoubledExtras := bits.OnesCount64(uint64(whitePawns)) - bits.OnesCount8(whiteFileBits)
	blackDoubledExtras := bits.OnesCount64(uint64(blackPawns)) - bits.OnesCount8(blackFileBits)
	score -= 8 * whiteDoubledExtras
	score += 8 * blackDoubledExtras

	whiteIsolatedFiles := whiteFileBits &^ ((whiteFileBits << 1) | (whiteFileBits >> 1))
	blackIsolatedFiles := blackFileBits &^ ((blackFileBits << 1) | (blackFileBits >> 1))
	whiteIsolatedMask := board.Bitboard(uint64(whiteIsolatedFiles) * fileFillMask)
	blackIsolatedMask := board.Bitboard(uint64(blackIsolatedFiles) * fileFillMask)
	whiteIsolatedCount := bits.OnesCount64(uint64(whitePawns & whiteIsolatedMask))
	blackIsolatedCount := bits.OnesCount64(uint64(blackPawns & blackIsolatedMask))
	score -= 10 * whiteIsolatedCount
	score += 10 * blackIsolatedCount

	whiteMob := countMobility(b, board.White)
	blackMob := countMobility(b, board.Black)
	score += whiteMob
	score -= blackMob

	for bb := b.Pieces[board.WhitePawn]; bb != 0; {
		sq := popLSB(&bb)
		if isPassedPawn(b, sq, board.White) {
			rank := sq / 8
			score += passedPawnBonus[rank]
		}
	}
	for bb := b.Pieces[board.BlackPawn]; bb != 0; {
		sq := popLSB(&bb)
		if isPassedPawn(b, sq, board.Black) {
			rank := sq / 8
			score -= passedPawnBonus[7-rank]
		}
	}

	nonPawnMaterial := 0
	nonPawnMaterial += bits.OnesCount64(uint64(b.Pieces[board.WhiteKnight])) * KnightValue
	nonPawnMaterial += bits.OnesCount64(uint64(b.Pieces[board.WhiteBishop])) * BishopValue
	nonPawnMaterial += bits.OnesCount64(uint64(b.Pieces[board.WhiteRook])) * RookValue
	nonPawnMaterial += bits.OnesCount64(uint64(b.Pieces[board.WhiteQueen])) * QueenValue
	nonPawnMaterial += bits.OnesCount64(uint64(b.Pieces[board.BlackKnight])) * KnightValue
	nonPawnMaterial += bits.OnesCount64(uint64(b.Pieces[board.BlackBishop])) * BishopValue
	nonPawnMaterial += bits.OnesCount64(uint64(b.Pieces[board.BlackRook])) * RookValue
	nonPawnMaterial += bits.OnesCount64(uint64(b.Pieces[board.BlackQueen])) * QueenValue

	if nonPawnMaterial > 2000 {

		whiteSafety := 0
		blackSafety := 0

		if b.Pieces[board.WhiteKing] != 0 {
			kingSq := bits.TrailingZeros64(uint64(b.Pieces[board.WhiteKing]))
			file := kingSq % 8
			rank := kingSq / 8

			if rank <= 1 {
				for _, df := range [3]int{-1, 0, 1} {
					f := file + df
					r := rank + 1

					if f < 0 || f > 7 || r > 7 {
						continue
					}

					sq := r*8 + f

					if b.Pieces[board.WhitePawn]&(1<<sq) != 0 {
						whiteSafety += 10
					} else {
						whiteSafety -= 5
					}
				}
			}
		}

		if b.Pieces[board.BlackKing] != 0 {
			kingSq := bits.TrailingZeros64(uint64(b.Pieces[board.BlackKing]))
			file := kingSq % 8
			rank := kingSq / 8

			if rank >= 6 {
				for _, df := range [3]int{-1, 0, 1} {
					f := file + df
					r := rank - 1

					if f < 0 || f > 7 || r < 0 {
						continue
					}

					sq := r*8 + f

					if b.Pieces[board.BlackPawn]&(1<<sq) != 0 {
						blackSafety += 10
					} else {
						blackSafety -= 5
					}
				}
			}
		}

		score += whiteSafety - blackSafety
	}

	if b.SideToMove == board.White {
		return score
	}
	return -score
}

func popLSB(bb *board.Bitboard) int {
	lsb := uint64(*bb) & (^(uint64(*bb)) + 1)
	sq := bits.TrailingZeros64(lsb)
	*bb &= *bb - 1
	return sq
}

func countMobility(b *board.Board, side int) int {
	own, enemyKing, offset := b.OccupancyWhite, b.Pieces[board.BlackKing], board.WhitePawn
	if side == board.Black {
		own, enemyKing, offset = b.OccupancyBlack, b.Pieces[board.WhiteKing], board.BlackPawn
	}
	available := ^(own | enemyKing)
	mobility := 0
	for bb := b.Pieces[offset+board.WhiteKnight]; bb != 0; {
		targets := board.KnightTargets(popLSB(&bb)) & available
		mobility += 4 * bits.OnesCount64(uint64(targets))
	}
	for bb := b.Pieces[offset+board.WhiteBishop]; bb != 0; {
		targets := board.BishopTargets(popLSB(&bb), b.OccupancyBoth) & available
		mobility += 4 * bits.OnesCount64(uint64(targets))
	}
	for bb := b.Pieces[offset+board.WhiteRook]; bb != 0; {
		targets := board.RookTargets(popLSB(&bb), b.OccupancyBoth) & available
		mobility += 2 * bits.OnesCount64(uint64(targets))
	}
	for bb := b.Pieces[offset+board.WhiteQueen]; bb != 0; {
		sq := popLSB(&bb)
		targets := board.BishopTargets(sq, b.OccupancyBoth) | board.RookTargets(sq, b.OccupancyBoth)
		mobility += bits.OnesCount64(uint64(targets & available))
	}
	return mobility
}

// A passed pawn has no enemy pawn ahead on its own or either adjacent file.
var passedPawnMasks [2][64]board.Bitboard

func init() {
	for square := 0; square < 64; square++ {
		file := square % 8
		files := uint64(0x0101010101010101) << file
		if file > 0 {
			files |= uint64(0x0101010101010101) << (file - 1)
		}
		if file < 7 {
			files |= uint64(0x0101010101010101) << (file + 1)
		}
		rank := square / 8
		passedPawnMasks[board.White][square] = board.Bitboard(files & (^uint64(0) << ((rank + 1) * 8)))
		passedPawnMasks[board.Black][square] = board.Bitboard(files & ((uint64(1) << (rank * 8)) - 1))
	}
}

func isPassedPawn(b *board.Board, square, side int) bool {
	enemy := board.BlackPawn
	if side == board.Black {
		enemy = board.WhitePawn
	}
	return b.Pieces[enemy]&passedPawnMasks[side][square] == 0
}
