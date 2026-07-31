package session

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"bingo/internal/config"
	"bingo/internal/profile"
)

const (
	Size       = 5
	CellCount  = Size * Size
	FreeIndex  = 12
	FreeLabel  = "FREE"
)

// Game is the bingo board and mark state for the current period.
type Game struct {
	Date   string   `json:"date"`
	Cells  []string `json:"cells"`
	Marked []bool   `json:"marked"`
	Won    bool     `json:"won"`
}

// Wins tracks total bingos and unique win days.
type Wins struct {
	Count int      `json:"count"`
	Dates []string `json:"dates"`
}

// State is the API payload for the UI.
type State struct {
	Name     string   `json:"name"`
	Date     string   `json:"date"`
	Cells    []string `json:"cells"`
	Marked   []bool   `json:"marked"`
	Won      bool     `json:"won"`
	WinCount int      `json:"winCount"`
	Period   string   `json:"period"`
}

// Store persists session and wins under the config directory.
type Store struct {
	cfg *config.Config
}

func NewStore(cfg *config.Config) *Store {
	return &Store{cfg: cfg}
}

// LoadOrCreate returns the session for the current period, creating a new board if needed.
func (s *Store) LoadOrCreate(pool []string, period profile.Period) (*Game, error) {
	key := PeriodKey(period)
	if g, err := s.readSession(); err == nil && g.Date == key && len(g.Cells) == CellCount && len(g.Marked) == CellCount {
		return g, nil
	}
	g, err := newGame(key, pool)
	if err != nil {
		return nil, err
	}
	if err := s.writeSession(g); err != nil {
		return nil, err
	}
	return g, nil
}

func newGame(date string, pool []string) (*Game, error) {
	if len(pool) < 24 {
		return nil, fmt.Errorf("need at least 24 terms to build a board")
	}
	shuffled := append([]string(nil), pool...)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	cells := make([]string, CellCount)
	marked := make([]bool, CellCount)
	idx := 0
	for i := 0; i < CellCount; i++ {
		if i == FreeIndex {
			cells[i] = FreeLabel
			marked[i] = true
			continue
		}
		cells[i] = shuffled[idx]
		idx++
	}
	return &Game{Date: date, Cells: cells, Marked: marked, Won: false}, nil
}

// ToggleMark flips a cell (center stays marked), checks win, persists.
func (s *Store) ToggleMark(g *Game, index int, period profile.Period) (*State, error) {
	if index < 0 || index >= CellCount {
		return nil, fmt.Errorf("cell index out of range: %d", index)
	}
	if index != FreeIndex {
		g.Marked[index] = !g.Marked[index]
	} else {
		g.Marked[FreeIndex] = true
	}

	if !g.Won && isBingo(g.Marked) {
		g.Won = true
		if err := s.recordWin(CalendarDay()); err != nil {
			return nil, err
		}
	}

	if err := s.writeSession(g); err != nil {
		return nil, err
	}
	return s.State(g, period)
}

// ResetTally reshuffles the current board from the term pool and clears marks/won.
// Wins history is left unchanged.
func (s *Store) ResetTally(g *Game, pool []string, period profile.Period) (*State, error) {
	key := g.Date
	if key == "" {
		key = PeriodKey(period)
	}
	next, err := newGame(key, pool)
	if err != nil {
		return nil, err
	}
	*g = *next
	if err := s.writeSession(g); err != nil {
		return nil, err
	}
	return s.State(g, period)
}

// RetagForPeriod updates the board's period key to match period without reshuffling
// cells or clearing marks. Used when the player changes daily/weekly preference.
func (s *Store) RetagForPeriod(g *Game, period profile.Period) (*State, error) {
	key := PeriodKey(period)
	if g.Date != key {
		g.Date = key
		if err := s.writeSession(g); err != nil {
			return nil, err
		}
	}
	return s.State(g, period)
}

// State builds the UI state including win count, player name, and reset period.
func (s *Store) State(g *Game, period profile.Period) (*State, error) {
	wins, err := s.readWins()
	if err != nil {
		return nil, err
	}
	name, err := profile.Get(s.cfg)
	if err != nil {
		return nil, err
	}
	return &State{
		Name:     name,
		Date:     g.Date,
		Cells:    g.Cells,
		Marked:   g.Marked,
		Won:      g.Won,
		WinCount: wins.total(),
		Period:   string(period),
	}, nil
}

func isBingo(marked []bool) bool {
	// Rows
	for r := 0; r < Size; r++ {
		ok := true
		for c := 0; c < Size; c++ {
			if !marked[r*Size+c] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	// Columns
	for c := 0; c < Size; c++ {
		ok := true
		for r := 0; r < Size; r++ {
			if !marked[r*Size+c] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	// Diagonals
	ok := true
	for i := 0; i < Size; i++ {
		if !marked[i*Size+i] {
			ok = false
			break
		}
	}
	if ok {
		return true
	}
	ok = true
	for i := 0; i < Size; i++ {
		if !marked[i*Size+(Size-1-i)] {
			ok = false
			break
		}
	}
	return ok
}

func (s *Store) recordWin(date string) error {
	wins, err := s.readWins()
	if err != nil {
		return err
	}
	wins.Count = wins.total() + 1
	seen := false
	for _, d := range wins.Dates {
		if d == date {
			seen = true
			break
		}
	}
	if !seen {
		wins.Dates = append(wins.Dates, date)
	}
	return s.writeWins(wins)
}

func (w *Wins) total() int {
	if w.Count > 0 {
		return w.Count
	}
	// Migrate older wins.json that only stored unique dates.
	return len(w.Dates)
}

func (s *Store) readSession() (*Game, error) {
	raw, err := os.ReadFile(s.cfg.SessionPath)
	if err != nil {
		return nil, err
	}
	var g Game
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Store) writeSession(g *Game) error {
	raw, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.cfg.SessionPath, raw, 0600)
}

func (s *Store) readWins() (*Wins, error) {
	raw, err := os.ReadFile(s.cfg.WinsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Wins{Dates: []string{}}, nil
		}
		return nil, err
	}
	var w Wins
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, err
	}
	if w.Dates == nil {
		w.Dates = []string{}
	}
	if w.Count == 0 && len(w.Dates) > 0 {
		w.Count = len(w.Dates)
	}
	return &w, nil
}

func (s *Store) writeWins(w *Wins) error {
	raw, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.cfg.WinsPath, raw, 0600)
}
