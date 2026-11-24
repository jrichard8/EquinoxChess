package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jrichard8/MarsGopherChess/internal/api"
	"github.com/jrichard8/MarsGopherChess/internal/board"
	"github.com/jrichard8/MarsGopherChess/internal/movegen"
	"github.com/jrichard8/MarsGopherChess/internal/search"
	"github.com/jrichard8/MarsGopherChess/internal/uci"
)

var game = api.NewGame()

type moveRequest struct {
	Move string `json:"move"`
}

type engineRequest struct {
	Depth int `json:"depth"`
}

type engineResponse struct {
	BestMove string `json:"bestmove"`
	Score    int    `json:"score"`
	Depth    int    `json:"depth"`
	Nodes    uint64 `json:"nodes"`
}

type fenResponse struct {
	FEN        string   `json:"fen"`
	LegalMoves []string `json:"legal_moves"`
	InCheck    bool     `json:"in_check"`
	Checkmate  bool     `json:"checkmate"`
	SideToMove string   `json:"side_to_move"`
}

func newGameHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	game.NewGame()
	log.Printf("[SERVER] newgame")
	w.WriteHeader(http.StatusOK)
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req moveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[MOVE] error endpoint=/move err=%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("[PLAYER] move=%s", req.Move)
	if err := game.PlayMove(req.Move); err != nil {
		log.Printf("[PLAYER] illegal move=%s error=%v", req.Move, err)
		log.Printf("[MOVE] error endpoint=/move err=%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("[PLAYER] applied move=%s", req.Move)
	log.Printf("[STATE] fen=%s", game.FEN())
	w.WriteHeader(http.StatusOK)
}

func engineHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req engineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ENGINE] error endpoint=/engine err=%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("[ENGINE] searching depth=%d", req.Depth)
	best, score, depth, nodes := game.Search.SearchRoot(game.Board, req.Depth)
	if best == 0 {
		log.Printf("[ENGINE] error endpoint=/engine err=%v", "no move found")
		http.Error(w, "no move found", http.StatusBadRequest)
		return
	}
	game.Board.MakeMove(best)
	m := uci.MoveToUCI(best)
	log.Printf("[ENGINE] bestmove=%s", m)
	log.Printf("[STATE] fen=%s", game.FEN())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(engineResponse{
		BestMove: m,
		Score:    score,
		Depth:    depth,
		Nodes:    nodes,
	})
}

func fenHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fen := game.FEN()
	inCheck := false
	checkmate := false
	sideToMove := ""
	legalMoves := []string{}
	if game.Board != nil {
		inCheck = search.InCheck(game.Board)
		var moves [movegen.MaxMoves]uint32
		count := movegen.GenerateLegalMoves(game.Board, &moves)
		if count == 0 && inCheck {
			checkmate = true
		}
		legalMoves = make([]string, 0, count)
		for i := 0; i < count; i++ {
			legalMoves = append(legalMoves, uci.MoveToUCI(moves[i]))
		}
		if game.Board.SideToMove == board.White {
			sideToMove = "white"
		} else {
			sideToMove = "black"
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(fenResponse{
		FEN:        fen,
		LegalMoves: legalMoves,
		InCheck:    inCheck,
		Checkmate:  checkmate,
		SideToMove: sideToMove,
	})
}

func main() {
	http.HandleFunc("/newgame", newGameHandler)
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/engine", engineHandler)
	http.HandleFunc("/fen", fenHandler)

	fileServer := http.FileServer(http.Dir("web"))
	http.Handle("/", fileServer)

	log.Printf("[SERVER] started")
	log.Printf("[SERVER] listening=:8080")
	log.Printf("[SERVER] endpoints: POST /newgame, POST /move, POST /engine, GET /fen")
	log.Printf("[SERVER] webroot=/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
