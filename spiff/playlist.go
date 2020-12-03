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

package spiff

import (
	"encoding/json"
)

// See the following specifications:
//  https://www.xspf.org/xspf-v1.html
//  https://www.xspf.org/jspf/

type Playlist struct {
	Spiff    `json:"playlist"`
	Index    int     `json:"index"`
	Position float64 `json:"position"`
}

type Spiff struct {
	Title    string  `json:"title"`
	Creator  string  `json:"creator,omitempty"`
	Image    string  `json:"image,omitempty"`
	Location string  `json:"location,omitempty"`
	Date     string  `json:"date,omitempty"` // "2005-01-08T17:10:47-05:00",
	Entries  []Entry `json:"track"`
}

type Entry struct {
	Ref        string   `json:"$ref,omitempty"`
	Creator    string   `json:"creator,omitempty"`
	Album      string   `json:"album,omitempty"`
	Title      string   `json:"title,omitempty"`
	Image      string   `json:"image,omitempty"`
	Location   []string `json:"location,omitempty"`
	Identifier []string `json:"identifier,omitempty"`
	Size       []int64  `json:"size,omitempty"`
}

func NewPlaylist() *Playlist {
	return &Playlist{Spiff{"", "", "", "", "", []Entry{}}, -1, 0}
}

func Unmarshal(data []byte) (*Playlist, error) {
	var playlist Playlist
	err := json.Unmarshal(data, &playlist)
	return &playlist, err
}

func (playlist *Playlist) Marshal() ([]byte, error) {
	data, err := json.Marshal(playlist)
	return data, err
}
