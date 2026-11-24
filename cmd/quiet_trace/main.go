//go:build searchtrace && quietverify && !trace1850

// quiet_trace measures how often the root verification reaches the arbiter's defense.
// Pass a loss analysis and its cold UCI replay. Each output row must reproduce the
// candidate's UCI move and nodes before its root trace can be counted.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"EquinoxChess/internal/api"
	"EquinoxChess/internal/board"
	"EquinoxChess/internal/search"
	"EquinoxChess/internal/uci"
)

type key struct {
	Round int
	Color string
}

type turningPoint struct {
	Opening string
	Prefix  []string
}

type replayCase struct {
	key
	Depth     int
	Reference string
	Candidate struct {
		Move  string
		Nodes uint64
	}
}

type observation struct {
	Round       int      `json:"round"`
	Color       string   `json:"color"`
	Reference   string   `json:"reference"`
	Depth       int      `json:"depth"`
	FinalModes  []string `json:"final_depth_reference_modes"`
	FinalChecks []string `json:"final_depth_verified_moves"`
	RefChecked  bool     `json:"reference_verified_any_iteration"`
	RefFull     bool     `json:"reference_full_depth_final_iteration"`
}

func read(path string, dest any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func observe(p turningPoint, c replayCase) (observation, error) {
	o := observation{Round: c.Round, Color: c.Color, Depth: c.Depth, Reference: c.Reference}
	g := api.NewGame()
	b, err := board.FromFEN(p.Opening)
	if err != nil {
		return o, err
	}
	g.Board = b
	for _, move := range p.Prefix {
		if err := g.PlayMove(move); err != nil {
			return o, err
		}
	}
	trace := &search.SearchTrace{}
	g.Search.Trace = trace
	best, _, depth, nodes := g.Search.SearchRoot(g.Board, c.Depth)
	if depth != c.Depth || uci.MoveToUCI(best) != c.Candidate.Move || nodes != c.Candidate.Nodes {
		return o, fmt.Errorf("trace and UCI disagree for round %d %s at D%d", c.Round, c.Color, c.Depth)
	}
	for _, row := range trace.RootMoveResults {
		move := uci.MoveToUCI(row.Move)
		check := row.SearchMode == "root_quiet_verification"
		if check && move == c.Reference {
			o.RefChecked = true
		}
		if row.Iteration != depth {
			continue
		}
		if check {
			o.FinalChecks = append(o.FinalChecks, move)
		}
		if move == c.Reference {
			o.FinalModes = append(o.FinalModes, row.SearchMode)
			if row.ChildDepth >= depth-1 {
				o.RefFull = true
			}
		}
	}
	return o, nil
}

func run(lossesFile, replayFile string) error {
	var analysis struct {
		Losses []struct {
			key
			Point *turningPoint `json:"turning_point"`
		}
	}
	var replay struct{ Cases []replayCase }
	if err := read(lossesFile, &analysis); err != nil {
		return err
	}
	if err := read(replayFile, &replay); err != nil {
		return err
	}
	points := make(map[key]*turningPoint)
	for _, loss := range analysis.Losses {
		points[loss.key] = loss.Point
	}
	rows := make([]observation, 0, len(replay.Cases))
	for _, c := range replay.Cases {
		p := points[c.key]
		if p == nil {
			return fmt.Errorf("missing loss for %+v", c.key)
		}
		row, err := observe(*p, c)
		if err != nil {
			return err
		}
		rows = append(rows, row)
	}
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	return out.Encode(rows)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: quiet_trace losses.json quiet-replay.json")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
