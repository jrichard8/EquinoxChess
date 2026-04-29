package movegen

import "EquinoxChess/internal/board"

var knightAttacks [64]board.Bitboard
var kingAttacks [64]board.Bitboard

func init() {
	for sq := 0; sq < 64; sq++ {
		file := sq % 8
		rank := sq / 8

		var n board.Bitboard
		knightDeltas := [8][2]int{{1, 2}, {2, 1}, {-1, 2}, {-2, 1}, {1, -2}, {2, -1}, {-1, -2}, {-2, -1}}
		for i := 0; i < len(knightDeltas); i++ {
			f := file + knightDeltas[i][0]
			r := rank + knightDeltas[i][1]
			if f >= 0 && f < 8 && r >= 0 && r < 8 {
				n |= board.Bitboard(1) << (r*8 + f)
			}
		}
		knightAttacks[sq] = n

		var k board.Bitboard
		kingDeltas := [8][2]int{{1, 1}, {1, 0}, {1, -1}, {0, 1}, {0, -1}, {-1, 1}, {-1, 0}, {-1, -1}}
		for i := 0; i < len(kingDeltas); i++ {
			f := file + kingDeltas[i][0]
			r := rank + kingDeltas[i][1]
			if f >= 0 && f < 8 && r >= 0 && r < 8 {
				k |= board.Bitboard(1) << (r*8 + f)
			}
		}
		kingAttacks[sq] = k
	}
}
