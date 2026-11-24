package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jrichard8/MarsGopherChess/internal/matchstats"
)

type config struct {
	engineAPath    string
	engineBPath    string
	engineAName    string
	engineBName    string
	engineAOptions uciOptions
	engineBOptions uciOptions
	pgnPath        string
	openings       string
	openingCount   int
	timeControl    string
	concurrency    int
	cutechess      string
}

type uciOptions []string

func (options *uciOptions) String() string {
	if options == nil {
		return ""
	}
	return strings.Join(*options, ", ")
}

func (options *uciOptions) Set(value string) error {
	if err := validateUCIOption(value); err != nil {
		return err
	}
	*options = append(*options, value)
	return nil
}

type metadata struct {
	CreatedAtUTC     time.Time `json:"created_at_utc"`
	GitCommitAtStart string    `json:"git_commit_at_start,omitempty"`
	EngineAName      string    `json:"engine_a_name"`
	EngineBName      string    `json:"engine_b_name"`
	EngineAPath      string    `json:"engine_a_path"`
	EngineBPath      string    `json:"engine_b_path"`
	OpeningsPath     string    `json:"openings_path"`
	PGNPath          string    `json:"pgn_path"`
	CutechessPath    string    `json:"cutechess_path"`
	CutechessVersion string    `json:"cutechess_version"`
	EngineASHA256    string    `json:"engine_a_sha256"`
	EngineBSHA256    string    `json:"engine_b_sha256"`
	OpeningsSHA256   string    `json:"openings_sha256"`
	CutechessSHA256  string    `json:"cutechess_sha256"`
	PGNSHA256        string    `json:"pgn_sha256"`
	OpeningCount     int       `json:"opening_count"`
	TimeControl      string    `json:"time_control"`
	Concurrency      int       `json:"concurrency"`
	CutechessArgs    []string  `json:"cutechess_args"`
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.engineAPath, "engine-a", "", "path to engine A executable")
	flag.StringVar(&cfg.engineBPath, "engine-b", "", "path to engine B executable")
	flag.StringVar(&cfg.engineAName, "name-a", "A", "engine A name in the PGN")
	flag.StringVar(&cfg.engineBName, "name-b", "B", "engine B name in the PGN")
	flag.Var(&cfg.engineAOptions, "option-a", "engine A UCI option as Name=Value (repeatable)")
	flag.Var(&cfg.engineBOptions, "option-b", "engine B UCI option as Name=Value (repeatable)")
	flag.StringVar(&cfg.pgnPath, "pgn", "", "new PGN output path")
	flag.StringVar(&cfg.openings, "openings", "tools/elo/openings-v1.epd", "EPD opening suite")
	flag.IntVar(&cfg.openingCount, "openings-count", 0, "number of sequential openings (0 means all)")
	flag.StringVar(&cfg.timeControl, "tc", "1+0.1", "cutechess time control")
	flag.IntVar(&cfg.concurrency, "concurrency", 1, "number of concurrent games")
	flag.StringVar(&cfg.cutechess, "cutechess", "cutechess-cli", "cutechess-cli path or command")
	flag.Parse()

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "elo_match: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	if cfg.engineAPath == "" || cfg.engineBPath == "" || cfg.pgnPath == "" {
		return fmt.Errorf("-engine-a, -engine-b, and -pgn are required")
	}
	if strings.TrimSpace(cfg.engineAName) == "" || strings.TrimSpace(cfg.engineBName) == "" {
		return fmt.Errorf("engine names must not be empty")
	}
	if cfg.engineAName == cfg.engineBName {
		return fmt.Errorf("engine names must be different")
	}
	if strings.TrimSpace(cfg.timeControl) == "" {
		return fmt.Errorf("time control must not be empty")
	}
	if cfg.concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}
	if err := validateUCIOptions("-option-a", cfg.engineAOptions); err != nil {
		return err
	}
	if err := validateUCIOptions("-option-b", cfg.engineBOptions); err != nil {
		return err
	}

	engineAPath, err := executablePath(cfg.engineAPath)
	if err != nil {
		return fmt.Errorf("engine A: %w", err)
	}
	engineBPath, err := executablePath(cfg.engineBPath)
	if err != nil {
		return fmt.Errorf("engine B: %w", err)
	}
	cutechessPath, err := exec.LookPath(cfg.cutechess)
	if err != nil {
		return fmt.Errorf("find cutechess: %w", err)
	}
	cutechessPath, err = filepath.Abs(cutechessPath)
	if err != nil {
		return fmt.Errorf("resolve cutechess path: %w", err)
	}
	openingsPath, err := inputFilePath(cfg.openings)
	if err != nil {
		return fmt.Errorf("openings: %w", err)
	}
	pgnPath, err := outputPath(cfg.pgnPath)
	if err != nil {
		return fmt.Errorf("PGN: %w", err)
	}
	metaPath := pgnPath + ".meta.json"
	if err := requireAbsent(metaPath); err != nil {
		return fmt.Errorf("metadata: %w", err)
	}

	availableOpenings, err := countOpenings(openingsPath)
	if err != nil {
		return err
	}
	openingCount := cfg.openingCount
	if openingCount == 0 {
		openingCount = availableOpenings
	}
	if openingCount < 1 || openingCount > availableOpenings {
		return fmt.Errorf("openings-count must be between 1 and %d (got %d)", availableOpenings, openingCount)
	}
	gitCommitAtStart := currentGitCommit()
	cutechessVersion, err := commandVersion(cutechessPath)
	if err != nil {
		return err
	}
	engineAHash, err := hashFile(engineAPath)
	if err != nil {
		return err
	}
	engineBHash, err := hashFile(engineBPath)
	if err != nil {
		return err
	}
	openingsHash, err := hashFile(openingsPath)
	if err != nil {
		return err
	}
	cutechessHash, err := hashFile(cutechessPath)
	if err != nil {
		return err
	}

	args := buildCutechessArgs(cfg, engineAPath, engineBPath, openingsPath, pgnPath, openingCount)
	command := exec.Command(cutechessPath, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	if err := command.Run(); err != nil {
		return fmt.Errorf("cutechess failed: %w", err)
	}
	for _, immutable := range []struct {
		name string
		path string
		hash string
	}{
		{name: "engine A", path: engineAPath, hash: engineAHash},
		{name: "engine B", path: engineBPath, hash: engineBHash},
		{name: "opening suite", path: openingsPath, hash: openingsHash},
		{name: "cutechess", path: cutechessPath, hash: cutechessHash},
	} {
		if err := requireUnchanged(immutable.name, immutable.path, immutable.hash); err != nil {
			return err
		}
	}

	report, err := analyzePGN(pgnPath, cfg.engineAName, cfg.engineBName)
	if err != nil {
		return fmt.Errorf("validate match PGN: %w", err)
	}
	if report.Games != 2*openingCount || report.Pairs != openingCount || report.Openings != openingCount {
		return fmt.Errorf(
			"match does not cover requested openings: got %d games, %d pairs, %d unique openings; want %d, %d, %d",
			report.Games, report.Pairs, report.Openings, 2*openingCount, openingCount, openingCount,
		)
	}
	pgnHash, err := hashFile(pgnPath)
	if err != nil {
		return err
	}

	meta := metadata{
		CreatedAtUTC:     time.Now().UTC(),
		GitCommitAtStart: gitCommitAtStart,
		EngineAName:      cfg.engineAName,
		EngineBName:      cfg.engineBName,
		EngineAPath:      engineAPath,
		EngineBPath:      engineBPath,
		OpeningsPath:     openingsPath,
		PGNPath:          pgnPath,
		CutechessPath:    cutechessPath,
		CutechessVersion: cutechessVersion,
		EngineASHA256:    engineAHash,
		EngineBSHA256:    engineBHash,
		OpeningsSHA256:   openingsHash,
		CutechessSHA256:  cutechessHash,
		PGNSHA256:        pgnHash,
		OpeningCount:     openingCount,
		TimeControl:      cfg.timeControl,
		Concurrency:      cfg.concurrency,
		CutechessArgs:    args,
	}
	if err := writeMetadata(metaPath, meta); err != nil {
		return err
	}

	fmt.Println(matchstats.FormatReport(report))
	fmt.Printf("metadata: %s\n", metaPath)
	return nil
}

func validateUCIOption(option string) error {
	name, value, ok := strings.Cut(option, "=")
	if !ok {
		return fmt.Errorf("must use Name=Value")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("option name must not be empty")
	}
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("option value must not be empty")
	}
	return nil
}

func validateUCIOptions(flagName string, options uciOptions) error {
	for _, option := range options {
		if err := validateUCIOption(option); err != nil {
			return fmt.Errorf("%s %q: %w", flagName, option, err)
		}
	}
	return nil
}

func buildCutechessArgs(cfg config, engineAPath, engineBPath, openingsPath, pgnPath string, openingCount int) []string {
	args := []string{"-engine", "cmd=" + engineAPath, "name=" + cfg.engineAName, "proto=uci"}
	for _, option := range cfg.engineAOptions {
		args = append(args, "option."+option)
	}
	args = append(args, "-engine", "cmd="+engineBPath, "name="+cfg.engineBName, "proto=uci")
	for _, option := range cfg.engineBOptions {
		args = append(args, "option."+option)
	}
	return append(args,
		"-each", "tc="+cfg.timeControl, "restart=off",
		"-games", "2",
		"-rounds", fmt.Sprint(openingCount),
		"-repeat", "2",
		"-openings", "file="+openingsPath, "format=epd", "order=sequential", "policy=default",
		"-tournament", "round-robin",
		"-concurrency", fmt.Sprint(cfg.concurrency),
		"-pgnout", pgnPath, "fi",
	)
}

func executablePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", abs)
	}
	if info.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("%s is not executable", abs)
	}
	return abs, nil
}

func inputFilePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", abs)
	}
	return abs, nil
}

func outputPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err := requireAbsent(abs); err != nil {
		return "", err
	}
	parent := filepath.Dir(abs)
	info, err := os.Stat(parent)
	if err != nil {
		return "", fmt.Errorf("output directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("output parent %s is not a directory", parent)
	}
	return abs, nil
}

func requireAbsent(path string) error {
	_, err := os.Lstat(path)
	switch {
	case err == nil:
		return fmt.Errorf("refusing to overwrite %s", path)
	case os.IsNotExist(err):
		return nil
	default:
		return err
	}
}

func countOpenings(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("read openings: %w", err)
	}
	if count == 0 {
		return 0, fmt.Errorf("opening suite is empty")
	}
	return count, nil
}

func analyzePGN(path, engineA, engineB string) (matchstats.Report, error) {
	f, err := os.Open(path)
	if err != nil {
		return matchstats.Report{}, err
	}
	defer f.Close()
	games, err := matchstats.ParsePGN(f)
	if err != nil {
		return matchstats.Report{}, err
	}
	return matchstats.Analyze(games, engineA, engineB)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func requireUnchanged(name, path, beforeHash string) error {
	afterHash, err := hashFile(path)
	if err != nil {
		return fmt.Errorf("verify %s: %w", name, err)
	}
	if afterHash != beforeHash {
		return fmt.Errorf("%s changed while the match was running", name)
	}
	return nil
}

func commandVersion(path string) (string, error) {
	command := exec.Command(path, "-version")
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("read cutechess version: %w: %s", err, strings.TrimSpace(string(output)))
	}
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("read cutechess version: empty output")
	}
	return version, nil
}

func currentGitCommit() string {
	command := exec.Command("git", "rev-parse", "HEAD")
	command.Stderr = io.Discard
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func writeMetadata(path string, meta metadata) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create metadata: %w", err)
	}
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(meta); err != nil {
		_ = f.Close()
		return fmt.Errorf("write metadata: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close metadata: %w", err)
	}
	return nil
}
