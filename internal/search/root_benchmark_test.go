package search

import (
	"testing"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/tt"
)

// Reuse a UCI-sized table, clearing it between positions. Timings include
// that reset and all iterations up to the requested depth.
func BenchmarkColdRoot(b *testing.B) {
	for _, tc := range []struct{ name, fen string }{
		{"middlegame", "rn1q1rk1/pb2bppp/1p2pn2/2ppN3/3P4/2NBP3/PPPB1PPP/R2Q1RK1 w - - 2 9"},
		{"ending", "8/8/RP1B4/3b4/8/P3k2p/3r4/6K1 w - - 15 64"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			pos, err := board.FromFEN(tc.fen)
			if err != nil {
				b.Fatal(err)
			}
			table := tt.NewTT(1 << 20)
			b.ReportAllocs()
			var nodes uint64
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				clear(table.Entries)
				s := Search{TT: table}
				var depth int
				_, _, depth, nodes = s.SearchRoot(pos, 7)
				if depth != 7 {
					b.Fatalf("completed depth %d", depth)
				}
			}
			b.ReportMetric(float64(nodes), "nodes/op")
		})
	}
}
