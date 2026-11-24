package board

import (
	"testing"

	"github.com/jrichard8/MarsGopherChess/internal/move"
)

func assertBoardStateEqual(t *testing.T, got *Board, want *Board, label string) {
	t.Helper()
	if got.Pieces != want.Pieces {
		t.Fatalf("%s: pieces mismatch", label)
	}
	if got.OccupancyWhite != want.OccupancyWhite {
		t.Fatalf("%s: occupancy white mismatch", label)
	}
	if got.OccupancyBlack != want.OccupancyBlack {
		t.Fatalf("%s: occupancy black mismatch", label)
	}
	if got.OccupancyBoth != want.OccupancyBoth {
		t.Fatalf("%s: occupancy both mismatch", label)
	}
	if got.SideToMove != want.SideToMove {
		t.Fatalf("%s: side to move mismatch", label)
	}
	if got.CastlingRights != want.CastlingRights {
		t.Fatalf("%s: castling rights mismatch", label)
	}
	if got.EnPassant != want.EnPassant {
		t.Fatalf("%s: en passant mismatch", label)
	}
	if got.HalfmoveClock != want.HalfmoveClock {
		t.Fatalf("%s: halfmove clock mismatch", label)
	}
	if got.FullmoveNumber != want.FullmoveNumber {
		t.Fatalf("%s: fullmove number mismatch", label)
	}
	if got.ZobristKey != want.ZobristKey {
		t.Fatalf("%s: zobrist mismatch", label)
	}
	if got.ply != want.ply {
		t.Fatalf("%s: ply mismatch", label)
	}
}

func TestMakeUnmakeNormalMove(t *testing.T) {
	b, err := FromFEN(StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	before := *b

	m := move.NewMove(G1, F3, WhiteKnight, move.NoPiece, move.NoPiece, 0)
	b.MakeMove(m)
	b.UnmakeMove()

	assertBoardStateEqual(t, b, &before, "normal")
}

func TestMakeUnmakeCapture(t *testing.T) {
	fen := "8/8/8/3p4/4P3/8/8/8 w - - 0 1"
	b, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	before := *b

	m := move.NewMove(E4, D5, WhitePawn, BlackPawn, move.NoPiece, 0)
	b.MakeMove(m)
	b.UnmakeMove()

	assertBoardStateEqual(t, b, &before, "capture")
}

func TestMakeUnmakePromotion(t *testing.T) {
	fen := "8/P7/8/8/8/8/8/8 w - - 0 1"
	b, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	before := *b

	m := move.NewMove(A7, A8, WhitePawn, move.NoPiece, WhiteQueen, 0)
	b.MakeMove(m)
	b.UnmakeMove()

	assertBoardStateEqual(t, b, &before, "promotion")
}

func TestMakeUnmakeDoublePawnPush(t *testing.T) {
	b, err := FromFEN(StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	before := *b

	m := move.NewMove(D2, D4, WhitePawn, move.NoPiece, move.NoPiece, move.FlagDoublePawnPush)
	b.MakeMove(m)
	b.UnmakeMove()

	assertBoardStateEqual(t, b, &before, "double-push")
}

func TestMakeUnmakeCastling(t *testing.T) {
	fen := "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"
	b, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}
	before := *b

	m := move.NewMove(E1, G1, WhiteKing, move.NoPiece, move.NoPiece, move.FlagCastling)
	b.MakeMove(m)
	b.UnmakeMove()

	assertBoardStateEqual(t, b, &before, "castling")
}
