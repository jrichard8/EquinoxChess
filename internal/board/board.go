package board

import (
	"math/bits"

	"EquinoxChess/internal/zobrist"
)

// Bitboard is a 64-bit board representation.
type Bitboard uint64

const (
	White = 0
	Black = 1
)

const (
	A1 = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
)

const (
	WhitePawn = iota
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
	BlackPawn
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing
)

const (
	CastleWhiteKing  = 1 << 0
	CastleWhiteQueen = 1 << 1
	CastleBlackKing  = 1 << 2
	CastleBlackQueen = 1 << 3
)

const NoSquare = -1
const MaxDepth = 1024

const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// Board holds all state needed to describe a chess position.
type Board struct {
	Pieces         [12]Bitboard
	OccupancyWhite Bitboard
	OccupancyBlack Bitboard
	OccupancyBoth  Bitboard
	SideToMove     int
	CastlingRights uint8
	EnPassant      int
	HalfmoveClock  int
	FullmoveNumber int
	ZobristKey     uint64
	undoStack      [MaxDepth]Undo
	ply            int
}

type Undo struct {
	Move           uint32
	CastlingRights uint8
	EnPassant      int
	HalfmoveClock  int
	Zobrist        uint64
}

// NewBoard returns a board initialized to the standard starting position.
func NewBoard() *Board {
	b, err := FromFEN(StartFEN)
	if err != nil {
		return &Board{}
	}
	return b
}

func (b *Board) updateOccupancies() {
	var white Bitboard
	for i := WhitePawn; i <= WhiteKing; i++ {
		white |= b.Pieces[i]
	}
	var black Bitboard
	for i := BlackPawn; i <= BlackKing; i++ {
		black |= b.Pieces[i]
	}
	b.OccupancyWhite = white
	b.OccupancyBlack = black
	b.OccupancyBoth = white | black
}

// ComputeZobrist recalculates the full hash from the current board state.
func (b *Board) ComputeZobrist() uint64 {
	var key uint64
	for piece := 0; piece < len(b.Pieces); piece++ {
		bb := b.Pieces[piece]
		for bb != 0 {
			sq := popLSB(&bb)
			key ^= zobrist.Piece[piece][sq]
		}
	}

	key ^= zobrist.Castling[b.CastlingRights]
	if b.EnPassant != NoSquare {
		key ^= zobrist.EnPassant[b.EnPassant]
	}
	if b.SideToMove == Black {
		key ^= zobrist.Side
	}

	b.ZobristKey = key
	return key
}

// popLSB removes and returns the index of the least-significant 1 bit.
// Bit-twiddling keeps this allocation-free and fast.
func popLSB(bb *Bitboard) int {
	lsb := uint64(*bb) & (^(uint64(*bb)) + 1)
	sq := bits.TrailingZeros64(lsb)
	*bb &= *bb - 1
	return sq
}

// IsRepetition checks if the current position is a draw by threefold repetition
// from the engine's current search path.
func (b *Board) IsRepetition() bool {
	// Check positions in the undo stack
	// Only check back to the last irreversible move (capture or pawn move)
	limit := b.HalfmoveClock
	if limit > b.ply {
		limit = b.ply
	}

	// Check every other position (same side to move)
	repetitions := 0
	for i := 2; i <= limit; i += 2 {
		if b.ply >= i && b.undoStack[b.ply-i].Zobrist == b.ZobristKey {
			repetitions++
			if repetitions >= 2 {
				return true
			}
		}
	}
	return false
}

// IsInsufficientMaterial checks if the position has insufficient material to mate.
// Returns true for K vs K, K+N vs K, K+B vs K, K+B vs K+B (same color bishops).
func (b *Board) IsInsufficientMaterial() bool {
	// If there are pawns, queens, or rooks, there's sufficient material
	if b.Pieces[WhitePawn] != 0 || b.Pieces[BlackPawn] != 0 ||
		b.Pieces[WhiteQueen] != 0 || b.Pieces[BlackQueen] != 0 ||
		b.Pieces[WhiteRook] != 0 || b.Pieces[BlackRook] != 0 {
		return false
	}

	whiteKnights := bits.OnesCount64(uint64(b.Pieces[WhiteKnight]))
	blackKnights := bits.OnesCount64(uint64(b.Pieces[BlackKnight]))
	whiteBishops := bits.OnesCount64(uint64(b.Pieces[WhiteBishop]))
	blackBishops := bits.OnesCount64(uint64(b.Pieces[BlackBishop]))

	totalMinor := whiteKnights + blackKnights + whiteBishops + blackBishops

	// K vs K
	if totalMinor == 0 {
		return true
	}

	// K+N vs K or K+B vs K
	if totalMinor == 1 {
		return true
	}

	// K+B vs K+B with same color bishops
	if totalMinor == 2 && whiteBishops == 1 && blackBishops == 1 {
		// Check if bishops are on same color squares
		whiteBishopSq := bits.TrailingZeros64(uint64(b.Pieces[WhiteBishop]))
		blackBishopSq := bits.TrailingZeros64(uint64(b.Pieces[BlackBishop]))

		// Square color is determined by (rank + file) % 2
		whiteSquareColor := (whiteBishopSq/8 + whiteBishopSq%8) % 2
		blackSquareColor := (blackBishopSq/8 + blackBishopSq%8) % 2

		if whiteSquareColor == blackSquareColor {
			return true
		}
	}

	return false
}
