package main

import (
	"testing"
	"time"

	"EquinoxChess/internal/board"
)

func TestParseTimeLimitMovetime(t *testing.T) {
	got := parseTimeLimit([]string{"go", "movetime", "123"}, board.White)
	if got != 123*time.Millisecond {
		t.Fatalf("got %s want 123ms", got)
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
