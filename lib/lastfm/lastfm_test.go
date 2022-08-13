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

package lastfm

import (
	"testing"
	lfm "github.com/shkh/lastfm-go/lastfm"
	"github.com/defsub/takeout/config"
)

func TestLast(t *testing.T) {
}

func TestScrobble(t *testing.T) {
	_, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	//lfm.Api.

	//api := lfm.New(config.LastFM.Key, config.LastFM.Secret)
	//api.Track.Scrobble()
}
