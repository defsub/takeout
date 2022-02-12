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
	results, err := m.MovieSearch("cowboys and aliens")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	for _, r := range results {
		d := date.ParseDate(r.ReleaseDate)
		fmt.Printf("%d %s (%d)\n", r.ID, r.Title, d.Year())
		fmt.Printf("  %s\n", m.OriginalPoster(r.PosterPath))
		for _, g := range r.GenreIDs {
			fmt.Printf("  %s\n", m.MovieGenre(g))
		}
	}
}

func TestMovieDetail(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	movie, err := m.MovieDetail(503736) // arm of the dead
	if err != nil {
		t.Errorf("%s\n", err)
	}
	fmt.Printf("%s (%s)\n", movie.Title, movie.ReleaseDate)
	fmt.Printf("%+v\n", movie)
}

func TestMovieCredits(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	credits, err := m.MovieCredits(503736) // arm of the dead
	if err != nil {
		t.Errorf("%s\n", err)
	}
	for _, c := range credits.Cast {
		fmt.Printf("%s - %s\n", c.Name, c.Character)
	}
	for _, c := range credits.Crew {
		fmt.Printf("%s - %s\n", c.Name, c.Job)
	}
}

func TestMovieReleases(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	releases, err := m.MovieReleases(503736) // arm of the dead
	if err != nil {
		t.Errorf("%s\n", err)
	}
	for k, v := range releases {
		fmt.Printf("%s - %+v\n", k, v)
	}
}

func TestMovieReleaseType(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	r, err := m.MovieReleaseType(503736, "US", TypeDigital) // arm of the dead
	if err != nil {
		t.Errorf("%s\n", err)
	}
	fmt.Printf("%s - %s\n", r.Certification, r.Date)
}

func TestPerson(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	p, err := m.PersonDetail(287) // brad pitt
	if err != nil {
		t.Errorf("%s\n", err)
	}
	fmt.Printf("%s - %s\n", p.Name, p.Birthday)
	fmt.Printf("%s\n", p.Biography)
}

func TestTVSearch(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	results, err := m.TVSearch("the shining")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	for _, r := range results {
		d := date.ParseDate(r.FirstAirDate)
		fmt.Printf("%d %s (%d)\n", r.ID, r.Name, d.Year())
		fmt.Printf("  %s\n", m.OriginalPoster(r.PosterPath))
		for _, g := range r.GenreIDs {
			fmt.Printf("  %s\n", m.TVGenre(g))
		}
	}
}

func TestTVDetail(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	tv, err := m.TVDetail(1867) // game of thrones
	if err != nil {
		t.Errorf("%s\n", err)
	}
	fmt.Printf("%s (%s)\n", tv.Name, tv.FirstAirDate)
	fmt.Printf("%+v\n", tv)
}

func TestEpisodeDetail(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	episode, err := m.EpisodeDetail(1399, 1, 1) // game of thrones
	if err != nil {
		t.Errorf("%s\n", err)
	}
	fmt.Printf("%d %s (%s)\n", episode.ID, episode.Name, episode.AirDate)
	fmt.Printf("%+v\n", episode)
}

func TestEpisodeCredits(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	if config.TMDB.Key == "" {
		t.Errorf("no key\n")
	}
	m := NewTMDB(config)
	credits, err := m.EpisodeCredits(1399, 1, 1) // game of thrones
	if err != nil {
		t.Errorf("%s\n", err)
	}
	for _, c := range credits.Cast {
		fmt.Printf("cast: %s - %s\n", c.Name, c.Character)
	}
	for _, c := range credits.Crew {
		fmt.Printf("crew: %s - %s\n", c.Name, c.Job)
	}
	for _, c := range credits.Guests {
		fmt.Printf("guest: %s - %s\n", c.Name, c.Character)
	}
}
