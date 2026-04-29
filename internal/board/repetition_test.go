package board

import (
	"testing"

	"EquinoxChess/internal/move"
)

func playMoves(b *Board, moves ...uint32) {
	for _, m := range moves {
		b.MakeMove(m)
	}
}

func knightMove(from, to, side int) uint32 {
	piece := WhiteKnight
	if side == Black {
		piece = BlackKnight
	}
	return move.NewMove(from, to, piece, move.NoPiece, move.NoPiece, 0)
}

func TestIsRepetitionRequiresThreefold(t *testing.T) {
	t.Helper()
	b, err := FromFEN(StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	playMoves(
		b,
		knightMove(G1, F3, White),
		knightMove(G8, F6, Black),
		knightMove(F3, G1, White),
		knightMove(F6, G8, Black),
	)
	if b.IsRepetition() {
		t.Fatalf("expected first repeated position to not count as draw")
	}

	playMoves(
		b,
		knightMove(G1, F3, White),
		knightMove(G8, F6, Black),
		knightMove(F3, G1, White),
		knightMove(F6, G8, Black),
	)
	if !b.IsRepetition() {
		t.Fatalf("expected third occurrence to count as repetition draw")
	}
}

func TestIsRepetitionStopsAtIrreversibleMove(t *testing.T) {
	b, err := FromFEN(StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	playMoves(
		b,
		knightMove(G1, F3, White),
		knightMove(G8, F6, Black),
		knightMove(F3, G1, White),
		knightMove(F6, G8, Black),
		move.NewMove(E2, E4, WhitePawn, move.NoPiece, move.NoPiece, move.FlagDoublePawnPush),
		move.NewMove(E7, E5, BlackPawn, move.NoPiece, move.NoPiece, move.FlagDoublePawnPush),
		knightMove(G1, F3, White),
		knightMove(G8, F6, Black),
		knightMove(F3, G1, White),
		knightMove(F6, G8, Black),
	)

	if b.IsRepetition() {
		t.Fatalf("expected irreversible pawn moves to reset repetition history")
	}
}
