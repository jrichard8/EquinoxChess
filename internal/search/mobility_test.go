package search

import (
	"math/rand"
	"testing"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/move"
	"github.com/jrichard8/MarsGopherChess/internal/movegen"
)

// The September evaluator counted weighted pseudo-legal moves for both colors.
// Check blockers, attacked kings and multiple sliders against that same oracle.
func TestMobilityMatchesMoveGenerator(t *testing.T) {
	rng := rand.New(rand.NewSource(1850))
	for game := 0; game < 16; game++ {
		b := board.NewBoard()
		for ply := 0; ply < 100; ply++ {
			var moves [movegen.MaxMoves]uint32
			for _, side := range []int{board.White, board.Black} {
				view := *b
				view.SideToMove = side
				n := movegen.GeneratePseudoLegalMoves(&view, &moves)
				want := 0
				for _, m := range moves[:n] {
					switch move.MovingPiece(m) {
					case board.WhiteKnight, board.BlackKnight, board.WhiteBishop, board.BlackBishop:
						want += 4
					case board.WhiteRook, board.BlackRook:
						want += 2
					case board.WhiteQueen, board.BlackQueen:
						want++
					}
				}
				if got := countMobility(b, side); got != want {
					t.Fatalf("game %d ply %d side %d: got %d want %d", game, ply, side, got, want)
				}
			}
			n := movegen.GenerateLegalMoves(b, &moves)
			if n == 0 {
				break
			}
			b.MakeMove(moves[rng.Intn(n)])
		}
	}
}
