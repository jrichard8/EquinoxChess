package board

import "testing"

func TestPieceIndexOrder(t *testing.T) {
	if WhitePawn != 0 {
		t.Fatalf("WhitePawn: got %d want 0", WhitePawn)
	}
	if WhiteKnight != 1 {
		t.Fatalf("WhiteKnight: got %d want 1", WhiteKnight)
	}
	if WhiteBishop != 2 {
		t.Fatalf("WhiteBishop: got %d want 2", WhiteBishop)
	}
	if WhiteRook != 3 {
		t.Fatalf("WhiteRook: got %d want 3", WhiteRook)
	}
	if WhiteQueen != 4 {
		t.Fatalf("WhiteQueen: got %d want 4", WhiteQueen)
	}
	if WhiteKing != 5 {
		t.Fatalf("WhiteKing: got %d want 5", WhiteKing)
	}
	if BlackPawn != 6 {
		t.Fatalf("BlackPawn: got %d want 6", BlackPawn)
	}
	if BlackKnight != 7 {
		t.Fatalf("BlackKnight: got %d want 7", BlackKnight)
	}
	if BlackBishop != 8 {
		t.Fatalf("BlackBishop: got %d want 8", BlackBishop)
	}
	if BlackRook != 9 {
		t.Fatalf("BlackRook: got %d want 9", BlackRook)
	}
	if BlackQueen != 10 {
		t.Fatalf("BlackQueen: got %d want 10", BlackQueen)
	}
	if BlackKing != 11 {
		t.Fatalf("BlackKing: got %d want 11", BlackKing)
	}

	if len(Board{}.Pieces) != 12 {
		t.Fatalf("Pieces length: got %d want 12", len(Board{}.Pieces))
	}
}
