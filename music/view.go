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

package music

type HomeView struct {
	Added    []Release
	Released []Release
}

type ArtistsView struct {
	Artists []Artist
}

type ArtistView struct {
	Artist   Artist
	Releases []Release
	Popular  []Track
	Singles  []Track
	Similar  []Artist
}

type PopularView struct {
	Artist   Artist
	Popular  []Track
}

type SinglesView struct {
	Artist   Artist
	Singles  []Track
}

type ReleaseView struct {
	Artist  Artist
	Release Release
	Tracks  []Track
	Similar []Release
}

type SearchView struct {
	Artists  []Artist
	Releases []Release
	Tracks   []Track
}

type PlayView struct {
	Tracks   []Track
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
	view.Popular = m.artistPopularTracks(artist.Name, nil)
	n := 5
	if len(view.Popular) > n {
		view.Popular = view.Popular[:n]
	}
	view.Singles = m.artistSingleTracks(artist.Name, nil)
	if len(view.Singles) > n {
		view.Singles = view.Singles[:n]
	}

	view.Similar = m.similarArtists(&artist)
	if len(view.Similar) > 10 {
		view.Similar = view.Similar[:10]
	}

	return view
}

func (m *Music) PopularView(artist Artist) *PopularView {
	view := &PopularView{}
	view.Artist = artist
	view.Popular = m.artistPopularTracks(artist.Name, nil)
	if len(view.Popular) > 25 {
		view.Popular = view.Popular[:25]
	}
	return view
}

func (m *Music) SinglesView(artist Artist) *SinglesView {
	view := &SinglesView{}
	view.Artist = artist
	view.Singles = m.artistSingleTracks(artist.Name, nil)
	if len(view.Singles) > 25 {
		view.Singles = view.Singles[:25]
	}
	return view
}

func (m *Music) ReleaseView(release Release) *ReleaseView {
	view := &ReleaseView{}
	view.Release = release
	view.Artist = *m.artist(release.Artist)
	view.Tracks = m.releaseTracks(release)
	view.Similar = m.similarReleases(&view.Artist, release)
	return view
}

func (m *Music) SearchView(query string) *SearchView {
	view := &SearchView{}
	artists, releases, tracks := m.search(query)
	view.Artists = artists
	view.Releases = releases
	view.Tracks = tracks
	return view
}

func (m *Music) PlayView(tracks []Track) *PlayView {
	view := &PlayView{}
	view.Tracks = tracks
	return view
}
