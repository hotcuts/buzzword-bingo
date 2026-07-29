package terms

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"bingo/data"
	"bingo/internal/config"
)

const MinTerms = 24

// Load returns custom terms from the config dir if present, otherwise embedded defaults.
func Load(cfg *config.Config) ([]string, error) {
	pool, source, err := loadPool(cfg)
	if err != nil {
		return nil, err
	}
	if len(pool) < MinTerms {
		return nil, fmt.Errorf("need at least %d terms, found %d in %s", MinTerms, len(pool), source)
	}
	return pool, nil
}

// CustomExists reports whether a user terms file is configured.
func CustomExists(cfg *config.Config) bool {
	_, err := os.Stat(cfg.TermsPath)
	return err == nil
}

// SetFile installs or replaces the custom terms file from path.
func SetFile(cfg *config.Config, path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read terms file: %w", err)
	}
	pool := parse(raw)
	if len(pool) < MinTerms {
		return 0, fmt.Errorf("need at least %d terms, found %d in %s", MinTerms, len(pool), path)
	}
	if err := os.WriteFile(cfg.TermsPath, raw, 0600); err != nil {
		return 0, fmt.Errorf("write terms: %w", err)
	}
	return len(pool), nil
}

// Add appends unique terms to the custom file, seeding from defaults if needed.
func Add(cfg *config.Config, termsToAdd ...string) (int, error) {
	if len(termsToAdd) == 0 {
		return 0, fmt.Errorf("no terms provided")
	}
	pool, err := ensureMutable(cfg)
	if err != nil {
		return 0, err
	}

	seen := make(map[string]struct{}, len(pool))
	for _, t := range pool {
		seen[t] = struct{}{}
	}

	added := 0
	for _, t := range termsToAdd {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		pool = append(pool, t)
		seen[t] = struct{}{}
		added++
	}
	if added == 0 {
		return len(pool), fmt.Errorf("no new terms to add (already present or empty)")
	}
	if err := savePool(cfg, pool); err != nil {
		return 0, err
	}
	return len(pool), nil
}

// Remove deletes terms from the custom file, seeding from defaults if needed.
func Remove(cfg *config.Config, termsToRemove ...string) (int, error) {
	if len(termsToRemove) == 0 {
		return 0, fmt.Errorf("no terms provided")
	}
	pool, err := ensureMutable(cfg)
	if err != nil {
		return 0, err
	}

	drop := make(map[string]struct{}, len(termsToRemove))
	for _, t := range termsToRemove {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		drop[t] = struct{}{}
	}
	if len(drop) == 0 {
		return 0, fmt.Errorf("no terms provided")
	}

	for t := range drop {
		found := false
		for _, p := range pool {
			if p == t {
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("term not found: %q", t)
		}
	}

	next := make([]string, 0, len(pool))
	for _, p := range pool {
		if _, ok := drop[p]; !ok {
			next = append(next, p)
		}
	}
	if len(next) < MinTerms {
		return 0, fmt.Errorf("removing those terms would leave %d terms (need at least %d)", len(next), MinTerms)
	}
	if err := savePool(cfg, next); err != nil {
		return 0, err
	}
	return len(next), nil
}

// Reset deletes the custom terms file so play uses embedded defaults.
func Reset(cfg *config.Config) error {
	if !CustomExists(cfg) {
		return fmt.Errorf("already using embedded defaults (no custom terms file)")
	}
	if err := os.Remove(cfg.TermsPath); err != nil {
		return fmt.Errorf("remove terms: %w", err)
	}
	return nil
}

// Replace writes a full custom terms list (deduped, trimmed).
func Replace(cfg *config.Config, list []string) (int, error) {
	seen := make(map[string]struct{}, len(list))
	pool := make([]string, 0, len(list))
	for _, t := range list {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		pool = append(pool, t)
	}
	if len(pool) < MinTerms {
		return 0, fmt.Errorf("need at least %d terms, found %d", MinTerms, len(pool))
	}
	if err := savePool(cfg, pool); err != nil {
		return 0, err
	}
	return len(pool), nil
}

// Snapshot returns the current term pool and whether a custom file is in use.
func Snapshot(cfg *config.Config) (pool []string, custom bool, err error) {
	pool, _, err = loadPool(cfg)
	if err != nil {
		return nil, false, err
	}
	return pool, CustomExists(cfg), nil
}

func loadPool(cfg *config.Config) ([]string, string, error) {
	if CustomExists(cfg) {
		raw, err := os.ReadFile(cfg.TermsPath)
		if err != nil {
			return nil, cfg.TermsPath, fmt.Errorf("read terms: %w", err)
		}
		return parse(raw), cfg.TermsPath, nil
	}
	return parse(data.DefaultTerms), "embedded defaults", nil
}

func ensureMutable(cfg *config.Config) ([]string, error) {
	if CustomExists(cfg) {
		raw, err := os.ReadFile(cfg.TermsPath)
		if err != nil {
			return nil, fmt.Errorf("read terms: %w", err)
		}
		return parse(raw), nil
	}
	pool := parse(data.DefaultTerms)
	if err := savePool(cfg, pool); err != nil {
		return nil, err
	}
	return pool, nil
}

func savePool(cfg *config.Config, pool []string) error {
	if len(pool) < MinTerms {
		return fmt.Errorf("need at least %d terms, found %d", MinTerms, len(pool))
	}
	var b strings.Builder
	b.WriteString("# Custom bingo terms\n")
	for _, t := range pool {
		b.WriteString(t)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(cfg.TermsPath, []byte(b.String()), 0600); err != nil {
		return fmt.Errorf("write terms: %w", err)
	}
	return nil
}

func parse(raw []byte) []string {
	var out []string
	sc := bufio.NewScanner(bytes.NewReader(raw))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}
