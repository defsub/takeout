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
	"time"
)

// Format the current time with the given layout and return whether or not it
// matches the provided format.
func Match(layout, format string) bool {
	return MatchTime(layout, format, time.Now())
}

// Format the given time with layout and return whether or not it matches the
// provided format.
func MatchTime(layout, format string, t time.Time) bool {
	result := t.Format(layout)
	return result == format
}
