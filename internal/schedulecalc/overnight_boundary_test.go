package schedulecalc

import (
	"errors"
	"testing"
)

func TestAddMinutesRejectsPastMidnight(t *testing.T) {
	if _, err := AddMinutes(Clock{Hour: 23, Minute: 30}, 60); !errors.Is(err, ErrInvalidClock) {
		t.Fatalf("AddMinutes should reject past-midnight result, got %v", err)
	}
}

func TestDurationMinutesRejectsBackwardRange(t *testing.T) {
	if _, err := DurationMinutes(Clock{Hour: 12}, Clock{Hour: 11}); !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("DurationMinutes should reject backward range, got %v", err)
	}
}

func TestClockRangeFitsRejectsZeroLength(t *testing.T) {
	if ClockRangeFits("09:00", "09:00", "08:00", "20:00") {
		t.Fatal("zero-length shift should not fit")
	}
}
