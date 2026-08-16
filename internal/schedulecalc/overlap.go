package schedulecalc

import (
	"fmt"
	"sort"
)

type Conflict struct {
	EmployeeID   int64
	EmployeeName string
	WorkDate     string
	SpanIDs      []int64
}

type spanWithInterval struct {
	span     Span
	interval Interval
}

func Overlaps(a, b Interval) bool {
	return a.Start.Minutes() < b.End.Minutes() && b.Start.Minutes() < a.End.Minutes()
}

func parseSpans(spans []Span) []spanWithInterval {
	parsed := make([]spanWithInterval, 0, len(spans))
	for _, span := range spans {
		interval, err := NewInterval(span.Start, span.End)
		if err != nil {
			continue
		}
		parsed = append(parsed, spanWithInterval{span: span, interval: interval})
	}
	return parsed
}

func groupSpans(spans []Span) map[string][]spanWithInterval {
	parsed := parseSpans(spans)
	groups := make(map[string][]spanWithInterval)
	for _, item := range parsed {
		key := fmt.Sprintf("%d:%s", item.span.EmployeeID, item.span.WorkDate)
		groups[key] = append(groups[key], item)
	}
	return groups
}

func DetectOverlaps(spans []Span) []Conflict {
	groups := groupSpans(spans)
	conflicts := make([]Conflict, 0)

	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			if group[i].interval.Start.Minutes() == group[j].interval.Start.Minutes() {
				return group[i].interval.End.Minutes() < group[j].interval.End.Minutes()
			}
			return group[i].interval.Start.Minutes() < group[j].interval.Start.Minutes()
		})

		conflictIDs := make(map[int64]bool)
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				if Overlaps(group[i].interval, group[j].interval) {
					conflictIDs[group[i].span.ID] = true
					conflictIDs[group[j].span.ID] = true
				}
			}
		}
		if len(conflictIDs) == 0 {
			continue
		}

		involved := make([]int64, 0, len(conflictIDs))
		for _, item := range group {
			if conflictIDs[item.span.ID] {
				involved = append(involved, item.span.ID)
			}
		}
		sort.Slice(involved, func(i, j int) bool { return involved[i] < involved[j] })
		conflicts = append(conflicts, Conflict{
			EmployeeID:   group[0].span.EmployeeID,
			EmployeeName: group[0].span.EmployeeName,
			WorkDate:     group[0].span.WorkDate,
			SpanIDs:      involved,
		})
	}

	sort.Slice(conflicts, func(i, j int) bool {
		if conflicts[i].EmployeeID == conflicts[j].EmployeeID {
			return conflicts[i].WorkDate < conflicts[j].WorkDate
		}
		return conflicts[i].EmployeeID < conflicts[j].EmployeeID
	})
	return conflicts
}

func MergeOverlaps(spans []Span) []Span {
	groups := groupSpans(spans)
	merged := make([]Span, 0, len(spans))

	for _, group := range groups {
		sort.Slice(group, func(i, j int) bool {
			if group[i].interval.Start.Minutes() == group[j].interval.Start.Minutes() {
				return group[i].interval.End.Minutes() < group[j].interval.End.Minutes()
			}
			return group[i].interval.Start.Minutes() < group[j].interval.Start.Minutes()
		})

		current := group[0]
		for i := 1; i < len(group); i++ {
			next := group[i]
			if Overlaps(current.interval, next.interval) {
				if next.interval.End.Minutes() > current.interval.End.Minutes() {
					current.interval.End = next.interval.End
					current.span.End = next.span.End
				}
				continue
			}
			merged = append(merged, current.span)
			current = next
		}
		merged = append(merged, current.span)
	}

	sort.Slice(merged, func(i, j int) bool {
		if merged[i].EmployeeID == merged[j].EmployeeID {
			if merged[i].WorkDate == merged[j].WorkDate {
				return merged[i].Start < merged[j].Start
			}
			return merged[i].WorkDate < merged[j].WorkDate
		}
		return merged[i].EmployeeID < merged[j].EmployeeID
	})
	return merged
}
