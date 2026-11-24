package search

import (
	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/movegen"
	"github.com/jrichard8/MarsGopherChess/internal/tt"
)

// SearchTrace is an opt-in diagnostic recorder for one root move and its
// immediate reply. A nil Search.Trace keeps normal engine runs on the existing
// path. Tracing is observational unless an explicit searchtrace-only
// counterfactual is enabled for one exact move and path.
type SearchTrace struct {
	RootMove                  uint32
	ReplyMove                 uint32
	ForceTargetReplyFullDepth bool
	TargetReplyExtraPlies     int
	ForceTargetReplyIteration int
	ForceBridgeStepNoLMR      bool
	BridgeNoLMRPly            int
	ForceBridgeNoLMRIteration int
	ForceBridgeNoFutility     bool
	ForceBridgeNoFutilityIter int
	ProbeD3CheckChains        bool
	ProbeD3CheckIteration     int
	ProbeD5ReducedFull        bool
	ProbeD5Iteration          int
	ForceGeneralD3D1Checks    bool
	ForceGeneralD5D3D1Checks  bool
	ForceGeneralD3D1Iteration int
	GeneralD5AblationPaths    [][]uint32
	BridgePath                []uint32
	BridgeMove                uint32
	BridgeLine                []uint32
	Events                    []TraceEvent
	Continuations             []TraceContinuation
	RootMoveResults           []TraceRootMoveResult
	D3CheckProbes             []TraceD3CheckProbe
	D3CheckFutilityChains     []TraceD3CheckFutilityChain
	D5ReducedFullProbes       []TraceD5ReducedFullProbe
	GeneralD3D1Events         []TraceGeneralD3D1Event

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
	recordD1Checks  bool
	d3ProbeOrigin   traceD3ProbeOrigin
	activeD3D1Ply   int
}

// TraceD3CheckProbe records an isolated full-depth shadow search for one
// baseline LMR-eligible quiet check at remaining depth 3. The shadow never
// modifies the live search board, TT, heuristics, node count, or result.
type TraceD3CheckProbe struct {
	Iteration         int      `json:"iteration"`
	Pass              string   `json:"pass"`
	Ply               int      `json:"ply"`
	Move              uint32   `json:"move"`
	OrderIndex        int      `json:"order_index"`
	LegalIndex        int      `json:"legal_index"`
	Alpha             int      `json:"alpha"`
	Beta              int      `json:"beta"`
	Reduction         int      `json:"reduction"`
	ReducedChildDepth int      `json:"reduced_child_depth"`
	FullChildDepth    int      `json:"full_child_depth"`
	FullDepthScore    int      `json:"full_depth_score"`
	ShadowNodes       uint64   `json:"shadow_nodes"`
	FutilityChecks    int      `json:"futility_checks"`
	Line              []uint32 `json:"line"`
}

// TraceD3CheckFutilityChain is one legal quiet check that the unchanged
// depth-1 futility predicate pruned inside a D3 quiet-check shadow.
type TraceD3CheckFutilityChain struct {
	Iteration  int      `json:"iteration"`
	Pass       string   `json:"pass"`
	D3Ply      int      `json:"d3_ply"`
	D3Move     uint32   `json:"d3_move"`
	D3Order    int      `json:"d3_order_index"`
	D3Legal    int      `json:"d3_legal_index"`
	D1Ply      int      `json:"d1_ply"`
	D1Move     uint32   `json:"d1_move"`
	D1Order    int      `json:"d1_order_index"`
	Alpha      int      `json:"alpha"`
	Beta       int      `json:"beta"`
	StaticEval int      `json:"static_eval"`
	Margin     int      `json:"futility_margin"`
	Line       []uint32 `json:"line"`
}

type traceD3ProbeOrigin struct {
	ply        int
	move       uint32
	orderIndex int
	legalIndex int
}

// TraceD5ReducedFullProbe compares the normal sequential LMR search with an
// independent full-depth search from the exact same board, TT, and heuristic
// snapshot. Nested D3 probes are observational shadows of the fresh full run.
type TraceD5ReducedFullProbe struct {
	Iteration                    int        `json:"iteration"`
	Pass                         string     `json:"pass"`
	Ply                          int        `json:"ply"`
	Move                         uint32     `json:"move"`
	OrderIndex                   int        `json:"order_index"`
	LegalIndex                   int        `json:"legal_index"`
	Alpha                        int        `json:"alpha"`
	Beta                         int        `json:"beta"`
	Reduction                    int        `json:"reduction"`
	ReducedChildDepth            int        `json:"reduced_child_depth"`
	FullChildDepth               int        `json:"full_child_depth"`
	ReducedScore                 int        `json:"reduced_score"`
	FreshFullScore               int        `json:"fresh_full_score"`
	SequentialScore              int        `json:"sequential_score"`
	FullMinusReduced             int        `json:"full_minus_reduced"`
	SequentialMinusFreshFull     int        `json:"sequential_minus_fresh_full"`
	ReducedExceedsAlpha          bool       `json:"reduced_exceeds_alpha"`
	FreshFullExceedsAlpha        bool       `json:"fresh_full_exceeds_alpha"`
	AlphaCrossingDisagrees       bool       `json:"alpha_crossing_disagrees"`
	SequentialDidFullResearch    bool       `json:"sequential_did_full_research"`
	SequentialReducedNodes       uint64     `json:"sequential_reduced_nodes"`
	SequentialFullNodes          uint64     `json:"sequential_full_nodes"`
	FreshFullNodes               uint64     `json:"fresh_full_nodes"`
	D3ShadowNodes                uint64     `json:"d3_shadow_nodes"`
	ParentTTMove                 uint32     `json:"parent_tt_move"`
	ParentPVMove                 uint32     `json:"parent_pv_move"`
	ParentMoveIsTT               bool       `json:"parent_move_is_tt"`
	ParentMoveIsPV               bool       `json:"parent_move_is_pv"`
	ChildTTKeyMatch              bool       `json:"child_tt_key_match"`
	ChildTTHalfmoveMatch         bool       `json:"child_tt_halfmove_match"`
	ChildTTEntryDepth            int        `json:"child_tt_entry_depth"`
	ChildTTStoredScore           int        `json:"child_tt_stored_score"`
	ChildTTFlag                  string     `json:"child_tt_flag"`
	ChildTTMove                  uint32     `json:"child_tt_move"`
	ChildTTReducedDepthEnough    bool       `json:"child_tt_reduced_depth_enough"`
	ChildTTFullDepthEnough       bool       `json:"child_tt_full_depth_enough"`
	ChildTTReducedCutoff         bool       `json:"child_tt_reduced_cutoff"`
	ChildTTFullCutoff            bool       `json:"child_tt_full_cutoff"`
	InitialFirstOrderedMove      uint32     `json:"initial_first_ordered_move"`
	AfterReducedTTMove           uint32     `json:"after_reduced_tt_move"`
	AfterReducedFirstOrderedMove uint32     `json:"after_reduced_first_ordered_move"`
	FreshFullTTMove              uint32     `json:"fresh_full_tt_move"`
	FreshFullFirstOrderedMove    uint32     `json:"fresh_full_first_ordered_move"`
	ChildTTEntryChanged          bool       `json:"child_tt_entry_changed"`
	FreshFullChildTTEntryChanged bool       `json:"fresh_full_child_tt_entry_changed"`
	ReducedFreshChildTTDiffers   bool       `json:"reduced_fresh_child_tt_differs"`
	ReducedTTEntriesChanged      int        `json:"reduced_tt_entries_changed"`
	FreshFullTTEntriesChanged    int        `json:"fresh_full_tt_entries_changed"`
	ReducedFreshTTEntriesDiffer  int        `json:"reduced_fresh_tt_entries_differ"`
	ReducedKillersChanged        bool       `json:"reduced_killers_changed"`
	FreshFullKillersChanged      bool       `json:"fresh_full_killers_changed"`
	ReducedFreshKillersDiffer    bool       `json:"reduced_fresh_killers_differ"`
	ReducedHistoryChanged        bool       `json:"reduced_history_changed"`
	FreshFullHistoryChanged      bool       `json:"fresh_full_history_changed"`
	ReducedFreshHistoryDiffers   bool       `json:"reduced_fresh_history_differs"`
	FirstOrderedMoveChanged      bool       `json:"first_ordered_move_changed"`
	FreshFirstOrderedMoveChanged bool       `json:"fresh_first_ordered_move_changed"`
	ReducedFreshOrderDiffers     bool       `json:"reduced_fresh_order_differs"`
	D3ProbeCount                 int        `json:"d3_probe_count"`
	D3D1ChainCount               int        `json:"d3_d1_chain_count"`
	Line                         []uint32   `json:"line"`
	ReducedLine                  []uint32   `json:"reduced_line"`
	FreshFullLine                []uint32   `json:"fresh_full_line"`
	SequentialLine               []uint32   `json:"sequential_line"`
	D3Lines                      [][]uint32 `json:"d3_lines"`
	D3D1Lines                    [][]uint32 `json:"d3_d1_lines"`
}

// TraceGeneralD3D1Event records an application of the generalized, coupled
// final-iteration counterfactual. The optional D5 stage is independent; the
// D1 exemption remains active only two plies below a D3 quiet check that this
// same counterfactual searched full-depth.
type TraceGeneralD3D1Event struct {
	Iteration  int      `json:"iteration"`
	Pass       string   `json:"pass"`
	Stage      string   `json:"stage"`
	Ply        int      `json:"ply"`
	Depth      int      `json:"depth"`
	Move       uint32   `json:"move"`
	OrderIndex int      `json:"order_index"`
	LegalIndex int      `json:"legal_index"`
	Alpha      int      `json:"alpha"`
	Beta       int      `json:"beta"`
	StaticEval int      `json:"static_eval"`
	ChildDepth int      `json:"child_depth"`
	Reduction  int      `json:"reduction"`
	Line       []uint32 `json:"line"`
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
	s.Trace.D3CheckProbes = nil
	s.Trace.D3CheckFutilityChains = nil
	s.Trace.D5ReducedFullProbes = nil
	s.Trace.GeneralD3D1Events = nil
	s.Trace.iteration = 0
	s.Trace.pass = ""
	s.Trace.path = [MaxDepth]uint32{}
	s.Trace.targetLineQSeen = false
	s.Trace.recordD1Checks = false
	s.Trace.d3ProbeOrigin = traceD3ProbeOrigin{}
	s.Trace.activeD3D1Ply = -1
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
	if bridgeMove, bridgeRole, bridgeOK := s.traceBridgeTargetAt(ply); bridgeOK {
		return bridgeMove, bridgeRole, true
	}
	return 0, "", false
}

func (s *Search) traceBridgeTargetAt(ply int) (uint32, string, bool) {
	if !searchTraceCompiled || s.Trace == nil || len(s.Trace.BridgeLine) < 3 ||
		ply < 2 || ply >= len(s.Trace.BridgeLine) {
		return 0, "", false
	}
	for index := 0; index < ply; index++ {
		if s.Trace.path[index] != s.Trace.BridgeLine[index] {
			return 0, "", false
		}
	}
	role := "bridge_step"
	if ply == len(s.Trace.BridgeLine)-1 {
		role = "bridge_move"
	}
	return s.Trace.BridgeLine[ply], role, true
}

func (s *Search) traceBridgePathMatches(ply int) bool {
	if !searchTraceCompiled || s.Trace == nil || s.Trace.BridgeMove == 0 ||
		len(s.Trace.BridgePath) < 2 || ply != len(s.Trace.BridgePath) {
		return false
	}
	for index, expected := range s.Trace.BridgePath {
		if s.Trace.path[index] != expected {
			return false
		}
	}
	return true
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

func (s *Search) traceBridgeStepNoLMR(ply int, m uint32) bool {
	if !searchTraceCompiled || s.Trace == nil || !s.Trace.ForceBridgeStepNoLMR ||
		(s.Trace.ForceBridgeNoLMRIteration != 0 && s.Trace.iteration != s.Trace.ForceBridgeNoLMRIteration) ||
		ply != s.Trace.BridgeNoLMRPly || ply < 2 || ply >= len(s.Trace.BridgeLine)-1 ||
		m != s.Trace.BridgeLine[ply] {
		return false
	}
	for index := 0; index <= ply; index++ {
		if s.Trace.path[index] != s.Trace.BridgeLine[index] {
			return false
		}
	}
	return true
}

func (s *Search) traceBridgeNoFutility(ply int, m uint32) bool {
	return searchTraceCompiled && s.Trace != nil && s.Trace.ForceBridgeNoFutility &&
		(s.Trace.ForceBridgeNoFutilityIter == 0 || s.Trace.iteration == s.Trace.ForceBridgeNoFutilityIter) &&
		m == s.Trace.BridgeMove && s.traceBridgePathMatches(ply)
}

func (s *Search) traceD3CheckProbeEnabled() bool {
	return searchTraceCompiled && s.Trace != nil && s.Trace.ProbeD3CheckChains &&
		(s.Trace.ProbeD3CheckIteration == 0 || s.Trace.iteration == s.Trace.ProbeD3CheckIteration)
}

func (s *Search) traceD5ReducedFullProbeEnabled() bool {
	return searchTraceCompiled && s.Trace != nil && s.Trace.ProbeD5ReducedFull &&
		(s.Trace.ProbeD5Iteration == 0 || s.Trace.iteration == s.Trace.ProbeD5Iteration)
}

func (s *Search) traceGeneralD3D1Enabled() bool {
	return searchTraceCompiled && s.Trace != nil &&
		(s.Trace.ForceGeneralD3D1Checks || s.Trace.ForceGeneralD5D3D1Checks) &&
		(s.Trace.ForceGeneralD3D1Iteration == 0 || s.Trace.iteration == s.Trace.ForceGeneralD3D1Iteration)
}

func (s *Search) traceGeneralD5Enabled() bool {
	return searchTraceCompiled && s.Trace != nil && s.Trace.ForceGeneralD5D3D1Checks &&
		(s.Trace.ForceGeneralD3D1Iteration == 0 || s.Trace.iteration == s.Trace.ForceGeneralD3D1Iteration)
}

func (s *Search) traceGeneralD5Ablated(ply int, m uint32) bool {
	if !s.traceGeneralD5Enabled() || s.Trace == nil || ply < 0 {
		return false
	}
	for _, ablationPath := range s.Trace.GeneralD5AblationPaths {
		if len(ablationPath) != ply+1 || ablationPath[ply] != m {
			continue
		}
		matches := true
		for index := 0; index < ply; index++ {
			if s.Trace.path[index] != ablationPath[index] {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func (s *Search) recordTraceGeneralD5Ablation(
	ply int,
	m uint32,
	orderIndex int,
	legalIndex int,
	alpha int,
	beta int,
	childDepth int,
	reduction int,
) {
	if !searchTraceCompiled || s.Trace == nil {
		return
	}
	s.Trace.GeneralD3D1Events = append(s.Trace.GeneralD3D1Events, TraceGeneralD3D1Event{
		Iteration:  s.Trace.iteration,
		Pass:       s.Trace.pass,
		Stage:      "d5_check_ablation_lmr_preserved",
		Ply:        ply,
		Depth:      5,
		Move:       m,
		OrderIndex: orderIndex,
		LegalIndex: legalIndex,
		Alpha:      alpha,
		Beta:       beta,
		ChildDepth: childDepth,
		Reduction:  reduction,
		Line:       append([]uint32(nil), s.Trace.path[:ply+1]...),
	})
}

func (s *Search) recordTraceGeneralD5Check(
	ply int,
	m uint32,
	orderIndex int,
	legalIndex int,
	alpha int,
	beta int,
	childDepth int,
) {
	if !s.traceGeneralD5Enabled() || s.Trace == nil {
		return
	}
	s.Trace.GeneralD3D1Events = append(s.Trace.GeneralD3D1Events, TraceGeneralD3D1Event{
		Iteration:  s.Trace.iteration,
		Pass:       s.Trace.pass,
		Stage:      "d5_check_forced_full_depth",
		Ply:        ply,
		Depth:      5,
		Move:       m,
		OrderIndex: orderIndex,
		LegalIndex: legalIndex,
		Alpha:      alpha,
		Beta:       beta,
		ChildDepth: childDepth,
		Line:       append([]uint32(nil), s.Trace.path[:ply+1]...),
	})
}

func (s *Search) beginTraceGeneralD3Check(
	ply int,
	m uint32,
	orderIndex int,
	legalIndex int,
	alpha int,
	beta int,
	childDepth int,
) int {
	if !s.traceGeneralD3D1Enabled() || s.Trace == nil {
		return -1
	}
	previous := s.Trace.activeD3D1Ply
	s.Trace.activeD3D1Ply = ply
	s.Trace.GeneralD3D1Events = append(s.Trace.GeneralD3D1Events, TraceGeneralD3D1Event{
		Iteration:  s.Trace.iteration,
		Pass:       s.Trace.pass,
		Stage:      "d3_check_forced_full_depth",
		Ply:        ply,
		Depth:      3,
		Move:       m,
		OrderIndex: orderIndex,
		LegalIndex: legalIndex,
		Alpha:      alpha,
		Beta:       beta,
		ChildDepth: childDepth,
		Line:       append([]uint32(nil), s.Trace.path[:ply+1]...),
	})
	return previous
}

func (s *Search) endTraceGeneralD3Check(previous int) {
	if searchTraceCompiled && s.Trace != nil {
		s.Trace.activeD3D1Ply = previous
	}
}

func traceMoveIsLegalCheck(b *board.Board, m uint32) bool {
	b.MakeMove(m)
	legalCheck := !IsMovedSideInCheck(b) && InCheck(b)
	b.UnmakeMove()
	return legalCheck
}

func (s *Search) traceGeneralD1NoFutility(
	b *board.Board,
	ply int,
	m uint32,
	orderIndex int,
	alpha int,
	beta int,
	staticEval int,
) bool {
	if !s.traceGeneralD3D1Enabled() || s.Trace == nil || s.Trace.activeD3D1Ply < 0 ||
		ply != s.Trace.activeD3D1Ply+2 || !traceMoveIsLegalCheck(b, m) {
		return false
	}
	line := append([]uint32(nil), s.Trace.path[:ply]...)
	line = append(line, m)
	s.Trace.GeneralD3D1Events = append(s.Trace.GeneralD3D1Events, TraceGeneralD3D1Event{
		Iteration:  s.Trace.iteration,
		Pass:       s.Trace.pass,
		Stage:      "d1_check_forced_no_futility",
		Ply:        ply,
		Depth:      1,
		Move:       m,
		OrderIndex: orderIndex,
		LegalIndex: -1,
		Alpha:      alpha,
		Beta:       beta,
		StaticEval: staticEval,
		ChildDepth: 0,
		Line:       line,
	})
	return true
}

func cloneTraceTT(source *tt.TranspositionTable) *tt.TranspositionTable {
	if source == nil {
		return nil
	}
	return &tt.TranspositionTable{
		Entries: append([]tt.TTEntry(nil), source.Entries...),
		Mask:    source.Mask,
	}
}

func traceTTFlag(flag uint8) string {
	switch flag {
	case tt.TTExact:
		return "exact"
	case tt.TTAlpha:
		return "alpha"
	case tt.TTBeta:
		return "beta"
	default:
		return "unknown"
	}
}

func traceTTEntryAt(table *tt.TranspositionTable, b *board.Board) tt.TTEntry {
	if table == nil || len(table.Entries) == 0 {
		return tt.TTEntry{}
	}
	return table.Entries[b.ZobristKey&table.Mask]
}

func traceTTDifferenceCount(left, right *tt.TranspositionTable) int {
	if left == nil || right == nil {
		if left == right {
			return 0
		}
		return 1
	}
	limit := len(left.Entries)
	if len(right.Entries) < limit {
		limit = len(right.Entries)
	}
	differences := 0
	for index := 0; index < limit; index++ {
		if left.Entries[index] != right.Entries[index] {
			differences++
		}
	}
	if len(left.Entries) > limit {
		differences += len(left.Entries) - limit
	}
	if len(right.Entries) > limit {
		differences += len(right.Entries) - limit
	}
	return differences
}

func traceTTProbe(
	table *tt.TranspositionTable,
	b *board.Board,
	depth int,
	alpha int,
	beta int,
	ply int,
) (int, uint32, bool) {
	if table == nil {
		return 0, 0, false
	}
	return table.Probe(b.ZobristKey, b.HalfmoveClock, depth, alpha, beta, ply, MateThreshold)
}

func traceFirstOrderedMove(s *Search, b *board.Board, ttMove uint32) uint32 {
	var moves [movegen.MaxMoves]uint32
	count := movegen.GeneratePseudoLegalMoves(b, &moves)
	if count == 0 {
		return 0
	}
	if ttMove != 0 {
		for index := 0; index < count; index++ {
			if moves[index] == ttMove {
				moves[0], moves[index] = moves[index], moves[0]
				break
			}
		}
	}
	if s.Ply == 0 && s.PVMove[0] != 0 {
		for index := 0; index < count; index++ {
			if moves[index] == s.PVMove[0] {
				moves[0], moves[index] = moves[index], moves[0]
				break
			}
		}
	}
	return moves[selectBestMove(s, &moves, 0, count)]
}

func traceShadowPVLine(trace *SearchTrace, ply int, prefix []uint32) []uint32 {
	line := append([]uint32(nil), prefix...)
	if trace == nil || ply < 0 || ply >= MaxDepth {
		return line
	}
	length := trace.pvLength[ply]
	if length > MaxDepth {
		length = MaxDepth
	}
	for index := ply; index < length; index++ {
		if trace.pvMoves[ply][index] != 0 {
			line = append(line, trace.pvMoves[ply][index])
		}
	}
	return line
}

func cloneTraceSearch(source *Search, trace *SearchTrace) Search {
	return Search{
		Ply:         source.Ply,
		MaxDepth:    source.MaxDepth,
		KillerMoves: source.KillerMoves,
		History:     source.History,
		TT:          cloneTraceTT(source.TT),
		PVMove:      source.PVMove,
		Trace:       trace,
	}
}

func (s *Search) runTraceD5ReducedFullProbe(
	b *board.Board,
	ply int,
	m uint32,
	orderIndex int,
	legalIndex int,
	alpha int,
	beta int,
	reduction int,
	reducedChildDepth int,
	parentTTMove uint32,
	parentPVMove uint32,
) {
	if !s.traceD5ReducedFullProbeEnabled() || s.Trace == nil {
		return
	}
	fullChildDepth := 4
	prefix := append([]uint32(nil), s.Trace.path[:ply+1]...)
	initialEntry := traceTTEntryAt(s.TT, b)
	childTTKeyMatch := initialEntry.Key == b.ZobristKey
	childTTHalfmoveMatch := childTTKeyMatch && initialEntry.HalfmoveClock == b.HalfmoveClock
	_, childTTMove, reducedCutoff := traceTTProbe(
		s.TT,
		b,
		reducedChildDepth,
		-alpha-1,
		-alpha,
		s.Ply,
	)
	_, _, fullCutoff := traceTTProbe(
		s.TT,
		b,
		fullChildDepth,
		-beta,
		-alpha,
		s.Ply,
	)
	initialFirstMove := traceFirstOrderedMove(s, b, childTTMove)

	sequentialTrace := &SearchTrace{
		iteration: s.Trace.iteration,
		pass:      s.Trace.pass,
		path:      s.Trace.path,
	}
	sequentialBoard := *b
	sequential := cloneTraceSearch(s, sequentialTrace)
	reducedChildScore := sequential.Negamax(
		&sequentialBoard,
		reducedChildDepth,
		-alpha-1,
		-alpha,
	)
	reducedScore := -reducedChildScore
	reducedNodes := sequential.Nodes
	reducedLine := traceShadowPVLine(sequentialTrace, sequential.Ply, prefix)
	afterReducedEntry := traceTTEntryAt(sequential.TT, b)
	afterReducedTT := cloneTraceTT(sequential.TT)
	afterReducedKillers := sequential.KillerMoves
	afterReducedHistory := sequential.History
	_, afterReducedTTMove, _ := traceTTProbe(
		sequential.TT,
		b,
		fullChildDepth,
		-beta,
		-alpha,
		sequential.Ply,
	)
	afterReducedFirstMove := traceFirstOrderedMove(&sequential, b, afterReducedTTMove)
	sequentialScore := reducedScore
	sequentialDidFull := reducedScore > alpha
	if sequentialDidFull {
		sequentialChildScore := sequential.Negamax(
			&sequentialBoard,
			fullChildDepth,
			-beta,
			-alpha,
		)
		sequentialScore = -sequentialChildScore
	}
	sequentialFullNodes := sequential.Nodes - reducedNodes
	sequentialLine := traceShadowPVLine(sequentialTrace, sequential.Ply, prefix)

	fullTrace := &SearchTrace{
		ProbeD3CheckChains:    true,
		ProbeD3CheckIteration: s.Trace.iteration,
		iteration:             s.Trace.iteration,
		pass:                  s.Trace.pass,
		path:                  s.Trace.path,
	}
	fullBoard := *b
	full := cloneTraceSearch(s, fullTrace)
	freshFullChildScore := full.Negamax(&fullBoard, fullChildDepth, -beta, -alpha)
	freshFullScore := -freshFullChildScore
	freshFullLine := traceShadowPVLine(fullTrace, full.Ply, prefix)
	freshFullEntry := traceTTEntryAt(full.TT, b)
	_, freshFullTTMove, _ := traceTTProbe(
		full.TT,
		b,
		fullChildDepth,
		-beta,
		-alpha,
		full.Ply,
	)
	freshFullFirstMove := traceFirstOrderedMove(&full, b, freshFullTTMove)
	d3Lines := make([][]uint32, 0, len(fullTrace.D3CheckProbes))
	var d3ShadowNodes uint64
	for _, probe := range fullTrace.D3CheckProbes {
		d3Lines = append(d3Lines, append([]uint32(nil), probe.Line...))
		d3ShadowNodes += probe.ShadowNodes
	}
	d3D1Lines := make([][]uint32, 0, len(fullTrace.D3CheckFutilityChains))
	for _, chain := range fullTrace.D3CheckFutilityChains {
		d3D1Lines = append(d3D1Lines, append([]uint32(nil), chain.Line...))
	}

	childTTFlag := ""
	if childTTKeyMatch {
		childTTFlag = traceTTFlag(initialEntry.Flag)
	}
	s.Trace.D5ReducedFullProbes = append(s.Trace.D5ReducedFullProbes, TraceD5ReducedFullProbe{
		Iteration:                    s.Trace.iteration,
		Pass:                         s.Trace.pass,
		Ply:                          ply,
		Move:                         m,
		OrderIndex:                   orderIndex,
		LegalIndex:                   legalIndex,
		Alpha:                        alpha,
		Beta:                         beta,
		Reduction:                    reduction,
		ReducedChildDepth:            reducedChildDepth,
		FullChildDepth:               fullChildDepth,
		ReducedScore:                 reducedScore,
		FreshFullScore:               freshFullScore,
		SequentialScore:              sequentialScore,
		FullMinusReduced:             freshFullScore - reducedScore,
		SequentialMinusFreshFull:     sequentialScore - freshFullScore,
		ReducedExceedsAlpha:          reducedScore > alpha,
		FreshFullExceedsAlpha:        freshFullScore > alpha,
		AlphaCrossingDisagrees:       (reducedScore > alpha) != (freshFullScore > alpha),
		SequentialDidFullResearch:    sequentialDidFull,
		SequentialReducedNodes:       reducedNodes,
		SequentialFullNodes:          sequentialFullNodes,
		FreshFullNodes:               full.Nodes,
		D3ShadowNodes:                d3ShadowNodes,
		ParentTTMove:                 parentTTMove,
		ParentPVMove:                 parentPVMove,
		ParentMoveIsTT:               parentTTMove != 0 && parentTTMove == m,
		ParentMoveIsPV:               parentPVMove != 0 && parentPVMove == m,
		ChildTTKeyMatch:              childTTKeyMatch,
		ChildTTHalfmoveMatch:         childTTHalfmoveMatch,
		ChildTTEntryDepth:            initialEntry.Depth,
		ChildTTStoredScore:           initialEntry.Score,
		ChildTTFlag:                  childTTFlag,
		ChildTTMove:                  childTTMove,
		ChildTTReducedDepthEnough:    childTTHalfmoveMatch && initialEntry.Depth >= reducedChildDepth,
		ChildTTFullDepthEnough:       childTTHalfmoveMatch && initialEntry.Depth >= fullChildDepth,
		ChildTTReducedCutoff:         reducedCutoff,
		ChildTTFullCutoff:            fullCutoff,
		InitialFirstOrderedMove:      initialFirstMove,
		AfterReducedTTMove:           afterReducedTTMove,
		AfterReducedFirstOrderedMove: afterReducedFirstMove,
		FreshFullTTMove:              freshFullTTMove,
		FreshFullFirstOrderedMove:    freshFullFirstMove,
		ChildTTEntryChanged:          initialEntry != afterReducedEntry,
		FreshFullChildTTEntryChanged: initialEntry != freshFullEntry,
		ReducedFreshChildTTDiffers:   afterReducedEntry != freshFullEntry,
		ReducedTTEntriesChanged:      traceTTDifferenceCount(s.TT, afterReducedTT),
		FreshFullTTEntriesChanged:    traceTTDifferenceCount(s.TT, full.TT),
		ReducedFreshTTEntriesDiffer:  traceTTDifferenceCount(afterReducedTT, full.TT),
		ReducedKillersChanged:        s.KillerMoves != afterReducedKillers,
		FreshFullKillersChanged:      s.KillerMoves != full.KillerMoves,
		ReducedFreshKillersDiffer:    afterReducedKillers != full.KillerMoves,
		ReducedHistoryChanged:        s.History != afterReducedHistory,
		FreshFullHistoryChanged:      s.History != full.History,
		ReducedFreshHistoryDiffers:   afterReducedHistory != full.History,
		FirstOrderedMoveChanged:      initialFirstMove != afterReducedFirstMove,
		FreshFirstOrderedMoveChanged: initialFirstMove != freshFullFirstMove,
		ReducedFreshOrderDiffers:     afterReducedFirstMove != freshFullFirstMove,
		D3ProbeCount:                 len(fullTrace.D3CheckProbes),
		D3D1ChainCount:               len(fullTrace.D3CheckFutilityChains),
		Line:                         prefix,
		ReducedLine:                  reducedLine,
		FreshFullLine:                freshFullLine,
		SequentialLine:               sequentialLine,
		D3Lines:                      d3Lines,
		D3D1Lines:                    d3D1Lines,
	})
}

func (s *Search) runTraceD3CheckProbe(
	b *board.Board,
	ply int,
	m uint32,
	orderIndex int,
	legalIndex int,
	alpha int,
	beta int,
	reduction int,
	reducedChildDepth int,
) {
	if !s.traceD3CheckProbeEnabled() || s.Trace == nil {
		return
	}
	line := append([]uint32(nil), s.Trace.path[:ply+1]...)
	origin := traceD3ProbeOrigin{
		ply:        ply,
		move:       m,
		orderIndex: orderIndex,
		legalIndex: legalIndex,
	}
	shadowTrace := &SearchTrace{
		recordD1Checks: true,
		d3ProbeOrigin:  origin,
		iteration:      s.Trace.iteration,
		pass:           s.Trace.pass,
		path:           s.Trace.path,
	}
	shadowBoard := *b
	shadow := Search{
		Ply:         s.Ply,
		MaxDepth:    s.MaxDepth,
		KillerMoves: s.KillerMoves,
		History:     s.History,
		TT:          cloneTraceTT(s.TT),
		PVMove:      s.PVMove,
		Trace:       shadowTrace,
	}
	childScore := shadow.Negamax(&shadowBoard, 2, -beta, -alpha)
	s.Trace.D3CheckFutilityChains = append(
		s.Trace.D3CheckFutilityChains,
		shadowTrace.D3CheckFutilityChains...,
	)
	s.Trace.D3CheckProbes = append(s.Trace.D3CheckProbes, TraceD3CheckProbe{
		Iteration:         s.Trace.iteration,
		Pass:              s.Trace.pass,
		Ply:               ply,
		Move:              m,
		OrderIndex:        orderIndex,
		LegalIndex:        legalIndex,
		Alpha:             alpha,
		Beta:              beta,
		Reduction:         reduction,
		ReducedChildDepth: reducedChildDepth,
		FullChildDepth:    2,
		FullDepthScore:    -childScore,
		ShadowNodes:       shadow.Nodes,
		FutilityChecks:    len(shadowTrace.D3CheckFutilityChains),
		Line:              line,
	})
}

func (s *Search) recordTraceD1FutilityCheck(
	b *board.Board,
	ply int,
	m uint32,
	orderIndex int,
	alpha int,
	beta int,
	staticEval int,
) {
	if !searchTraceCompiled || s.Trace == nil || !s.Trace.recordD1Checks ||
		ply != s.Trace.d3ProbeOrigin.ply+2 {
		return
	}
	if !traceMoveIsLegalCheck(b, m) {
		return
	}
	line := append([]uint32(nil), s.Trace.path[:ply]...)
	line = append(line, m)
	s.Trace.D3CheckFutilityChains = append(s.Trace.D3CheckFutilityChains, TraceD3CheckFutilityChain{
		Iteration:  s.Trace.iteration,
		Pass:       s.Trace.pass,
		D3Ply:      s.Trace.d3ProbeOrigin.ply,
		D3Move:     s.Trace.d3ProbeOrigin.move,
		D3Order:    s.Trace.d3ProbeOrigin.orderIndex,
		D3Legal:    s.Trace.d3ProbeOrigin.legalIndex,
		D1Ply:      ply,
		D1Move:     m,
		D1Order:    orderIndex,
		Alpha:      alpha,
		Beta:       beta,
		StaticEval: staticEval,
		Margin:     150,
		Line:       line,
	})
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
