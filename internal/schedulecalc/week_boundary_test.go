package schedulecalc

import (
	"testing"
	"time"
)

func TestWeekdayIndexMondayBased(t *testing.T) {
	monday := time.Date(2026, 8, 17, 0, 0, 0, 0, time.Local)
	sunday := time.Date(2026, 8, 23, 0, 0, 0, 0, time.Local)
	if got := WeekdayIndex(monday); got != 0 {
		t.Fatalf("Monday index = %d, want 0", got)
	}
	if got := WeekdayIndex(sunday); got != 6 {
		t.Fatalf("Sunday index = %d, want 6", got)
	}
}

func TestParseWeekRangeNormalizesToMonday(t *testing.T) {
	start, end, err := ParseWeekRange("2026-08-19")
	if err != nil {
		t.Fatalf("ParseWeekRange returned error: %v", err)
	}
	if got := DateKey(start); got != "2026-08-17" {
		t.Fatalf("week start = %s, want 2026-08-17", got)
	}
	if got := DateKey(end); got != "2026-08-23" {
		t.Fatalf("week end = %s, want 2026-08-23", got)
	}
}

func TestParseClockAcceptsMidnight(t *testing.T) {
	if _, err := ParseClock("00:00"); err != nil {
		t.Fatalf("ParseClock should accept midnight, got %v", err)
	}
}

func TestClockRangeFitsRejectsZeroLength(t *testing.T) {
	if ClockRangeFits("09:00", "09:00", "08:00", "20:00") {
		t.Fatal("zero-length shift should not fit")
	}
}
