// Package matchstats parses paired-match PGNs and reports an engine's result.
package matchstats

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"EquinoxChess/internal/board"
)

// Game contains the PGN tags needed to analyse one game.
type Game struct {
	Tags map[string]string
}

// ColorStats is a W-D-L breakdown for one color.
type ColorStats struct {
	Games  int
	Wins   int
	Draws  int
	Losses int
}

// TerminationStats counts abnormal PGN terminations attributed to an engine.
type TerminationStats struct {
	TimeForfeits int
	IllegalMoves int
	Disconnects  int
	Crashes      int
	Stalls       int
	Other        int
}

// Report summarizes a paired match from EngineA's point of view.
type Report struct {
	EngineA             string
	EngineB             string
	Games               int
	Pairs               int
	Openings            int
	Wins                int
	Draws               int
	Losses              int
	Score               float64
	Elo                 float64
	EloLow95            float64
	EloHigh95           float64
	ScoreLow95          float64
	ScoreHigh95         float64
	AsWhite             ColorStats
	AsBlack             ColorStats
	EngineATerminations TerminationStats
	EngineBTerminations TerminationStats
	Interpretation      string
}

// ParsePGN reads PGN tag blocks. Movetext is intentionally ignored.
func ParsePGN(r io.Reader) ([]Game, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	var games []Game
	current := make(map[string]string)
	inMovetext := false
	lineNumber := 0

	finishGame := func() {
		if len(current) == 0 {
			return
		}
		games = append(games, Game{Tags: current})
		current = make(map[string]string)
		inMovetext = false
	}

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			name, value, err := parseTagLine(line)
			if err != nil {
				return nil, fmt.Errorf("PGN line %d: %w", lineNumber, err)
			}
			if inMovetext || (name == "Event" && current["Event"] != "") {
				finishGame()
			}
			if _, exists := current[name]; exists {
				return nil, fmt.Errorf("PGN line %d: duplicate %q tag", lineNumber, name)
			}
			current[name] = value
			continue
		}
		if line != "" && len(current) > 0 {
			inMovetext = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read PGN: %w", err)
	}
	finishGame()
	if len(games) == 0 {
		return nil, fmt.Errorf("PGN contains no games")
	}
	return games, nil
}

func parseTagLine(line string) (string, string, error) {
	if !strings.HasSuffix(line, "]") {
		return "", "", fmt.Errorf("malformed tag: missing closing bracket")
	}
	inside := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]"))
	separator := strings.IndexAny(inside, " \t")
	if separator <= 0 {
		return "", "", fmt.Errorf("malformed tag")
	}
	name := inside[:separator]
	quoted := strings.TrimSpace(inside[separator:])
	value, err := strconv.Unquote(quoted)
	if err != nil {
		return "", "", fmt.Errorf("malformed %q tag value: %w", name, err)
	}
	return name, value, nil
}

// Analyze validates color-swapped pairs and computes statistics for engineA.
func Analyze(games []Game, engineA, engineB string) (Report, error) {
	if strings.TrimSpace(engineA) == "" || strings.TrimSpace(engineB) == "" {
		return Report{}, fmt.Errorf("engine names must not be empty")
	}
	if engineA == engineB {
		return Report{}, fmt.Errorf("engine names must be different")
	}
	if len(games) == 0 {
		return Report{}, fmt.Errorf("cannot analyse an empty match")
	}

	byRound := make(map[string][]Game)
	for index, game := range games {
		for _, tag := range []string{"Round", "White", "Black", "Result", "FEN"} {
			if strings.TrimSpace(game.Tags[tag]) == "" {
				return Report{}, fmt.Errorf("game %d is missing required %q tag", index+1, tag)
			}
		}
		white, black := game.Tags["White"], game.Tags["Black"]
		if !((white == engineA && black == engineB) || (white == engineB && black == engineA)) {
			return Report{}, fmt.Errorf("game %d has unexpected players %q vs %q", index+1, white, black)
		}
		if _, err := whiteScore(game.Tags["Result"]); err != nil {
			return Report{}, fmt.Errorf("game %d: %w", index+1, err)
		}
		byRound[game.Tags["Round"]] = append(byRound[game.Tags["Round"]], game)
	}

	rounds := make([]string, 0, len(byRound))
	for round := range byRound {
		rounds = append(rounds, round)
	}
	sort.Strings(rounds)

	report := Report{EngineA: engineA, EngineB: engineB, Games: len(games), Pairs: len(rounds)}
	pairScores := make([]float64, 0, len(rounds))
	openings := make(map[string]struct{})
	for _, round := range rounds {
		pair := byRound[round]
		if len(pair) != 2 {
			return Report{}, fmt.Errorf("round %q has %d games, expected exactly 2", round, len(pair))
		}

		fen0, err := normalizeFEN(pair[0].Tags["FEN"])
		if err != nil {
			return Report{}, fmt.Errorf("round %q first game: %w", round, err)
		}
		fen1, err := normalizeFEN(pair[1].Tags["FEN"])
		if err != nil {
			return Report{}, fmt.Errorf("round %q second game: %w", round, err)
		}
		if fen0 != fen1 {
			return Report{}, fmt.Errorf("round %q uses different opening FENs", round)
		}
		openings[fen0] = struct{}{}

		if pair[0].Tags["White"] != pair[1].Tags["Black"] || pair[0].Tags["Black"] != pair[1].Tags["White"] {
			return Report{}, fmt.Errorf("round %q does not swap engine colors", round)
		}

		points := 0.0
		for _, game := range pair {
			gamePoints, err := pointsForEngine(game, engineA)
			if err != nil {
				return Report{}, fmt.Errorf("round %q: %w", round, err)
			}
			points += gamePoints
			updateReport(&report, game, engineA, gamePoints)
			updateTerminationStats(&report, game)
		}
		pairScores = append(pairScores, points/2.0)
	}

	report.Openings = len(openings)
	report.Score = (float64(report.Wins) + 0.5*float64(report.Draws)) / float64(report.Games)
	report.Elo = scoreToElo(report.Score)
	report.ScoreLow95, report.ScoreHigh95 = pairedHoeffdingInterval(pairScores)
	report.EloLow95 = scoreToElo(report.ScoreLow95)
	report.EloHigh95 = scoreToElo(report.ScoreHigh95)
	switch {
	case report.EloLow95 > 0:
		report.Interpretation = "gain"
	case report.EloHigh95 < 0:
		report.Interpretation = "perte"
	default:
		report.Interpretation = "inconclusif"
	}
	return report, nil
}

func normalizeFEN(fen string) (string, error) {
	fields := strings.Fields(fen)
	if len(fields) != 6 {
		return "", fmt.Errorf("invalid FEN %q: expected exactly 6 fields", fen)
	}
	normalized := strings.Join(fields, " ")
	if _, err := board.FromFEN(normalized); err != nil {
		return "", fmt.Errorf("invalid FEN %q: %w", fen, err)
	}
	return normalized, nil
}

func whiteScore(result string) (float64, error) {
	switch result {
	case "1-0":
		return 1, nil
	case "0-1":
		return 0, nil
	case "1/2-1/2":
		return 0.5, nil
	default:
		return 0, fmt.Errorf("invalid or unfinished result %q", result)
	}
}

func pointsForEngine(game Game, engine string) (float64, error) {
	whitePoints, err := whiteScore(game.Tags["Result"])
	if err != nil {
		return 0, err
	}
	switch engine {
	case game.Tags["White"]:
		return whitePoints, nil
	case game.Tags["Black"]:
		return 1 - whitePoints, nil
	default:
		return 0, fmt.Errorf("engine %q is not in game", engine)
	}
}

func updateReport(report *Report, game Game, engine string, points float64) {
	var color *ColorStats
	if game.Tags["White"] == engine {
		color = &report.AsWhite
	} else {
		color = &report.AsBlack
	}
	color.Games++
	switch points {
	case 1:
		report.Wins++
		color.Wins++
	case 0.5:
		report.Draws++
		color.Draws++
	default:
		report.Losses++
		color.Losses++
	}
}

type terminationKind int

const (
	terminationNormal terminationKind = iota
	terminationTimeForfeit
	terminationIllegalMove
	terminationDisconnect
	terminationCrash
	terminationStall
	terminationOther
)

func classifyTermination(value string) terminationKind {
	value = strings.ToLower(strings.TrimSpace(value))
	switch {
	case value == "", value == "normal":
		return terminationNormal
	case strings.Contains(value, "time forfeit"):
		return terminationTimeForfeit
	case strings.Contains(value, "illegal move"):
		return terminationIllegalMove
	case strings.Contains(value, "disconnect"):
		return terminationDisconnect
	case strings.Contains(value, "crash"):
		return terminationCrash
	case strings.Contains(value, "stall"):
		return terminationStall
	default:
		return terminationOther
	}
}

func updateTerminationStats(report *Report, game Game) {
	kind := classifyTermination(game.Tags["Termination"])
	if kind == terminationNormal {
		return
	}

	var loser string
	switch game.Tags["Result"] {
	case "1-0":
		loser = game.Tags["Black"]
	case "0-1":
		loser = game.Tags["White"]
	default:
		return
	}

	stats := &report.EngineBTerminations
	if loser == report.EngineA {
		stats = &report.EngineATerminations
	}
	switch kind {
	case terminationTimeForfeit:
		stats.TimeForfeits++
	case terminationIllegalMove:
		stats.IllegalMoves++
	case terminationDisconnect:
		stats.Disconnects++
	case terminationCrash:
		stats.Crashes++
	case terminationStall:
		stats.Stalls++
	case terminationOther:
		stats.Other++
	}
}

// pairedHoeffdingInterval returns a conservative bilateral 95% Hoeffding
// interval for independent pair scores bounded to [0, 1].
func pairedHoeffdingInterval(scores []float64) (float64, float64) {
	mean := 0.0
	for _, score := range scores {
		mean += score
	}
	mean /= float64(len(scores))
	margin := math.Sqrt(math.Log(2/0.05) / (2 * float64(len(scores))))
	return math.Max(0, mean-margin), math.Min(1, mean+margin)
}

func scoreToElo(score float64) float64 {
	switch {
	case score <= 0:
		return math.Inf(-1)
	case score >= 1:
		return math.Inf(1)
	default:
		return 400 * math.Log10(score/(1-score))
	}
}

// FormatReport returns a concise human-readable summary.
func FormatReport(report Report) string {
	return fmt.Sprintf(
		"%s vs %s: %d games, %d pairs, %d unique openings\n"+
			"W-D-L %d-%d-%d; score %.2f%%; Elo %+.1f (Hoeffding 95%% CI %+.1f to %+.1f): %s\n"+
			"%s as White %d-%d-%d; as Black %d-%d-%d\n"+
			"Abnormal terminations by losing engine:\n"+
			"  %s: time forfeits %d, illegal moves %d, disconnects %d, crashes %d, stalls %d, other/unknown %d\n"+
			"  %s: time forfeits %d, illegal moves %d, disconnects %d, crashes %d, stalls %d, other/unknown %d",
		report.EngineA, report.EngineB, report.Games, report.Pairs, report.Openings,
		report.Wins, report.Draws, report.Losses, 100*report.Score,
		report.Elo, report.EloLow95, report.EloHigh95, report.Interpretation,
		report.EngineA,
		report.AsWhite.Wins, report.AsWhite.Draws, report.AsWhite.Losses,
		report.AsBlack.Wins, report.AsBlack.Draws, report.AsBlack.Losses,
		report.EngineA,
		report.EngineATerminations.TimeForfeits, report.EngineATerminations.IllegalMoves,
		report.EngineATerminations.Disconnects, report.EngineATerminations.Crashes,
		report.EngineATerminations.Stalls, report.EngineATerminations.Other,
		report.EngineB,
		report.EngineBTerminations.TimeForfeits, report.EngineBTerminations.IllegalMoves,
		report.EngineBTerminations.Disconnects, report.EngineBTerminations.Crashes,
		report.EngineBTerminations.Stalls, report.EngineBTerminations.Other,
	)
}
