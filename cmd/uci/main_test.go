package main

import (
	"testing"
	"time"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/search"
)

func TestParseTimeLimitMovetime(t *testing.T) {
	got := parseTimeLimit([]string{"go", "movetime", "123"}, board.White)
	if got != 123*time.Millisecond {
		t.Fatalf("got %s want 123ms", got)
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
	if limit != 50*time.Millisecond {
		t.Fatalf("limit = %s, want 50ms", limit)
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
