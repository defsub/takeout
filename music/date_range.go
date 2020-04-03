// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// Takeout is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

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
