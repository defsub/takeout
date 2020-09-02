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
	jsonpatch "github.com/evanphx/json-patch"
)

const (
	ContentType = "application/json-patch+json"
)

func Patch(data []byte, patch []byte) ([]byte, error) {
	jp, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		return data, err
	}
	data, err = jp.Apply(data)
	return data, err
}

func Compare(before, after []byte) (bool, error) {
	var p1, p2 Playlist
	var err error

	// TODO this doesn't seem very efficient. Need to compare only the
	// playlist tracks to see if they changed. First need to unmarshal
	// both, then remarshal just the tracks, then compare.

	err = json.Unmarshal(before, &p1)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(after, &p2)
	if err != nil {
		return false, err
	}

	s1, err := json.Marshal(p1.Spiff)
	if err != nil {
		return false, err
	}
	s2, err := json.Marshal(p2.Spiff)
	if err != nil {
		return false, err
	}

	// And finally compare
	v := jsonpatch.Equal(s1, s2)

	return v, nil
}
