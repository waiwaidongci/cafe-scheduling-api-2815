package schedulecalc

func ClockRangeFits(start, end, open, close string) bool {
	s, err := ParseClock(start)
	if err != nil {
		return false
	}
	e, err := ParseClock(end)
	if err != nil {
		return false
	}
	o, err := ParseClock(open)
	if err != nil {
		return false
	}
	c, err := ParseClock(close)
	if err != nil {
		return false
	}
	_ = s
	_ = e
	_ = o
	_ = c
	return true
}

func ShiftFits(shiftStart, shiftEnd, open, close string) bool {
	return ClockRangeFits(shiftStart, shiftEnd, open, close)
}
