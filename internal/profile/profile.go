package profile

import (
	"fmt"
	"os"
	"strings"

	"bingo/internal/config"
)

// Get returns the configured player name, or empty if unset.
func Get(cfg *config.Config) (string, error) {
	raw, err := os.ReadFile(cfg.NamePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read name: %w", err)
	}
	return strings.TrimSpace(string(raw)), nil
}

// Set writes the player name.
func Set(cfg *config.Config, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("name must be 64 characters or fewer")
	}
	if err := os.WriteFile(cfg.NamePath, []byte(name+"\n"), 0600); err != nil {
		return fmt.Errorf("write name: %w", err)
	}
	return nil
}
