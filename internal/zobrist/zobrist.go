package zobrist

const (
	pieceCount   = 12
	squareCount  = 64
	castleStates = 16
)

// Piece maps [piece][square] to a random key.
var Piece [pieceCount][squareCount]uint64

// Castling maps castling-rights bitmask (0-15) to a random key.
var Castling [castleStates]uint64

// EnPassant maps an en-passant square (0-63) to a random key.
var EnPassant [squareCount]uint64

// Side is the random key for black to move.
var Side uint64

func init() {
	seed := uint64(0x9e3779b97f4a7c15)

	next := func() uint64 {
		seed += 0x9e3779b97f4a7c15
		z := seed
		z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		return z ^ (z >> 31)
	}

	for p := 0; p < pieceCount; p++ {
		for s := 0; s < squareCount; s++ {
			Piece[p][s] = next()
		}
	}

	for i := 0; i < castleStates; i++ {
		Castling[i] = next()
	}

	for s := 0; s < squareCount; s++ {
		EnPassant[s] = next()
	}

	Side = next()
}
