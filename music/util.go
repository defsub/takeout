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

package music

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

func split(s string) []string {
	if len(s) == 0 {
		// TODO fix this
		return make([]string, 0)
	}
	a := strings.Split(s, ",")
	for i := range a {
		a[i] = strings.Trim(a[i], " ")
	}
	return a
}

// Parse a date string to time in format yyyy-mm-dd, yyyy-mm, yyyy.
func parseDate(date string) (t time.Time) {
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

func atoi(a string) int {
	i, err := strconv.Atoi(a)
	if err != nil {
		i = 0
	}
	return i
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

var doubleQuotedRegexp = regexp.MustCompile(`"(.*)"`)

func unquote(s string) string {
	matches := doubleQuotedRegexp.FindStringSubmatch(s)
	if matches != nil {
		return matches[1]
	}
	return s
}
