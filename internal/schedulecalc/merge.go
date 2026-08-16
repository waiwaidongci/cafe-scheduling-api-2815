package schedulecalc

import "sort"

func sortSpansByStart(group []spanWithInterval) {
	sort.Slice(group, func(i, j int) bool {
		if group[i].interval.Start.Minutes() == group[j].interval.Start.Minutes() {
			return group[i].interval.End.Minutes() < group[j].interval.End.Minutes()
		}
		return group[i].interval.Start.Minutes() < group[j].interval.Start.Minutes()
	})
}

func mergeIntervals(intervals []Interval) []Interval {
	if len(intervals) < 2 {
		return intervals
	}

	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i].Start.Minutes() == intervals[j].Start.Minutes() {
			return intervals[i].End.Minutes() < intervals[j].End.Minutes()
		}
		return intervals[i].Start.Minutes() < intervals[j].Start.Minutes()
	})

	merged := make([]Interval, 0, len(intervals))
	current := intervals[0]
	for _, next := range intervals[1:] {
		if Overlaps(current, next) {
			if next.End.Minutes() > current.End.Minutes() {
				current.End = next.End
			}
			continue
		}
		merged = append(merged, current)
		current = next
	}
	merged = append(merged, current)
	return merged
}
