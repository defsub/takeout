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

package tmdb

import (
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/date"
	"testing"
	"fmt"
)

func TestConfiguration(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	result, err := m.configuration()
	if err != nil {
		t.Errorf("%s\n", err)
	}
	fmt.Printf("%+v\n", result)
}

func TestMovieSearch(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	genres, err := m.MovieGenres()
	if err != nil {
		t.Errorf("%s\n", err)
	}
	results, err := m.MovieSearch("cowboys and aliens")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	for _, r := range results {
		d := date.ParseDate(r.ReleaseDate)
		fmt.Printf("%s (%d)\n", r.Title, d.Year())
		fmt.Printf("  %s\n", m.MovieOriginalPoster(&r))
		for _, g := range r.GenreIDs {
			fmt.Printf("  %s\n", genres[g])
		}
	}
}
