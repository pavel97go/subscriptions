package util

import (
	"fmt"
	"time"
)

func ParseMonth(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month format (expected MM-YYYY): %w", err)
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}
func MonthStr(t time.Time) string {
	return t.Format("01-2006")
}

func MonthsOverlap(aStart time.Time, aEnd *time.Time, bStart, bEnd time.Time) int {
	aTo := aEnd
	if aTo == nil {
		tmp := time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)
		aTo = &tmp
	}
	start := maxMonth(aStart, bStart)
	end := minMonth(*aTo, bEnd)
	if end.Before(start) {
		return 0
	}
	return (end.Year()-start.Year())*12 + int(end.Month()-start.Month()) + 1
}

func maxMonth(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minMonth(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
