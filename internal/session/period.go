package session

import (
	"fmt"
	"time"

	"bingo/internal/profile"
)

// now returns the current time; tests may override via SetNowForTest.
var now = time.Now

// SetNowForTest overrides the clock used for period keys. Pass nil to restore.
func SetNowForTest(fn func() time.Time) {
	if fn == nil {
		now = time.Now
		return
	}
	now = fn
}

// PeriodKey returns the current board period key for the given period mode.
func PeriodKey(period profile.Period) string {
	return PeriodKeyAt(period, now())
}

// PeriodKeyAt returns the board period key at t.
// Daily keys are YYYY-MM-DD; weekly keys are ISO week YYYY-Www (Monday start).
func PeriodKeyAt(period profile.Period, t time.Time) string {
	switch period {
	case profile.PeriodWeekly:
		year, week := t.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	default:
		return CalendarDayAt(t)
	}
}

// CalendarDay returns today's date as YYYY-MM-DD (for win history).
func CalendarDay() string {
	return CalendarDayAt(now())
}

// CalendarDayAt formats t as YYYY-MM-DD.
func CalendarDayAt(t time.Time) string {
	return t.Format("2006-01-02")
}
