package search

import (
	"time"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/move"
	"EquinoxChess/internal/movegen"
	"EquinoxChess/internal/tt"
)

func (s *Search) Negamax(b *board.Board, depth int, alpha int, beta int) int {
	s.Nodes++
	if s.TimeLimit > 0 && (s.Nodes&255) == 0 {
		if time.Since(s.StartTime) >= s.TimeLimit {
			s.Stop.Store(true)
			return 0
		}
	}
	if s.Stop.Load() {
		return alpha
	}
	if s.Ply >= MaxDepth-1 {
		return Evaluate(b)
	}

	// Draw detection: must be done before expanding node
	if b.HalfmoveClock >= 100 {
		return DrawScore
	}
	if s.Ply > 0 && b.IsRepetition() {
		return DrawScore
	}
	if b.IsInsufficientMaterial() {
		return DrawScore
	}

	if depth == 0 {
		return s.Quiescence(b, alpha, beta)
	}
	originalAlpha := alpha
	originalBeta := beta
	var ttMove uint32
	if s.TT != nil {
		score, move, cutoff := s.TT.Probe(b.ZobristKey, b.HalfmoveClock, depth, alpha, beta, s.Ply, MateThreshold)
		if move != 0 {
			ttMove = move
		}
		if cutoff {
			if s.Ply == 0 && move != 0 {
				s.PVMove[0] = move
			}
			return score
		}
	}

	inCheck := InCheck(b)
	if depth >= 5 && !inCheck && hasNonPawnMaterial(b, b.SideToMove) && beta < MateScore-1000 {
		R := 2
		if depth > R+1 {
			b.MakeNullMove()
			s.Ply++
			score := -s.Negamax(b, depth-1-R, -beta, -beta+1)
			s.Ply--
			b.UnmakeNullMove()
			if s.Stop.Load() {
				return alpha
			}
			if score >= beta {
				verify := -s.Negamax(b, depth-1, -beta, -alpha)
				if verify >= beta {
					return beta
				}
			}
		}
	}

	var moves [movegen.MaxMoves]uint32
	count := movegen.GeneratePseudoLegalMoves(b, &moves)
	if count == 0 {
		if InCheck(b) {
			return -MateScore + s.Ply
		}
		return 0
	}

	if ttMove != 0 {
		for i := 0; i < count; i++ {
			if moves[i] == ttMove {
				moves[0], moves[i] = moves[i], moves[0]
				break
			}
		}
	}

	if s.Ply == 0 && s.PVMove[0] != 0 {
		for i := 0; i < count; i++ {
			if moves[i] == s.PVMove[0] {
				moves[0], moves[i] = moves[i], moves[0]
				break
			}
		}
	}

	staticEval := 0
	if depth == 1 && !inCheck && (beta-alpha) > 1 {
		staticEval = Evaluate(b)
	}

	foundLegal := false
	bestMove := uint32(0)
	for i := 0; i < count; i++ {
		bestIdx := selectBestMove(s, &moves, i, count)
		moves[i], moves[bestIdx] = moves[bestIdx], moves[i]

		m := moves[i]
		// Store current ply before incrementing for child search
		currentPly := s.Ply
		killer1 := uint32(0)
		killer2 := uint32(0)
		if currentPly < MaxDepth {
			killer1 = s.KillerMoves[currentPly][0]
			killer2 = s.KillerMoves[currentPly][1]
		}

		if depth == 1 && !inCheck && (beta-alpha) > 1 {
			if !move.IsCapture(m) && !move.IsPromotion(m) {
				isKiller := false
				isKiller = killer1 == m || killer2 == m
				if !isKiller && staticEval+150 <= alpha {
					continue
				}
			}
		}
		b.MakeMove(m)
		s.Ply++
		if IsMovedSideInCheck(b) {
			s.Ply--
			b.UnmakeMove()
			continue
		}
		foundLegal = true
		score := -Infinity
		if terminalScore, terminal := s.scoreTerminalChild(b); terminal {
			score = terminalScore
		} else if depth >= 3 && i >= 3 && !inCheck && !move.IsCapture(m) && !move.IsPromotion(m) && m != killer1 && m != killer2 && (beta-alpha) > 1 {
			reduction := 1
			if depth >= 6 {
				reduction = 2
			}
			if depth >= 10 {
				reduction = 3
			}
			reducedDepth := depth - 1 - reduction
			if reducedDepth < 1 {
				reducedDepth = 1
			}
			score = -s.Negamax(b, reducedDepth, -alpha-1, -alpha)
			if score > alpha {
				score = -s.Negamax(b, depth-1, -beta, -alpha)
			}
		} else {
			score = -s.Negamax(b, depth-1, -beta, -alpha)
		}
		s.Ply--
		b.UnmakeMove()

		if score >= beta {
			if s.Stop.Load() {
				return alpha
			}
			if s.Ply == 0 {
				s.PVMove[0] = m
			}
			if s.TT != nil {
				s.TT.Store(b.ZobristKey, b.HalfmoveClock, depth, beta, tt.TTBeta, m, currentPly, MateThreshold)
			}
			updateKillers(s, m)
			return beta
		}
		if score > alpha {
			alpha = score
			updateHistory(s, m, depth)
			bestMove = m
			if s.Ply == 0 {
				s.PVMove[0] = m
			}
		}
	}

	if !foundLegal {
		if InCheck(b) {
			return -MateScore + s.Ply
		}
		return 0
	}

	finalScore := alpha
	if s.Stop.Load() {
		return alpha
	}
	if s.TT != nil && !s.Stop.Load() {
		flag := tt.TTExact
		if finalScore <= originalAlpha {
			flag = tt.TTAlpha
		} else if finalScore >= originalBeta {
			flag = tt.TTBeta
		} else {
			flag = tt.TTExact
		}
		s.TT.Store(b.ZobristKey, b.HalfmoveClock, depth, finalScore, flag, bestMove, s.Ply, MateThreshold)
	}
	return finalScore
}

func (s *Search) SearchRoot(b *board.Board, maxDepth int) (uint32, int, int, uint64) {
	s.Nodes = 0
	s.Ply = 0
	s.PVMove[0] = 0
	if maxDepth > MaxDepth-1 {
		maxDepth = MaxDepth - 1
	}
	bestMove, bestScore := rootFallback(b)
	bestDepth := 0
	if maxDepth <= 0 {
		return bestMove, bestScore, bestDepth, s.Nodes
	}

	prevScore := -Infinity
	for depth := 1; depth <= maxDepth; depth++ {
		alpha := -Infinity
		beta := Infinity
		if depth > 1 {
			window := 50
			alpha = prevScore - window
			beta = prevScore + window
		}

		score := s.Negamax(b, depth, alpha, beta)
		if !s.Stop.Load() && (score <= alpha || score >= beta) {
			score = s.Negamax(b, depth, -Infinity, Infinity)
		}

		if s.Stop.Load() {
			s.PVMove[0] = bestMove
			break
		}

		prevScore = score
		bestScore = score
		bestDepth = depth
		if s.PVMove[0] != 0 {
			bestMove = s.PVMove[0]
		}
		s.DecayHistory()
	}

	return bestMove, bestScore, bestDepth, s.Nodes
}

func rootFallback(b *board.Board) (uint32, int) {
	if b.HalfmoveClock >= 100 || b.IsInsufficientMaterial() {
		return 0, DrawScore
	}

	var moves [movegen.MaxMoves]uint32
	count := movegen.GenerateLegalMoves(b, &moves)
	if count > 0 {
		return moves[0], Evaluate(b)
	}
	if InCheck(b) {
		return 0, -MateScore
	}
	return 0, DrawScore
}

func (s *Search) Quiescence(b *board.Board, alpha int, beta int) int {
	s.Nodes++
	if s.TimeLimit > 0 && (s.Nodes&255) == 0 {
		if time.Since(s.StartTime) >= s.TimeLimit {
			s.Stop.Store(true)
			return 0
		}
	}
	if s.Stop.Load() {
		return alpha
	}
	if s.Ply >= MaxDepth-1 {
		return Evaluate(b)
	}
	inCheck := InCheck(b)
	var moves [movegen.MaxMoves]uint32
	count := movegen.GeneratePseudoLegalMoves(b, &moves)

	if !inCheck && !hasAnyLegalMove(b, &moves, count) {
		return DrawScore
	}

	// When in check, we must search all evasions and cannot use stand-pat
	if !inCheck {
		standPat := Evaluate(b)
		if standPat >= beta {
			return beta
		}
		if standPat > alpha {
			alpha = standPat
		}
	}

	foundLegal := false
	for i := 0; i < count; i++ {
		m := moves[i]
		// When not in check, only consider captures
		// When in check, consider all evasions (king moves, captures, blocks)
		if !inCheck && !move.IsCapture(m) {
			continue
		}
		if !inCheck && !move.IsPromotion(m) {
			captured := move.CapturedPiece(m)
			standPat := Evaluate(b)
			optimisticScore := standPat + pieceValues[captured] + 200
			if optimisticScore <= alpha {
				continue
			}
		}
		b.MakeMove(m)
		s.Ply++
		if IsMovedSideInCheck(b) {
			s.Ply--
			b.UnmakeMove()
			continue
		}
		foundLegal = true
		score := -s.Quiescence(b, -beta, -alpha)
		s.Ply--
		b.UnmakeMove()
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	// If in check and no legal moves found, it's checkmate
	if inCheck && !foundLegal {
		return -MateScore + s.Ply
	}

	return alpha
}

func (s *Search) scoreTerminalChild(b *board.Board) (int, bool) {
	var childMoves [movegen.MaxMoves]uint32
	childCount := movegen.GeneratePseudoLegalMoves(b, &childMoves)
	if hasAnyLegalMove(b, &childMoves, childCount) {
		return 0, false
	}
	if InCheck(b) {
		return MateScore - s.Ply, true
	}
	return DrawScore, true
}

func hasAnyLegalMove(b *board.Board, moves *[movegen.MaxMoves]uint32, count int) bool {
	for i := 0; i < count; i++ {
		m := moves[i]
		b.MakeMove(m)
		legal := !IsMovedSideInCheck(b)
		b.UnmakeMove()
		if legal {
			return true
		}
	}
	return false
}
