package board

import "testing"

func orPieces(pieces [12]Bitboard, start, end int) Bitboard {
	var out Bitboard
	for i := start; i <= end; i++ {
		out |= pieces[i]
	}
	return out
}

func TestOccupancyRecompute(t *testing.T) {
	cases := []struct {
		name string
		fen  string
	}{
		{
			name: "start",
			fen:  StartFEN,
		},
		{
			name: "dispersed",
			fen:  "8/8/3p4/8/3P4/8/8/4K2k w Kq e6 3 10",
		},
	}

	for _, tc := range cases {
		b, err := FromFEN(tc.fen)
		if err != nil {
			t.Fatalf("%s: FromFEN failed: %v", tc.name, err)
		}

		white := orPieces(b.Pieces, WhitePawn, WhiteKing)
		black := orPieces(b.Pieces, BlackPawn, BlackKing)

		if b.OccupancyWhite != white {
			t.Fatalf("%s: white occupancy mismatch", tc.name)
		}
		if b.OccupancyBlack != black {
			t.Fatalf("%s: black occupancy mismatch", tc.name)
		}
		if b.OccupancyBoth != (white|black) {
			t.Fatalf("%s: both occupancy mismatch", tc.name)
		}
	}
}
