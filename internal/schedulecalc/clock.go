package schedulecalc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidClock     = errors.New("invalid clock")
	ErrInvalidTimeRange = errors.New("invalid time range")
)

type Clock struct {
	Hour   int
	Minute int
}

func ParseClock(raw string) (Clock, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 2 {
		return Clock{}, ErrInvalidClock
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return Clock{}, ErrInvalidClock
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return Clock{}, ErrInvalidClock
	}

	return Clock{Hour: hour, Minute: minute}, nil
}

func (c Clock) Minutes() int {
	return c.Hour*60 + c.Minute
}

func (c Clock) String() string {
	return fmt.Sprintf("%02d:%02d", c.Hour, c.Minute)
}

func DurationMinutes(start, end Clock) (int, error) {
	if end.Minutes() <= start.Minutes() {
		return end.Minutes() - start.Minutes() + 24*60, nil
	}
	return end.Minutes() - start.Minutes(), nil
}

func AddMinutes(c Clock, delta int) (Clock, error) {
	total := c.Minutes() + delta
	if total < 0 {
		return Clock{}, ErrInvalidClock
	}
	return Clock{Hour: (total / 60) % 24, Minute: total % 60}, nil
}
