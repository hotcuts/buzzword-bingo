package updatecheck_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"bingo/internal/updatecheck"
)

func writeCache(t *testing.T, path string, checkedAt time.Time, latestTag string) {
	t.Helper()
	data, err := json.Marshal(map[string]any{
		"checked_at": checkedAt.UTC().Format(time.RFC3339),
		"latest_tag": latestTag,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestMaybeNotify_newerVersionPrintsSuggestion(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/repos/hotcuts/buzzword-bingo/releases/latest" {
			t.Errorf("path = %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v0.2.0"})
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	cachePath := filepath.Join(dir, "update-check.json")
	var out bytes.Buffer
	c := &updatecheck.Checker{
		Client:     srv.Client(),
		APIBaseURL: srv.URL,
		CachePath:  cachePath,
		Out:        &out,
		Now:        func() time.Time { return time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC) },
	}
	c.MaybeNotify("0.1.0")

	got := out.String()
	if !strings.Contains(got, "Update available: v0.2.0") {
		t.Fatalf("missing update line: %q", got)
	}
	if !strings.Contains(got, "you have v0.1.0") {
		t.Fatalf("missing current version: %q", got)
	}
	if !strings.Contains(got, "bingo update") {
		t.Fatalf("missing update command: %q", got)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d, want 1", hits.Load())
	}

	raw, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte(`"latest_tag":"v0.2.0"`)) && !bytes.Contains(raw, []byte(`"latest_tag": "v0.2.0"`)) {
		t.Fatalf("cache not updated: %s", raw)
	}
}

func TestMaybeNotify_sameVersionQuiet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v0.1.0"})
	}))
	t.Cleanup(srv.Close)

	var out bytes.Buffer
	c := &updatecheck.Checker{
		Client:     srv.Client(),
		APIBaseURL: srv.URL,
		CachePath:  filepath.Join(t.TempDir(), "update-check.json"),
		Out:        &out,
		Now:        func() time.Time { return time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC) },
	}
	c.MaybeNotify("0.1.0")

	if out.Len() != 0 {
		t.Fatalf("expected quiet, got %q", out.String())
	}
}

func TestMaybeNotify_httpErrorQuiet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	var out bytes.Buffer
	c := &updatecheck.Checker{
		Client:     srv.Client(),
		APIBaseURL: srv.URL,
		CachePath:  filepath.Join(t.TempDir(), "update-check.json"),
		Out:        &out,
		Now:        func() time.Time { return time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC) },
	}
	c.MaybeNotify("0.1.0")

	if out.Len() != 0 {
		t.Fatalf("expected quiet, got %q", out.String())
	}
}

func TestMaybeNotify_cacheHitSkipsNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v9.9.9"})
	}))
	t.Cleanup(srv.Close)

	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	cachePath := filepath.Join(t.TempDir(), "update-check.json")
	writeCache(t, cachePath, now.Add(-1*time.Hour), "v0.2.0")

	var out bytes.Buffer
	c := &updatecheck.Checker{
		Client:     srv.Client(),
		APIBaseURL: srv.URL,
		CachePath:  cachePath,
		Out:        &out,
		Now:        func() time.Time { return now },
	}
	c.MaybeNotify("0.1.0")

	if hits.Load() != 0 {
		t.Fatalf("expected no network, hits = %d", hits.Load())
	}
	if !strings.Contains(out.String(), "Update available: v0.2.0") {
		t.Fatalf("expected cached suggestion: %q", out.String())
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		remote, local string
		want          bool
	}{
		{"v0.2.0", "0.1.0", true},
		{"0.2.0", "v0.1.0", true},
		{"0.1.0", "0.1.0", false},
		{"v0.1.0", "0.2.0", false},
		{"1.0.0", "0.9.9", true},
		{"0.1.10", "0.1.9", true},
		{"bad", "0.1.0", false},
		{"0.1.0", "bad", false},
	}
	for _, tc := range cases {
		if got := updatecheck.IsNewer(tc.remote, tc.local); got != tc.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tc.remote, tc.local, got, tc.want)
		}
	}
}
