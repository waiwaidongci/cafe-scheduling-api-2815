package schedulecalc

import "testing"

func TestMergeOverlapsMergesContinuousSpans(t *testing.T) {
	spans := []Span{
		{ID: 1, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "08:00", End: "12:00"},
		{ID: 2, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "11:00", End: "14:00"},
		{ID: 3, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "16:00", End: "18:00"},
	}

	merged := MergeOverlaps(spans)
	if len(merged) != 2 {
		t.Fatalf("expected 2 merged spans, got %d: %+v", len(merged), merged)
	}
	if merged[0].Start != "08:00" || merged[0].End != "14:00" {
		t.Fatalf("first merged span should be 08:00-14:00, got %s-%s", merged[0].Start, merged[0].End)
	}
	if merged[1].Start != "16:00" || merged[1].End != "18:00" {
		t.Fatalf("second merged span should be 16:00-18:00, got %s-%s", merged[1].Start, merged[1].End)
	}
}

func TestDetectOverlapsIncludesAllConflictingSpans(t *testing.T) {
	spans := []Span{
		{ID: 1, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "08:00", End: "12:00"},
		{ID: 2, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "11:00", End: "14:00"},
		{ID: 3, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "10:00", End: "13:00"},
	}

	conflicts := DetectOverlaps(spans)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if len(conflicts[0].SpanIDs) != 3 {
		t.Fatalf("expected 3 involved spans, got %v", conflicts[0].SpanIDs)
	}
}

func TestOverlapFreeWeeklyHoursDoesNotDoubleCount(t *testing.T) {
	spans := []Span{
		{ID: 1, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "08:00", End: "12:00"},
		{ID: 2, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", Start: "11:00", End: "14:00"},
	}

	hours := OverlapFreeWeeklyHours(spans)
	if len(hours) != 1 {
		t.Fatalf("expected 1 employee row, got %d", len(hours))
	}
	if hours[0].TotalMinutes != 360 {
		t.Fatalf("expected 360 unique minutes, got %d", hours[0].TotalMinutes)
	}
}
