// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// Takeout is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for
// more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package date

import (
	"time"
)

// Parse a date string to time in format yyyy-mm-dd, yyyy-mm, yyyy.
func ParseDate(date string) (t time.Time) {
	if date == "" {
		return t
	}
	var err error
	// TODO is this done with a single call?
	t, err = time.Parse("2006-1-2", date)
	if err != nil {
		t, err = time.Parse("2006-1", date)
		if err != nil {
			t, err = time.Parse("2006", date)
			if err != nil {
				t = time.Time{}
			}
		}
	}
	return t
}

const (
	RFC1123_1  = "Mon, _2 Jan 2006 15:04:05 MST"
	RFC1123_2  = "Mon, 2 Jan 2006 15:04:05 MST"
	RFC1123Z_1 = "Mon, _2 Jan 2006 15:04:05 -0700"
	RFC1123Z_2 = "Mon, 2 Jan 2006 15:04:05 -0700"
)

// Mon, 02 Jan 2006 15:04:05 MST
// Tue, 07 Dec 2021 19:57:22 -0500
// Fri, 6 Nov 2020 19:32:35 +0000
func ParseRFC1123(date string) (t time.Time) {
	if date == "" {
		return t
	}
	var err error
	layouts := []string{time.RFC1123, time.RFC1123Z, RFC1123_1, RFC1123_2, RFC1123Z_1, RFC1123Z_2}
	for _, layout := range layouts {
		t, err = time.Parse(layout, date)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

const (
	Simple12 = "Jan 02 03:04 PM"
	Simple24 = "Jan 02 15:04"
)

func Format(t time.Time) string {
	return t.Format(Simple12)
}

func FormatJson(t time.Time) string {
	return t.Format(time.RFC3339)
}
