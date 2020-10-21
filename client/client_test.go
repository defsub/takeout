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

package client

import (
	"testing"
)

type mbzTag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type mbzGenre struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type mbzArtist struct {
	Name           string     `json:"name"`
	SortName       string     `json:"sort-name"`
	Disambiguation string     `json:"disambiguation"`
	Type           string     `json:"type"`
	Genres         []mbzGenre `json:"genres"`
	Tags           []mbzTag   `json:"tags"`
}

func TestClient(t *testing.T) {
	urls := []string{
		"http://musicbrainz.org/ws/2/artist/5b11f4ce-a62d-471e-81fc-a69a8278c7da?inc=aliases&fmt=json",
		"http://musicbrainz.org/ws/2/artist/5b11f4ce-a62d-471e-81fc-a69a8278c7da?inc=aliases&fmt=json",
		"http://musicbrainz.org/ws/2/artist/5b11f4ce-a62d-471e-81fc-a69a8278c7da?inc=aliases&fmt=json",
		"http://musicbrainz.org/ws/2/artist/ba0d6274-db14-4ef5-b28d-657ebde1a396?inc=aliases&fmt=json",
		"http://musicbrainz.org/ws/2/artist/ba0d6274-db14-4ef5-b28d-657ebde1a396?inc=aliases&fmt=json",
		"http://musicbrainz.org/ws/2/artist/ba0d6274-db14-4ef5-b28d-657ebde1a396?inc=aliases&fmt=json",
	}

	for i := 0; i < len(urls); i++ {
		url := urls[i]
		var artist mbzArtist
		GetJson(url, &artist)
		t.Logf("%v\n", artist)
	}
}
