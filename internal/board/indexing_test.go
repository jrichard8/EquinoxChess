package board

import "testing"

func TestSquareConstantsMapping(t *testing.T) {
	if A1 != 0 {
		t.Fatalf("A1: got %d want 0", A1)
	}
	if B1 != 1 {
		t.Fatalf("B1: got %d want 1", B1)
	}
	if H1 != 7 {
		t.Fatalf("H1: got %d want 7", H1)
	}
	if A2 != 8 {
		t.Fatalf("A2: got %d want 8", A2)
	}
	if H2 != 15 {
		t.Fatalf("H2: got %d want 15", H2)
	}
	if A8 != 56 {
		t.Fatalf("A8: got %d want 56", A8)
	}
	if H8 != 63 {
		t.Fatalf("H8: got %d want 63", H8)
	}
	if E1-A1 != 4 {
		t.Fatalf("E1-A1: got %d want 4", E1-A1)
	}
	if E2-E1 != 8 {
		t.Fatalf("E2-E1: got %d want 8", E2-E1)
	}
}

func TestSquareArithmeticSanity(t *testing.T) {
	sq := D4
	file := sq % 8
	rank := sq / 8
	if file != 3 {
		t.Fatalf("file: got %d want 3", file)
	}
	if rank != 3 {
		t.Fatalf("rank: got %d want 3", rank)
	}
}
