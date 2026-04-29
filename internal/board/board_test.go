package board

import "testing"

func bb(squares ...int) Bitboard {
	var out Bitboard
	for _, sq := range squares {
		out |= Bitboard(1) << sq
	}
	return out
}

func TestFromFENInitialPosition(t *testing.T) {
	b, err := FromFEN(StartFEN)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	if b.SideToMove != White {
		t.Fatalf("side to move: got %d want %d", b.SideToMove, White)
	}

	if b.CastlingRights != (CastleWhiteKing|CastleWhiteQueen|CastleBlackKing|CastleBlackQueen) {
		t.Fatalf("castling rights: got %d", b.CastlingRights)
	}

	if b.EnPassant != NoSquare {
		t.Fatalf("en passant: got %d", b.EnPassant)
	}

	if b.ZobristKey == 0 {
		t.Fatalf("zobrist key should be non-zero")
	}

	if b.Pieces[WhitePawn] != bb(A2, B2, C2, D2, E2, F2, G2, H2) {
		t.Fatalf("white pawns mismatch")
	}
	if b.Pieces[WhiteKnight] != bb(B1, G1) {
		t.Fatalf("white knights mismatch")
	}
	if b.Pieces[WhiteBishop] != bb(C1, F1) {
		t.Fatalf("white bishops mismatch")
	}
	if b.Pieces[WhiteRook] != bb(A1, H1) {
		t.Fatalf("white rooks mismatch")
	}
	if b.Pieces[WhiteQueen] != bb(D1) {
		t.Fatalf("white queen mismatch")
	}
	if b.Pieces[WhiteKing] != bb(E1) {
		t.Fatalf("white king mismatch")
	}

	if b.Pieces[BlackPawn] != bb(A7, B7, C7, D7, E7, F7, G7, H7) {
		t.Fatalf("black pawns mismatch")
	}
	if b.Pieces[BlackKnight] != bb(B8, G8) {
		t.Fatalf("black knights mismatch")
	}
	if b.Pieces[BlackBishop] != bb(C8, F8) {
		t.Fatalf("black bishops mismatch")
	}
	if b.Pieces[BlackRook] != bb(A8, H8) {
		t.Fatalf("black rooks mismatch")
	}
	if b.Pieces[BlackQueen] != bb(D8) {
		t.Fatalf("black queen mismatch")
	}
	if b.Pieces[BlackKing] != bb(E8) {
		t.Fatalf("black king mismatch")
	}

	whiteOcc := bb(A1, B1, C1, D1, E1, F1, G1, H1, A2, B2, C2, D2, E2, F2, G2, H2)
	blackOcc := bb(A8, B8, C8, D8, E8, F8, G8, H8, A7, B7, C7, D7, E7, F7, G7, H7)
	if b.OccupancyWhite != whiteOcc {
		t.Fatalf("white occupancy mismatch")
	}
	if b.OccupancyBlack != blackOcc {
		t.Fatalf("black occupancy mismatch")
	}
	if b.OccupancyBoth != (whiteOcc | blackOcc) {
		t.Fatalf("both occupancy mismatch")
	}
}

func TestFromFENCustomPosition(t *testing.T) {
	fen := "rnbqkbnr/pppp1ppp/8/4p3/3P4/8/PPP1PPPP/RNBQKBNR b Kq d3 0 2"
	b, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN failed: %v", err)
	}

	if b.SideToMove != Black {
		t.Fatalf("side to move: got %d want %d", b.SideToMove, Black)
	}

	if b.CastlingRights != (CastleWhiteKing | CastleBlackQueen) {
		t.Fatalf("castling rights: got %d", b.CastlingRights)
	}

	if b.EnPassant != D3 {
		t.Fatalf("en passant: got %d want %d", b.EnPassant, D3)
	}

	if b.ZobristKey == 0 {
		t.Fatalf("zobrist key should be non-zero")
	}

	if b.Pieces[WhitePawn] != bb(A2, B2, C2, D4, E2, F2, G2, H2) {
		t.Fatalf("white pawns mismatch")
	}
	if b.Pieces[WhiteKnight] != bb(B1, G1) {
		t.Fatalf("white knights mismatch")
	}
	if b.Pieces[WhiteBishop] != bb(C1, F1) {
		t.Fatalf("white bishops mismatch")
	}
	if b.Pieces[WhiteRook] != bb(A1, H1) {
		t.Fatalf("white rooks mismatch")
	}
	if b.Pieces[WhiteQueen] != bb(D1) {
		t.Fatalf("white queen mismatch")
	}
	if b.Pieces[WhiteKing] != bb(E1) {
		t.Fatalf("white king mismatch")
	}

	if b.Pieces[BlackPawn] != bb(A7, B7, C7, D7, E5, F7, G7, H7) {
		t.Fatalf("black pawns mismatch")
	}
	if b.Pieces[BlackKnight] != bb(B8, G8) {
		t.Fatalf("black knights mismatch")
	}
	if b.Pieces[BlackBishop] != bb(C8, F8) {
		t.Fatalf("black bishops mismatch")
	}
	if b.Pieces[BlackRook] != bb(A8, H8) {
		t.Fatalf("black rooks mismatch")
	}
	if b.Pieces[BlackQueen] != bb(D8) {
		t.Fatalf("black queen mismatch")
	}
	if b.Pieces[BlackKing] != bb(E8) {
		t.Fatalf("black king mismatch")
	}

	whiteOcc := bb(A1, B1, C1, D1, E1, F1, G1, H1, A2, B2, C2, D4, E2, F2, G2, H2)
	blackOcc := bb(A8, B8, C8, D8, E8, F8, G8, H8, A7, B7, C7, D7, E5, F7, G7, H7)
	if b.OccupancyWhite != whiteOcc {
		t.Fatalf("white occupancy mismatch")
	}
	if b.OccupancyBlack != blackOcc {
		t.Fatalf("black occupancy mismatch")
	}
	if b.OccupancyBoth != (whiteOcc | blackOcc) {
		t.Fatalf("both occupancy mismatch")
	}
}
