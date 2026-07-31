package profile

import (
	"fmt"
	"os"
	"strings"

	"bingo/internal/config"
)

// Period controls how often a new board is created.
type Period string

const (
	PeriodDaily  Period = "daily"
	PeriodWeekly Period = "weekly"
)

// ParsePeriod validates and normalizes a period string.
func ParsePeriod(s string) (Period, error) {
	switch Period(strings.ToLower(strings.TrimSpace(s))) {
	case PeriodDaily:
		return PeriodDaily, nil
	case PeriodWeekly:
		return PeriodWeekly, nil
	default:
		return "", fmt.Errorf("period must be %q or %q", PeriodDaily, PeriodWeekly)
	}
}

// GetPeriod returns the configured reset period, defaulting to daily if unset.
func GetPeriod(cfg *config.Config) (Period, error) {
	raw, err := os.ReadFile(cfg.PeriodPath)
	if err != nil {
		if os.IsNotExist(err) {
			return PeriodDaily, nil
		}
		return "", fmt.Errorf("read period: %w", err)
	}
	s := strings.TrimSpace(string(raw))
	if s == "" {
		return PeriodDaily, nil
	}
	p, err := ParsePeriod(s)
	if err != nil {
		return PeriodDaily, nil
	}
	return p, nil
}

// SetPeriod writes the board reset period.
func SetPeriod(cfg *config.Config, period Period) error {
	p, err := ParsePeriod(string(period))
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfg.PeriodPath, []byte(string(p)+"\n"), 0600); err != nil {
		return fmt.Errorf("write period: %w", err)
	}
	return nil
}
