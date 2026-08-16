package schedulecalc

import "sort"

type EmployeeHours struct {
	EmployeeID   int64
	EmployeeName string
	ShiftCount   int
	TotalMinutes int
}

func WeeklyHours(spans []Span) []EmployeeHours {
	return weeklyHours(spans, false)
}

func OverlapFreeWeeklyHours(spans []Span) []EmployeeHours {
	return weeklyHours(spans, true)
}

func weeklyHours(spans []Span, mergeOverlaps bool) []EmployeeHours {
	byEmployee := make(map[int64][]spanWithInterval)
	for _, item := range parseSpans(spans) {
		byEmployee[item.span.EmployeeID] = append(byEmployee[item.span.EmployeeID], item)
	}

	hours := make([]EmployeeHours, 0, len(byEmployee))
	for employeeID, items := range byEmployee {
		total := 0
		count := len(items)
		if mergeOverlaps {
			total = mergedMinutes(items)
		} else {
			for _, item := range items {
				if duration, err := DurationMinutes(item.interval.Start, item.interval.End); err == nil {
					total += duration
				}
			}
		}
		hours = append(hours, EmployeeHours{
			EmployeeID:   employeeID,
			EmployeeName: items[0].span.EmployeeName,
			ShiftCount:   count,
			TotalMinutes: total,
		})
	}

	sort.Slice(hours, func(i, j int) bool { return hours[i].EmployeeID < hours[j].EmployeeID })
	return hours
}

func mergedMinutes(items []spanWithInterval) int {
	byDate := make(map[string][]spanWithInterval)
	for _, item := range items {
		byDate[item.span.WorkDate] = append(byDate[item.span.WorkDate], item)
	}

	total := 0
	for _, group := range byDate {
		intervals := make([]Interval, 0, len(group))
		for _, item := range group {
			intervals = append(intervals, item.interval)
		}
		for _, interval := range mergeIntervals(intervals) {
			if duration, err := DurationMinutes(interval.Start, interval.End); err == nil {
				total += duration
			}
		}
	}
	return total
}
