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
	"github.com/defsub/takeout/spiff"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

const (
	typeArtist = "artist" // Songs by single artist
	typeGenre  = "genre"  // Songs from one or more genres
	typeMix    = "mix"    // Songs from many artists
	typePeriod = "period" // Songs from one or more time periods
)

func (s *Station) visible(user *auth.User) bool {
	return s.User == user.Name || s.Shared
}

func (m *Music) ClearStations() {
	m.clearStationPlaylists()
}

func (m *Music) CreateStations() {
	artists, _ := m.favoriteArtists(10)
	for _, v := range artists {
		a := m.artist(v)
		if a == nil {
			continue
		}
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   typeArtist,
			Name:   fmt.Sprintf("%s Singles", a.Name),
			Ref:    fmt.Sprintf(`/music/artists/%d/shuffle`, a.ID)}
		m.createStation(&station)

		station = Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   typeMix,
			Name:   fmt.Sprintf("%s Mix", a.Name),
			Ref:    fmt.Sprintf(`/music/artists/%d/mix`, a.ID)}
		m.createStation(&station)
	}

	for _, g := range m.config.Music.RadioGenres {
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   typeGenre,
			Name:   strings.Title(g),
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(fmt.Sprintf(`+genre:"%s" +type:"single" -artist:"Various Artists"`, g)))}
		m.createStation(&station)
	}

}

func (m *Music) stationRefresh(s *Station, user *auth.User) {
	if len(s.Playlist) == 0 {
		m.resolveStation(s, user)
	}
	// TODO force refresh
}

func (m *Music) resolveStation(s *Station, user *auth.User) {
	plist := spiff.NewPlaylist()
	// Creator, Image
	plist.Spiff.Title = s.Name
	plist.Entries = []spiff.Entry{{Ref: s.Ref}}
	m.Resolve(user, plist)
	if plist.Entries == nil {
		plist.Entries = []spiff.Entry{}
	}
	s.Playlist, _ = plist.Marshal()
	m.updateStation(s)
}

func shuffle(tracks []Track) []Track {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tracks), func(i, j int) { tracks[i], tracks[j] = tracks[j], tracks[i] })
	return tracks
}

func (m *Music) artistMix(artist Artist, depth int, breadth int) []Track {
	var station []Track
	tracks := m.artistPopularTracks(artist, depth)
	station = append(station, tracks...)
	artists := m.similarArtists(&artist, breadth)
	for _, a := range artists {
		tracks = m.artistPopularTracks(a)
		station = append(station, tracks...)
	}
	return shuffle(station)
}

func (m *Music) artistShuffle(artist Artist, depth int) []Track {
	tracks := m.artistPopularTracks(artist, depth)
	return shuffle(tracks)
}

func (m *Music) artistDeep(artist Artist, depth int) []Track {
	tracks := m.artistDeepTracks(artist, depth)
	return shuffle(tracks)
}

func (m *Music) genreRadio(genres []string, depth int) []Track {
	var query string
	query += fmt.Sprintf("+type:single")
	for _, g := range genres {
		query += fmt.Sprintf(` +genre:"%s"`, g)
	}
	tracks := m.Search(query, depth)
	return shuffle(tracks)
}

func (m *Music) decadeRadio(dstart, dend int, depth int) []Track {
	var query string
	query += fmt.Sprintf("+type:single")
	query += fmt.Sprintf(` +date:>="%d-01-01" +date:<="%d-12-31"`, dstart, dend)
	tracks := m.Search(query, depth)
	return shuffle(tracks)
}

func (m *Music) decadeGenreRadio(dstart, dend int, genres []string, depth int) []Track {
	var query string
	query += fmt.Sprintf("+type:single")
	query += fmt.Sprintf(` +date:>="%d-01-01" +date:<="%d-12-31"`, dstart, dend)
	for _, g := range genres {
		query += fmt.Sprintf(` +genre:"%s"`, g)
	}
	tracks := m.Search(query, depth)
	return shuffle(tracks)
}
