package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"sync"

	"bingo/internal/config"
	"bingo/internal/profile"
	"bingo/internal/session"
	"bingo/internal/terms"
)

// Server serves the bingo UI and JSON API.
type Server struct {
	cfg   *config.Config
	store *session.Store
	game  *session.Game
	pool  []string
	mu    sync.Mutex
}

func New(cfg *config.Config, store *session.Store, game *session.Game, pool []string) *Server {
	return &Server{cfg: cfg, store: store, game: game, pool: pool}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", s.handleState)
	mux.HandleFunc("POST /api/mark", s.handleMark)
	mux.HandleFunc("POST /api/reset", s.handleReset)
	mux.HandleFunc("GET /api/terms", s.handleTermsGet)
	mux.HandleFunc("POST /api/terms/add", s.handleTermsAdd)
	mux.HandleFunc("POST /api/terms/remove", s.handleTermsRemove)
	mux.HandleFunc("PUT /api/terms", s.handleTermsReplace)
	mux.HandleFunc("POST /api/terms/reset", s.handleTermsReset)
	mux.HandleFunc("PUT /api/name", s.handleName)
	mux.HandleFunc("POST /api/reset-all", s.handleResetAll)

	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))
	return mux
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.store.State(s.game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, st)
}

type markRequest struct {
	Index int `json:"index"`
}

func (s *Server) handleMark(w http.ResponseWriter, r *http.Request) {
	var req markRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.store.ToggleMark(s.game, req.Index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, st)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.store.ResetTally(s.game, s.pool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, st)
}

type termsResponse struct {
	Terms  []string `json:"terms"`
	Custom bool     `json:"custom"`
	Count  int      `json:"count"`
}

type termsListRequest struct {
	Terms []string `json:"terms"`
}

func (s *Server) handleTermsGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeTerms(w)
}

func (s *Server) handleTermsAdd(w http.ResponseWriter, r *http.Request) {
	var req termsListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := terms.Add(s.cfg, req.Terms...); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.reloadPool(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeTerms(w)
}

func (s *Server) handleTermsRemove(w http.ResponseWriter, r *http.Request) {
	var req termsListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := terms.Remove(s.cfg, req.Terms...); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.reloadPool(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeTerms(w)
}

func (s *Server) handleTermsReplace(w http.ResponseWriter, r *http.Request) {
	var req termsListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := terms.Replace(s.cfg, req.Terms); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.reloadPool(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeTerms(w)
}

func (s *Server) handleTermsReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := terms.Reset(s.cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.reloadPool(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeTerms(w)
}

type nameRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleName(w http.ResponseWriter, r *http.Request) {
	var req nameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := profile.Set(s.cfg, req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	st, err := s.store.State(s.game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, st)
}

func (s *Server) handleResetAll(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := config.ResetAll(s.cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.reloadPool(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	game, err := s.store.LoadOrCreate(s.pool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	*s.game = *game
	st, err := s.store.State(s.game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, st)
}

func (s *Server) reloadPool() error {
	pool, err := terms.Load(s.cfg)
	if err != nil {
		return err
	}
	s.pool = pool
	return nil
}

func (s *Server) writeTerms(w http.ResponseWriter) {
	pool, custom, err := terms.Snapshot(s.cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pool == nil {
		pool = []string{}
	}
	writeJSON(w, termsResponse{
		Terms:  pool,
		Custom: custom,
		Count:  len(pool),
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
