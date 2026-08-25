package search

import "EquinoxChess/internal/tt"

// SearchTrace is an opt-in diagnostic recorder for one root move and its
// immediate reply. A nil Search.Trace keeps normal engine runs on the existing
// path; tracing never changes scores, ordering, pruning, or search depth.
type SearchTrace struct {
	RootMove  uint32
	ReplyMove uint32
	Events    []TraceEvent

	iteration       int
	pass            string
	path            [MaxDepth]uint32
	targetLineQSeen bool
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
	s.Trace.iteration = 0
	s.Trace.pass = ""
	s.Trace.path = [MaxDepth]uint32{}
	s.Trace.targetLineQSeen = false
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
