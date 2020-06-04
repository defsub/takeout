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
	Similar  []Artist
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
	if len(view.Popular) > 5 {
		view.Popular = view.Popular[:5]
	}

	view.Similar = m.similarArtists(&artist)
	if len(view.Similar) > 10 {
		view.Similar = view.Similar[:10]
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
