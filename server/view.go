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

package server

import (
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/music"
)

type HomeView struct {
	Added    []music.Release
	Released []music.Release
}

type ArtistsView struct {
	Artists []music.Artist
}

type ArtistView struct {
	Artist     music.Artist
	Image      string
	Background string
	Releases   []music.Release
	Popular    []music.Track
	Singles    []music.Track
	Similar    []music.Artist
}

type PopularView struct {
	Artist  music.Artist
	Popular []music.Track
}

type SinglesView struct {
	Artist  music.Artist
	Singles []music.Track
}

type ReleaseView struct {
	Artist  music.Artist
	Release music.Release
	Tracks  []music.Track
	Singles []music.Track
	Popular []music.Track
	Similar []music.Release
}

type SearchView struct {
	Artists  []music.Artist
	Releases []music.Release
	Tracks   []music.Track
	Query    string
	Hits     int
}

type RadioView struct {
	Artist  []music.Station
	Genre   []music.Station
	Similar []music.Station
	Period  []music.Station
	Series  []music.Station
	Other   []music.Station
}

func (handler *MusicHandler) homeView(m *music.Music) *HomeView {
	view := &HomeView{}
	view.Added = m.RecentlyAdded()
	view.Released = m.RecentlyReleased()
	return view
}

func (handler *MusicHandler) artistsView(m *music.Music) *ArtistsView {
	view := &ArtistsView{}
	view.Artists = m.Artists()
	return view
}

func (handler *MusicHandler) artistView(m *music.Music, artist music.Artist) *ArtistView {
	view := &ArtistView{}
	view.Artist = artist
	view.Releases = m.ArtistReleases(&artist)
	view.Popular = m.ArtistPopularTracks(artist)
	n := 5
	if len(view.Popular) > n {
		view.Popular = view.Popular[:n]
	}
	view.Singles = m.ArtistSingleTracks(artist)
	if len(view.Singles) > n {
		view.Singles = view.Singles[:n]
	}
	view.Similar = m.SimilarArtists(&artist)
	view.Image = m.ArtistImage(&artist)
	view.Background = m.ArtistBackground(&artist)
	return view
}

func (handler *MusicHandler) popularView(m *music.Music, artist music.Artist) *PopularView {
	view := &PopularView{}
	view.Artist = artist
	view.Popular = m.ArtistPopularTracks(artist)
	limit := handler.config.Music.PopularLimit
	if len(view.Popular) > limit {
		view.Popular = view.Popular[:limit]
	}
	return view
}

func (handler *MusicHandler) singlesView(m *music.Music, artist music.Artist) *SinglesView {
	view := &SinglesView{}
	view.Artist = artist
	view.Singles = m.ArtistSingleTracks(artist)
	limit := handler.config.Music.SinglesLimit
	if len(view.Singles) > limit {
		view.Singles = view.Singles[:limit]
	}
	return view
}

func (handler *MusicHandler) releaseView(m *music.Music, release music.Release) *ReleaseView {
	view := &ReleaseView{}
	view.Release = release
	view.Artist = *m.Artist(release.Artist)
	view.Tracks = m.ReleaseTracks(release)
	view.Singles = m.ReleaseSingles(release)
	view.Popular = m.ReleasePopular(release)
	view.Similar = m.SimilarReleases(&view.Artist, release)
	return view
}

func (handler *MusicHandler) searchView(m *music.Music, query string) *SearchView {
	view := &SearchView{}
	artists, releases, _ := m.Query(query)
	view.Artists = artists
	view.Releases = releases
	view.Query = query
	view.Tracks = m.Search(query)
	view.Hits = len(view.Artists) + len(view.Releases) + len(view.Tracks)
	return view
}

func (handler *MusicHandler) radioView(m *music.Music, user *auth.User) *RadioView {
	view := &RadioView{}
	for _, s := range m.Stations(user) {
		switch s.Type {
		case music.TypeArtist:
			view.Artist = append(view.Artist, s)
		case music.TypeGenre:
			view.Genre = append(view.Genre, s)
		case music.TypeSimilar:
			view.Similar = append(view.Similar, s)
		case music.TypePeriod:
			view.Period = append(view.Period, s)
		case music.TypeSeries:
			view.Series = append(view.Series, s)
		default:
			view.Other = append(view.Other, s)
		}
	}
	return view
}
