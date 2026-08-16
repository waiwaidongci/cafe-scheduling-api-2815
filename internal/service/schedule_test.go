package service

import (
	"testing"

	"cafe-scheduling-api/internal/model"
)

func TestParseWeekRangeNormalizesToMonday(t *testing.T) {
	start, end, err := parseWeekRange("2026-08-19")
	if err != nil {
		t.Fatalf("parseWeekRange returned error: %v", err)
	}
	if start.Format("2006-01-02") != "2026-08-17" {
		t.Fatalf("expected normalized Monday, got %s", start.Format("2006-01-02"))
	}
	if end.Format("2006-01-02") != "2026-08-23" {
		t.Fatalf("expected Sunday, got %s", end.Format("2006-01-02"))
	}
}

func TestDetectOverlaps(t *testing.T) {
	shifts := []model.Shift{
		{ID: 1, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", StartTime: "08:00", EndTime: "12:00"},
		{ID: 2, EmployeeID: 10, EmployeeName: "Alice", WorkDate: "2026-08-17", StartTime: "11:00", EndTime: "14:00"},
		{ID: 4, EmployeeID: 20, EmployeeName: "Bob", WorkDate: "2026-08-17", StartTime: "08:00", EndTime: "12:00"},
	}

	conflicts := detectOverlaps(shifts)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].EmployeeID != 10 {
		t.Fatalf("expected employee 10, got %d", conflicts[0].EmployeeID)
	}
	if len(conflicts[0].Shifts) != 2 {
		t.Fatalf("expected 2 involved shifts, got %d", len(conflicts[0].Shifts))
	}
}

func TestCalculateWeeklyHours(t *testing.T) {
	shifts := []model.Shift{
		{EmployeeID: 10, EmployeeName: "Alice", StartTime: "08:00", EndTime: "12:00"},
		{EmployeeID: 10, EmployeeName: "Alice", StartTime: "13:00", EndTime: "17:00"},
		{EmployeeID: 20, EmployeeName: "Bob", StartTime: "09:00", EndTime: "12:30"},
	}

	stats := calculateWeeklyHours(shifts)
	if len(stats) != 2 {
		t.Fatalf("expected 2 employees, got %d", len(stats))
	}
	if stats[0].EmployeeID != 10 || stats[0].TotalHours != 8 {
		t.Fatalf("unexpected Alice stats: %+v", stats[0])
	}
	if stats[1].EmployeeID != 20 || stats[1].TotalHours != 3.5 {
		t.Fatalf("unexpected Bob stats: %+v", stats[1])
	}
}

func TestParseClock(t *testing.T) {
	tests := []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{raw: "08:00", want: 480},
		{raw: "23:59", want: 1439},
		{raw: "8:00", wantErr: true},
		{raw: "24:00", wantErr: true},
		{raw: "08:60", wantErr: true},
		{raw: "0800", wantErr: true},
	}

	for _, tt := range tests {
		got, err := parseClock(tt.raw)
		if tt.wantErr && err == nil {
			t.Fatalf("expected error for %q", tt.raw)
		}
		if !tt.wantErr && (err != nil || got != tt.want) {
			t.Fatalf("parseClock(%q) = %d, %v; want %d", tt.raw, got, err, tt.want)
		}
	}
}

func TestShiftFitsBusinessHours(t *testing.T) {
	shiftType := model.ShiftType{StartTime: "09:00", EndTime: "13:00"}
	hours := model.BusinessHour{OpenTime: "08:00", CloseTime: "20:00"}
	if !shiftFits(shiftType, hours) {
		t.Fatal("expected shift to fit business hours")
	}

	hours.CloseTime = "12:00"
	if shiftFits(shiftType, hours) {
		t.Fatal("expected shift not to fit business hours")
	}
}
