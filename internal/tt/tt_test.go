package tt

import "testing"

func TestProbeSameHalfmovePreservesCutoffsAndMateNormalization(t *testing.T) {
	const (
		key           = uint64(0x123456789abcdef0)
		halfmoveClock = 47
		depth         = 7
		move          = uint32(0x10203040)
		mateThreshold = 31000
	)

	tests := []struct {
		name      string
		score     int
		flag      uint8
		alpha     int
		beta      int
		storePly  int
		probePly  int
		wantScore int
	}{
		{name: "exact", score: 42, flag: TTExact, alpha: -100, beta: 100, wantScore: 42},
		{name: "alpha", score: -25, flag: TTAlpha, alpha: -20, beta: 100, wantScore: -25},
		{name: "beta", score: 25, flag: TTBeta, alpha: -100, beta: 20, wantScore: 25},
		{
			name:      "positive mate",
			score:     31990,
			flag:      TTExact,
			alpha:     -50000,
			beta:      50000,
			storePly:  5,
			probePly:  2,
			wantScore: 31993,
		},
		{
			name:      "negative mate",
			score:     -31990,
			flag:      TTExact,
			alpha:     -50000,
			beta:      50000,
			storePly:  5,
			probePly:  2,
			wantScore: -31993,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			table := NewTT(16)
			table.Store(key, halfmoveClock, depth, test.score, test.flag, move, test.storePly, mateThreshold)

			gotScore, gotMove, cutoff := table.Probe(
				key,
				halfmoveClock,
				depth,
				test.alpha,
				test.beta,
				test.probePly,
				mateThreshold,
			)
			if !cutoff {
				t.Fatal("Probe did not return a cutoff")
			}
			if gotScore != test.wantScore {
				t.Fatalf("score = %d, want %d", gotScore, test.wantScore)
			}
			if gotMove != move {
				t.Fatalf("move = %#x, want %#x", gotMove, move)
			}
		})
	}
}

func TestProbeDifferentHalfmoveReturnsMoveWithoutCutoff(t *testing.T) {
	const (
		key             = uint64(0x23456789abcdef01)
		storedHalfmove  = 12
		currentHalfmove = 13
		depth           = 8
		move            = uint32(0x50607080)
		mateThreshold   = 31000
	)

	table := NewTT(16)
	table.Store(key, storedHalfmove, depth, 123, TTExact, move, 0, mateThreshold)

	score, gotMove, cutoff := table.Probe(
		key,
		currentHalfmove,
		depth,
		-500,
		500,
		0,
		mateThreshold,
	)
	if cutoff {
		t.Fatal("Probe returned a cutoff for a different halfmove clock")
	}
	if score != 0 {
		t.Fatalf("score = %d, want 0", score)
	}
	if gotMove != move {
		t.Fatalf("move = %#x, want %#x", gotMove, move)
	}
}

func TestStoreUpdatesHalfmoveClock(t *testing.T) {
	const (
		key           = uint64(0x3456789abcdef012)
		firstHalfmove = 21
		lastHalfmove  = 22
		depth         = 6
		firstMove     = uint32(0x11112222)
		lastMove      = uint32(0x33334444)
		mateThreshold = 31000
	)

	table := NewTT(16)
	table.Store(key, firstHalfmove, depth, 10, TTExact, firstMove, 0, mateThreshold)
	table.Store(key, lastHalfmove, depth, 20, TTExact, lastMove, 0, mateThreshold)

	entry := table.Entries[key&table.Mask]
	if entry.HalfmoveClock != lastHalfmove {
		t.Fatalf("stored halfmove clock = %d, want %d", entry.HalfmoveClock, lastHalfmove)
	}

	if score, move, cutoff := table.Probe(key, firstHalfmove, depth, -100, 100, 0, mateThreshold); cutoff || score != 0 || move != lastMove {
		t.Fatalf("old halfmove Probe = (%d, %#x, %t), want (0, %#x, false)", score, move, cutoff, lastMove)
	}
	if score, move, cutoff := table.Probe(key, lastHalfmove, depth, -100, 100, 0, mateThreshold); !cutoff || score != 20 || move != lastMove {
		t.Fatalf("new halfmove Probe = (%d, %#x, %t), want (20, %#x, true)", score, move, cutoff, lastMove)
	}
}
