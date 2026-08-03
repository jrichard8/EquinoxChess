package matchstats

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

const fenOne = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -"
const fenTwo = "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq -"

func TestParseAndAnalyzePairedPGN(t *testing.T) {
	pgn := `[Event "paired"]
[Round "1"]
[White "Candidate"]
[Black "Baseline"]
[Result "1-0"]
[FEN "` + fenOne + ` 0 1"]
[Termination "normal"]

1. e4 e5 1-0

[Event "paired"]
[Round "1"]
[White "Baseline"]
[Black "Candidate"]
[Result "1/2-1/2"]
[FEN "` + fenOne + ` 0 1"]

1/2-1/2

[Event "paired"]
[Round "2"]
[White "Candidate"]
[Black "Baseline"]
[Result "0-1"]
[FEN "` + fenTwo + ` 3 5"]

0-1

[Event "paired"]
[Round "2"]
[White "Baseline"]
[Black "Candidate"]
[Result "0-1"]
[FEN "` + fenTwo + ` 3 5"]

0-1
`
	games, err := ParsePGN(strings.NewReader(pgn))
	if err != nil {
		t.Fatalf("ParsePGN() error = %v", err)
	}
	if got, want := len(games), 4; got != want {
		t.Fatalf("len(games) = %d, want %d", got, want)
	}
	if got := games[0].Tags["Termination"]; got != "normal" {
		t.Fatalf("Termination = %q, want normal", got)
	}

	report, err := Analyze(games, "Candidate", "Baseline")
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report.Games != 4 || report.Pairs != 2 || report.Openings != 2 {
		t.Fatalf("counts = games %d, pairs %d, openings %d", report.Games, report.Pairs, report.Openings)
	}
	if report.Wins != 2 || report.Draws != 1 || report.Losses != 1 {
		t.Fatalf("W-D-L = %d-%d-%d, want 2-1-1", report.Wins, report.Draws, report.Losses)
	}
	if math.Abs(report.Score-0.625) > 1e-12 {
		t.Fatalf("score = %v, want 0.625", report.Score)
	}
	if report.AsWhite.Wins != 1 || report.AsWhite.Losses != 1 {
		t.Fatalf("white breakdown = %+v", report.AsWhite)
	}
	if report.AsBlack.Wins != 1 || report.AsBlack.Draws != 1 {
		t.Fatalf("black breakdown = %+v", report.AsBlack)
	}
	if report.Interpretation != "inconclusif" {
		t.Fatalf("interpretation = %q, want inconclusif", report.Interpretation)
	}
	if report.EngineATerminations != (TerminationStats{}) || report.EngineBTerminations != (TerminationStats{}) {
		t.Fatalf("normal and absent termination stats = A %+v, B %+v; want zero", report.EngineATerminations, report.EngineBTerminations)
	}
}

func TestAnalyzeClassifiesAbnormalTerminationsByLosingEngine(t *testing.T) {
	games := []Game{
		gameWithTermination(1, "Candidate", "Baseline", "0-1", fenOne+" 0 1", "time forfeit"),
		gameWithTermination(1, "Baseline", "Candidate", "0-1", fenOne+" 0 1", "normal"),
		gameWithTermination(2, "Candidate", "Baseline", "1-0", fenOne+" 0 1", "illegal move"),
		gameWithTermination(2, "Baseline", "Candidate", "1-0", fenOne+" 0 1", "normal"),
		gameWithTermination(3, "Candidate", "Baseline", "0-1", fenOne+" 0 1", "engine disconnected"),
		gameWithTermination(3, "Baseline", "Candidate", "0-1", fenOne+" 0 1", "engine crash"),
		gameWithTermination(4, "Candidate", "Baseline", "0-1", fenOne+" 0 1", "stalled connection"),
		gameWithTermination(4, "Baseline", "Candidate", "0-1", fenOne+" 0 1", "unexpected arena failure"),
	}

	report, err := Analyze(games, "Candidate", "Baseline")
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	wantA := TerminationStats{TimeForfeits: 1, Disconnects: 1, Stalls: 1}
	wantB := TerminationStats{IllegalMoves: 1, Crashes: 1, Other: 1}
	if report.EngineATerminations != wantA {
		t.Fatalf("Candidate termination stats = %+v, want %+v", report.EngineATerminations, wantA)
	}
	if report.EngineBTerminations != wantB {
		t.Fatalf("Baseline termination stats = %+v, want %+v", report.EngineBTerminations, wantB)
	}
	if report.Wins != 4 || report.Draws != 0 || report.Losses != 4 {
		t.Fatalf("W-D-L = %d-%d-%d, want 4-0-4", report.Wins, report.Draws, report.Losses)
	}
}

func TestAnalyzePGNWithoutTerminationTagRemainsCompatible(t *testing.T) {
	games := []Game{
		game(1, "Candidate", "Baseline", "1-0", fenOne+" 0 1"),
		game(1, "Baseline", "Candidate", "0-1", fenOne+" 0 1"),
	}
	report, err := Analyze(games, "Candidate", "Baseline")
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report.EngineATerminations != (TerminationStats{}) || report.EngineBTerminations != (TerminationStats{}) {
		t.Fatalf("termination stats = A %+v, B %+v; want zero", report.EngineATerminations, report.EngineBTerminations)
	}
}

func TestFormatReportIncludesTerminationCounters(t *testing.T) {
	report := Report{
		EngineA:             "Candidate",
		EngineB:             "Baseline",
		EngineATerminations: TerminationStats{TimeForfeits: 1, Disconnects: 2},
		EngineBTerminations: TerminationStats{IllegalMoves: 3, Crashes: 4, Stalls: 5, Other: 6},
	}
	formatted := FormatReport(report)
	for _, want := range []string{
		"Abnormal terminations by losing engine:",
		"Candidate: time forfeits 1, illegal moves 0, disconnects 2, crashes 0, stalls 0, other/unknown 0",
		"Baseline: time forfeits 0, illegal moves 3, disconnects 0, crashes 4, stalls 5, other/unknown 6",
	} {
		if !strings.Contains(formatted, want) {
			t.Fatalf("FormatReport() = %q, want containing %q", formatted, want)
		}
	}
}

func TestHoeffdingIntervalRemainsConservativeForFourConsistentPairs(t *testing.T) {
	var games []Game
	for round := 1; round <= 4; round++ {
		games = append(games,
			game(round, "Candidate", "Baseline", "1-0", fenOne+" 0 1"),
			game(round, "Baseline", "Candidate", "1/2-1/2", fenOne+" 0 1"),
		)
	}
	report, err := Analyze(games, "Candidate", "Baseline")
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	wantElo := 400 * math.Log10(3)
	if math.Abs(report.Elo-wantElo) > 1e-9 {
		t.Fatalf("Elo = %v, want %v", report.Elo, wantElo)
	}
	if report.ScoreLow95 >= report.Score || report.ScoreHigh95 <= report.Score {
		t.Fatalf("Hoeffding interval = [%v, %v], want non-degenerate around %v", report.ScoreLow95, report.ScoreHigh95, report.Score)
	}
	if report.Interpretation != "inconclusif" {
		t.Fatalf("interpretation = %q, want inconclusif", report.Interpretation)
	}
}

func TestHoeffdingIntervalReportsGainForStrongLargeSample(t *testing.T) {
	var games []Game
	for round := 1; round <= 32; round++ {
		games = append(games,
			game(round, "Candidate", "Baseline", "1-0", fenOne+" 0 1"),
			game(round, "Baseline", "Candidate", "0-1", fenOne+" 0 1"),
		)
	}
	report, err := Analyze(games, "Candidate", "Baseline")
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	wantLow := 1 - math.Sqrt(math.Log(2/0.05)/(2*32))
	if math.Abs(report.ScoreLow95-wantLow) > 1e-12 || report.ScoreHigh95 != 1 {
		t.Fatalf("score interval = [%v, %v], want [%v, 1]", report.ScoreLow95, report.ScoreHigh95, wantLow)
	}
	if report.Interpretation != "gain" {
		t.Fatalf("interpretation = %q, want gain", report.Interpretation)
	}
}

func TestAnalyzeRejectsInvalidPair(t *testing.T) {
	tests := []struct {
		name  string
		games []Game
		want  string
	}{
		{
			name: "colors not swapped",
			games: []Game{
				game(1, "Candidate", "Baseline", "1-0", fenOne+" 0 1"),
				game(1, "Candidate", "Baseline", "0-1", fenOne+" 0 1"),
			},
			want: "does not swap",
		},
		{
			name: "different FEN",
			games: []Game{
				game(1, "Candidate", "Baseline", "1-0", fenOne+" 0 1"),
				game(1, "Baseline", "Candidate", "0-1", fenTwo+" 0 1"),
			},
			want: "different opening FENs",
		},
		{
			name: "different FEN counters",
			games: []Game{
				game(1, "Candidate", "Baseline", "1-0", fenOne+" 0 1"),
				game(1, "Baseline", "Candidate", "0-1", fenOne+" 1 1"),
			},
			want: "different opening FENs",
		},
		{
			name: "unfinished game",
			games: []Game{
				game(1, "Candidate", "Baseline", "*", fenOne+" 0 1"),
				game(1, "Baseline", "Candidate", "0-1", fenOne+" 0 1"),
			},
			want: "unfinished result",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Analyze(tt.games, "Candidate", "Baseline")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Analyze() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestParsePGNRejectsMalformedTag(t *testing.T) {
	_, err := ParsePGN(strings.NewReader("[Event bad]\n"))
	if err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("ParsePGN() error = %v, want malformed tag error", err)
	}
}

func game(round int, white, black, result, fen string) Game {
	return Game{Tags: map[string]string{
		"Round":  strconv.Itoa(round),
		"White":  white,
		"Black":  black,
		"Result": result,
		"FEN":    fen,
	}}
}

func gameWithTermination(round int, white, black, result, fen, termination string) Game {
	game := game(round, white, black, result, fen)
	game.Tags["Termination"] = termination
	return game
}
