package search

import (
	"testing"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/tt"
)

func BenchmarkNegamax(b *testing.B) {
	pos, err := board.FromFEN(board.StartFEN)
	if err != nil {
		b.Fatalf("FromFEN failed: %v", err)
	}
	var s Search
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Nodes = 0
		s.Ply = 0
		_ = s.Negamax(pos, 3, -Infinity, Infinity)
	}
}

func TestSearchReturnsLegalMoveFromStartPosition(t *testing.T) {
	pos, err := board.FromFEN(board.StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	s := Search{TT: tt.NewTT(1 << 10)}
	best, score, depth, _ := s.SearchRoot(pos, 2)

	if best == 0 {
		t.Fatalf("expected a legal move, got none")
	}
	if depth == 0 {
		t.Fatalf("expected search depth > 0")
	}
	if score <= -Infinity || score >= Infinity {
		t.Fatalf("expected bounded score, got %d", score)
	}
}

func TestSearchRecognizesStalemateAsDraw(t *testing.T) {
	pos, err := board.FromFEN("7k/5Q2/7K/8/8/8/8/8 b - - 0 1")
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	s := Search{TT: tt.NewTT(1 << 10)}
	best, score, depth, _ := s.SearchRoot(pos, 2)

	if best != 0 {
		t.Fatalf("expected no best move in stalemate, got %d", best)
	}
	if depth == 0 {
		t.Fatalf("expected search to inspect at least one depth")
	}
	if score != DrawScore {
		t.Fatalf("expected draw score in stalemate, got %d", score)
	}
}

func TestQuiescenceRecognizesStalemateAsDraw(t *testing.T) {
	pos, err := board.FromFEN("7k/5Q2/7K/8/8/8/8/8 b - - 0 1")
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	s := Search{TT: tt.NewTT(1 << 10)}
	score := s.Quiescence(pos, -Infinity, Infinity)

	if score != DrawScore {
		t.Fatalf("expected quiescence draw score in stalemate, got %d", score)
	}
}
