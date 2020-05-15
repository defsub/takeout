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

type ArtistView struct {
	artist   *Artist
	releases []Release
	popular  []Track
	similar  []Artist
}

func (m *Music) ArtistView(artist string) *ArtistView {
	view := &ArtistView{}
	view.artist = m.artist(artist)
	view.releases = m.artistReleases(view.artist)
	view.popular = m.artistPopularTracks(artist, nil)
	if len(view.popular) > 10 {
		view.popular = view.popular[:10]
	}
	view.similar = 	m.similarArtists(view.artist)
	if len(view.similar) > 10 {
		view.similar = view.similar[:10]
	}
	return view
}
