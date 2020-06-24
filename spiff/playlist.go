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

package spiff

import (
	"encoding/json"
)

type Playlist struct {
	Spiff    `json:"playlist"`
	Position uint `json:"position,omitempty"`
}

type Spiff struct {
	Title   string  `json:"title"`
	Entries []Entry `json:"track"`
}

type Entry struct {
	Ref        string   `json:"$ref,omitempty"`
	Creator    string   `json:"creator,omitempty"`
	Album      string   `json:"album,omitempty"`
	Title      string   `json:"title,omitempty"`
	Image      string   `json:"image,omitempty"`
	Location   []string `json:"location,omitempty"`
	Identifier []string `json:"identifier,omitempty"`
}

func NewPlaylist() *Playlist {
	return &Playlist{Spiff{"", []Entry{}}, 0}
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
