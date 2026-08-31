package search

import "EquinoxChess/internal/tt"

// SearchTrace is an opt-in diagnostic recorder for one root move and its
// immediate reply. A nil Search.Trace keeps normal engine runs on the existing
// path. Tracing is observational unless ForceTargetReplyFullDepth explicitly
// enables the searchtrace-only counterfactual for the exact reply.
type SearchTrace struct {
	RootMove                  uint32
	ReplyMove                 uint32
	ForceTargetReplyFullDepth bool
	TargetReplyExtraPlies     int
	ForceTargetReplyIteration int
	Events                    []TraceEvent
	Continuations             []TraceContinuation
	RootMoveResults           []TraceRootMoveResult

	iteration       int
	pass            string
	path            [MaxDepth]uint32
	targetLineQSeen bool
	pvMoves         [MaxDepth][MaxDepth]uint32
	pvLength        [MaxDepth]int
	pvBestScore     [MaxDepth]int
	pvLeafPly       [MaxDepth]int
	pvLeafScore     [MaxDepth]int
	pvLeafReason    [MaxDepth]string
}

// TraceContinuation captures the best explored line after the exact target
// reply in one iterative-deepening pass. Scores use the POV of the side that
// plays ReplyMove; leaf scores use the POV at the leaf itself.
type TraceContinuation struct {
	Iteration   int      `json:"iteration"`
	Pass        string   `json:"pass"`
	SearchMode  string   `json:"search_mode"`
	ParentDepth int      `json:"parent_depth"`
	ChildDepth  int      `json:"child_depth"`
	Reduction   int      `json:"reduction"`
	Extension   int      `json:"extension"`
	Alpha       int      `json:"alpha"`
	Beta        int      `json:"beta"`
	ChildAlpha  int      `json:"child_alpha"`
	ChildBeta   int      `json:"child_beta"`
	Score       int      `json:"score"`
	Line        []uint32 `json:"line"`
	LeafPly     int      `json:"leaf_ply"`
	LeafScore   int      `json:"leaf_score"`
	LeafReason  string   `json:"leaf_reason"`
}

// TraceRootMoveResult records the alpha-beta result returned for every legal
// root move. A fail-low is an upper bound and a fail-high is a lower bound;
// callers must retain that qualification when comparing root moves.
type TraceRootMoveResult struct {
	Iteration   int    `json:"iteration"`
	Pass        string `json:"pass"`
	Move        uint32 `json:"move"`
	OrderIndex  int    `json:"order_index"`
	LegalIndex  int    `json:"legal_index"`
	SearchMode  string `json:"search_mode"`
	ParentDepth int    `json:"parent_depth"`
	ChildDepth  int    `json:"child_depth"`
	Reduction   int    `json:"reduction"`
	Extension   int    `json:"extension"`
	Alpha       int    `json:"alpha"`
	Beta        int    `json:"beta"`
	ChildAlpha  int    `json:"child_alpha"`
	ChildBeta   int    `json:"child_beta"`
	Score       int    `json:"score"`
	ScoreBound  string `json:"score_bound"`
}

// TraceEvent records how the target move progressed through one search node.
// Indices are -1 when the corresponding stage was not reached.
type TraceEvent struct {
	Iteration       int    `json:"iteration"`
	Pass            string `json:"pass"`
	Role            string `json:"role"`
	Node            string `json:"node"`
	Stage           string `json:"stage"`
	Ply             int    `json:"ply"`
	Depth           int    `json:"depth"`
	Alpha           int    `json:"alpha"`
	Beta            int    `json:"beta"`
	InCheck         bool   `json:"in_check"`
	TargetMove      uint32 `json:"target_move"`
	CurrentMove     uint32 `json:"current_move"`
	PseudoIndex     int    `json:"pseudo_index"`
	OrderIndex      int    `json:"order_index"`
	LegalIndex      int    `json:"legal_index"`
	GeneratedMoves  int    `json:"generated_moves"`
	MoveScore       int    `json:"move_score"`
	Score           int    `json:"score"`
	Reduction       int    `json:"reduction"`
	Extension       int    `json:"extension"`
	ChildDepth      int    `json:"child_depth"`
	TargetIsTT      bool   `json:"target_is_tt"`
	TargetIsPV      bool   `json:"target_is_pv"`
	TargetIsKiller  bool   `json:"target_is_killer"`
	TargetProcessed bool   `json:"target_processed"`
	TTKeyMatch      bool   `json:"tt_key_match"`
	TTHalfmoveMatch bool   `json:"tt_halfmove_match"`
	TTDepthEnough   bool   `json:"tt_depth_enough"`
	TTEntryDepth    int    `json:"tt_entry_depth"`
	TTStoredScore   int    `json:"tt_stored_score"`
	TTFlag          string `json:"tt_flag"`
}

// SearchTraceCompiled reports whether this binary was built with
// `-tags searchtrace`. Normal engine builds compile the hot-path hooks away.
func SearchTraceCompiled() bool {
	return searchTraceCompiled
}

func addTTEntryToTrace(event *TraceEvent, entry tt.TTEntry, key uint64, halfmoveClock, depth int) {
	event.TTKeyMatch = entry.Key == key
	if !event.TTKeyMatch {
		return
	}
	event.TTHalfmoveMatch = entry.HalfmoveClock == halfmoveClock
	event.TTDepthEnough = entry.Depth >= depth
	event.TTEntryDepth = entry.Depth
	event.TTStoredScore = entry.Score
	switch entry.Flag {
	case tt.TTExact:
		event.TTFlag = "exact"
	case tt.TTAlpha:
		event.TTFlag = "alpha"
	case tt.TTBeta:
		event.TTFlag = "beta"
	default:
		event.TTFlag = "unknown"
	}
}

func (s *Search) resetTrace() {
	if !searchTraceCompiled || s.Trace == nil {
		return
	}
	s.Trace.Events = nil
	s.Trace.Continuations = nil
	s.Trace.RootMoveResults = nil
	s.Trace.iteration = 0
	s.Trace.pass = ""
	s.Trace.path = [MaxDepth]uint32{}
	s.Trace.targetLineQSeen = false
	s.Trace.pvMoves = [MaxDepth][MaxDepth]uint32{}
	for ply := 0; ply < MaxDepth; ply++ {
		s.Trace.pvLength[ply] = ply
		s.Trace.pvBestScore[ply] = -Infinity
		s.Trace.pvLeafPly[ply] = -1
		s.Trace.pvLeafScore[ply] = 0
		s.Trace.pvLeafReason[ply] = ""
	}
}

func (s *Search) startTracePass(iteration int, pass string) {
	if !searchTraceCompiled || s.Trace == nil {
		return
	}
	s.Trace.iteration = iteration
	s.Trace.pass = pass
	s.Trace.targetLineQSeen = false
}

func (s *Search) traceTarget() (uint32, string, bool) {
	return s.traceTargetAt(s.Ply)
}

func (s *Search) traceTargetAt(ply int) (uint32, string, bool) {
	if !searchTraceCompiled || s.Trace == nil {
		return 0, "", false
	}
	if ply == 0 {
		return s.Trace.RootMove, "root_move", true
	}
	if ply == 1 && s.Trace.path[0] == s.Trace.RootMove {
		return s.Trace.ReplyMove, "reply_move", true
	}
	return 0, "", false
}

func (s *Search) newTraceEvent(node, stage string, depth, alpha, beta int) (TraceEvent, bool) {
	return s.newTraceEventAt(s.Ply, node, stage, depth, alpha, beta)
}

func (s *Search) newTraceEventAt(ply int, node, stage string, depth, alpha, beta int) (TraceEvent, bool) {
	target, role, ok := s.traceTargetAt(ply)
	if !ok {
		return TraceEvent{}, false
	}
	return TraceEvent{
		Iteration:   s.Trace.iteration,
		Pass:        s.Trace.pass,
		Role:        role,
		Node:        node,
		Stage:       stage,
		Ply:         ply,
		Depth:       depth,
		Alpha:       alpha,
		Beta:        beta,
		TargetMove:  target,
		PseudoIndex: -1,
		OrderIndex:  -1,
		LegalIndex:  -1,
		ChildDepth:  -1,
	}, true
}

func (s *Search) appendTrace(event TraceEvent) {
	if searchTraceCompiled && s.Trace != nil {
		s.Trace.Events = append(s.Trace.Events, event)
	}
}

func (s *Search) pushTraceMove(ply int, m uint32) {
	if searchTraceCompiled && s.Trace != nil && ply >= 0 && ply < MaxDepth {
		s.Trace.path[ply] = m
	}
}

func (s *Search) popTraceMove(ply int) {
	if searchTraceCompiled && s.Trace != nil && ply >= 0 && ply < MaxDepth {
		s.Trace.path[ply] = 0
	}
}

func (s *Search) traceTargetReplyDepthAdjustment(ply int, m uint32) (bool, int) {
	if !searchTraceCompiled || s.Trace == nil || !s.Trace.ForceTargetReplyFullDepth ||
		(s.Trace.ForceTargetReplyIteration != 0 && s.Trace.iteration != s.Trace.ForceTargetReplyIteration) ||
		ply != 1 || s.Trace.path[0] != s.Trace.RootMove || m != s.Trace.ReplyMove {
		return false, 0
	}
	extraPlies := s.Trace.TargetReplyExtraPlies
	if extraPlies < 0 {
		extraPlies = 0
	}
	if extraPlies > 1 {
		extraPlies = 1
	}
	return true, extraPlies
}

func (s *Search) resetTracePVAt(ply int, qsearch bool) {
	if !searchTraceCompiled || s.Trace == nil || ply < 0 || ply >= MaxDepth {
		return
	}
	s.Trace.pvMoves[ply] = [MaxDepth]uint32{}
	s.Trace.pvLength[ply] = ply
	s.Trace.pvBestScore[ply] = -Infinity
	s.Trace.pvLeafPly[ply] = -1
	s.Trace.pvLeafScore[ply] = 0
	s.Trace.pvLeafReason[ply] = ""
	if qsearch {
		s.Trace.pvLeafPly[ply] = ply
		s.Trace.pvLeafReason[ply] = "qsearch"
	}
}

func (s *Search) setTraceLeafAt(ply, score int, reason string) {
	if !searchTraceCompiled || s.Trace == nil || ply < 0 || ply >= MaxDepth {
		return
	}
	s.Trace.pvBestScore[ply] = score
	s.Trace.pvLength[ply] = ply
	s.Trace.pvLeafPly[ply] = ply
	s.Trace.pvLeafScore[ply] = score
	s.Trace.pvLeafReason[ply] = reason
}

func (s *Search) considerTraceStandPat(ply, score int) {
	if !searchTraceCompiled || s.Trace == nil || ply < 0 || ply >= MaxDepth ||
		score <= s.Trace.pvBestScore[ply] {
		return
	}
	s.Trace.pvBestScore[ply] = score
	s.Trace.pvLength[ply] = ply
	s.Trace.pvLeafPly[ply] = ply
	s.Trace.pvLeafScore[ply] = score
	s.Trace.pvLeafReason[ply] = "qsearch_stand_pat"
}

func (s *Search) considerTracePV(ply int, m uint32, score int) {
	if !searchTraceCompiled || s.Trace == nil || ply < 0 || ply+1 >= MaxDepth ||
		score <= s.Trace.pvBestScore[ply] {
		return
	}
	s.Trace.pvBestScore[ply] = score
	s.Trace.pvMoves[ply] = [MaxDepth]uint32{}
	s.Trace.pvMoves[ply][ply] = m
	childLength := s.Trace.pvLength[ply+1]
	if childLength < ply+1 {
		childLength = ply + 1
	}
	if childLength > MaxDepth {
		childLength = MaxDepth
	}
	for index := ply + 1; index < childLength; index++ {
		s.Trace.pvMoves[ply][index] = s.Trace.pvMoves[ply+1][index]
	}
	s.Trace.pvLength[ply] = childLength
	s.Trace.pvLeafPly[ply] = s.Trace.pvLeafPly[ply+1]
	s.Trace.pvLeafScore[ply] = s.Trace.pvLeafScore[ply+1]
	s.Trace.pvLeafReason[ply] = s.Trace.pvLeafReason[ply+1]
}

func (s *Search) recordTraceContinuation(
	ply int,
	m uint32,
	mode string,
	parentDepth int,
	childDepth int,
	reduction int,
	extension int,
	alpha int,
	beta int,
	childAlpha int,
	childBeta int,
	score int,
) {
	if !searchTraceCompiled || s.Trace == nil || ply != 1 ||
		s.Trace.path[0] != s.Trace.RootMove || m != s.Trace.ReplyMove {
		return
	}
	line := []uint32{s.Trace.RootMove, s.Trace.ReplyMove}
	childPly := ply + 1
	childLength := s.Trace.pvLength[childPly]
	if childLength > MaxDepth {
		childLength = MaxDepth
	}
	for index := childPly; index < childLength; index++ {
		line = append(line, s.Trace.pvMoves[childPly][index])
	}
	s.Trace.Continuations = append(s.Trace.Continuations, TraceContinuation{
		Iteration:   s.Trace.iteration,
		Pass:        s.Trace.pass,
		SearchMode:  mode,
		ParentDepth: parentDepth,
		ChildDepth:  childDepth,
		Reduction:   reduction,
		Extension:   extension,
		Alpha:       alpha,
		Beta:        beta,
		ChildAlpha:  childAlpha,
		ChildBeta:   childBeta,
		Score:       score,
		Line:        line,
		LeafPly:     s.Trace.pvLeafPly[childPly],
		LeafScore:   s.Trace.pvLeafScore[childPly],
		LeafReason:  s.Trace.pvLeafReason[childPly],
	})
}

func (s *Search) recordTraceRootMoveResult(
	ply int,
	m uint32,
	orderIndex int,
	legalIndex int,
	mode string,
	parentDepth int,
	childDepth int,
	reduction int,
	extension int,
	alpha int,
	beta int,
	childAlpha int,
	childBeta int,
	score int,
) {
	if !searchTraceCompiled || s.Trace == nil || ply != 0 {
		return
	}
	bound := "exact"
	if score <= alpha {
		bound = "upper"
	} else if score >= beta {
		bound = "lower"
	}
	s.Trace.RootMoveResults = append(s.Trace.RootMoveResults, TraceRootMoveResult{
		Iteration:   s.Trace.iteration,
		Pass:        s.Trace.pass,
		Move:        m,
		OrderIndex:  orderIndex,
		LegalIndex:  legalIndex,
		SearchMode:  mode,
		ParentDepth: parentDepth,
		ChildDepth:  childDepth,
		Reduction:   reduction,
		Extension:   extension,
		Alpha:       alpha,
		Beta:        beta,
		ChildAlpha:  childAlpha,
		ChildBeta:   childBeta,
		Score:       score,
		ScoreBound:  bound,
	})
}

func (s *Search) traceTargetLineQuiescence(alpha, beta int, inCheck bool) {
	if !searchTraceCompiled || s.Trace == nil || s.Trace.targetLineQSeen || s.Ply < 2 {
		return
	}
	if s.Trace.path[0] != s.Trace.RootMove || s.Trace.path[1] != s.Trace.ReplyMove {
		return
	}
	s.Trace.targetLineQSeen = true
	s.Trace.Events = append(s.Trace.Events, TraceEvent{
		Iteration:   s.Trace.iteration,
		Pass:        s.Trace.pass,
		Role:        "target_line",
		Node:        "quiescence",
		Stage:       "enter_after_reply",
		Ply:         s.Ply,
		Depth:       0,
		Alpha:       alpha,
		Beta:        beta,
		InCheck:     inCheck,
		PseudoIndex: -1,
		OrderIndex:  -1,
		LegalIndex:  -1,
		ChildDepth:  -1,
	})
}

func traceMoveIndex(moves *[256]uint32, count int, target uint32) int {
	for i := 0; i < count; i++ {
		if moves[i] == target {
			return i
		}
	}
	return -1
}
