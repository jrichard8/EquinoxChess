package perft

import (
	"EquinoxChess/internal/board"
	"EquinoxChess/internal/movegen"
)

func Perft(b *board.Board, depth int) uint64 {
	if depth == 0 {
		return 1
	}

	var moves [movegen.MaxMoves]uint32
	count := movegen.GenerateLegalMoves(b, &moves)
	var nodes uint64
	for i := 0; i < count; i++ {
		b.MakeMove(moves[i])
		nodes += Perft(b, depth-1)
		b.UnmakeMove()
	}
	return nodes
}
