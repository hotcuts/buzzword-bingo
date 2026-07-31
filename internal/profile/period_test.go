package profile

import (
	"os"
	"path/filepath"
	"testing"

	"bingo/internal/config"
)

func TestParsePeriod(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want Period
		ok   bool
	}{
		{"daily", PeriodDaily, true},
		{"weekly", PeriodWeekly, true},
		{"DAILY", PeriodDaily, true},
		{" Weekly ", PeriodWeekly, true},
		{"", "", false},
		{"monthly", "", false},
	}
	for _, tc := range cases {
		got, err := ParsePeriod(tc.in)
		if tc.ok {
			if err != nil {
				t.Fatalf("ParsePeriod(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("ParsePeriod(%q) = %q, want %q", tc.in, got, tc.want)
			}
			continue
		}
		if err == nil {
			t.Fatalf("ParsePeriod(%q) = %q, want error", tc.in, got)
		}
	}
}

func TestGetPeriodDefaultsDaily(t *testing.T) {
	t.Parallel()
	cfg := testCfg(t)
	got, err := GetPeriod(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got != PeriodDaily {
		t.Fatalf("GetPeriod = %q, want daily", got)
	}
}

func TestSetAndGetPeriod(t *testing.T) {
	t.Parallel()
	cfg := testCfg(t)
	if err := SetPeriod(cfg, PeriodWeekly); err != nil {
		t.Fatal(err)
	}
	got, err := GetPeriod(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got != PeriodWeekly {
		t.Fatalf("GetPeriod = %q, want weekly", got)
	}
	raw, err := os.ReadFile(cfg.PeriodPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "weekly\n" {
		t.Fatalf("file contents = %q", raw)
	}
}

func TestSetPeriodRejectsInvalid(t *testing.T) {
	t.Parallel()
	cfg := testCfg(t)
	if err := SetPeriod(cfg, Period("monthly")); err == nil {
		t.Fatal("expected error for monthly")
	}
}

func testCfg(t *testing.T) *config.Config {
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
