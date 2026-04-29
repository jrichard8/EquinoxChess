package perft

import (
	"testing"

	"EquinoxChess/internal/board"
)

func TestPerftInitialPosition(t *testing.T) {
	b, err := board.FromFEN(board.StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	cases := []struct {
		depth int
		want  uint64
	}{
		{1, 20},
		{2, 400},
		{3, 8902},
		{4, 197281},
		{5, 4865609},
	}

	for _, tc := range cases {
		got := Perft(b, tc.depth)
		if got != tc.want {
			t.Fatalf("depth %d: got %d want %d", tc.depth, got, tc.want)
		}
	}
}

func TestPerftKiwipete(t *testing.T) {
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	cases := []struct {
		depth int
		want  uint64
	}{
		{1, 48},
		{2, 2039},
		{3, 97862},
		{4, 4085603},
	}

	for _, tc := range cases {
		got := Perft(b, tc.depth)
		if got != tc.want {
			t.Fatalf("depth %d: got %d want %d", tc.depth, got, tc.want)
		}
	}
}

func TestPerftEnPassantPosition(t *testing.T) {
	fen := "8/8/8/3pP3/8/8/8/4K2k w - d6 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	got := Perft(b, 1)
	if got != 7 {
		t.Fatalf("depth 1: got %d want 7", got)
	}
}

func TestPerftCastlingPosition(t *testing.T) {
	fen := "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"
	b, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	got := Perft(b, 1)
	if got != 26 {
		t.Fatalf("depth 1: got %d want 26", got)
	}
}
