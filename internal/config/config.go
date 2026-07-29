package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirPerm os.FileMode = 0700

// Config holds paths under the owner-only config directory.
type Config struct {
	Dir         string
	NamePath    string
	TermsPath   string
	SessionPath string
	WinsPath    string
}

// Ensure creates ~/.config/bingo with mode 0700 and returns paths.
func Ensure() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".config", "bingo")
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	if err := os.Chmod(dir, dirPerm); err != nil {
		return nil, fmt.Errorf("chmod config dir: %w", err)
	}
	return &Config{
		Dir:         dir,
		NamePath:    filepath.Join(dir, "name"),
		TermsPath:   filepath.Join(dir, "terms.txt"),
		SessionPath: filepath.Join(dir, "session.json"),
		WinsPath:    filepath.Join(dir, "wins.json"),
	}, nil
}

// ResetAll removes local bingo config files (name, terms, session, wins).
// The config directory itself is kept with mode 0700.
func ResetAll(cfg *Config) error {
	paths := []string{cfg.NamePath, cfg.TermsPath, cfg.SessionPath, cfg.WinsPath}
	for _, p := range paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", p, err)
		}
	}
	if err := os.Chmod(cfg.Dir, dirPerm); err != nil {
		return fmt.Errorf("chmod config dir: %w", err)
	}
	return nil
}
