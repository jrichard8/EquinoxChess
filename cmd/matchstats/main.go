package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jrichard8/MarsGopherChess/internal/matchstats"
)

func main() {
	pgnPath := flag.String("pgn", "", "PGN file produced by a paired match")
	engineA := flag.String("engine-a", "", "exact PGN name of engine A")
	engineB := flag.String("engine-b", "", "exact PGN name of engine B")
	flag.Parse()

	if err := run(*pgnPath, *engineA, *engineB); err != nil {
		fmt.Fprintf(os.Stderr, "matchstats: %v\n", err)
		os.Exit(1)
	}
}

func run(pgnPath, engineA, engineB string) error {
	if pgnPath == "" || engineA == "" || engineB == "" {
		return fmt.Errorf("-pgn, -engine-a, and -engine-b are required")
	}
	f, err := os.Open(pgnPath)
	if err != nil {
		return fmt.Errorf("open PGN: %w", err)
	}
	defer f.Close()

	games, err := matchstats.ParsePGN(f)
	if err != nil {
		return err
	}
	report, err := matchstats.Analyze(games, engineA, engineB)
	if err != nil {
		return err
	}
	fmt.Println(matchstats.FormatReport(report))
	return nil
}
