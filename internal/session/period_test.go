package session

import (
	"testing"
	"time"

	"bingo/internal/profile"
)

func TestPeriodKeyDaily(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 31, 15, 30, 0, 0, time.Local)
	got := PeriodKeyAt(profile.PeriodDaily, now)
	if got != "2026-07-31" {
		t.Fatalf("daily key = %q, want 2026-07-31", got)
	}
}

func TestPeriodKeyWeeklyMondayBoundary(t *testing.T) {
	t.Parallel()
	// Sunday 2026-08-02 is still ISO week 31; Monday 2026-08-03 starts week 32.
	sun := time.Date(2026, 8, 2, 23, 0, 0, 0, time.Local)
	mon := time.Date(2026, 8, 3, 0, 0, 0, 0, time.Local)
	if got := PeriodKeyAt(profile.PeriodWeekly, sun); got != "2026-W31" {
		t.Fatalf("Sunday key = %q, want 2026-W31", got)
	}
	if got := PeriodKeyAt(profile.PeriodWeekly, mon); got != "2026-W32" {
		t.Fatalf("Monday key = %q, want 2026-W32", got)
	}
}

func TestPeriodKeyWeeklyYearEndISO(t *testing.T) {
	t.Parallel()
	dec31 := time.Date(2026, 12, 31, 12, 0, 0, 0, time.Local)
	jan1 := time.Date(2027, 1, 1, 12, 0, 0, 0, time.Local)
	jan4 := time.Date(2027, 1, 4, 12, 0, 0, 0, time.Local)

	if got := PeriodKeyAt(profile.PeriodWeekly, dec31); got != "2026-W53" {
		t.Fatalf("Dec 31 key = %q, want 2026-W53", got)
	}
	if got := PeriodKeyAt(profile.PeriodWeekly, jan1); got != "2026-W53" {
		t.Fatalf("Jan 1 key = %q, want 2026-W53", got)
	}
	if got := PeriodKeyAt(profile.PeriodWeekly, jan4); got != "2027-W01" {
		t.Fatalf("Jan 4 key = %q, want 2027-W01", got)
	}
}

func TestCalendarDay(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 31, 3, 0, 0, 0, time.Local)
	if got := CalendarDayAt(now); got != "2026-07-31" {
		t.Fatalf("CalendarDayAt = %q, want 2026-07-31", got)
	}
}
