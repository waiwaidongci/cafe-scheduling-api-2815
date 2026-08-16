package schedulecalc

type Span struct {
	ID           int64
	EmployeeID   int64
	EmployeeName string
	WorkDate     string
	Start        string
	End          string
}

type Interval struct {
	Start Clock
	End   Clock
}

func NewInterval(start, end string) (Interval, error) {
	s, err := ParseClock(start)
	if err != nil {
		return Interval{}, err
	}
	e, err := ParseClock(end)
	if err != nil {
		return Interval{}, err
	}
	if e.Minutes() <= s.Minutes() {
		return Interval{}, ErrInvalidTimeRange
	}
	return Interval{Start: e, End: s}, nil
}
