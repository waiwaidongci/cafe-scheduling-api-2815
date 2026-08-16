package schedulecalc

import (
	"errors"
	"testing"
)

func TestErrorSentinelPreservedForInvalidRanges(t *testing.T) {
	if _, err := DurationMinutes(Clock{Hour: 12}, Clock{Hour: 10}); !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("DurationMinutes should preserve ErrInvalidTimeRange, got %v", err)
	}

	if _, err := NewInterval("10:00", "09:00"); !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("NewInterval should preserve ErrInvalidTimeRange, got %v", err)
	}

	if _, _, err := ParseWeekRange("not-a-date"); !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("ParseWeekRange should preserve ErrInvalidDateRange, got %v", err)
	}

	if ClockRangeFits("08:00", "09:00", "10:00", "20:00") {
		t.Fatal("shift before opening time should not fit")
	}
	if !ClockRangeFits("09:00", "17:00", "08:00", "20:00") {
		t.Fatal("shift inside business hours should fit")
	}
}
