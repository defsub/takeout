package music

import (
	"fmt"
	"time"
)

type DateRange struct {
	after  time.Time
	before time.Time
}

func NewDateRange(a, b time.Time) *DateRange {
	return &DateRange{after: a, before: b}
}

func (d *DateRange) IsZero() bool {
	return d.after.IsZero() && d.before.IsZero()
}

func (d *DateRange) AfterDate() string {
	return ymd(d.after)
}

func (d *DateRange) BeforeDate() string {
	return ymd(d.before)
}

func ymd(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
}
