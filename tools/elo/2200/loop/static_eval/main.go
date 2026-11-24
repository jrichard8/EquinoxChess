// Command static_eval reads FENs, writing one static score in White's perspective per line.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/search"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	for scanner.Scan() {
		b, err := board.FromFEN(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
		score := search.Evaluate(b)
		if b.SideToMove == board.Black {
			score = -score
		}
		fmt.Fprintln(out, score)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if err := out.Flush(); err != nil {
		log.Fatal(err)
	}
}
