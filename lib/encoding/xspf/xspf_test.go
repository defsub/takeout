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

package xspf

import (
	"testing"
	"os"
)

type Track struct {
	Artist       string `spiff:"creator"`
	Release      string `spiff:"album"`
	TrackNum     uint   `spiff:"tracknum"`
	Title        string `spiff:"title"`
	Location     string `spiff:"location"`
}

func TestXml(t *testing.T) {
	e := NewXMLEncoder(os.Stdout)
	e.Header("test title")
	var track Track
	track.Artist = "My Artist"
	track.Release = "My Release"
	track.TrackNum = 1
	track.Title = "My Title"
	track.Location = "https://a/b/c"
	e.Encode(track)
	track.TrackNum = 2
	e.Encode(track)
	track.TrackNum = 3
	e.Encode(track)
	e.Footer()
}

func TestJson(t *testing.T) {
	e := NewJsonEncoder(os.Stdout)
	e.Header("test title")
	var track Track
	track.Artist = "My Artist"
	track.Release = "My Release"
	track.TrackNum = 1
	track.Title = "My Title"
	track.Location = "https://a/b/c"
	e.Encode(track)
	track.TrackNum = 2
	e.Encode(track)
	track.TrackNum = 3
	e.Encode(track)
	e.Footer()
}
