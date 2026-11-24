package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"EquinoxChess/internal/api"
	"EquinoxChess/internal/board"
	"EquinoxChess/internal/movegen"
	"EquinoxChess/internal/search"
	"EquinoxChess/internal/uci"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	game := api.NewGame()
	var searchDone chan struct{}
	var searching atomic.Bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		cmd := fields[0]

		switch cmd {
		case "uci":
			fmt.Println("id name Owl-Gopher")
			fmt.Println("id author Julien Richard")
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			game.NewGame()
		case "position":
			game = handlePosition(fields, game)
		case "go":
			depth, timeLimit := parseSearchLimits(fields, game.Board.SideToMove)
			if searching.Load() {
				game.Search.Stop.Store(true)
				if searchDone != nil {
					<-searchDone
				}
			}
			game.Search.Stop.Store(false)
			game.Search.StartTime = time.Now()
			game.Search.SoftTimeLimit, game.Search.TimeLimit = iterationBudgets(fields, timeLimit)
			searching.Store(true)
			searchDone = make(chan struct{})
			go func(done chan struct{}) {
				defer func() {
					searching.Store(false)
					close(done)
				}()
				defer func() {
					if r := recover(); r != nil {
						fmt.Fprintf(os.Stderr, "PANIC in search: %v\n", r)
						_ = os.Stderr.Sync()
						fmt.Printf("bestmove %s\n", fallbackBestMove(game.Board))
						_ = os.Stdout.Sync()
					}
				}()
				best, score, bestDepth, nodes := game.Search.SearchRoot(game.Board, depth)
				fmt.Printf("info depth %d %s nodes %d\n", bestDepth, formatScore(score), nodes)
				moveStr := "0000"
				if best != 0 {
					moveStr = uci.MoveToUCI(best)
				}
				if best == 0 {
					var moves [movegen.MaxMoves]uint32
					if movegen.GenerateLegalMoves(game.Board, &moves) > 0 {
						best = moves[0]
						moveStr = uci.MoveToUCI(best)
					} else {
						fmt.Fprintf(os.Stderr, "WARNING: bestmove is zero\n")
						_ = os.Stderr.Sync()
					}
				}
				fmt.Printf("bestmove %s\n", moveStr)
				_ = os.Stdout.Sync()
				fmt.Fprintf(os.Stderr, "Search completed normally\n")
				_ = os.Stderr.Sync()
			}(searchDone)
		case "stop":
			if searching.Load() {
				game.Search.Stop.Store(true)
				if searchDone != nil {
					<-searchDone
				}
			}
		case "quit":
			if searching.Load() {
				game.Search.Stop.Store(true)
				if searchDone != nil {
					<-searchDone
				}
			}
			return
		}
	}
}

func handlePosition(fields []string, game *api.Game) *api.Game {
	if len(fields) < 2 {
		return game
	}
	idx := 1
	if fields[idx] == "startpos" {
		// "position" replaces the board, not the game. Keep search state so TT,
		// history, and killers remain useful until "ucinewgame" resets them.
		game.Board = board.NewBoard()
		idx++
	} else if fields[idx] == "fen" {
		if len(fields) < idx+7 {
			return game
		}
		fen := strings.Join(fields[idx+1:idx+7], " ")
		b, err := board.FromFEN(fen)
		if err == nil {
			game.Board = b
		}
		idx += 7
	}
	if idx < len(fields) && fields[idx] == "moves" {
		idx++
		for ; idx < len(fields); idx++ {
			_ = game.PlayMove(fields[idx])
		}
	}
	return game
}

// Clock allocation leaves room to finish one more iteration. Explicit
// movetime and emergency allocations retain a single hard deadline.
func iterationBudgets(fields []string, budget time.Duration) (soft, hard time.Duration) {
	if budget < 10*time.Millisecond || hasGoToken(fields, "movetime") {
		return 0, budget
	}
	return budget, 2 * budget
}

func parseSearchLimits(fields []string, side int) (int, time.Duration) {
	timeLimit := parseTimeLimit(fields, side)
	if depth, ok := parseExplicitDepth(fields); ok {
		return depth, timeLimit
	}
	if timeLimit > 0 || hasGoToken(fields, "infinite") {
		return search.MaxDepth - 1, timeLimit
	}
	return 4, 0
}

func parseDepth(fields []string) int {
	if depth, ok := parseExplicitDepth(fields); ok {
		return depth
	}
	return 4
}

func parseExplicitDepth(fields []string) (int, bool) {
	for i := 0; i+1 < len(fields); i++ {
		if fields[i] == "depth" {
			if d, err := strconv.Atoi(fields[i+1]); err == nil {
				return d, true
			}
		}
	}
	return 0, false
}

func hasGoToken(fields []string, token string) bool {
	for i := 1; i < len(fields); i++ {
		if fields[i] == token {
			return true
		}
	}
	return false
}

func parseTimeLimit(fields []string, side int) time.Duration {
	var wtime, btime, winc, binc int64
	for i := 0; i+1 < len(fields); i++ {
		switch fields[i] {
		case "movetime":
			if v, err := strconv.ParseInt(fields[i+1], 10, 64); err == nil {
				if v < 1 {
					v = 1
				}
				return time.Duration(v) * time.Millisecond
			}
		case "wtime":
			if v, err := strconv.ParseInt(fields[i+1], 10, 64); err == nil {
				wtime = v
			}
		case "btime":
			if v, err := strconv.ParseInt(fields[i+1], 10, 64); err == nil {
				btime = v
			}
		case "winc":
			if v, err := strconv.ParseInt(fields[i+1], 10, 64); err == nil {
				winc = v
			}
		case "binc":
			if v, err := strconv.ParseInt(fields[i+1], 10, 64); err == nil {
				binc = v
			}
		}
	}

	var remaining, inc int64
	if side == board.White {
		remaining, inc = wtime, winc
	} else {
		remaining, inc = btime, binc
	}
	if remaining <= 0 && inc <= 0 {
		return 0
	}
	if inc > 0 {
		const (
			movesHorizon      int64 = 40
			incrementShareNum int64 = 3
			incrementShareDen int64 = 4
			clockCapDivisor   int64 = 10
			moveOverheadMs    int64 = 15
		)

		// Spread the clock over the horizon, then reserve anti-flag headroom.
		allocationMs := remaining/movesHorizon + incrementShareNum*inc/incrementShareDen
		if clockCapMs := remaining / clockCapDivisor; allocationMs > clockCapMs {
			allocationMs = clockCapMs
		}
		allocationMs -= moveOverheadMs
		if allocationMs < 1 {
			allocationMs = 1
		}
		return time.Duration(allocationMs) * time.Millisecond
	}
	allocationMs := remaining / 50
	if allocationMs < 1 {
		allocationMs = 1
	}
	return time.Duration(allocationMs) * time.Millisecond
}

func fallbackBestMove(b *board.Board) string {
	var moves [movegen.MaxMoves]uint32
	if movegen.GenerateLegalMoves(b, &moves) == 0 {
		return "0000"
	}
	return uci.MoveToUCI(moves[0])
}

func formatScore(score int) string {
	if score > search.MateThreshold {
		plies := search.MateScore - score
		return fmt.Sprintf("score mate %d", mateMoves(plies))
	}
	if score < -search.MateThreshold {
		plies := search.MateScore + score
		return fmt.Sprintf("score mate -%d", mateMoves(plies))
	}
	return fmt.Sprintf("score cp %d", score)
}

func mateMoves(plies int) int {
	moves := (plies + 1) / 2
	if moves < 1 {
		return 1
	}
	return moves
}
