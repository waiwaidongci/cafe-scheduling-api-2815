package schedulecalc

import "time"

func WeekdayIndex(t time.Time) int {
	return (int(t.Weekday()) + 6) % 7
}

func rangeFits(start, end, open, close Clock) bool {
	return start.Minutes() >= open.Minutes() &&
		end.Minutes() <= close.Minutes() &&
		end.Minutes() > start.Minutes()
}
