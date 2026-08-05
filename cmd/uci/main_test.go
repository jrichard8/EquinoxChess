package main

import (
	"testing"
	"time"

	"EquinoxChess/internal/api"
	"EquinoxChess/internal/board"
	"EquinoxChess/internal/search"
	"EquinoxChess/internal/tt"
)

func TestHandlePositionStartposPreservesSearchState(t *testing.T) {
	game := api.NewGame()
	ttBefore := game.Search.TT
	const (
		ttKey      = uint64(0x13579bdf2468ace0)
		ttHalfmove = 73
		ttDepth    = 6
		ttScore    = 321
		ttMove     = uint32(0x11223344)
	)
	game.Search.TT.Store(ttKey, ttHalfmove, ttDepth, ttScore, tt.TTExact, ttMove, 0, search.MateThreshold)
	game.Search.History[3][27] = 1234
	game.Search.KillerMoves[5][0] = 0x10203040

	got := handlePosition(
		[]string{"position", "startpos", "moves", "e2e4", "e7e5", "g1f3"},
		game,
	)

	if got != game {
		t.Fatal("handlePosition replaced the game")
	}
	if got.Search.TT != ttBefore {
		t.Fatal("handlePosition replaced the transposition table")
	}
	if score, move, cutoff := got.Search.TT.Probe(ttKey, ttHalfmove, ttDepth, -search.Infinity, search.Infinity, 0, search.MateThreshold); !cutoff || score != ttScore || move != ttMove {
		t.Fatalf("TT sentinel Probe = (%d, %#x, %t), want (%d, %#x, true)", score, move, cutoff, ttScore, ttMove)
	}
	if got.Search.History[3][27] != 1234 {
		t.Fatalf("history sentinel = %d, want 1234", got.Search.History[3][27])
	}
	if got.Search.KillerMoves[5][0] != 0x10203040 {
		t.Fatalf("killer sentinel = %#x, want %#x", got.Search.KillerMoves[5][0], uint32(0x10203040))
	}
	const wantFEN = "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 1"
	if gotFEN := got.FEN(); gotFEN != wantFEN {
		t.Fatalf("FEN = %q, want %q", gotFEN, wantFEN)
	}
}

func TestHandlePositionFENPreservesSearchState(t *testing.T) {
	game := api.NewGame()
	ttBefore := game.Search.TT
	const (
		ttKey      = uint64(0x2468ace013579bdf)
		ttHalfmove = 81
		ttDepth    = 7
		ttScore    = -234
		ttMove     = uint32(0x55667788)
	)
	game.Search.TT.Store(ttKey, ttHalfmove, ttDepth, ttScore, tt.TTExact, ttMove, 0, search.MateThreshold)
	game.Search.History[8][42] = 5678
	game.Search.KillerMoves[9][1] = 0x50607080

	got := handlePosition(
		[]string{
			"position", "fen",
			"rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR", "w", "KQkq", "e6", "0", "2",
			"moves", "g1f3", "b8c6",
		},
		game,
	)

	if got != game {
		t.Fatal("handlePosition replaced the game")
	}
	if got.Search.TT != ttBefore {
		t.Fatal("handlePosition replaced the transposition table")
	}
	if score, move, cutoff := got.Search.TT.Probe(ttKey, ttHalfmove, ttDepth, -search.Infinity, search.Infinity, 0, search.MateThreshold); !cutoff || score != ttScore || move != ttMove {
		t.Fatalf("TT sentinel Probe = (%d, %#x, %t), want (%d, %#x, true)", score, move, cutoff, ttScore, ttMove)
	}
	if got.Search.History[8][42] != 5678 {
		t.Fatalf("history sentinel = %d, want 5678", got.Search.History[8][42])
	}
	if got.Search.KillerMoves[9][1] != 0x50607080 {
		t.Fatalf("killer sentinel = %#x, want %#x", got.Search.KillerMoves[9][1], uint32(0x50607080))
	}
	const wantFEN = "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 2"
	if gotFEN := got.FEN(); gotFEN != wantFEN {
		t.Fatalf("FEN = %q, want %q", gotFEN, wantFEN)
	}
}

func TestNewGameResetsSearchState(t *testing.T) {
	game := api.NewGame()
	ttBefore := game.Search.TT
	game.Search.History[1][2] = 99
	game.Search.KillerMoves[3][0] = 0x1234

	game.NewGame()

	if game.Search.TT == ttBefore {
		t.Fatal("NewGame retained the transposition table")
	}
	if game.Search.History[1][2] != 0 {
		t.Fatalf("history sentinel survived NewGame: %d", game.Search.History[1][2])
	}
	if game.Search.KillerMoves[3][0] != 0 {
		t.Fatalf("killer sentinel survived NewGame: %#x", game.Search.KillerMoves[3][0])
	}
}

func TestParseTimeLimitMovetime(t *testing.T) {
	got := parseTimeLimit([]string{"go", "movetime", "123"}, board.White)
	if got != 123*time.Millisecond {
		t.Fatalf("got %s want 123ms", got)
	}
}

func TestParseTimeLimitClockWithIncrement(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		side   int
		want   time.Duration
	}{
		{
			name:   "white 1s plus 100ms",
			fields: []string{"go", "wtime", "1000", "btime", "60000", "winc", "100", "binc", "250"},
			side:   board.White,
			want:   85 * time.Millisecond,
		},
		{
			name:   "black 1s plus 100ms",
			fields: []string{"go", "wtime", "60000", "btime", "1000", "winc", "250", "binc", "100"},
			side:   board.Black,
			want:   85 * time.Millisecond,
		},
		{
			name:   "white 60s plus 100ms",
			fields: []string{"go", "wtime", "60000", "btime", "20", "winc", "100", "binc", "500"},
			side:   board.White,
			want:   1560 * time.Millisecond,
		},
		{
			name:   "black 60s plus 100ms",
			fields: []string{"go", "wtime", "20", "btime", "60000", "winc", "500", "binc", "100"},
			side:   board.Black,
			want:   1560 * time.Millisecond,
		},
		{
			name:   "white 100ms cap",
			fields: []string{"go", "wtime", "100", "btime", "60000", "winc", "100", "binc", "500"},
			side:   board.White,
			want:   time.Millisecond,
		},
		{
			name:   "black 100ms cap",
			fields: []string{"go", "wtime", "60000", "btime", "100", "winc", "500", "binc", "100"},
			side:   board.Black,
			want:   time.Millisecond,
		},
		{
			name:   "white 16ms pre-overhead boundary",
			fields: []string{"go", "wtime", "160", "btime", "60000", "winc", "100", "binc", "500"},
			side:   board.White,
			want:   time.Millisecond,
		},
		{
			name:   "black 17ms pre-overhead boundary",
			fields: []string{"go", "wtime", "60000", "btime", "170", "winc", "500", "binc", "100"},
			side:   board.Black,
			want:   2 * time.Millisecond,
		},
		{
			name:   "white 20ms cap",
			fields: []string{"go", "wtime", "20", "btime", "60000", "winc", "100", "binc", "500"},
			side:   board.White,
			want:   time.Millisecond,
		},
		{
			name:   "black 20ms cap",
			fields: []string{"go", "wtime", "60000", "btime", "20", "winc", "500", "binc", "100"},
			side:   board.Black,
			want:   time.Millisecond,
		},
		{
			name:   "white exhausted clock floor",
			fields: []string{"go", "wtime", "0", "btime", "60000", "winc", "100", "binc", "500"},
			side:   board.White,
			want:   time.Millisecond,
		},
		{
			name:   "black exhausted clock floor",
			fields: []string{"go", "wtime", "60000", "btime", "0", "winc", "500", "binc", "100"},
			side:   board.Black,
			want:   time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTimeLimit(tt.fields, tt.side); got != tt.want {
				t.Fatalf("parseTimeLimit(%v, %d) = %s, want %s", tt.fields, tt.side, got, tt.want)
			}
		})
	}
}

func TestParseTimeLimitClockWithoutIncrement(t *testing.T) {
	got := parseTimeLimit([]string{"go", "wtime", "1000", "winc", "0"}, board.White)
	if got != 20*time.Millisecond {
		t.Fatalf("got %s want 20ms", got)
	}
}

func TestParseSearchLimitsDefaultsToDepthFourWithoutClock(t *testing.T) {
	depth, limit := parseSearchLimits([]string{"go"}, board.White)
	if depth != 4 {
		t.Fatalf("depth = %d, want 4", depth)
	}
	if limit != 0 {
		t.Fatalf("limit = %s, want 0", limit)
	}
}

func TestParseSearchLimitsKeepsExplicitDepth(t *testing.T) {
	depth, limit := parseSearchLimits([]string{"go", "depth", "6", "movetime", "123"}, board.White)
	if depth != 6 {
		t.Fatalf("depth = %d, want 6", depth)
	}
	if limit != 123*time.Millisecond {
		t.Fatalf("limit = %s, want 123ms", limit)
	}
}

func TestParseSearchLimitsUsesMaxDepthForMovetime(t *testing.T) {
	depth, limit := parseSearchLimits([]string{"go", "movetime", "123"}, board.White)
	if depth != search.MaxDepth-1 {
		t.Fatalf("depth = %d, want %d", depth, search.MaxDepth-1)
	}
	if limit != 123*time.Millisecond {
		t.Fatalf("limit = %s, want 123ms", limit)
	}
}

func TestParseSearchLimitsUsesMaxDepthForClock(t *testing.T) {
	depth, limit := parseSearchLimits([]string{"go", "wtime", "60000", "btime", "60000", "winc", "100", "binc", "100"}, board.White)
	if depth != search.MaxDepth-1 {
		t.Fatalf("depth = %d, want %d", depth, search.MaxDepth-1)
	}
	if limit != 1560*time.Millisecond {
		t.Fatalf("limit = %s, want 1560ms", limit)
	}
}

func TestFallbackBestMoveReturnsLegalMove(t *testing.T) {
	b := board.NewBoard()
	if got := fallbackBestMove(b); got == "" || got == "0000" {
		t.Fatalf("expected legal fallback move, got %q", got)
	}
}

func TestFormatScoreUsesMateNotationForMateScores(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{
		{name: "centipawns", score: 42, want: "score cp 42"},
		{name: "mate for engine", score: 31999, want: "score mate 1"},
		{name: "mate against engine", score: -31999, want: "score mate -1"},
		{name: "mate not clipped", score: -31997, want: "score mate -2"},
		{name: "mate score boundary", score: -32000, want: "score mate -1"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := formatScore(tt.score); got != tt.want {
				t.Fatalf("formatScore(%d) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}
