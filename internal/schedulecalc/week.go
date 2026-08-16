package schedulecalc

import (
	"errors"
	"time"
)

var ErrInvalidDateRange = errors.New("invalid date range")

func WeekdayIndex(t time.Time) int {
	return (int(t.Weekday()) + 6) % 7
}

func NormalizeWeekStart(t time.Time) time.Time {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return start.AddDate(0, 0, -WeekdayIndex(start))
}

func ParseWeekRange(raw string) (time.Time, time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	start := NormalizeWeekStart(t)
	return start, start.AddDate(0, 0, 6), nil
}

func DateKey(t time.Time) string {
	return t.Format("2006-01-02")
}
