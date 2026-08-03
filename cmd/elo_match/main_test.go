package main

import (
	"flag"
	"reflect"
	"testing"
)

func TestUCIOptionsAreRepeatableAndSeparatedByEngine(t *testing.T) {
	var engineAOptions, engineBOptions uciOptions
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.Var(&engineAOptions, "option-a", "")
	flags.Var(&engineBOptions, "option-b", "")

	err := flags.Parse([]string{
		"-option-a", "Threads=1",
		"-option-b", "UCI_LimitStrength=true",
		"-option-a", "Hash=16",
		"-option-b", "UCI_Elo=1600",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if want := (uciOptions{"Threads=1", "Hash=16"}); !reflect.DeepEqual(engineAOptions, want) {
		t.Fatalf("engine A options = %v, want %v", engineAOptions, want)
	}
	if want := (uciOptions{"UCI_LimitStrength=true", "UCI_Elo=1600"}); !reflect.DeepEqual(engineBOptions, want) {
		t.Fatalf("engine B options = %v, want %v", engineBOptions, want)
	}
}

func TestUCIOptionsRejectEmptyAndMalformedValues(t *testing.T) {
	tests := []string{
		"",
		"Threads",
		"=1",
		" =1",
		"Threads=",
		"Threads= ",
	}
	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			var options uciOptions
			if err := options.Set(value); err == nil {
				t.Fatalf("Set(%q) succeeded, want error", value)
			}
			if len(options) != 0 {
				t.Fatalf("options = %v after rejected value, want empty", options)
			}
		})
	}
}

func TestBuildCutechessArgsPlacesOptionsInExactEngineOrder(t *testing.T) {
	cfg := config{
		engineAName:    "Candidate",
		engineBName:    "Stockfish-1600",
		engineAOptions: uciOptions{"Threads=1", "Hash=16"},
		engineBOptions: uciOptions{"UCI_LimitStrength=true", "UCI_Elo=1600", "Threads=1", "Hash=16"},
		timeControl:    "1+0.1",
		concurrency:    2,
	}
	got := buildCutechessArgs(cfg, "/engine-a", "/engine-b", "/openings.epd", "/match.pgn", 32)
	want := []string{
		"-engine", "cmd=/engine-a", "name=Candidate", "proto=uci",
		"option.Threads=1", "option.Hash=16",
		"-engine", "cmd=/engine-b", "name=Stockfish-1600", "proto=uci",
		"option.UCI_LimitStrength=true", "option.UCI_Elo=1600", "option.Threads=1", "option.Hash=16",
		"-each", "tc=1+0.1", "restart=off",
		"-games", "2",
		"-rounds", "32",
		"-repeat", "2",
		"-openings", "file=/openings.epd", "format=epd", "order=sequential", "policy=default",
		"-tournament", "round-robin",
		"-concurrency", "2",
		"-pgnout", "/match.pgn", "fi",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCutechessArgs() =\n%q\nwant\n%q", got, want)
	}
}

func TestBuildCutechessArgsHasNoDefaultUCIOptions(t *testing.T) {
	cfg := config{
		engineAName: "A",
		engineBName: "B",
		timeControl: "1+0.1",
		concurrency: 1,
	}
	got := buildCutechessArgs(cfg, "/engine-a", "/engine-b", "/openings.epd", "/match.pgn", 1)
	want := []string{
		"-engine", "cmd=/engine-a", "name=A", "proto=uci",
		"-engine", "cmd=/engine-b", "name=B", "proto=uci",
		"-each", "tc=1+0.1", "restart=off",
		"-games", "2",
		"-rounds", "1",
		"-repeat", "2",
		"-openings", "file=/openings.epd", "format=epd", "order=sequential", "policy=default",
		"-tournament", "round-robin",
		"-concurrency", "1",
		"-pgnout", "/match.pgn", "fi",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildCutechessArgs() =\n%q\nwant\n%q", got, want)
	}
}
