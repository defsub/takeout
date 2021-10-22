// Copyright (C) 2021 The Takeout Authors.
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
	"testing"
)

func TestMatch(t *testing.T) {
	if MatchTime("Mon 02", "Fri 13", ParseDate("2021-08-13")) {
		t.Logf("it's friday the 13th\n")
	}

	if MatchTime("Jan 02", "Dec 31", ParseDate("2021-12-31")) {
		t.Logf("new years eve\n")
	}

	if MatchTime("01/02/2006", "12/31/1998", ParseDate("1998-12-31")) {
		t.Logf("new years eve 1999\n")
	}

	if MatchTime("Jan 02", "Jan 01", ParseDate("2021-01-01")) {
		t.Logf("new years day\n")
	}

	if MatchTime("January 02", "July 04", ParseDate("2021-07-04")) {
		t.Logf("it's july 4th\n")
	}

	if MatchTime("Jan", "Dec", ParseDate("2021-12-12")) {
		t.Logf("it's xmas time")
	}

	if MatchTime("Jan", "Oct", ParseDate("2021-10-15")) {
		t.Logf("it's halloween time")
	}
}
