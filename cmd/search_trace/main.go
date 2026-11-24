// Command search_trace replays selected tactical positions with opt-in search
// instrumentation. It is an offline diagnostic and does not change engine play.
package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/move"
	"github.com/jrichard8/MarsGopherChess/internal/search"
	"github.com/jrichard8/MarsGopherChess/internal/tt"
	"github.com/jrichard8/MarsGopherChess/internal/uci"
)

type corpusCase struct {
	Name           string `json:"name"`
	Cohort         string `json:"cohort"`
	GameIndex      int    `json:"game_index"`
	Ply            int    `json:"ply"`
	FEN            string `json:"fen"`
	RootMoveUCI    string `json:"root_move_uci"`
	RootMoveSAN    string `json:"root_move_san"`
	ReplyMoveUCI   string `json:"reply_move_uci"`
	ReplyMoveSAN   string `json:"reply_move_san"`
	Severe         bool   `json:"severe"`
	QueenNetDanger int    `json:"queen_net_danger"`
	Runs           []run  `json:"runs"`
}

type run struct {
	RequestedDepth                    int                   `json:"requested_depth"`
	CompletedDepth                    int                   `json:"completed_depth"`
	BestMoveUCI                       string                `json:"best_move_uci"`
	Score                             int                   `json:"score"`
	Nodes                             uint64                `json:"nodes"`
	RootMoveBest                      bool                  `json:"root_move_best"`
	CounterfactualMode                string                `json:"counterfactual_mode"`
	CounterfactualApplied             bool                  `json:"counterfactual_applied"`
	BridgeTraceApplied                bool                  `json:"bridge_trace_applied"`
	BridgeNoLMRApplied                bool                  `json:"bridge_no_lmr_applied"`
	BridgeNoFutilityApplied           bool                  `json:"bridge_no_futility_applied"`
	StructuralProbeApplied            bool                  `json:"structural_probe_applied"`
	StructuralProbeShadowNodes        uint64                `json:"structural_probe_shadow_nodes"`
	D5ReducedFullProbeApplied         bool                  `json:"d5_reduced_full_probe_applied"`
	D5ReducedFullShadowNodes          uint64                `json:"d5_reduced_full_shadow_nodes"`
	D5ShadowDeterminismVerified       bool                  `json:"d5_shadow_determinism_verified"`
	GeneralD3D1Applied                bool                  `json:"general_d3_d1_applied"`
	GeneralD5D3D1Applied              bool                  `json:"general_d5_d3_d1_applied"`
	GeneralD5AblationApplied          bool                  `json:"general_d5_ablation_applied"`
	GeneralD5AblationLineUCI          []string              `json:"general_d5_ablation_line_uci"`
	Verification                      string                `json:"verification"`
	NonInterferenceVerified           bool                  `json:"non_interference_verified"`
	CounterfactualDeterminismVerified bool                  `json:"counterfactual_determinism_verified"`
	Summary                           runSummary            `json:"summary"`
	Continuations                     []continuationOutput  `json:"continuations"`
	FinalContinuation                 *continuationOutput   `json:"final_continuation"`
	FinalRootMoveResults              []rootMoveOutput      `json:"final_root_move_results"`
	RootMargin                        *rootMarginOutput     `json:"root_margin"`
	Events                            []traceEventOutput    `json:"events"`
	D3CheckProbes                     []d3CheckProbeOutput  `json:"d3_check_probes"`
	D3CheckFutilityChains             []d3CheckChainOutput  `json:"d3_check_futility_chains"`
	D5ReducedFullProbes               []d5ReducedFullOutput `json:"d5_reduced_full_probes"`
	GeneralD3D1Events                 []generalD3D1Output   `json:"general_d3_d1_events"`
}

type d3CheckProbeOutput struct {
	search.TraceD3CheckProbe
	MoveUCI string   `json:"move_uci"`
	LineUCI []string `json:"line_uci"`
}

type d3CheckChainOutput struct {
	search.TraceD3CheckFutilityChain
	D3MoveUCI string   `json:"d3_move_uci"`
	D1MoveUCI string   `json:"d1_move_uci"`
	LineUCI   []string `json:"line_uci"`
}

type d5ReducedFullOutput struct {
	search.TraceD5ReducedFullProbe
	MoveUCI                         string     `json:"move_uci"`
	ParentTTMoveUCI                 string     `json:"parent_tt_move_uci"`
	ParentPVMoveUCI                 string     `json:"parent_pv_move_uci"`
	ChildTTMoveUCI                  string     `json:"child_tt_move_uci"`
	InitialFirstOrderedMoveUCI      string     `json:"initial_first_ordered_move_uci"`
	AfterReducedTTMoveUCI           string     `json:"after_reduced_tt_move_uci"`
	AfterReducedFirstOrderedMoveUCI string     `json:"after_reduced_first_ordered_move_uci"`
	FreshFullTTMoveUCI              string     `json:"fresh_full_tt_move_uci"`
	FreshFullFirstOrderedMoveUCI    string     `json:"fresh_full_first_ordered_move_uci"`
	LineUCI                         []string   `json:"line_uci"`
	ReducedLineUCI                  []string   `json:"reduced_line_uci"`
	FreshFullLineUCI                []string   `json:"fresh_full_line_uci"`
	SequentialLineUCI               []string   `json:"sequential_line_uci"`
	D3LinesUCI                      [][]string `json:"d3_lines_uci"`
	D3D1LinesUCI                    [][]string `json:"d3_d1_lines_uci"`
}

type generalD3D1Output struct {
	search.TraceGeneralD3D1Event
	MoveUCI string   `json:"move_uci"`
	LineUCI []string `json:"line_uci"`
}

type traceEventOutput struct {
	search.TraceEvent
	TargetMoveUCI  string `json:"target_move_uci"`
	CurrentMoveUCI string `json:"current_move_uci"`
}

type continuationOutput struct {
	search.TraceContinuation
	LineUCI           []string `json:"line_uci"`
	AfterReplyUCI     []string `json:"after_reply_uci"`
	ScoreRootPOV      int      `json:"score_root_pov"`
	LeafAfterReplyPly int      `json:"leaf_after_reply_ply"`
	LeafScoreRootPOV  int      `json:"leaf_score_root_pov"`
}

type rootMoveOutput struct {
	search.TraceRootMoveResult
	MoveUCI string `json:"move_uci"`
}

type rootMarginOutput struct {
	HistoricalMove                  rootMoveOutput `json:"historical_move"`
	BestCompetitor                  rootMoveOutput `json:"best_competitor"`
	HistoricalMinusBestCompetitorCP int            `json:"historical_minus_best_competitor_cp"`
	MarginBound                     string         `json:"margin_bound"`
	BoundMeaning                    string         `json:"bound_meaning"`
}

type roleOutcome struct {
	Events         int            `json:"events"`
	StageCounts    map[string]int `json:"stage_counts"`
	PseudoIndexes  []int          `json:"pseudo_indexes"`
	OrderIndexes   []int          `json:"order_indexes"`
	LegalIndexes   []int          `json:"legal_indexes"`
	TargetAsTT     bool           `json:"target_as_tt"`
	TargetAsPV     bool           `json:"target_as_pv"`
	TargetAsKiller bool           `json:"target_as_killer"`
}

type runSummary struct {
	RootAllIterations           roleOutcome `json:"root_all_iterations"`
	ReplyAllIterations          roleOutcome `json:"reply_all_iterations"`
	RootFinalIteration          roleOutcome `json:"root_final_iteration"`
	ReplyFinalIteration         roleOutcome `json:"reply_final_iteration"`
	FinalPass                   string      `json:"final_pass"`
	RootFinalPass               roleOutcome `json:"root_final_pass"`
	ReplyFinalPass              roleOutcome `json:"reply_final_pass"`
	BridgeAllIterations         roleOutcome `json:"bridge_all_iterations"`
	BridgeFinalIteration        roleOutcome `json:"bridge_final_iteration"`
	BridgeFinalPass             roleOutcome `json:"bridge_final_pass"`
	BridgeStepsAllIterations    roleOutcome `json:"bridge_steps_all_iterations"`
	BridgeStepsFinalIteration   roleOutcome `json:"bridge_steps_final_iteration"`
	BridgeStepsFinalPass        roleOutcome `json:"bridge_steps_final_pass"`
	TargetLineQEntriesAll       int         `json:"target_line_qsearch_entries_all"`
	TargetLineQEntriesFinal     int         `json:"target_line_qsearch_entries_final"`
	TargetLineQEntriesFinalPass int         `json:"target_line_qsearch_entries_final_pass"`
}

type aggregate struct {
	Cases                     int            `json:"cases"`
	RootMoveBest              int            `json:"root_move_best"`
	HistoricalBadMoveSelected int            `json:"historical_bad_move_selected"`
	BadMoveAvoided            int            `json:"bad_move_avoided"`
	LossRecognized            int            `json:"loss_recognized"`
	Metrics                   map[string]int `json:"metrics"`
}

type report struct {
	SchemaVersion int                             `json:"schema_version"`
	Inputs        map[string]map[string]string    `json:"inputs"`
	Parameters    map[string]any                  `json:"parameters"`
	Counts        map[string]int                  `json:"counts"`
	Aggregates    map[string]map[string]aggregate `json:"aggregates_by_depth_and_cohort"`
	Structural    structuralProbeSummary          `json:"structural_probe_summary"`
	D5ReducedFull d5ReducedFullSummary            `json:"d5_reduced_full_summary"`
	GeneralD3D1   generalD3D1Summary              `json:"general_d3_d1_summary"`
	Cases         []corpusCase                    `json:"cases"`
}

type structuralProbeSummary struct {
	Applied            bool     `json:"applied"`
	D3ProbeOccurrences int      `json:"d3_probe_occurrences"`
	UniqueD3Paths      int      `json:"unique_d3_paths"`
	CasesWithD3Probes  int      `json:"cases_with_d3_probes"`
	ChainOccurrences   int      `json:"chain_occurrences"`
	UniqueChains       int      `json:"unique_chains"`
	CasesWithChains    int      `json:"cases_with_chains"`
	ChainCaseNames     []string `json:"chain_case_names"`
	ShadowNodes        uint64   `json:"shadow_nodes"`
}

type d5ReducedFullSummary struct {
	Applied                        bool     `json:"applied"`
	Occurrences                    int      `json:"occurrences"`
	UniquePaths                    int      `json:"unique_paths"`
	CasesWithProbes                int      `json:"cases_with_probes"`
	ReducedCrossesAlpha            int      `json:"reduced_crosses_alpha"`
	FreshFullCrossesAlpha          int      `json:"fresh_full_crosses_alpha"`
	AlphaCrossingDisagreements     int      `json:"alpha_crossing_disagreements"`
	FullScoreDifferences           int      `json:"full_score_differences"`
	SequentialFullResearches       int      `json:"sequential_full_researches"`
	SequentialFreshDifferences     int      `json:"sequential_fresh_differences"`
	ParentTTMoves                  int      `json:"parent_tt_moves"`
	ParentPVMoves                  int      `json:"parent_pv_moves"`
	InitialChildTTHits             int      `json:"initial_child_tt_hits"`
	InitialReducedTTCutoffs        int      `json:"initial_reduced_tt_cutoffs"`
	InitialFullTTCutoffs           int      `json:"initial_full_tt_cutoffs"`
	ChildTTEntriesChanged          int      `json:"child_tt_entries_changed"`
	FreshFullChildTTEntriesChanged int      `json:"fresh_full_child_tt_entries_changed"`
	ReducedFreshChildTTDifferences int      `json:"reduced_fresh_child_tt_differences"`
	ReducedTTEntriesChanged        int      `json:"reduced_tt_entries_changed"`
	FreshFullTTEntriesChanged      int      `json:"fresh_full_tt_entries_changed"`
	ReducedFreshTTEntriesDiffer    int      `json:"reduced_fresh_tt_entries_differ"`
	FirstOrderedMovesChanged       int      `json:"first_ordered_moves_changed"`
	FreshFirstOrderedMovesChanged  int      `json:"fresh_first_ordered_moves_changed"`
	ReducedFreshOrderDifferences   int      `json:"reduced_fresh_order_differences"`
	ReducedKillersChanged          int      `json:"reduced_killers_changed"`
	FreshFullKillersChanged        int      `json:"fresh_full_killers_changed"`
	ReducedFreshKillersDiffer      int      `json:"reduced_fresh_killers_differ"`
	ReducedHistoryChanged          int      `json:"reduced_history_changed"`
	FreshFullHistoryChanged        int      `json:"fresh_full_history_changed"`
	ReducedFreshHistoryDiffers     int      `json:"reduced_fresh_history_differs"`
	D3Probes                       int      `json:"d3_probes"`
	D3D1Chains                     int      `json:"d3_d1_chains"`
	CasesWithD3D1Chains            int      `json:"cases_with_d3_d1_chains"`
	D3D1ChainCaseNames             []string `json:"d3_d1_chain_case_names"`
	ShadowNodes                    uint64   `json:"shadow_nodes"`
}

type generalD3D1Summary struct {
	Applied               bool     `json:"applied"`
	D5Applications        int      `json:"d5_applications"`
	D5Ablations           int      `json:"d5_ablations"`
	D3Applications        int      `json:"d3_applications"`
	D1Applications        int      `json:"d1_applications"`
	CasesWithD5           int      `json:"cases_with_d5"`
	CasesWithD5Ablation   int      `json:"cases_with_d5_ablation"`
	CasesWithD3           int      `json:"cases_with_d3"`
	CasesWithD1           int      `json:"cases_with_d1"`
	D5CaseNames           []string `json:"d5_case_names"`
	D5AblationCaseNames   []string `json:"d5_ablation_case_names"`
	D1CaseNames           []string `json:"d1_case_names"`
	UniqueD5Paths         int      `json:"unique_d5_paths"`
	UniqueD5AblationPaths int      `json:"unique_d5_ablation_paths"`
	UniqueD3Paths         int      `json:"unique_d3_paths"`
	UniqueD1Paths         int      `json:"unique_d1_paths"`
}

const lossRecognitionThresholdCP = -300

func main() {
	input := flag.String("input", "", "king-zone tactical CSV")
	output := flag.String("output", "", "JSON trace output")
	depthText := flag.String("depths", "4,6", "comma-separated fixed depths")
	ttSize := flag.Int("tt-size", 1<<16, "fresh transposition-table entries per replay")
	expectedPrimary := flag.Int("expected-primary", 19, "required primary quiet-check cases")
	expectedCanaries := flag.Int("expected-canaries", 13, "required exact canary cases")
	counterfactualMode := flag.String(
		"counterfactual-mode",
		"observational",
		"target reply mode: observational, full-depth, or plus-one",
	)
	counterfactualCaseText := flag.String(
		"counterfactual-cases",
		"",
		"comma-separated exact case names affected by a non-observational mode",
	)
	bridgeCase := flag.String(
		"bridge-case",
		"",
		"exact corpus case whose deeper bridge move is traced",
	)
	bridgeLineText := flag.String(
		"bridge-line",
		"",
		"comma-separated exact UCI line ending with the bridge move",
	)
	bridgeNoLMRMove := flag.String(
		"bridge-no-lmr-move",
		"",
		"exact intermediate quiet-check move exempted from LMR on the bridge path",
	)
	bridgeNoFutilityMove := flag.String(
		"bridge-no-futility-move",
		"",
		"exact final bridge quiet check exempted from an otherwise applicable futility prune",
	)
	probeD3CheckChains := flag.Bool(
		"probe-d3-check-chains",
		false,
		"run isolated full-depth shadows for all final-iteration depth-3 LMR quiet checks and record depth-1 futility-pruned quiet checks",
	)
	probeD5ReducedFull := flag.Bool(
		"probe-d5-reduced-full",
		false,
		"compare isolated reduced, fresh full-depth, and sequential reduced/full searches for every final-iteration depth-5 LMR quiet check",
	)
	generalD3D1Checks := flag.Bool(
		"general-d3-d1-checks",
		false,
		"coupled final-iteration counterfactual: full-depth D3 LMR quiet checks and no futility for their legal D1 quiet-check descendants",
	)
	generalD5D3D1Checks := flag.Bool(
		"general-d5-d3-d1-checks",
		false,
		"allowlist-free final-iteration counterfactual: full-depth D5 and D3 LMR quiet checks, with coupled no-futility for legal D1 quiet-check descendants of forced D3 checks",
	)
	d5AblationCase := flag.String(
		"d5-ablation-case",
		"",
		"exact corpus case in which one normally forced D5 quiet-check path keeps its baseline LMR",
	)
	d5AblationLineText := flag.String(
		"d5-ablation-line",
		"",
		"exact two-move UCI root/check path removed from the generalized D5 exemption",
	)
	flag.Parse()
	if !search.SearchTraceCompiled() {
		fatal(errors.New("binary must be built with -tags searchtrace"))
	}
	if *input == "" || *output == "" {
		fatal(errors.New("-input and -output are required"))
	}
	depths, err := parseDepths(*depthText)
	if err != nil {
		fatal(err)
	}
	cases, err := loadCorpus(*input)
	if err != nil {
		fatal(err)
	}
	primary, canaries := cohortCounts(cases)
	if primary != *expectedPrimary || canaries != *expectedCanaries {
		fatal(fmt.Errorf(
			"corpus mismatch: got primary=%d canaries=%d, want primary=%d canaries=%d",
			primary, canaries, *expectedPrimary, *expectedCanaries,
		))
	}
	counterfactualExtraPlies, err := parseCounterfactualMode(*counterfactualMode)
	if err != nil {
		fatal(err)
	}
	counterfactualCases, counterfactualCaseNames, err := parseCounterfactualCases(
		*counterfactualCaseText,
		*counterfactualMode,
		cases,
	)
	if err != nil {
		fatal(err)
	}
	bridgeLine, err := parseBridgeLineText(*bridgeLineText)
	if err != nil {
		fatal(err)
	}
	if err := validateBridgeSelection(
		*bridgeCase,
		bridgeLine,
		*counterfactualMode,
		counterfactualCases,
		cases,
	); err != nil {
		fatal(err)
	}
	if err := validateBridgeNoLMRSelection(*bridgeNoLMRMove, *bridgeCase, bridgeLine); err != nil {
		fatal(err)
	}
	if err := validateBridgeNoFutilitySelection(
		*bridgeNoFutilityMove,
		*bridgeNoLMRMove,
		*bridgeCase,
		bridgeLine,
	); err != nil {
		fatal(err)
	}
	if *probeD3CheckChains && (*counterfactualMode == "plus-one" || *bridgeCase != "") {
		fatal(errors.New("-probe-d3-check-chains supports observational or full-depth mode without a bridge case"))
	}
	if *generalD3D1Checks && (*counterfactualMode != "full-depth" ||
		len(counterfactualCaseNames) == 0 || *bridgeCase != "" || *probeD3CheckChains) {
		fatal(errors.New("-general-d3-d1-checks requires selected full-depth target prerequisites and cannot be combined with bridge or structural-probe modes"))
	}
	if err := validateGeneralD5D3D1Selection(
		*generalD5D3D1Checks,
		*counterfactualMode,
		counterfactualCaseNames,
		*bridgeCase,
		*probeD3CheckChains,
		*generalD3D1Checks,
	); err != nil {
		fatal(err)
	}
	d5AblationLine, err := parseD5AblationLineText(*d5AblationLineText)
	if err != nil {
		fatal(err)
	}
	if err := validateD5AblationSelection(
		*generalD5D3D1Checks,
		*d5AblationCase,
		d5AblationLine,
		cases,
	); err != nil {
		fatal(err)
	}
	if err := validateD5ReducedFullSelection(
		*probeD5ReducedFull,
		*counterfactualMode,
		counterfactualCaseNames,
		*bridgeCase,
		*probeD3CheckChains,
		*generalD3D1Checks,
		*generalD5D3D1Checks,
		*d5AblationCase,
		d5AblationLine,
	); err != nil {
		fatal(err)
	}

	for caseIndex := range cases {
		for _, depth := range depths {
			targeted := counterfactualCases[cases[caseIndex].Name]
			result, err := replay(
				&cases[caseIndex],
				depth,
				*ttSize,
				*counterfactualMode,
				targeted,
				counterfactualExtraPlies,
				cases[caseIndex].Name == *bridgeCase,
				bridgeLine,
				*bridgeNoLMRMove,
				*bridgeNoFutilityMove,
				*probeD3CheckChains,
				*probeD5ReducedFull,
				*generalD3D1Checks,
				*generalD5D3D1Checks,
				cases[caseIndex].Name == *d5AblationCase,
				d5AblationLine,
			)
			if err != nil {
				fatal(fmt.Errorf("%s depth %d: %w", cases[caseIndex].Name, depth, err))
			}
			cases[caseIndex].Runs = append(cases[caseIndex].Runs, result)
		}
	}

	inputHash, err := sha256File(*input)
	if err != nil {
		fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		fatal(err)
	}
	executableHash, err := sha256File(executable)
	if err != nil {
		fatal(err)
	}
	r := report{
		SchemaVersion: 11,
		Inputs: map[string]map[string]string{
			"csv":        {"path": *input, "sha256": inputHash},
			"executable": {"path": executable, "sha256": executableHash},
		},
		Parameters: map[string]any{
			"depths":                        depths,
			"tt_entries":                    *ttSize,
			"fresh_search_and_tt_per_run":   true,
			"non_interference_check":        "traced and untraced best move, score, completed depth, and nodes must match",
			"go_version":                    runtime.Version(),
			"primary_selection":             "midgame PRE non-check, no canary, severe, queen_net_danger > 0, invisible quiet-check reply",
			"canary_selection":              "every exact canary row",
			"bad_move_avoidance_definition": "best move differs from the historical Equinox root move",
			"loss_recognition_definition":   "root score from side-to-move POV is <= loss_recognition_threshold_cp",
			"loss_recognition_threshold_cp": lossRecognitionThresholdCP,
			"counterfactual_mode":           *counterfactualMode,
			"counterfactual_cases":          counterfactualCaseNames,
			"counterfactual_scope":          "only named RootMove/ReplyMove pairs at ply 1 of the requested final iteration that give check; full-depth requires current LMR eligibility, plus-one cases are baseline-LMR preselected and remain extended if their killer/order state changes causally",
			"continuation_definition":       "best explored continuation after the exact reply, with propagated qsearch/terminal horizon",
			"root_margin_definition":        "historical root move score minus best competitor in the decisive pass; alpha-beta score bounds are retained",
			"bridge_case":                   *bridgeCase,
			"bridge_line_uci":               bridgeLine,
			"bridge_trace_scope":            "exact path prefix and exact final bridge move only",
			"bridge_no_lmr_move_uci":        *bridgeNoLMRMove,
			"bridge_no_lmr_scope":           "only the unique named intermediate move after the exact bridge prefix, in the requested final iteration, if it is normally LMR-eligible and gives check; child depth changes from reduced to normal with no extension",
			"bridge_no_futility_move_uci":   *bridgeNoFutilityMove,
			"bridge_no_futility_scope":      "only the exact final bridge move after the complete path prefix, in the requested final iteration, and only when the unchanged depth-1 futility predicate would prune it; the validated move is a legal quiet check",
			"probe_d3_check_chains":         *probeD3CheckChains,
			"structural_probe_scope":        "every LMR-eligible quiet check at remaining depth 3 in the requested final iteration, after any selected exact target-reply full-depth prerequisite; each is replayed at normal child depth 2 in an isolated board/search/TT shadow, and legal quiet checks rejected by the unchanged depth-1 futility predicate are recorded",
			"probe_d5_reduced_full":         *probeD5ReducedFull,
			"d5_reduced_full_scope":         "every normally LMR-eligible quiet check at remaining depth 5 in the requested final iteration; from the exact same child board, TT, killers, and history snapshot, compare a reduced depth-3 null-window search, a fresh full depth-4 search, and the normal sequential reduced then conditional full re-search; record alpha crossings, TT/order state changes, PV lines, and D3-to-D1 structural chains found only in the fresh full shadow",
			"general_d3_d1_checks":          *generalD3D1Checks,
			"general_d3_d1_scope":           "coupled final-iteration counterfactual after selected exact target-reply full-depth prerequisites: every normally LMR-eligible quiet check at remaining depth 3 is searched at normal child depth 2; only its legal quiet-check descendants that meet the unchanged depth-1 futility predicate bypass that prune",
			"general_d5_d3_d1_checks":       *generalD5D3D1Checks,
			"general_d5_d3_d1_scope":        "allowlist-free final-iteration counterfactual: every normally LMR-eligible quiet check at remaining depths 5 and 3 is searched at normal child depths 4 and 2; only legal quiet-check descendants exactly two plies below a D3 check forced by this same candidate bypass the unchanged depth-1 futility predicate",
			"d5_ablation_case":              *d5AblationCase,
			"d5_ablation_line_uci":          d5AblationLine,
			"d5_ablation_scope":             "diagnostic-only removal of the generalized D5 exemption for one exact legal root/quiet-check path in one exact corpus case; the move retains normal LMR eligibility, reduction, null-window search, and possible full re-search while all D3-to-D1 coupling remains generalized",
		},
		Counts: map[string]int{
			"cases":         len(cases),
			"primary_cases": primary,
			"canary_cases":  canaries,
			"runs":          len(cases) * len(depths),
		},
		Cases:         cases,
		Aggregates:    buildAggregates(cases, depths),
		Structural:    buildStructuralProbeSummary(cases, *probeD3CheckChains),
		D5ReducedFull: buildD5ReducedFullSummary(cases, *probeD5ReducedFull),
		GeneralD3D1:   buildGeneralD3D1Summary(cases, *generalD3D1Checks || *generalD5D3D1Checks),
	}
	encoded, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("search_trace: %d cases (%d primary, %d canaries), %d runs, mode=%s targets=%d -> %s\n",
		len(cases), primary, canaries, len(cases)*len(depths), *counterfactualMode,
		len(counterfactualCaseNames), *output)
}

func validateGeneralD5D3D1Selection(
	applied bool,
	counterfactualMode string,
	counterfactualCaseNames []string,
	bridgeCase string,
	structuralProbeApplied bool,
	generalD3D1Applied bool,
) error {
	if !applied {
		return nil
	}
	if counterfactualMode != "observational" || len(counterfactualCaseNames) != 0 ||
		bridgeCase != "" || structuralProbeApplied || generalD3D1Applied {
		return errors.New("-general-d5-d3-d1-checks requires observational mode without a case allowlist, bridge, structural probe, or -general-d3-d1-checks")
	}
	return nil
}

func validateD5ReducedFullSelection(
	applied bool,
	counterfactualMode string,
	counterfactualCaseNames []string,
	bridgeCase string,
	structuralProbeApplied bool,
	generalD3D1Applied bool,
	generalD5D3D1Applied bool,
	d5AblationCase string,
	d5AblationLine []string,
) error {
	if !applied {
		return nil
	}
	if counterfactualMode != "observational" || len(counterfactualCaseNames) != 0 ||
		bridgeCase != "" || structuralProbeApplied || generalD3D1Applied ||
		generalD5D3D1Applied || d5AblationCase != "" || len(d5AblationLine) != 0 {
		return errors.New("-probe-d5-reduced-full requires observational mode without case allowlists, bridge, D3 probe, generalized candidates, or D5 ablation")
	}
	return nil
}

func parseD5AblationLineText(value string) ([]string, error) {
	line := []string{}
	for _, part := range strings.Split(value, ",") {
		moveText := strings.TrimSpace(part)
		if moveText != "" {
			line = append(line, moveText)
		}
	}
	if value != "" && len(line) == 0 {
		return nil, errors.New("D5 ablation line contains no moves")
	}
	return line, nil
}

func validateD5AblationSelection(
	generalD5D3D1Applied bool,
	ablationCase string,
	ablationLine []string,
	cases []corpusCase,
) error {
	if ablationCase == "" && len(ablationLine) == 0 {
		return nil
	}
	if !generalD5D3D1Applied {
		return errors.New("D5 ablation requires -general-d5-d3-d1-checks")
	}
	if ablationCase == "" || len(ablationLine) != 2 {
		return errors.New("D5 ablation requires one case and an exact two-move root/check line")
	}
	for _, tc := range cases {
		if tc.Name == ablationCase {
			return nil
		}
	}
	return fmt.Errorf("D5 ablation case %q is not in the selected corpus", ablationCase)
}

func parseAndValidateD5AblationLine(tc *corpusCase, lineUCI []string) ([]uint32, error) {
	if len(lineUCI) != 2 {
		return nil, errors.New("D5 ablation line must contain exactly one root move and one reply")
	}
	position, err := board.FromFEN(tc.FEN)
	if err != nil {
		return nil, fmt.Errorf("D5 ablation FromFEN: %w", err)
	}
	encoded := make([]uint32, 0, len(lineUCI))
	for index, moveText := range lineUCI {
		m, err := uci.ParseUCIMove(position, moveText)
		if err != nil {
			return nil, fmt.Errorf("D5 ablation move %d %s: %w", index+1, moveText, err)
		}
		position.MakeMove(m)
		if search.IsMovedSideInCheck(position) {
			return nil, fmt.Errorf("D5 ablation move %d %s is illegal", index+1, moveText)
		}
		encoded = append(encoded, m)
	}
	reply := encoded[1]
	if move.IsCapture(reply) || move.IsPromotion(reply) || !search.InCheck(position) {
		return nil, errors.New("D5 ablation reply must be a legal quiet check")
	}
	return encoded, nil
}

func loadCorpus(path string) ([]corpusCase, error) {
	handle, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	reader := csv.NewReader(handle)
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	indexes := make(map[string]int, len(headers))
	for index, header := range headers {
		indexes[header] = index
	}
	required := []string{
		"game_index", "ply", "fen", "move_uci", "move_san", "reply_move_uci",
		"reply_move_san", "severe", "non_pawn_material", "queen_net_danger",
		"king_in_check", "reply_is_quiet_check_qsearch_invisible", "canary_name",
	}
	for _, field := range required {
		if _, ok := indexes[field]; !ok {
			return nil, fmt.Errorf("missing CSV field %q", field)
		}
	}
	var cases []corpusCase
	for line := 2; ; line++ {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV line %d: %w", line, err)
		}
		field := func(name string) string { return record[indexes[name]] }
		gameIndex, err := strconv.Atoi(field("game_index"))
		if err != nil {
			return nil, fmt.Errorf("CSV line %d game_index: %w", line, err)
		}
		ply, err := strconv.Atoi(field("ply"))
		if err != nil {
			return nil, fmt.Errorf("CSV line %d ply: %w", line, err)
		}
		nonPawnMaterial, err := strconv.Atoi(field("non_pawn_material"))
		if err != nil {
			return nil, fmt.Errorf("CSV line %d non_pawn_material: %w", line, err)
		}
		queenNetDanger, err := strconv.Atoi(field("queen_net_danger"))
		if err != nil {
			return nil, fmt.Errorf("CSV line %d queen_net_danger: %w", line, err)
		}
		canary := field("canary_name")
		severe := field("severe") == "True"
		primary := canary == "" && nonPawnMaterial > 2000 &&
			field("king_in_check") == "False" && severe && queenNetDanger > 0 &&
			field("reply_is_quiet_check_qsearch_invisible") == "True"
		if !primary && canary == "" {
			continue
		}
		name := canary
		cohort := "canary"
		if primary {
			name = fmt.Sprintf("Game%02d_ply%d", gameIndex, ply)
			cohort = "primary_quiet_check"
		}
		cases = append(cases, corpusCase{
			Name:           name,
			Cohort:         cohort,
			GameIndex:      gameIndex,
			Ply:            ply,
			FEN:            field("fen"),
			RootMoveUCI:    field("move_uci"),
			RootMoveSAN:    field("move_san"),
			ReplyMoveUCI:   field("reply_move_uci"),
			ReplyMoveSAN:   field("reply_move_san"),
			Severe:         severe,
			QueenNetDanger: queenNetDanger,
		})
	}
	return cases, nil
}

func replay(
	tc *corpusCase,
	depth int,
	ttSize int,
	counterfactualMode string,
	counterfactualApplied bool,
	targetReplyExtraPlies int,
	bridgeTraceApplied bool,
	bridgeLineUCI []string,
	bridgeNoLMRMoveUCI string,
	bridgeNoFutilityMoveUCI string,
	structuralProbeApplied bool,
	d5ReducedFullProbeApplied bool,
	generalD3D1Applied bool,
	generalD5D3D1Applied bool,
	d5AblationApplied bool,
	d5AblationLineUCI []string,
) (run, error) {
	position, err := board.FromFEN(tc.FEN)
	if err != nil {
		return run{}, err
	}
	rootMove, err := uci.ParseUCIMove(position, tc.RootMoveUCI)
	if err != nil {
		return run{}, fmt.Errorf("parse root move: %w", err)
	}
	position.MakeMove(rootMove)
	if search.IsMovedSideInCheck(position) {
		return run{}, errors.New("root move is illegal")
	}
	replyMove, err := uci.ParseUCIMove(position, tc.ReplyMoveUCI)
	if err != nil {
		return run{}, fmt.Errorf("parse reply move: %w", err)
	}
	position.MakeMove(replyMove)
	if search.IsMovedSideInCheck(position) {
		return run{}, errors.New("reply move is illegal")
	}
	position.UnmakeMove()
	position.UnmakeMove()
	var bridgePath []uint32
	var bridgeMove uint32
	var bridgeLine []uint32
	bridgeNoLMRPly := 0
	if bridgeTraceApplied {
		bridgePath, bridgeMove, err = parseAndValidateBridgeLine(tc, bridgeLineUCI)
		if err != nil {
			return run{}, err
		}
		bridgeLine = append(append([]uint32(nil), bridgePath...), bridgeMove)
		if bridgeNoLMRMoveUCI != "" {
			bridgeNoLMRPly, err = validateBridgeNoLMRMove(tc, bridgeLineUCI, bridgeNoLMRMoveUCI)
			if err != nil {
				return run{}, err
			}
		}
		if err := validateBridgeNoFutilitySelection(
			bridgeNoFutilityMoveUCI,
			bridgeNoLMRMoveUCI,
			tc.Name,
			bridgeLineUCI,
		); err != nil {
			return run{}, err
		}
	}
	bridgeNoLMRApplied := bridgeTraceApplied && bridgeNoLMRMoveUCI != ""
	bridgeNoFutilityApplied := bridgeTraceApplied && bridgeNoFutilityMoveUCI != ""
	var d5AblationPath []uint32
	if d5AblationApplied {
		d5AblationPath, err = parseAndValidateD5AblationLine(tc, d5AblationLineUCI)
		if err != nil {
			return run{}, err
		}
	}
	d5AblationPaths := [][]uint32(nil)
	if len(d5AblationPath) != 0 {
		d5AblationPaths = [][]uint32{d5AblationPath}
	}

	trace := &search.SearchTrace{
		RootMove:                  rootMove,
		ReplyMove:                 replyMove,
		ForceTargetReplyFullDepth: counterfactualApplied,
		TargetReplyExtraPlies:     targetReplyExtraPlies,
		ForceTargetReplyIteration: depth,
		ForceBridgeStepNoLMR:      bridgeNoLMRApplied,
		BridgeNoLMRPly:            bridgeNoLMRPly,
		ForceBridgeNoLMRIteration: depth,
		ForceBridgeNoFutility:     bridgeNoFutilityApplied,
		ForceBridgeNoFutilityIter: depth,
		ProbeD3CheckChains:        structuralProbeApplied,
		ProbeD3CheckIteration:     depth,
		ProbeD5ReducedFull:        d5ReducedFullProbeApplied,
		ProbeD5Iteration:          depth,
		ForceGeneralD3D1Checks:    generalD3D1Applied,
		ForceGeneralD5D3D1Checks:  generalD5D3D1Applied,
		ForceGeneralD3D1Iteration: depth,
		GeneralD5AblationPaths:    d5AblationPaths,
		BridgePath:                bridgePath,
		BridgeMove:                bridgeMove,
		BridgeLine:                bridgeLine,
	}
	worker := search.Search{TT: tt.NewTT(ttSize), Trace: trace}
	best, score, completedDepth, nodes := worker.SearchRoot(position, depth)
	verification := "observational trace equals untraced search"
	nonInterferenceVerified := false
	counterfactualDeterminismVerified := false
	verificationPosition, err := board.FromFEN(tc.FEN)
	if err != nil {
		return run{}, fmt.Errorf("verification FromFEN: %w", err)
	}
	d5ShadowDeterminismVerified := false
	if d5ReducedFullProbeApplied {
		verification = "two independent D5 reduced/full shadows match exactly and traced search equals untraced search"
		verificationTrace := &search.SearchTrace{
			RootMove:           rootMove,
			ReplyMove:          replyMove,
			ProbeD5ReducedFull: true,
			ProbeD5Iteration:   depth,
		}
		verificationWorker := search.Search{TT: tt.NewTT(ttSize), Trace: verificationTrace}
		verificationBest, verificationScore, verificationDepth, verificationNodes :=
			verificationWorker.SearchRoot(verificationPosition, depth)
		if best != verificationBest || score != verificationScore || completedDepth != verificationDepth ||
			nodes != verificationNodes ||
			!reflect.DeepEqual(trace.D5ReducedFullProbes, verificationTrace.D5ReducedFullProbes) {
			return run{}, fmt.Errorf(
				"D5 shadow nondeterminism: first=(%s,%d,%d,%d,%d probes) second=(%s,%d,%d,%d,%d probes)",
				moveUCI(best), score, completedDepth, nodes, len(trace.D5ReducedFullProbes),
				moveUCI(verificationBest), verificationScore, verificationDepth, verificationNodes,
				len(verificationTrace.D5ReducedFullProbes),
			)
		}
		plainPosition, plainErr := board.FromFEN(tc.FEN)
		if plainErr != nil {
			return run{}, fmt.Errorf("plain verification FromFEN: %w", plainErr)
		}
		plain := search.Search{TT: tt.NewTT(ttSize)}
		plainBest, plainScore, plainDepth, plainNodes := plain.SearchRoot(plainPosition, depth)
		if best != plainBest || score != plainScore || completedDepth != plainDepth || nodes != plainNodes {
			return run{}, fmt.Errorf(
				"D5 trace interference: traced=(%s,%d,%d,%d) plain=(%s,%d,%d,%d)",
				moveUCI(best), score, completedDepth, nodes,
				moveUCI(plainBest), plainScore, plainDepth, plainNodes,
			)
		}
		d5ShadowDeterminismVerified = true
		nonInterferenceVerified = true
	} else if counterfactualApplied || generalD3D1Applied || generalD5D3D1Applied {
		verification = "two independent targeted counterfactual traces match exactly"
		verificationTrace := &search.SearchTrace{
			RootMove:                  rootMove,
			ReplyMove:                 replyMove,
			ForceTargetReplyFullDepth: counterfactualApplied,
			TargetReplyExtraPlies:     targetReplyExtraPlies,
			ForceTargetReplyIteration: depth,
			ForceBridgeStepNoLMR:      bridgeNoLMRApplied,
			BridgeNoLMRPly:            bridgeNoLMRPly,
			ForceBridgeNoLMRIteration: depth,
			ForceBridgeNoFutility:     bridgeNoFutilityApplied,
			ForceBridgeNoFutilityIter: depth,
			ProbeD3CheckChains:        structuralProbeApplied,
			ProbeD3CheckIteration:     depth,
			ProbeD5ReducedFull:        d5ReducedFullProbeApplied,
			ProbeD5Iteration:          depth,
			ForceGeneralD3D1Checks:    generalD3D1Applied,
			ForceGeneralD5D3D1Checks:  generalD5D3D1Applied,
			ForceGeneralD3D1Iteration: depth,
			GeneralD5AblationPaths:    cloneMoveLines(d5AblationPaths),
			BridgePath:                bridgePath,
			BridgeMove:                bridgeMove,
			BridgeLine:                bridgeLine,
		}
		verificationWorker := search.Search{TT: tt.NewTT(ttSize), Trace: verificationTrace}
		verificationBest, verificationScore, verificationDepth, verificationNodes :=
			verificationWorker.SearchRoot(verificationPosition, depth)
		if best != verificationBest || score != verificationScore || completedDepth != verificationDepth ||
			nodes != verificationNodes || !reflect.DeepEqual(trace.Continuations, verificationTrace.Continuations) ||
			!reflect.DeepEqual(trace.RootMoveResults, verificationTrace.RootMoveResults) ||
			!reflect.DeepEqual(trace.Events, verificationTrace.Events) ||
			!reflect.DeepEqual(trace.D3CheckProbes, verificationTrace.D3CheckProbes) ||
			!reflect.DeepEqual(trace.D3CheckFutilityChains, verificationTrace.D3CheckFutilityChains) ||
			!reflect.DeepEqual(trace.D5ReducedFullProbes, verificationTrace.D5ReducedFullProbes) ||
			!reflect.DeepEqual(trace.GeneralD3D1Events, verificationTrace.GeneralD3D1Events) {
			return run{}, fmt.Errorf(
				"counterfactual nondeterminism: first=(%s,%d,%d,%d) second=(%s,%d,%d,%d)",
				moveUCI(best), score, completedDepth, nodes,
				moveUCI(verificationBest), verificationScore, verificationDepth, verificationNodes,
			)
		}
		counterfactualDeterminismVerified = true
	} else {
		plain := search.Search{TT: tt.NewTT(ttSize)}
		plainBest, plainScore, plainDepth, plainNodes := plain.SearchRoot(verificationPosition, depth)
		if best != plainBest || score != plainScore || completedDepth != plainDepth || nodes != plainNodes {
			return run{}, fmt.Errorf(
				"trace interference: traced=(%s,%d,%d,%d) plain=(%s,%d,%d,%d)",
				moveUCI(best), score, completedDepth, nodes,
				moveUCI(plainBest), plainScore, plainDepth, plainNodes,
			)
		}
		nonInterferenceVerified = true
	}
	events := make([]traceEventOutput, 0, len(trace.Events))
	for _, event := range trace.Events {
		events = append(events, traceEventOutput{
			TraceEvent:     event,
			TargetMoveUCI:  moveUCI(event.TargetMove),
			CurrentMoveUCI: moveUCI(event.CurrentMove),
		})
	}
	d3CheckProbes := make([]d3CheckProbeOutput, 0, len(trace.D3CheckProbes))
	var structuralShadowNodes uint64
	for _, probe := range trace.D3CheckProbes {
		structuralShadowNodes += probe.ShadowNodes
		d3CheckProbes = append(d3CheckProbes, d3CheckProbeOutput{
			TraceD3CheckProbe: probe,
			MoveUCI:           moveUCI(probe.Move),
			LineUCI:           formatMoveLine(probe.Line),
		})
	}
	d3CheckChains := make([]d3CheckChainOutput, 0, len(trace.D3CheckFutilityChains))
	for _, chain := range trace.D3CheckFutilityChains {
		d3CheckChains = append(d3CheckChains, d3CheckChainOutput{
			TraceD3CheckFutilityChain: chain,
			D3MoveUCI:                 moveUCI(chain.D3Move),
			D1MoveUCI:                 moveUCI(chain.D1Move),
			LineUCI:                   formatMoveLine(chain.Line),
		})
	}
	d5ReducedFullProbes := make([]d5ReducedFullOutput, 0, len(trace.D5ReducedFullProbes))
	var d5ReducedFullShadowNodes uint64
	for _, probe := range trace.D5ReducedFullProbes {
		d3LinesUCI := make([][]string, 0, len(probe.D3Lines))
		for _, line := range probe.D3Lines {
			d3LinesUCI = append(d3LinesUCI, formatMoveLine(line))
		}
		d3D1LinesUCI := make([][]string, 0, len(probe.D3D1Lines))
		for _, line := range probe.D3D1Lines {
			d3D1LinesUCI = append(d3D1LinesUCI, formatMoveLine(line))
		}
		d5ReducedFullShadowNodes += probe.SequentialReducedNodes + probe.SequentialFullNodes +
			probe.FreshFullNodes + probe.D3ShadowNodes
		d5ReducedFullProbes = append(d5ReducedFullProbes, d5ReducedFullOutput{
			TraceD5ReducedFullProbe:         probe,
			MoveUCI:                         moveUCI(probe.Move),
			ParentTTMoveUCI:                 moveUCI(probe.ParentTTMove),
			ParentPVMoveUCI:                 moveUCI(probe.ParentPVMove),
			ChildTTMoveUCI:                  moveUCI(probe.ChildTTMove),
			InitialFirstOrderedMoveUCI:      moveUCI(probe.InitialFirstOrderedMove),
			AfterReducedTTMoveUCI:           moveUCI(probe.AfterReducedTTMove),
			AfterReducedFirstOrderedMoveUCI: moveUCI(probe.AfterReducedFirstOrderedMove),
			FreshFullTTMoveUCI:              moveUCI(probe.FreshFullTTMove),
			FreshFullFirstOrderedMoveUCI:    moveUCI(probe.FreshFullFirstOrderedMove),
			LineUCI:                         formatMoveLine(probe.Line),
			ReducedLineUCI:                  formatMoveLine(probe.ReducedLine),
			FreshFullLineUCI:                formatMoveLine(probe.FreshFullLine),
			SequentialLineUCI:               formatMoveLine(probe.SequentialLine),
			D3LinesUCI:                      d3LinesUCI,
			D3D1LinesUCI:                    d3D1LinesUCI,
		})
	}
	generalD3D1Events := make([]generalD3D1Output, 0, len(trace.GeneralD3D1Events))
	for _, event := range trace.GeneralD3D1Events {
		generalD3D1Events = append(generalD3D1Events, generalD3D1Output{
			TraceGeneralD3D1Event: event,
			MoveUCI:               moveUCI(event.Move),
			LineUCI:               formatMoveLine(event.Line),
		})
	}
	continuations := make([]continuationOutput, 0, len(trace.Continuations))
	for _, continuation := range trace.Continuations {
		continuations = append(continuations, formatContinuation(continuation))
	}
	summary := summarize(trace.Events, depth)
	finalContinuation := selectFinalContinuation(continuations, depth, summary.FinalPass)
	finalRootMoveResults := selectFinalRootMoveResults(trace.RootMoveResults, depth, summary.FinalPass)
	rootMargin := buildRootMargin(finalRootMoveResults, tc.RootMoveUCI, moveUCI(best))
	d5AblationOutputLine := []string(nil)
	if d5AblationApplied {
		d5AblationOutputLine = append(d5AblationOutputLine, d5AblationLineUCI...)
	}
	return run{
		RequestedDepth:                    depth,
		CompletedDepth:                    completedDepth,
		BestMoveUCI:                       moveUCI(best),
		Score:                             score,
		Nodes:                             nodes,
		RootMoveBest:                      best == rootMove,
		CounterfactualMode:                counterfactualMode,
		CounterfactualApplied:             counterfactualApplied,
		BridgeTraceApplied:                bridgeTraceApplied,
		BridgeNoLMRApplied:                bridgeNoLMRApplied,
		BridgeNoFutilityApplied:           bridgeNoFutilityApplied,
		StructuralProbeApplied:            structuralProbeApplied,
		StructuralProbeShadowNodes:        structuralShadowNodes,
		D5ReducedFullProbeApplied:         d5ReducedFullProbeApplied,
		D5ReducedFullShadowNodes:          d5ReducedFullShadowNodes,
		D5ShadowDeterminismVerified:       d5ShadowDeterminismVerified,
		GeneralD3D1Applied:                generalD3D1Applied,
		GeneralD5D3D1Applied:              generalD5D3D1Applied,
		GeneralD5AblationApplied:          d5AblationApplied,
		GeneralD5AblationLineUCI:          d5AblationOutputLine,
		Verification:                      verification,
		NonInterferenceVerified:           nonInterferenceVerified,
		CounterfactualDeterminismVerified: counterfactualDeterminismVerified,
		Summary:                           summary,
		Continuations:                     continuations,
		FinalContinuation:                 finalContinuation,
		FinalRootMoveResults:              finalRootMoveResults,
		RootMargin:                        rootMargin,
		Events:                            events,
		D3CheckProbes:                     d3CheckProbes,
		D3CheckFutilityChains:             d3CheckChains,
		D5ReducedFullProbes:               d5ReducedFullProbes,
		GeneralD3D1Events:                 generalD3D1Events,
	}, nil
}

func selectFinalRootMoveResults(
	results []search.TraceRootMoveResult,
	iteration int,
	pass string,
) []rootMoveOutput {
	selected := []rootMoveOutput{}
	for _, result := range results {
		if result.Iteration == iteration && result.Pass == pass {
			selected = append(selected, rootMoveOutput{
				TraceRootMoveResult: result,
				MoveUCI:             moveUCI(result.Move),
			})
		}
	}
	return selected
}

func buildRootMargin(results []rootMoveOutput, historicalMoveUCI, bestMoveUCI string) *rootMarginOutput {
	var historical *rootMoveOutput
	for index := range results {
		if results[index].MoveUCI == historicalMoveUCI {
			historical = &results[index]
			break
		}
	}
	if historical == nil {
		return nil
	}

	var competitor *rootMoveOutput
	if bestMoveUCI != historicalMoveUCI {
		for index := range results {
			if results[index].MoveUCI == bestMoveUCI {
				competitor = &results[index]
				break
			}
		}
	} else {
		for index := range results {
			candidate := &results[index]
			if candidate.MoveUCI == historicalMoveUCI {
				continue
			}
			if competitor == nil || candidate.Score > competitor.Score {
				competitor = candidate
			}
		}
	}
	if competitor == nil {
		return nil
	}

	marginBound := scoreDifferenceBound(historical.ScoreBound, competitor.ScoreBound)
	boundMeaning := "recorded score difference is not a formal bound"
	switch marginBound {
	case "exact":
		boundMeaning = "historical-minus-competitor margin is exact"
	case "lower":
		boundMeaning = "true historical-minus-competitor margin is at least the reported value"
	case "upper":
		boundMeaning = "true historical-minus-competitor margin is at most the reported value"
	}
	return &rootMarginOutput{
		HistoricalMove:                  *historical,
		BestCompetitor:                  *competitor,
		HistoricalMinusBestCompetitorCP: historical.Score - competitor.Score,
		MarginBound:                     marginBound,
		BoundMeaning:                    boundMeaning,
	}
}

func scoreDifferenceBound(leftBound, rightBound string) string {
	leftHasLower := leftBound == "exact" || leftBound == "lower"
	leftHasUpper := leftBound == "exact" || leftBound == "upper"
	rightHasLower := rightBound == "exact" || rightBound == "lower"
	rightHasUpper := rightBound == "exact" || rightBound == "upper"
	hasLower := leftHasLower && rightHasUpper
	hasUpper := leftHasUpper && rightHasLower
	if hasLower && hasUpper {
		return "exact"
	}
	if hasLower {
		return "lower"
	}
	if hasUpper {
		return "upper"
	}
	return "indeterminate"
}

func formatMoveLine(line []uint32) []string {
	result := make([]string, 0, len(line))
	for _, m := range line {
		result = append(result, moveUCI(m))
	}
	return result
}

func cloneMoveLines(lines [][]uint32) [][]uint32 {
	result := make([][]uint32, 0, len(lines))
	for _, line := range lines {
		result = append(result, append([]uint32(nil), line...))
	}
	return result
}

func buildStructuralProbeSummary(cases []corpusCase, applied bool) structuralProbeSummary {
	result := structuralProbeSummary{Applied: applied}
	uniqueD3 := map[string]bool{}
	uniqueChains := map[string]bool{}
	d3Cases := map[string]bool{}
	chainCases := map[string]bool{}
	for _, tc := range cases {
		for _, replay := range tc.Runs {
			result.ShadowNodes += replay.StructuralProbeShadowNodes
			result.D3ProbeOccurrences += len(replay.D3CheckProbes)
			result.ChainOccurrences += len(replay.D3CheckFutilityChains)
			if len(replay.D3CheckProbes) > 0 {
				d3Cases[tc.Name] = true
			}
			if len(replay.D3CheckFutilityChains) > 0 {
				chainCases[tc.Name] = true
			}
			for _, probe := range replay.D3CheckProbes {
				uniqueD3[tc.Name+"|"+strings.Join(probe.LineUCI, ",")] = true
			}
			for _, chain := range replay.D3CheckFutilityChains {
				uniqueChains[tc.Name+"|"+strings.Join(chain.LineUCI, ",")] = true
			}
		}
	}
	result.UniqueD3Paths = len(uniqueD3)
	result.UniqueChains = len(uniqueChains)
	result.CasesWithD3Probes = len(d3Cases)
	result.CasesWithChains = len(chainCases)
	for name := range chainCases {
		result.ChainCaseNames = append(result.ChainCaseNames, name)
	}
	sort.Strings(result.ChainCaseNames)
	return result
}

func buildD5ReducedFullSummary(cases []corpusCase, applied bool) d5ReducedFullSummary {
	result := d5ReducedFullSummary{Applied: applied}
	uniquePaths := map[string]bool{}
	probeCases := map[string]bool{}
	chainCases := map[string]bool{}
	for _, tc := range cases {
		for _, replay := range tc.Runs {
			result.ShadowNodes += replay.D5ReducedFullShadowNodes
			for _, probe := range replay.D5ReducedFullProbes {
				result.Occurrences++
				probeCases[tc.Name] = true
				uniquePaths[tc.Name+"|"+strings.Join(probe.LineUCI, ",")] = true
				if probe.ReducedExceedsAlpha {
					result.ReducedCrossesAlpha++
				}
				if probe.FreshFullExceedsAlpha {
					result.FreshFullCrossesAlpha++
				}
				if probe.AlphaCrossingDisagrees {
					result.AlphaCrossingDisagreements++
				}
				if probe.FullMinusReduced != 0 {
					result.FullScoreDifferences++
				}
				if probe.SequentialDidFullResearch {
					result.SequentialFullResearches++
				}
				if probe.SequentialMinusFreshFull != 0 {
					result.SequentialFreshDifferences++
				}
				if probe.ParentMoveIsTT {
					result.ParentTTMoves++
				}
				if probe.ParentMoveIsPV {
					result.ParentPVMoves++
				}
				if probe.ChildTTKeyMatch {
					result.InitialChildTTHits++
				}
				if probe.ChildTTReducedCutoff {
					result.InitialReducedTTCutoffs++
				}
				if probe.ChildTTFullCutoff {
					result.InitialFullTTCutoffs++
				}
				if probe.ChildTTEntryChanged {
					result.ChildTTEntriesChanged++
				}
				if probe.FreshFullChildTTEntryChanged {
					result.FreshFullChildTTEntriesChanged++
				}
				if probe.ReducedFreshChildTTDiffers {
					result.ReducedFreshChildTTDifferences++
				}
				result.ReducedTTEntriesChanged += probe.ReducedTTEntriesChanged
				result.FreshFullTTEntriesChanged += probe.FreshFullTTEntriesChanged
				result.ReducedFreshTTEntriesDiffer += probe.ReducedFreshTTEntriesDiffer
				if probe.FirstOrderedMoveChanged {
					result.FirstOrderedMovesChanged++
				}
				if probe.FreshFirstOrderedMoveChanged {
					result.FreshFirstOrderedMovesChanged++
				}
				if probe.ReducedFreshOrderDiffers {
					result.ReducedFreshOrderDifferences++
				}
				if probe.ReducedKillersChanged {
					result.ReducedKillersChanged++
				}
				if probe.FreshFullKillersChanged {
					result.FreshFullKillersChanged++
				}
				if probe.ReducedFreshKillersDiffer {
					result.ReducedFreshKillersDiffer++
				}
				if probe.ReducedHistoryChanged {
					result.ReducedHistoryChanged++
				}
				if probe.FreshFullHistoryChanged {
					result.FreshFullHistoryChanged++
				}
				if probe.ReducedFreshHistoryDiffers {
					result.ReducedFreshHistoryDiffers++
				}
				result.D3Probes += probe.D3ProbeCount
				result.D3D1Chains += probe.D3D1ChainCount
				if probe.D3D1ChainCount > 0 {
					chainCases[tc.Name] = true
				}
			}
		}
	}
	result.UniquePaths = len(uniquePaths)
	result.CasesWithProbes = len(probeCases)
	result.CasesWithD3D1Chains = len(chainCases)
	for name := range chainCases {
		result.D3D1ChainCaseNames = append(result.D3D1ChainCaseNames, name)
	}
	sort.Strings(result.D3D1ChainCaseNames)
	return result
}

func buildGeneralD3D1Summary(cases []corpusCase, applied bool) generalD3D1Summary {
	result := generalD3D1Summary{Applied: applied}
	d5Cases := map[string]bool{}
	d5AblationCases := map[string]bool{}
	d3Cases := map[string]bool{}
	d1Cases := map[string]bool{}
	uniqueD5 := map[string]bool{}
	uniqueD5Ablations := map[string]bool{}
	uniqueD3 := map[string]bool{}
	uniqueD1 := map[string]bool{}
	for _, tc := range cases {
		for _, replay := range tc.Runs {
			for _, event := range replay.GeneralD3D1Events {
				key := tc.Name + "|" + strings.Join(event.LineUCI, ",")
				switch event.Stage {
				case "d5_check_forced_full_depth":
					result.D5Applications++
					d5Cases[tc.Name] = true
					uniqueD5[key] = true
				case "d5_check_ablation_lmr_preserved":
					result.D5Ablations++
					d5AblationCases[tc.Name] = true
					uniqueD5Ablations[key] = true
				case "d3_check_forced_full_depth":
					result.D3Applications++
					d3Cases[tc.Name] = true
					uniqueD3[key] = true
				case "d1_check_forced_no_futility":
					result.D1Applications++
					d1Cases[tc.Name] = true
					uniqueD1[key] = true
				}
			}
		}
	}
	result.CasesWithD5 = len(d5Cases)
	result.CasesWithD5Ablation = len(d5AblationCases)
	result.CasesWithD3 = len(d3Cases)
	result.CasesWithD1 = len(d1Cases)
	result.UniqueD5Paths = len(uniqueD5)
	result.UniqueD5AblationPaths = len(uniqueD5Ablations)
	result.UniqueD3Paths = len(uniqueD3)
	result.UniqueD1Paths = len(uniqueD1)
	for name := range d5Cases {
		result.D5CaseNames = append(result.D5CaseNames, name)
	}
	for name := range d5AblationCases {
		result.D5AblationCaseNames = append(result.D5AblationCaseNames, name)
	}
	for name := range d1Cases {
		result.D1CaseNames = append(result.D1CaseNames, name)
	}
	sort.Strings(result.D5CaseNames)
	sort.Strings(result.D5AblationCaseNames)
	sort.Strings(result.D1CaseNames)
	return result
}

func formatContinuation(continuation search.TraceContinuation) continuationOutput {
	lineUCI := formatMoveLine(continuation.Line)
	afterReply := []string{}
	if len(lineUCI) > 2 {
		afterReply = append(afterReply, lineUCI[2:]...)
	}
	leafScoreRootPOV := continuation.LeafScore
	if continuation.LeafPly >= 0 && continuation.LeafPly%2 != 0 {
		leafScoreRootPOV = -leafScoreRootPOV
	}
	leafAfterReplyPly := -1
	if continuation.LeafPly >= 0 {
		leafAfterReplyPly = continuation.LeafPly - 2
	}
	return continuationOutput{
		TraceContinuation: continuation,
		LineUCI:           lineUCI,
		AfterReplyUCI:     afterReply,
		ScoreRootPOV:      -continuation.Score,
		LeafAfterReplyPly: leafAfterReplyPly,
		LeafScoreRootPOV:  leafScoreRootPOV,
	}
}

func selectFinalContinuation(
	continuations []continuationOutput,
	iteration int,
	pass string,
) *continuationOutput {
	var selected *continuationOutput
	for index := range continuations {
		continuation := &continuations[index]
		if continuation.Iteration == iteration && continuation.Pass == pass {
			selected = continuation
		}
	}
	return selected
}

func summarize(events []search.TraceEvent, finalIteration int) runSummary {
	finalPass := ""
	for _, event := range events {
		if event.Iteration == finalIteration && event.Pass != "" {
			finalPass = event.Pass
		}
	}
	return runSummary{
		RootAllIterations:           summarizeRole(events, "root_move", 0, ""),
		ReplyAllIterations:          summarizeRole(events, "reply_move", 0, ""),
		RootFinalIteration:          summarizeRole(events, "root_move", finalIteration, ""),
		ReplyFinalIteration:         summarizeRole(events, "reply_move", finalIteration, ""),
		FinalPass:                   finalPass,
		RootFinalPass:               summarizeRole(events, "root_move", finalIteration, finalPass),
		ReplyFinalPass:              summarizeRole(events, "reply_move", finalIteration, finalPass),
		BridgeAllIterations:         summarizeRole(events, "bridge_move", 0, ""),
		BridgeFinalIteration:        summarizeRole(events, "bridge_move", finalIteration, ""),
		BridgeFinalPass:             summarizeRole(events, "bridge_move", finalIteration, finalPass),
		BridgeStepsAllIterations:    summarizeRole(events, "bridge_step", 0, ""),
		BridgeStepsFinalIteration:   summarizeRole(events, "bridge_step", finalIteration, ""),
		BridgeStepsFinalPass:        summarizeRole(events, "bridge_step", finalIteration, finalPass),
		TargetLineQEntriesAll:       countEvents(events, "target_line", "enter_after_reply", 0, ""),
		TargetLineQEntriesFinal:     countEvents(events, "target_line", "enter_after_reply", finalIteration, ""),
		TargetLineQEntriesFinalPass: countEvents(events, "target_line", "enter_after_reply", finalIteration, finalPass),
	}
}

func summarizeRole(events []search.TraceEvent, role string, iteration int, pass string) roleOutcome {
	result := roleOutcome{StageCounts: map[string]int{}}
	pseudo := map[int]bool{}
	ordered := map[int]bool{}
	legal := map[int]bool{}
	for _, event := range events {
		if event.Role != role || (iteration != 0 && event.Iteration != iteration) ||
			(pass != "" && event.Pass != pass) {
			continue
		}
		result.Events++
		result.StageCounts[event.Node+"."+event.Stage]++
		result.TargetAsTT = result.TargetAsTT || event.TargetIsTT
		result.TargetAsPV = result.TargetAsPV || event.TargetIsPV
		result.TargetAsKiller = result.TargetAsKiller || event.TargetIsKiller
		if event.PseudoIndex >= 0 {
			pseudo[event.PseudoIndex] = true
		}
		if event.OrderIndex >= 0 {
			ordered[event.OrderIndex] = true
		}
		if event.LegalIndex >= 0 {
			legal[event.LegalIndex] = true
		}
	}
	result.PseudoIndexes = sortedKeys(pseudo)
	result.OrderIndexes = sortedKeys(ordered)
	result.LegalIndexes = sortedKeys(legal)
	return result
}

func countEvents(events []search.TraceEvent, role, stage string, iteration int, pass string) int {
	count := 0
	for _, event := range events {
		if event.Role == role && event.Stage == stage &&
			(iteration == 0 || event.Iteration == iteration) && (pass == "" || event.Pass == pass) {
			count++
		}
	}
	return count
}

func buildAggregates(cases []corpusCase, depths []int) map[string]map[string]aggregate {
	result := map[string]map[string]aggregate{}
	for _, depth := range depths {
		depthKey := strconv.Itoa(depth)
		result[depthKey] = map[string]aggregate{}
		for _, cohort := range []string{"all", "primary_quiet_check", "canary"} {
			entry := aggregate{Metrics: map[string]int{}}
			for _, tc := range cases {
				if cohort != "all" && tc.Cohort != cohort {
					continue
				}
				for _, replay := range tc.Runs {
					if replay.RequestedDepth != depth {
						continue
					}
					entry.Cases++
					if replay.RootMoveBest {
						entry.RootMoveBest++
						entry.HistoricalBadMoveSelected++
					} else {
						entry.BadMoveAvoided++
					}
					if replay.Score <= lossRecognitionThresholdCP {
						entry.LossRecognized++
					}
					addMetric := func(name string, condition bool) {
						if condition {
							entry.Metrics[name]++
						}
					}
					rootFinal := replay.Summary.RootFinalPass.StageCounts
					replyFinal := replay.Summary.ReplyFinalPass.StageCounts
					bridgeFinal := replay.Summary.BridgeFinalPass.StageCounts
					replyAll := replay.Summary.ReplyAllIterations.StageCounts
					addMetric("root_final_generated", rootFinal["negamax.generated"] > 0)
					addMetric("root_final_searched", rootFinal["negamax.search_complete"] > 0)
					addMetric("root_final_futility_pruned", rootFinal["negamax.futility_pruned"] > 0)
					addMetric("root_final_lmr_reduced", rootFinal["negamax.lmr_reduced"] > 0)
					addMetric("root_final_lmr_full_research", rootFinal["negamax.lmr_full_research"] > 0)
					addMetric("root_final_beta_cutoff_before", rootFinal["negamax.beta_cutoff_before_target"] > 0)
					addMetric("root_final_target_tt", replay.Summary.RootFinalPass.TargetAsTT)
					addMetric("root_final_target_pv", replay.Summary.RootFinalPass.TargetAsPV)
					addMetric("root_final_target_killer", replay.Summary.RootFinalPass.TargetAsKiller)
					addMetric("reply_final_negamax_enter", replyFinal["negamax.enter"] > 0)
					addMetric("reply_final_generated_main", replyFinal["negamax.generated"] > 0)
					addMetric("reply_final_ordered_main", replyFinal["negamax.ordered"] > 0)
					addMetric("reply_final_legal_main", replyFinal["negamax.legal"] > 0)
					addMetric("reply_final_searched_main", replyFinal["negamax.search_complete"] > 0)
					addMetric("reply_final_full_depth", replyFinal["negamax.full_depth"] > 0)
					addMetric("reply_final_lmr_reduced", replyFinal["negamax.lmr_reduced"] > 0)
					addMetric("reply_final_lmr_full_research", replyFinal["negamax.lmr_full_research"] > 0)
					addMetric("reply_final_futility_pruned", replyFinal["negamax.futility_pruned"] > 0)
					addMetric("reply_final_tt_cutoff", replyFinal["negamax.tt_cutoff"] > 0)
					addMetric("reply_final_null_cutoff", replyFinal["negamax.null_cutoff"] > 0)
					addMetric("reply_final_beta_cutoff_before", replyFinal["negamax.beta_cutoff_before_target"] > 0)
					addMetric("reply_final_target_tt", replay.Summary.ReplyFinalPass.TargetAsTT)
					addMetric("reply_final_target_killer", replay.Summary.ReplyFinalPass.TargetAsKiller)
					addMetric("bridge_final_generated", bridgeFinal["negamax.generated"] > 0)
					addMetric("bridge_final_ordered", bridgeFinal["negamax.ordered"] > 0)
					addMetric("bridge_final_futility_pruned", bridgeFinal["negamax.futility_pruned"] > 0)
					addMetric("bridge_final_legal", bridgeFinal["negamax.legal"] > 0)
					addMetric("bridge_final_searched", bridgeFinal["negamax.search_complete"] > 0)
					addMetric("bridge_final_beta_cutoff_before", bridgeFinal["negamax.beta_cutoff_before_target"] > 0)
					addMetric("reply_ever_qsearch_generated", replyAll["quiescence.generated"] > 0)
					addMetric("reply_ever_qsearch_not_generated", replyAll["quiescence.not_generated"] > 0)
					addMetric("reply_ever_qsearch_stand_pat_cutoff", replyAll["quiescence.stand_pat_cutoff"] > 0)
					addMetric("reply_ever_qsearch_beta_cutoff_before", replyAll["quiescence.beta_cutoff_before_target"] > 0)
					addMetric("reply_ever_qsearch_capture_filtered", replyAll["quiescence.capture_filtered"] > 0)
					addMetric("reply_ever_qsearch_delta_pruned", replyAll["quiescence.delta_pruned"] > 0)
					addMetric("reply_ever_qsearch_searched", replyAll["quiescence.search_complete"] > 0)
					addMetric("target_line_qsearch_entered_final", replay.Summary.TargetLineQEntriesFinalPass > 0)
				}
			}
			result[depthKey][cohort] = entry
		}
	}
	return result
}

func parseDepths(value string) ([]int, error) {
	seen := map[int]bool{}
	var depths []int
	for _, part := range strings.Split(value, ",") {
		depth, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || depth <= 0 || depth >= search.MaxDepth {
			return nil, fmt.Errorf("invalid depth %q", part)
		}
		if !seen[depth] {
			seen[depth] = true
			depths = append(depths, depth)
		}
	}
	if len(depths) == 0 {
		return nil, errors.New("no depths provided")
	}
	sort.Ints(depths)
	return depths, nil
}

func parseCounterfactualMode(mode string) (int, error) {
	switch mode {
	case "observational":
		return 0, nil
	case "full-depth":
		return 0, nil
	case "plus-one":
		return 1, nil
	default:
		return 0, fmt.Errorf("invalid counterfactual mode %q", mode)
	}
}

func parseCounterfactualCases(
	value string,
	mode string,
	cases []corpusCase,
) (map[string]bool, []string, error) {
	selected := map[string]bool{}
	for _, part := range strings.Split(value, ",") {
		name := strings.TrimSpace(part)
		if name != "" {
			selected[name] = true
		}
	}
	if mode == "observational" {
		if len(selected) != 0 {
			return nil, nil, errors.New("observational mode cannot name counterfactual cases")
		}
		return selected, []string{}, nil
	}
	if len(selected) == 0 {
		return nil, nil, fmt.Errorf("counterfactual mode %q requires -counterfactual-cases", mode)
	}
	available := make(map[string]bool, len(cases))
	for _, tc := range cases {
		available[tc.Name] = true
	}
	for name := range selected {
		if !available[name] {
			return nil, nil, fmt.Errorf("counterfactual case %q is not in the selected corpus", name)
		}
	}
	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	sort.Strings(names)
	return selected, names, nil
}

func parseBridgeLineText(value string) ([]string, error) {
	line := []string{}
	for _, part := range strings.Split(value, ",") {
		moveText := strings.TrimSpace(part)
		if moveText != "" {
			line = append(line, moveText)
		}
	}
	if value != "" && len(line) == 0 {
		return nil, errors.New("bridge line contains no moves")
	}
	return line, nil
}

func validateBridgeSelection(
	bridgeCase string,
	bridgeLine []string,
	counterfactualMode string,
	counterfactualCases map[string]bool,
	cases []corpusCase,
) error {
	if bridgeCase == "" {
		if len(bridgeLine) != 0 {
			return errors.New("-bridge-line requires -bridge-case")
		}
		return nil
	}
	if len(bridgeLine) < 3 {
		return errors.New("bridge tracing requires a line of at least three moves")
	}
	if counterfactualMode != "full-depth" || !counterfactualCases[bridgeCase] {
		return errors.New("bridge tracing requires the same case in full-depth counterfactual mode")
	}
	for _, tc := range cases {
		if tc.Name == bridgeCase {
			return nil
		}
	}
	return fmt.Errorf("bridge case %q is not in the selected corpus", bridgeCase)
}

func validateBridgeNoLMRSelection(targetMove, bridgeCase string, bridgeLine []string) error {
	if targetMove == "" {
		return nil
	}
	if bridgeCase == "" || len(bridgeLine) < 4 {
		return errors.New("-bridge-no-lmr-move requires a bridge case and line")
	}
	match := -1
	for index := 2; index < len(bridgeLine)-1; index++ {
		if bridgeLine[index] != targetMove {
			continue
		}
		if match >= 0 {
			return fmt.Errorf("bridge NoLMR move %q is not unique in the intermediate line", targetMove)
		}
		match = index
	}
	if match < 0 {
		return fmt.Errorf("bridge NoLMR move %q is not an intermediate move in the line", targetMove)
	}
	return nil
}

func validateBridgeNoFutilitySelection(
	targetMove string,
	noLMRMove string,
	bridgeCase string,
	bridgeLine []string,
) error {
	if targetMove == "" {
		return nil
	}
	if bridgeCase == "" || len(bridgeLine) < 3 {
		return errors.New("-bridge-no-futility-move requires a bridge case and line")
	}
	if noLMRMove == "" {
		return errors.New("-bridge-no-futility-move requires -bridge-no-lmr-move")
	}
	if bridgeLine[len(bridgeLine)-1] != targetMove {
		return fmt.Errorf("bridge NoFutility move %q must be the final bridge move", targetMove)
	}
	return nil
}

func validateBridgeNoLMRMove(tc *corpusCase, lineUCI []string, targetMove string) (int, error) {
	if err := validateBridgeNoLMRSelection(targetMove, tc.Name, lineUCI); err != nil {
		return 0, err
	}
	position, err := board.FromFEN(tc.FEN)
	if err != nil {
		return 0, fmt.Errorf("bridge NoLMR FromFEN: %w", err)
	}
	for index, moveText := range lineUCI {
		m, err := uci.ParseUCIMove(position, moveText)
		if err != nil {
			return 0, fmt.Errorf("bridge NoLMR move %d %s: %w", index+1, moveText, err)
		}
		position.MakeMove(m)
		if search.IsMovedSideInCheck(position) {
			return 0, fmt.Errorf("bridge NoLMR move %d %s is illegal", index+1, moveText)
		}
		if moveText == targetMove && index >= 2 && index < len(lineUCI)-1 {
			if move.IsCapture(m) || move.IsPromotion(m) || !search.InCheck(position) {
				return 0, errors.New("bridge NoLMR move must be a legal quiet check")
			}
			return index, nil
		}
	}
	return 0, fmt.Errorf("bridge NoLMR move %q was not reached", targetMove)
}

func parseAndValidateBridgeLine(tc *corpusCase, lineUCI []string) ([]uint32, uint32, error) {
	if len(lineUCI) < 3 || lineUCI[0] != tc.RootMoveUCI || lineUCI[1] != tc.ReplyMoveUCI {
		return nil, 0, errors.New("bridge line must begin with the case root and reply moves")
	}
	if len(lineUCI) >= search.MaxDepth {
		return nil, 0, errors.New("bridge line exceeds trace path capacity")
	}
	position, err := board.FromFEN(tc.FEN)
	if err != nil {
		return nil, 0, fmt.Errorf("bridge FromFEN: %w", err)
	}
	encoded := make([]uint32, 0, len(lineUCI))
	for index, moveText := range lineUCI {
		m, err := uci.ParseUCIMove(position, moveText)
		if err != nil {
			return nil, 0, fmt.Errorf("bridge move %d %s: %w", index+1, moveText, err)
		}
		position.MakeMove(m)
		if search.IsMovedSideInCheck(position) {
			return nil, 0, fmt.Errorf("bridge move %d %s is illegal", index+1, moveText)
		}
		encoded = append(encoded, m)
	}
	bridgeMove := encoded[len(encoded)-1]
	if move.IsCapture(bridgeMove) || move.IsPromotion(bridgeMove) || !search.InCheck(position) {
		return nil, 0, errors.New("final bridge move must be a legal quiet check")
	}
	bridgePath := append([]uint32(nil), encoded[:len(encoded)-1]...)
	return bridgePath, bridgeMove, nil
}

func cohortCounts(cases []corpusCase) (primary, canaries int) {
	for _, tc := range cases {
		if tc.Cohort == "canary" {
			canaries++
		} else if tc.Cohort == "primary_quiet_check" {
			primary++
		}
	}
	return primary, canaries
}

func sortedKeys(values map[int]bool) []int {
	result := make([]int, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Ints(result)
	return result
}

func moveUCI(m uint32) string {
	if m == 0 {
		return ""
	}
	return uci.MoveToUCI(m)
}

func sha256File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "search_trace:", err)
	os.Exit(1)
}
