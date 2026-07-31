package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"bingo/internal/config"
	"bingo/internal/profile"
	"bingo/internal/server"
	"bingo/internal/session"
	"bingo/internal/terms"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	return &config.Config{
		Dir:         dir,
		NamePath:    filepath.Join(dir, "name"),
		TermsPath:   filepath.Join(dir, "terms.txt"),
		SessionPath: filepath.Join(dir, "session.json"),
		WinsPath:    filepath.Join(dir, "wins.json"),
		PeriodPath:  filepath.Join(dir, "period"),
	}
}

func TestAPIs(t *testing.T) {
	cfg := testConfig(t)
	pool, err := terms.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	store := session.NewStore(cfg)
	game, err := store.LoadOrCreate(pool, profile.PeriodDaily)
	if err != nil {
		t.Fatal(err)
	}
	srv := server.New(cfg, store, game, pool)
	h := srv.Handler()

	// State
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/state", nil))
	if res.Code != 200 {
		t.Fatalf("state: %d %s", res.Code, res.Body.String())
	}
	var stateResp struct {
		Period string `json:"period"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &stateResp); err != nil {
		t.Fatal(err)
	}
	if stateResp.Period != "daily" {
		t.Fatalf("state period = %q, want daily", stateResp.Period)
	}

	// Index HTML embedded
	res = httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != 200 {
		t.Fatalf("index: %d", res.Code)
	}
	if !bytes.Contains(res.Body.Bytes(), []byte("root")) {
		t.Fatalf("index missing root")
	}

	// Terms get
	res = httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/terms", nil))
	if res.Code != 200 {
		t.Fatalf("terms get: %d %s", res.Code, res.Body.String())
	}
	var termsResp struct {
		Count  int  `json:"count"`
		Custom bool `json:"custom"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &termsResp); err != nil {
		t.Fatal(err)
	}
	if termsResp.Count < 24 || termsResp.Custom {
		t.Fatalf("unexpected terms: %+v", termsResp)
	}

	// Add term
	body, _ := json.Marshal(map[string]any{"terms": []string{"ZZ Unique Test Term"}})
	res = httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/terms/add", bytes.NewReader(body)))
	if res.Code != 200 {
		t.Fatalf("terms add: %d %s", res.Code, res.Body.String())
	}
	if _, err := os.Stat(cfg.TermsPath); err != nil {
		t.Fatalf("custom terms not created: %v", err)
	}

	// Name
	body, _ = json.Marshal(map[string]string{"name": "Tester"})
	res = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/name", bytes.NewReader(body))
	h.ServeHTTP(res, req)
	if res.Code != 200 {
		t.Fatalf("name: %d %s", res.Code, res.Body.String())
	}

	// Period — preference change must not reshuffle the board
	beforeCells := append([]string(nil), game.Cells...)
	body, _ = json.Marshal(map[string]string{"period": "weekly"})
	res = httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodPut, "/api/period", bytes.NewReader(body)))
	if res.Code != 200 {
		t.Fatalf("period: %d %s", res.Code, res.Body.String())
	}
	var periodResp struct {
		Period string   `json:"period"`
		Cells  []string `json:"cells"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &periodResp); err != nil {
		t.Fatal(err)
	}
	if periodResp.Period != "weekly" {
		t.Fatalf("period response = %q, want weekly", periodResp.Period)
	}
	if len(periodResp.Cells) != len(beforeCells) {
		t.Fatalf("period change altered cell count")
	}
	for i := range beforeCells {
		if periodResp.Cells[i] != beforeCells[i] {
			t.Fatalf("period change reshuffled board at %d: %q -> %q", i, beforeCells[i], periodResp.Cells[i])
		}
	}

	// Reset terms
	res = httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/terms/reset", nil))
	if res.Code != 200 {
		t.Fatalf("terms reset: %d %s", res.Code, res.Body.String())
	}
}
