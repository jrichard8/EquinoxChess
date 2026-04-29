package board

import "testing"

func TestZobristDeterministicForSamePosition(t *testing.T) {
	b1, err := FromFEN(StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	b2, err := FromFEN(StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	if b1.ZobristKey != b2.ZobristKey {
		t.Fatalf("zobrist keys differ: %d vs %d", b1.ZobristKey, b2.ZobristKey)
	}
}

func TestZobristChangesWithSideToMove(t *testing.T) {
	fenWhite := "8/8/8/8/8/8/8/8 w - - 0 1"
	fenBlack := "8/8/8/8/8/8/8/8 b - - 0 1"

	bWhite, err := FromFEN(fenWhite)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	bBlack, err := FromFEN(fenBlack)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	if bWhite.ZobristKey == bBlack.ZobristKey {
		t.Fatalf("expected different zobrist keys for side to move")
	}
}

func TestZobristChangesWithCastlingRights(t *testing.T) {
	fenNone := "8/8/8/8/8/8/8/8 w - - 0 1"
	fenK := "8/8/8/8/8/8/8/8 w K - 0 1"

	bNone, err := FromFEN(fenNone)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	bK, err := FromFEN(fenK)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	if bNone.ZobristKey == bK.ZobristKey {
		t.Fatalf("expected different zobrist keys for castling rights")
	}
}

func TestZobristChangesWithEnPassant(t *testing.T) {
	fenNone := "8/8/8/8/8/8/8/8 w - - 0 1"
	fenEP := "8/8/8/8/8/8/8/8 w - e3 0 1"

	bNone, err := FromFEN(fenNone)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	bEP, err := FromFEN(fenEP)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	if bNone.ZobristKey == bEP.ZobristKey {
		t.Fatalf("expected different zobrist keys for en passant")
	}
}

func TestZobristStableRecompute(t *testing.T) {
	b, err := FromFEN(StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	first := b.ComputeZobrist()
	second := b.ComputeZobrist()
	if first != second {
		t.Fatalf("zobrist recompute mismatch: %d vs %d", first, second)
	}
}

func TestComputeZobrist_PanicsOnInvalidCastlingRightsIndex(t *testing.T) {
	b := &Board{CastlingRights: 255, EnPassant: NoSquare}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on invalid castling rights index")
		}
	}()

	_ = b.ComputeZobrist()
}
