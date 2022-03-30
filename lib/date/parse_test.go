// Copyright (C) 2022 The Takeout Authors.
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
	"time"
)

func TestParse1123(t *testing.T) {
	s := "Fri, 6 Nov 2020 19:32:35 +0000"
	//s := "Fri,  6 Nov 2020 19:32:35 +0000"
	//s := "Fri, 06 Nov 2020 19:32:35 +0000"
	d := ParseRFC1123(s)
	t.Logf("got %s\n", d.String())
	if d.Day() != 6 {
		t.Errorf("wrong day got %d\n", d.Day())
	}
	if d.Month() != time.November {
		t.Errorf("wrong month got %s\n", d.Month())
	}
	if d.Year() != 2020 {
		t.Errorf("wrong year got %d\n", d.Year())
	}
}
