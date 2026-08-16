package schedulecalc

import "testing"

func TestWeeklyHoursDoesNotPanicOnPopulatedInput(t *testing.T) {
	spans := []Span{
		{ID: 1, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "08:00", End: "12:00"},
	}
	hours := WeeklyHours(spans)
	if len(hours) != 1 {
		t.Fatalf("expected 1 employee row, got %d", len(hours))
	}
	if hours[0].TotalMinutes != 240 {
		t.Fatalf("expected 240 minutes, got %d", hours[0].TotalMinutes)
	}
}

func TestOverlapFreeWeeklyHoursDoesNotPanic(t *testing.T) {
	spans := []Span{
		{ID: 1, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "08:00", End: "12:00"},
		{ID: 2, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "11:00", End: "14:00"},
	}
	hours := OverlapFreeWeeklyHours(spans)
	if len(hours) != 1 || hours[0].TotalMinutes != 360 {
		t.Fatalf("unexpected hours: %+v", hours)
	}
}

func TestMergeOverlapsDoesNotPanic(t *testing.T) {
	spans := []Span{
		{ID: 1, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "08:00", End: "12:00"},
		{ID: 2, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "11:00", End: "14:00"},
	}
	if got := len(MergeOverlaps(spans)); got != 1 {
		t.Fatalf("expected 1 merged span, got %d", got)
	}
}
