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

package pls

import (
	"testing"
)

func TestParse(t *testing.T) {
	data := `
[playlist]
numberofentries=3
File1=https://ice2.somafm.com/dronezone-128-aac
Title1=SomaFM: Drone Zone (#1): Served best chilled, safe with most medications. Atmospheric textures with minimal beats.
Length1=-1
File2=https://ice1.somafm.com/dronezone-128-aac
Title2=SomaFM: Drone Zone (#2): Served best chilled, safe with most medications. Atmospheric textures with minimal beats.
Length2=-1
File3=https://ice4.somafm.com/dronezone-128-aac
Title3=SomaFM: Drone Zone (#3): Served best chilled, safe with most medications. Atmospheric textures with minimal beats.
Length3=-1
Version=2

`
	playlist, err := parse(data)
	if err != nil {
		t.Error(err)
	}
	for _, v := range playlist.Entries {
		t.Logf("%d. %s %s (%d)\n", v.Index, v.File, v.Title, v.Length)
	}
}
