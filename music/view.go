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
	"fmt"

	"github.com/defsub/takeout/auth"
)

type HomeView struct {
	Added    []Release
	Released []Release
}

type ArtistsView struct {
	Artists []Artist
}

type ArtistView struct {
	Artist     Artist
	Image      string
	Background string
	Releases   []Release
	Popular    []Track
	Singles    []Track
	Similar    []Artist
}

type PopularView struct {
	Artist  Artist
	Popular []Track
}

type SinglesView struct {
	Artist  Artist
	Singles []Track
}

type ReleaseView struct {
	Artist  Artist
	Release Release
	Tracks  []Track
	Singles []Track
	Popular []Track
	Similar []Release
}

type SearchView struct {
	Artists  []Artist
	Releases []Release
	Tracks   []Track
	Query    string
	Hits     int
}

type RadioView struct {
	Artist  []Station
	Genre   []Station
	Similar []Station
	Period  []Station
	Other   []Station
}

func (m *Music) HomeView() *HomeView {
	view := &HomeView{}
	view.Added = m.recentlyAdded()
	view.Released = m.recentlyReleased()
	return view
}

func (m *Music) ArtistsView() *ArtistsView {
	view := &ArtistsView{}
	view.Artists = m.artists()
	return view
}

func (m *Music) ArtistView(artist Artist) *ArtistView {
	view := &ArtistView{}
	view.Artist = artist
	view.Releases = m.artistReleases(&artist)
	view.Popular = m.artistPopularTracks(artist)
	n := 5
	if len(view.Popular) > n {
		view.Popular = view.Popular[:n]
	}
	view.Singles = m.artistSingleTracks(artist)
	if len(view.Singles) > n {
		view.Singles = view.Singles[:n]
	}
	view.Similar = m.similarArtists(&artist)
	view.Image = m.artistImage(&artist)
	view.Background = m.artistBackground(&artist)
	return view
}

func (m *Music) PopularView(artist Artist) *PopularView {
	view := &PopularView{}
	view.Artist = artist
	view.Popular = m.artistPopularTracks(artist)
	limit := m.config.Music.PopularLimit
	if len(view.Popular) > limit {
		view.Popular = view.Popular[:limit]
	}
	return view
}

func (m *Music) SinglesView(artist Artist) *SinglesView {
	view := &SinglesView{}
	view.Artist = artist
	view.Singles = m.artistSingleTracks(artist)
	limit := m.config.Music.SinglesLimit
	if len(view.Singles) > limit {
		view.Singles = view.Singles[:limit]
	}
	return view
}

func (m *Music) ReleaseView(release Release) *ReleaseView {
	view := &ReleaseView{}
	view.Release = release
	view.Artist = *m.artist(release.Artist)
	view.Tracks = m.releaseTracks(release)
	view.Singles = m.releaseSingles(release)
	view.Popular = m.releasePopular(release)
	view.Similar = m.similarReleases(&view.Artist, release)
	return view
}

func (m *Music) SearchView(query string) *SearchView {
	view := &SearchView{}
	artists, releases, _ := m.search(query)
	view.Artists = artists
	view.Releases = releases
	view.Query = query
	view.Tracks = m.Search(query)
	view.Hits = len(view.Artists) + len(view.Releases) + len(view.Tracks)
	return view
}

func (m *Music) RadioView(user *auth.User) *RadioView {
	view := &RadioView{}
	for _, s := range m.stations(user) {
		switch s.Type {
		case typeArtist:
			view.Artist = append(view.Artist, s)
		case typeGenre:
			view.Genre = append(view.Genre, s)
		case typeSimilar:
			view.Similar = append(view.Similar, s)
		case typePeriod:
			view.Period = append(view.Period, s)
		default:
			view.Other = append(view.Other, s)
		}
	}
	return view
}

func trackLocation(t Track) string {
	return fmt.Sprintf("/api/tracks/%d/location", t.ID)
}
