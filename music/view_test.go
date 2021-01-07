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

package music

import (
	"testing"

	"github.com/defsub/takeout/config"
)

func TestArtistView(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	m := NewMusic(config)
	m.Open()
	// view := m.ArtistView("Black Sabbath")
	// if view == nil {
	// 	t.Errorf("view is nil")
	// }
	// if view.artist == nil {
	// 	t.Errorf("view artist is nil")
	// }
	// if view.releases == nil || len(view.releases) == 0 {
	// 	t.Errorf("view artist releases empty")
	// }
	// if view.popular == nil || len(view.popular) == 0 {
	// 	t.Errorf("view artist popular empty")
	// }
	// if view.similar == nil || len(view.similar) == 0 {
	// 	t.Errorf("view artist similar empty")
	// }

	// for _, r := range view.releases {
	// 	t.Logf("A: %d %s\n", r.Date.Year(), r.Name)
	// }

	// for _, tt := range view.popular {
	// 	t.Logf("P: %s\n", tt.Title)
	// }

	// for _, a := range view.similar {
	// 	t.Logf("S: %s\n", a.Name)
	// }
}
