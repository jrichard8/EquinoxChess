package search

import (
	"testing"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/tt"
	"EquinoxChess/internal/uci"
)

// Anti-blunder regression metrics (Phase 2 canary).
//
// Each FEN below is the position EquinoxChess actually faced just before
// playing a tactical blunder in matches/sf1600-run6-1+0.1-30.pgn (1+0.1
// against Stockfish limited to 1600 Elo). The historical move and its depth-4
// PGN score are retained so that two different questions stay independent:
//
//   - bad-move avoidance: does depth 6 choose something other than the
//     historical blunder?
//   - loss recognition: does depth 6 score the position at or below -300 cp
//     from the side-to-move POV (including a mate-against score)?
//
// The baseline protections are per case as well as aggregate. This prevents a
// newly fixed canary from hiding a regression on a canary that already passed.

type tacticalBlunderCase struct {
	name                 string
	fen                  string
	historicalBadMoveUCI string
	pgnDepth4Score       string
}

var tacticalBlunderCases = []tacticalBlunderCase{
	{"Game02_B_ply40", "r4r1k/1pp3pp/p1b1p3/3pP1Nn/P2P3P/qPP1p3/2Q2P2/1K2RBR1 b - - 3 20", "h7h6", "+1.48"},
	{"Game03_W_ply29", "r1b2rk1/4bppp/p1q2n2/1p2p3/3QP3/P1N5/1PP1BPPP/R3R1K1 w - - 0 15", "d4e5", "-1.54"},
	{"Game06_B_ply46", "2rr4/1ppN1pkp/8/p7/8/2Q1PN2/pB3P1K/3q4 b - - 1 23", "g7g6", "+3.04"},
	{"Game09_W_ply127", "8/8/RP1B4/3b4/8/P3k2p/3r4/6K1 w - - 15 64", "a6a7", "+0.82"},
	{"Game16_B_ply62", "4b2k/1Bp2ppp/P3p3/r7/2p3P1/8/5K1P/3R4 b - - 2 31", "c4c3", "+3.03"},
	{"Game17_W_ply17", "rn1q1rk1/pb2bppp/1p2pn2/2ppN3/3P4/2NBP3/PPPB1PPP/R2Q1RK1 w - - 2 9", "d1f3", "+1.69"},
	{"Game21_W_ply87", "8/8/8/1PN5/5kpp/8/5PP1/3r3K w - - 2 44", "h1h2", "-0.36"},
	{"Game22_B_ply36", "r2q1rk1/p1pb1ppp/1p2p3/2bpP1N1/4nBPP/2P5/PP1N3Q/R3KB1R b KQ - 5 18", "c5f2", "-0.03"},
	{"Game25_W_ply49", "r5k1/p1b3pp/7q/8/1B1Pn1P1/4P3/PPP1K3/R2Q2R1 w - - 5 25", "d4d5", "+4.86"},
	{"Game26_B_ply30", "r1b1kb1r/pp4pp/2p5/1N2p3/6p1/P2BP3/1PQ2PqP/2KR2R1 b kq - 1 15", "g2h2", "+2.48"},
	{"Game27_W_ply29", "r1b2rk1/2p1qpp1/2pp1b1p/p7/2B1PBn1/2N2Q2/PPP2PPP/R3R1K1 w - - 0 15", "a2a3", "+1.63"},
	{"Game28_B_ply48", "r3r1k1/1pp4p/3p1Q2/1qnP1R2/p3p3/PP4PP/2P3P1/4R1K1 b - - 0 24", "h7h6", "+1.64"},
	{"Game30_B_ply26", "r4rk1/pppbbppp/2n1p3/3pP3/3Pq2P/P1P5/2P1BPP1/R1BQ1KNR b - - 0 13", "e7h4", "+1.51"},
}

var baselineBadMoveAvoided = map[string]bool{
	"Game02_B_ply40":  true,
	"Game06_B_ply46":  true,
	"Game09_W_ply127": true,
	"Game16_B_ply62":  true,
	"Game22_B_ply36":  true,
	"Game25_W_ply49":  true,
	"Game26_B_ply30":  true,
	"Game27_W_ply29":  true,
	"Game28_B_ply48":  true,
	"Game30_B_ply26":  true,
}

var baselineLossRecognized = map[string]bool{
	"Game06_B_ply46": true,
	"Game21_W_ply87": true,
}

const (
	tacticalCanaryDepth             = 6
	tacticalLossThreshold           = -300
	minimumBadMoveAvoidanceCanaries = 10
	minimumLossRecognitionCanaries  = 2
)

type tacticalBlunderResult struct {
	testCase       tacticalBlunderCase
	bestMoveUCI    string
	score          int
	avoidedBadMove bool
	recognizedLoss bool
}

func TestSearchTacticalCanaryMetrics(t *testing.T) {
	results := searchTacticalBlunderCases(t)

	t.Run("BadMoveAvoidance", func(t *testing.T) {
		passed := 0
		for _, result := range results {
			if result.avoidedBadMove {
				passed++
			}
			t.Logf("%s: best=%s historical_bad=%s avoided=%t",
				result.testCase.name,
				result.bestMoveUCI,
				result.testCase.historicalBadMoveUCI,
				result.avoidedBadMove,
			)
			if baselineBadMoveAvoided[result.testCase.name] && !result.avoidedBadMove {
				t.Errorf("%s newly selects historical bad move %s",
					result.testCase.name, result.testCase.historicalBadMoveUCI)
			}
		}
		t.Logf("bad-move avoidance: %d/%d (minimum %d)",
			passed, len(results), minimumBadMoveAvoidanceCanaries)
		if passed < minimumBadMoveAvoidanceCanaries {
			t.Fatalf("bad-move avoidance regressed: got %d/%d, want at least %d/%d",
				passed, len(results), minimumBadMoveAvoidanceCanaries, len(results))
		}
	})

	t.Run("LossRecognition", func(t *testing.T) {
		passed := 0
		for _, result := range results {
			if result.recognizedLoss {
				passed++
			}
			t.Logf("%s: score=%d recognized=%t (threshold <= %d; PGN depth-4 score %s)",
				result.testCase.name,
				result.score,
				result.recognizedLoss,
				tacticalLossThreshold,
				result.testCase.pgnDepth4Score,
			)
			if baselineLossRecognized[result.testCase.name] && !result.recognizedLoss {
				t.Errorf("%s no longer recognizes the loss: score=%d, want <= %d",
					result.testCase.name, result.score, tacticalLossThreshold)
			}
		}
		t.Logf("loss recognition: %d/%d (minimum %d)",
			passed, len(results), minimumLossRecognitionCanaries)
		if passed < minimumLossRecognitionCanaries {
			t.Fatalf("loss recognition regressed: got %d/%d, want at least %d/%d",
				passed, len(results), minimumLossRecognitionCanaries, len(results))
		}
	})
}

func searchTacticalBlunderCases(t *testing.T) []tacticalBlunderResult {
	t.Helper()
	results := make([]tacticalBlunderResult, 0, len(tacticalBlunderCases))
	caseNames := make(map[string]bool, len(tacticalBlunderCases))
	for _, tc := range tacticalBlunderCases {
		if caseNames[tc.name] {
			t.Fatalf("duplicate tactical canary %s", tc.name)
		}
		caseNames[tc.name] = true
	}
	if len(baselineBadMoveAvoided) != minimumBadMoveAvoidanceCanaries {
		t.Fatalf("bad-move baseline fixture has %d cases, want %d",
			len(baselineBadMoveAvoided), minimumBadMoveAvoidanceCanaries)
	}
	if len(baselineLossRecognized) != minimumLossRecognitionCanaries {
		t.Fatalf("loss-recognition baseline fixture has %d cases, want %d",
			len(baselineLossRecognized), minimumLossRecognitionCanaries)
	}
	for name := range baselineBadMoveAvoided {
		if !caseNames[name] {
			t.Fatalf("bad-move baseline references unknown canary %s", name)
		}
	}
	for name := range baselineLossRecognized {
		if !caseNames[name] {
			t.Fatalf("loss-recognition baseline references unknown canary %s", name)
		}
	}
	for _, tc := range tacticalBlunderCases {
		pos, err := board.FromFEN(tc.fen)
		if err != nil {
			t.Fatalf("%s: FromFEN failed: %v", tc.name, err)
		}
		historicalBadMove, err := uci.ParseUCIMove(pos, tc.historicalBadMoveUCI)
		if err != nil {
			t.Fatalf("%s: historical bad move %s is invalid: %v",
				tc.name, tc.historicalBadMoveUCI, err)
		}

		s := Search{TT: tt.NewTT(1 << 16)}
		bestMove, score, depth, _ := s.SearchRoot(pos, tacticalCanaryDepth)
		if depth != tacticalCanaryDepth {
			t.Fatalf("%s: completed depth=%d, want %d", tc.name, depth, tacticalCanaryDepth)
		}
		results = append(results, tacticalBlunderResult{
			testCase:       tc,
			bestMoveUCI:    uci.MoveToUCI(bestMove),
			score:          score,
			avoidedBadMove: bestMove != uint32(historicalBadMove),
			recognizedLoss: score <= tacticalLossThreshold,
		})
	}
	return results
}
