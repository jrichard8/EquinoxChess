package main

import (
	"fmt"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/move"
	"github.com/jrichard8/MarsGopherChess/internal/movegen"
	"github.com/jrichard8/MarsGopherChess/internal/perft"
)

func main() {
	fen := "8/4P3/8/8/8/8/8/4k2K w - - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		panic(err)
	}

	var movesArr [movegen.MaxMoves]uint32
	count := movegen.GenerateLegalMoves(b, &movesArr)
	for i := 0; i < count; i++ {
		m := movesArr[i]
		uci := perft.MoveToUCI(m)
		var moveStr string
		if move.PromotionPiece(m) != move.NoPiece {
			moveStr = fmt.Sprintf("%s%c", uci[:4], uci[4])
		} else {
			moveStr = fmt.Sprintf("%s", uci[:4])
		}
		if len(moveStr) >= 4 && moveStr[:4] == "e7e8" {
			fmt.Println(moveStr)
		}
	}
}
