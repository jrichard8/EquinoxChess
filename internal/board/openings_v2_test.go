package board_test

import (
	"bufio"
	"math/bits"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"EquinoxChess/internal/api"
	"EquinoxChess/internal/board"
	"EquinoxChess/internal/movegen"
)

func TestOpeningsV2(t *testing.T) {
	v1 := readEPDLines(t, filepath.Join("..", "..", "tools", "elo", "openings-v1.epd"))
	v2 := readEPDLines(t, filepath.Join("..", "..", "tools", "elo", "openings-v2.epd"))
	if len(v1) != 32 {
		t.Fatalf("v1 has %d positions, want 32", len(v1))
	}
	if len(v2) != 64 {
		t.Fatalf("v2 has %d positions, want 64", len(v2))
	}
	for i := range v1 {
		if v2[i] != v1[i] {
			t.Fatalf("v2 line %d differs from the required v1 prefix", i+1)
		}
	}

	seen := make(map[uint64]int, len(v2))
	for i, line := range v2 {
		fen := epdFEN(t, line)
		position, err := board.FromFEN(fen)
		if err != nil {
			t.Fatalf("v2 line %d: FromFEN(%q): %v", i+1, fen, err)
		}
		if position.SideToMove != board.White || position.FullmoveNumber != 5 {
			t.Fatalf("v2 line %d: expected White to move after 8 plies, got %q", i+1, fen)
		}
		var legal [movegen.MaxMoves]uint32
		if movegen.GenerateLegalMoves(position, &legal) == 0 {
			t.Fatalf("v2 line %d: no legal move in %q", i+1, fen)
		}
		if previous, exists := seen[position.ZobristKey]; exists {
			t.Fatalf("v2 lines %d and %d are duplicate positions", previous+1, i+1)
		}
		seen[position.ZobristKey] = i
	}

	cases := []struct {
		name  string
		moves []string
	}{
		{"33 Queen's Gambit Accepted", []string{"d2d4", "d7d5", "c2c4", "d5c4", "g1f3", "g8f6", "e2e3", "e7e6"}},
		{"34 QGD Exchange", []string{"d2d4", "d7d5", "c2c4", "e7e6", "b1c3", "g8f6", "c4d5", "e6d5"}},
		{"35 QGD Tarrasch", []string{"d2d4", "d7d5", "c2c4", "e7e6", "b1c3", "c7c5", "c4d5", "e6d5"}},
		{"36 Slav Main Line", []string{"d2d4", "d7d5", "c2c4", "c7c6", "g1f3", "g8f6", "b1c3", "c8f5"}},
		{"37 Semi-Slav", []string{"d2d4", "d7d5", "c2c4", "e7e6", "b1c3", "c7c6", "g1f3", "d5c4"}},
		{"38 Catalan", []string{"d2d4", "g8f6", "c2c4", "e7e6", "g1f3", "d7d5", "g2g3", "c7c6"}},
		{"39 Nimzo-Indian", []string{"d2d4", "g8f6", "c2c4", "e7e6", "b1c3", "f8b4", "g1f3", "d7d5"}},
		{"40 Queen's Indian", []string{"d2d4", "g8f6", "c2c4", "e7e6", "g1f3", "b7b6", "g2g3", "c8a6"}},
		{"41 King's Indian", []string{"d2d4", "g8f6", "c2c4", "g7g6", "b1c3", "f8g7", "e2e4", "e8g8"}},
		{"42 Grunfeld Fianchetto", []string{"d2d4", "g8f6", "c2c4", "g7g6", "b1c3", "d7d5", "g2g3", "f8g7"}},
		{"43 Modern Benoni", []string{"d2d4", "g8f6", "c2c4", "c7c5", "d4d5", "e7e6", "g1f3", "e6d5"}},
		{"44 Dutch Classical", []string{"d2d4", "f7f5", "c2c4", "g8f6", "b1c3", "e7e6", "g1f3", "f8b4"}},
		{"45 English Four Knights", []string{"c2c4", "e7e5", "b1c3", "g8f6", "g1f3", "b8c6", "g2g3", "d7d5"}},
		{"46 Reti", []string{"g1f3", "d7d5", "c2c4", "e7e6", "g2g3", "g8f6", "f1g2", "f8e7"}},
		{"47 London System", []string{"d2d4", "d7d5", "g1f3", "g8f6", "c1f4", "e7e6", "e2e3", "f8d6"}},
		{"48 Trompowsky", []string{"d2d4", "g8f6", "c1g5", "e7e6", "b1d2", "d7d5", "e2e3", "f8e7"}},
		{"49 Ruy Lopez Berlin", []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1b5", "g8f6", "d2d3", "f8c5"}},
		{"50 Ruy Lopez Schliemann", []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1b5", "f7f5", "b1c3", "f5e4"}},
		{"51 Italian Game", []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "f8c5", "d2d3", "d7d6"}},
		{"52 Caro-Kann Panov", []string{"e2e4", "c7c6", "d2d4", "d7d5", "e4d5", "c6d5", "c2c4", "g8f6"}},
		{"53 Scotch Game", []string{"e2e4", "e7e5", "g1f3", "b8c6", "d2d4", "e5d4", "f3d4", "f8c5"}},
		{"54 Scotch Four Knights", []string{"e2e4", "e7e5", "g1f3", "b8c6", "b1c3", "g8f6", "d2d4", "e5d4"}},
		{"55 Petrov Three Knights", []string{"e2e4", "e7e5", "g1f3", "g8f6", "d2d3", "b8c6", "c2c3", "d7d5"}},
		{"56 Philidor Defense", []string{"e2e4", "e7e5", "g1f3", "d7d6", "d2d4", "b8d7", "f1c4", "c7c6"}},
		{"57 Center Game", []string{"e2e4", "e7e5", "d2d4", "e5d4", "d1d4", "b8c6", "d4e3", "g8f6"}},
		{"58 Bishop's Opening", []string{"e2e4", "e7e5", "f1c4", "g8f6", "d2d3", "f8c5", "g1f3", "d7d6"}},
		{"59 Sicilian Taimanov", []string{"e2e4", "c7c5", "g1f3", "e7e6", "d2d4", "c5d4", "f3d4", "b8c6"}},
		{"60 Sicilian Moscow", []string{"e2e4", "c7c5", "g1f3", "d7d6", "f1b5", "c8d7", "b5d7", "d8d7"}},
		{"61 Sicilian Alapin", []string{"e2e4", "c7c5", "c2c3", "g8f6", "e4e5", "f6d5", "d2d4", "c5d4"}},
		{"62 Sicilian Rossolimo", []string{"e2e4", "c7c5", "g1f3", "b8c6", "f1b5", "g7g6", "b5c6", "d7c6"}},
		{"63 Closed Sicilian", []string{"e2e4", "c7c5", "b1c3", "b8c6", "g2g3", "g7g6", "f1g2", "f8g7"}},
		{"64 French Winawer", []string{"e2e4", "e7e6", "d2d4", "d7d5", "b1c3", "f8b4", "e4e5", "c7c5"}},
	}

	if len(cases) != len(v2)-len(v1) {
		t.Fatalf("generated opening cases = %d, want %d", len(cases), len(v2)-len(v1))
	}
	for caseIndex, tc := range cases {
		caseIndex, tc := caseIndex, tc
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.moves) != 8 {
				t.Fatalf("moves = %d, want 8", len(tc.moves))
			}
			g := &api.Game{Board: board.NewBoard()}
			for _, uci := range tc.moves {
				if err := g.PlayMove(uci); err != nil {
					t.Fatalf("PlayMove(%q): %v", uci, err)
				}
			}
			if g.Board.SideToMove != board.White {
				t.Fatalf("side to move = %d, want White", g.Board.SideToMove)
			}
			if remaining := bits.OnesCount64(uint64(g.Board.OccupancyBoth)); remaining < 30 {
				t.Fatalf("position has %d pieces, want at least 30", remaining)
			}

			// MakeMove does not currently maintain FullmoveNumber; the EPD
			// contract records the standard FEN value after eight half-moves.
			g.Board.FullmoveNumber = 5
			got := g.FEN()
			want := epdFEN(t, v2[len(v1)+caseIndex])
			if got != want {
				t.Fatalf("generated FEN = %q, want %q", got, want)
			}
		})
	}
}

func readEPDLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return lines
}

func epdFEN(t *testing.T, line string) string {
	t.Helper()
	fields := strings.Fields(line)
	if len(fields) < 6 {
		t.Fatalf("EPD line has fewer than six FEN fields: %q", line)
	}
	return strings.Join(fields[:6], " ")
}
