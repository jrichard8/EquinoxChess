package tt

const (
	TTExact uint8 = iota
	TTAlpha
	TTBeta
)

type TTEntry struct {
	Key   uint64
	Depth int
	Score int
	Flag  uint8
	Move  uint32
}

type TranspositionTable struct {
	Entries []TTEntry
	Mask    uint64
}

func NewTT(size int) *TranspositionTable {
	if size <= 0 || (size&(size-1)) != 0 {
		size = 1 << 20
	}
	entries := make([]TTEntry, size)
	return &TranspositionTable{
		Entries: entries,
		Mask:    uint64(size - 1),
	}
}

func (tt *TranspositionTable) Probe(key uint64, depth int, alpha int, beta int, ply int, mateThreshold int) (score int, move uint32, cutoff bool) {
	if tt == nil || len(tt.Entries) == 0 {
		return 0, 0, false
	}
	entry := &tt.Entries[key&tt.Mask]
	if entry.Key != key {
		return 0, 0, false
	}
	move = entry.Move
	if entry.Depth < depth {
		return 0, move, false
	}

	// Denormalize mate scores relative to current ply
	adjustedScore := entry.Score
	if adjustedScore > mateThreshold {
		adjustedScore -= ply
	} else if adjustedScore < -mateThreshold {
		adjustedScore += ply
	}

	switch entry.Flag {
	case TTExact:
		return adjustedScore, move, true
	case TTAlpha:
		if adjustedScore <= alpha {
			return adjustedScore, move, true
		}
	case TTBeta:
		if adjustedScore >= beta {
			return adjustedScore, move, true
		}
	}
	return 0, move, false
}

func (tt *TranspositionTable) Store(key uint64, depth int, score int, flag uint8, move uint32, ply int, mateThreshold int) {
	if tt == nil || len(tt.Entries) == 0 {
		return
	}
	entry := &tt.Entries[key&tt.Mask]
	if entry.Key != 0 && entry.Key != key && flag != TTExact && depth < entry.Depth {
		return
	}

	// Normalize mate scores relative to current ply before storing
	adjustedScore := score
	if adjustedScore > mateThreshold {
		adjustedScore += ply
	} else if adjustedScore < -mateThreshold {
		adjustedScore -= ply
	}

	entry.Key = key
	entry.Depth = depth
	entry.Score = adjustedScore
	entry.Flag = flag
	entry.Move = move
}
