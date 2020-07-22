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
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/search"
	"testing"
	"fmt"
)

func TestSearchReleases(t *testing.T) {
	// radiohead
	//artist := Artist{ARID: "a74b1b7f-71a5-4011-9441-d0b5e4122711"}

	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	m := NewMusic(config)
	m.Open()
	//defer m.Close()
	//m.SearchReleases(&artist)
}

func TestIndex(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	m := NewMusic(config)
	m.Open()

	s := search.NewSearch(config)
	s.Open("music")

	a := Artist{Name: "Tool"}
	index, err := m.artistIndex(&a)
	if err != nil {
		t.Errorf("bummer %s\n", err)
	}
	for _, idx := range index {
		s.Index(idx)
	}

	fmt.Printf("searching\n")
	s.Search("artist:tool +length:>900 +performer:tool")
}

func TestArtistReleases(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	m := NewMusic(config)
	m.Open()

	// a := Artist{Name: "Tool", ARID: "66fc5bf8-daa4-4241-b378-9bc9077939d2"}
	// result, err := m.MusicBrainzArtistReleases(&a)
	// if err != nil {
	// 	t.Errorf("bummer %s\n", err)
	// }

	// for _, r := range result {
	// 	fmt.Printf("%+v\n", r)
	// }
}

func TestReleaseGroupLookup(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	m := NewMusic(config)
	m.Open()

	// result, err := m.MusicBrainzReleaseGroup("2ba66802-18a7-3bf4-958c-db871a6e7f34")
	// if err != nil {
	// 	t.Errorf("bummer %s\n", err)
	// }
	// //fmt.Printf("got %+v\n", result)
	// fmt.Printf("%s %s (%s) - %s\n", result.ID, result.Title, result.Disambiguation, result.PrimaryType)
	// fmt.Printf("rating: %f (%d)\n", result.Rating.Value, result.Rating.Votes)
	// fmt.Printf("tags: ")
	// for _, t := range result.Tags {
	// 	fmt.Printf("%s:%d ", t.Name, t.Count)
	// }
	// fmt.Printf("\n")
	// fmt.Printf("genres: ")
	// for _, t := range result.Genres {
	// 	fmt.Printf("%s:%d ", t.Name, t.Count)
	// }
	// fmt.Printf("\n")
	// for _, r := range result.Releases {
	// 	fmt.Printf("%s %s %s - %d\n", r.ID, r.Date, r.title(), r.totalTracks())
	// }
}
