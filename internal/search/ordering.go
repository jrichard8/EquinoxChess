package search

import (
	"EquinoxChess/internal/move"
	"EquinoxChess/internal/movegen"
)

const (
	captureScoreBase = 10000
	killerScoreBase  = 8000
	historyScoreBase = 1000
)

// MVV-LVA values used for move ordering only. Distinct from pieceValues used
// by Evaluate, where the king is 0 because its material value is irrelevant
// (it is never captured). For ordering, the king must be the *most expensive*
// attacker so that king-captures are tried last among attackers (we prefer
// pawn-captures, then minor pieces, etc.).
var mvvLvaValues = [12]int{
	PawnValue, KnightValue, BishopValue, RookValue, QueenValue, 10000,
	PawnValue, KnightValue, BishopValue, RookValue, QueenValue, 10000,
}

func selectBestMove(s *Search, moves *[movegen.MaxMoves]uint32, start int, count int) int {
	bestIdx := start
	bestScore := scoreMove(s, moves[start])
	for i := start + 1; i < count; i++ {
		score := scoreMove(s, moves[i])
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	return bestIdx
}

func scoreMove(s *Search, m uint32) int {
	captured := move.CapturedPiece(m)
	moving := move.MovingPiece(m)
	promo := move.PromotionPiece(m)

	if captured != move.NoPiece {
		score := captureScoreBase + mvvLvaValues[captured]*10 - mvvLvaValues[moving]
		if promo != move.NoPiece {
			score += pieceValues[promo] - PawnValue
		}
		return score
	}

	if promo != move.NoPiece {
		// Non-capture promotion: rank near top captures so a queen promotion
		// is tried before quiet history moves and killers.
		return captureScoreBase + mvvLvaValues[promo]*10 - mvvLvaValues[moving]
	}

	ply := s.Ply
	if ply < MaxDepth {
		if s.KillerMoves[ply][0] == m {
			return killerScoreBase + 2
		}
		if s.KillerMoves[ply][1] == m {
			return killerScoreBase + 1
		}
	}

	to := move.To(m)
	return historyScoreBase + s.History[moving][to]
}

func updateKillers(s *Search, m uint32) {
	if move.IsCapture(m) {
		return
	}
	ply := s.Ply
	if ply >= MaxDepth {
		return
	}
	if s.KillerMoves[ply][0] != m {
		s.KillerMoves[ply][1] = s.KillerMoves[ply][0]
		s.KillerMoves[ply][0] = m
	}
}

func updateHistory(s *Search, m uint32, depth int) {
	if move.IsCapture(m) {
		return
	}
	moving := move.MovingPiece(m)
	to := move.To(m)
	if moving < 0 || moving >= len(s.History) {
		return
	}
	if to < 0 || to >= 64 {
		return
	}
	inc := depth * depth
	s.History[moving][to] += inc
}

func (s *Search) DecayHistory() {
	for piece := 0; piece < len(s.History); piece++ {
		for sq := 0; sq < len(s.History[piece]); sq++ {
			v := s.History[piece][sq] / 2
			if v < 0 {
				v = 0
			}
			s.History[piece][sq] = v
		}
	}
}

// Order captures by victim and attacker value. Equal scores retain generator order.
// Quiet moves stay in the tail for the diagnostic trace.
func orderQuiescenceCaptures(s *Search, moves []uint32) int {
	count := 0
	var scores [movegen.MaxMoves]int
	for i, m := range moves {
		if !move.IsCapture(m) {
			continue
		}
		score := scoreMove(s, m)
		moves[i] = moves[count]
		j := count
		for j > 0 && scores[j-1] < score {
			moves[j], scores[j] = moves[j-1], scores[j-1]
			j--
		}
		moves[j], scores[j] = m, score
		count++
	}
	return count
}
