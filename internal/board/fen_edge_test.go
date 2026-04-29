package board

import "testing"

func TestFromFENInvalidInputs(t *testing.T) {
	cases := []struct {
		name string
		fen  string
	}{
		{
			name: "too-few-fields",
			fen:  "8/8/8/8/8/8/8/8 w -",
		},
		{
			name: "piece-placement-not-8-ranks",
			fen:  "8/8/8/8/8/8/8 w - - 0 1",
		},
		{
			name: "row-count-not-8",
			fen:  "9/8/8/8/8/8/8/8 w - - 0 1",
		},
		{
			name: "row-count-not-8-with-piece",
			fen:  "7p1/8/8/8/8/8/8/8 w - - 0 1",
		},
		{
			name: "invalid-piece-char",
			fen:  "8/8/8/8/8/8/8/X7 w - - 0 1",
		},
		{
			name: "invalid-side",
			fen:  "8/8/8/8/8/8/8/8 x - - 0 1",
		},
		{
			name: "invalid-castling-char",
			fen:  "8/8/8/8/8/8/8/8 w KQx - 0 1",
		},
		{
			name: "ep-length-not-2",
			fen:  "8/8/8/8/8/8/8/8 w - e 0 1",
		},
		{
			name: "ep-file-out-of-range",
			fen:  "8/8/8/8/8/8/8/8 w - i3 0 1",
		},
		{
			name: "ep-rank-out-of-range",
			fen:  "8/8/8/8/8/8/8/8 w - e9 0 1",
		},
		{
			name: "halfmove-not-numeric",
			fen:  "8/8/8/8/8/8/8/8 w - - x 1",
		},
		{
			name: "fullmove-not-numeric",
			fen:  "8/8/8/8/8/8/8/8 w - - 0 y",
		},
	}

	for _, tc := range cases {
		if _, err := FromFEN(tc.fen); err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
	}
}

func TestFromFENValidInputs(t *testing.T) {
	cases := []struct {
		name string
		fen  string
	}{
		{
			name: "dash-castling-and-ep",
			fen:  "8/8/8/8/8/8/8/8 w - - 0 1",
		},
		{
			name: "all-castling",
			fen:  "8/8/8/8/8/8/8/8 w KQkq - 0 1",
		},
		{
			name: "subset-castling",
			fen:  "8/8/8/8/8/8/8/8 b Kq - 0 1",
		},
		{
			name: "ep-d3",
			fen:  "8/8/8/8/8/8/8/8 w - d3 0 1",
		},
		{
			name: "ep-e6",
			fen:  "8/8/8/8/8/8/8/8 w - e6 0 1",
		},
	}

	for _, tc := range cases {
		if _, err := FromFEN(tc.fen); err != nil {
			t.Fatalf("%s: expected success, got error: %v", tc.name, err)
		}
	}
}

func TestFromFENPermissiveEnPassantA1(t *testing.T) {
	if _, err := FromFEN("8/8/8/8/8/8/8/8 w - a1 0 1"); err != nil {
		t.Fatalf("expected permissive en passant a1, got error: %v", err)
	}
}

func BenchmarkFromFEN(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := FromFEN(StartFEN); err != nil {
			b.Fatalf("FromFEN failed: %v", err)
		}
	}
}
