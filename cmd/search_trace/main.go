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
	"runtime"
	"sort"
	"strconv"
	"strings"

	"EquinoxChess/internal/board"
	"EquinoxChess/internal/search"
	"EquinoxChess/internal/tt"
	"EquinoxChess/internal/uci"
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
	RequestedDepth          int                `json:"requested_depth"`
	CompletedDepth          int                `json:"completed_depth"`
	BestMoveUCI             string             `json:"best_move_uci"`
	Score                   int                `json:"score"`
	Nodes                   uint64             `json:"nodes"`
	RootMoveBest            bool               `json:"root_move_best"`
	NonInterferenceVerified bool               `json:"non_interference_verified"`
	Summary                 runSummary         `json:"summary"`
	Events                  []traceEventOutput `json:"events"`
}

type traceEventOutput struct {
	search.TraceEvent
	TargetMoveUCI  string `json:"target_move_uci"`
	CurrentMoveUCI string `json:"current_move_uci"`
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
	TargetLineQEntriesAll       int         `json:"target_line_qsearch_entries_all"`
	TargetLineQEntriesFinal     int         `json:"target_line_qsearch_entries_final"`
	TargetLineQEntriesFinalPass int         `json:"target_line_qsearch_entries_final_pass"`
}

type aggregate struct {
	Cases        int            `json:"cases"`
	RootMoveBest int            `json:"root_move_best"`
	Metrics      map[string]int `json:"metrics"`
}

type report struct {
	SchemaVersion int                             `json:"schema_version"`
	Inputs        map[string]map[string]string    `json:"inputs"`
	Parameters    map[string]any                  `json:"parameters"`
	Counts        map[string]int                  `json:"counts"`
	Aggregates    map[string]map[string]aggregate `json:"aggregates_by_depth_and_cohort"`
	Cases         []corpusCase                    `json:"cases"`
}

func main() {
	input := flag.String("input", "", "king-zone tactical CSV")
	output := flag.String("output", "", "JSON trace output")
	depthText := flag.String("depths", "4,6", "comma-separated fixed depths")
	ttSize := flag.Int("tt-size", 1<<16, "fresh transposition-table entries per replay")
	expectedPrimary := flag.Int("expected-primary", 19, "required primary quiet-check cases")
	expectedCanaries := flag.Int("expected-canaries", 13, "required exact canary cases")
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

	for caseIndex := range cases {
		for _, depth := range depths {
			result, err := replay(&cases[caseIndex], depth, *ttSize)
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
		SchemaVersion: 1,
		Inputs: map[string]map[string]string{
			"csv":        {"path": *input, "sha256": inputHash},
			"executable": {"path": executable, "sha256": executableHash},
		},
		Parameters: map[string]any{
			"depths":                      depths,
			"tt_entries":                  *ttSize,
			"fresh_search_and_tt_per_run": true,
			"non_interference_check":      "traced and untraced best move, score, completed depth, and nodes must match",
			"go_version":                  runtime.Version(),
			"primary_selection":           "midgame PRE non-check, no canary, severe, queen_net_danger > 0, invisible quiet-check reply",
			"canary_selection":            "every exact canary row",
		},
		Counts: map[string]int{
			"cases":         len(cases),
			"primary_cases": primary,
			"canary_cases":  canaries,
			"runs":          len(cases) * len(depths),
		},
		Cases:      cases,
		Aggregates: buildAggregates(cases, depths),
	}
	encoded, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("search_trace: %d cases (%d primary, %d canaries), %d runs -> %s\n",
		len(cases), primary, canaries, len(cases)*len(depths), *output)
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

func replay(tc *corpusCase, depth, ttSize int) (run, error) {
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

	trace := &search.SearchTrace{RootMove: rootMove, ReplyMove: replyMove}
	worker := search.Search{TT: tt.NewTT(ttSize), Trace: trace}
	best, score, completedDepth, nodes := worker.SearchRoot(position, depth)
	plainPosition, err := board.FromFEN(tc.FEN)
	if err != nil {
		return run{}, fmt.Errorf("plain FromFEN: %w", err)
	}
	plain := search.Search{TT: tt.NewTT(ttSize)}
	plainBest, plainScore, plainDepth, plainNodes := plain.SearchRoot(plainPosition, depth)
	if best != plainBest || score != plainScore || completedDepth != plainDepth || nodes != plainNodes {
		return run{}, fmt.Errorf(
			"trace interference: traced=(%s,%d,%d,%d) plain=(%s,%d,%d,%d)",
			moveUCI(best), score, completedDepth, nodes,
			moveUCI(plainBest), plainScore, plainDepth, plainNodes,
		)
	}
	events := make([]traceEventOutput, 0, len(trace.Events))
	for _, event := range trace.Events {
		events = append(events, traceEventOutput{
			TraceEvent:     event,
			TargetMoveUCI:  moveUCI(event.TargetMove),
			CurrentMoveUCI: moveUCI(event.CurrentMove),
		})
	}
	return run{
		RequestedDepth:          depth,
		CompletedDepth:          completedDepth,
		BestMoveUCI:             moveUCI(best),
		Score:                   score,
		Nodes:                   nodes,
		RootMoveBest:            best == rootMove,
		NonInterferenceVerified: true,
		Summary:                 summarize(trace.Events, depth),
		Events:                  events,
	}, nil
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
					}
					addMetric := func(name string, condition bool) {
						if condition {
							entry.Metrics[name]++
						}
					}
					rootFinal := replay.Summary.RootFinalPass.StageCounts
					replyFinal := replay.Summary.ReplyFinalPass.StageCounts
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
