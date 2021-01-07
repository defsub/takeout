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
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/spiff"
)

const (
	typeArtist  = "artist"  // Songs by single artist
	typeGenre   = "genre"   // Songs from one or more genres
	typeSimilar = "similar" // Songs from similar artists
	typePeriod  = "period"  // Songs from one or more time periods
	typeSeries  = "series"  // Songs from one or more series (chart)
	typeOther   = "other"
)

func (s *Station) visible(user *auth.User) bool {
	return s.User == user.Name || s.Shared
}

func (m *Music) ClearStations() {
	m.clearStationPlaylists()
}

func (m *Music) CreateStations() {
	// artists, _ := m.favoriteArtists(25)
	// for _, v := range artists {
	// 	a := m.artist(v)
	// 	if a == nil {
	// 		continue
	// 	}
	// 	station := Station{
	// 		User:   TakeoutUser,
	// 		Shared: true,
	// 		Type:   typeArtist,
	// 		Name:   fmt.Sprintf("%s Top Tracks", a.Name),
	// 		Ref:    fmt.Sprintf(`/music/artists/%d/popular`, a.ID)}
	// 	m.createStation(&station)

	// 	station = Station{
	// 		User:   TakeoutUser,
	// 		Shared: true,
	// 		Type:   typeSimilar,
	// 		Name:   fmt.Sprintf("%s Similar", a.Name),
	// 		Ref:    fmt.Sprintf(`/music/artists/%d/similar`, a.ID)}
	// 	m.createStation(&station)
	// }

	for _, g := range m.config.Music.RadioGenres {
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   typeGenre,
			Name:   strings.Title(g),
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(fmt.Sprintf(`+genre:"%s" +popularity:<11 -artist:"Various Artists"`, g)))}
		m.createStation(&station)
	}

	decades := []int{1960, 1970, 1980, 1990, 2000, 2010, 2020}
	for _, d := range decades {
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   typePeriod,
			Name:   fmt.Sprintf("%ds Top Tracks", d),
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(fmt.Sprintf(
					`+date:>="%d-01-01" +date:<="%d-12-31" +popularity:<11`, d, d+9)))}
		m.createStation(&station)
	}

	for _, s := range m.config.Music.RadioSeries {
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   typeSeries,
			Name:   s,
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(fmt.Sprintf(`+series:"%s"`, s)))}
		m.createStation(&station)
	}

	for k, v := range m.config.Music.RadioOther {
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   typeSeries,
			Name:   k,
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(v))}
		m.createStation(&station)
	}
}

func (m *Music) stationRefresh(s *Station, user *auth.User) {
	m.resolveStation(s, user)
}

func (m *Music) resolveStation(s *Station, user *auth.User) {
	plist := spiff.NewPlaylist()
	// Image
	plist.Spiff.Location = fmt.Sprintf("%s/api/radio/%d", m.config.Server.URL, s.ID)
	plist.Spiff.Title = s.Name
	plist.Spiff.Creator = "Radio"
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

func (m *Music) artistSimilar(artist Artist, depth int, breadth int) []Track {
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

func contains(tracks []Track, t Track) bool {
	for _, v := range tracks {
		if v.Title == t.Title && v.Artist == t.Artist {
			return true
		}
	}
	return false
}

func (m *Music) artistShuffle(artist Artist, depth int) []Track {
	var tracks []Track
	// add 50% popular
	pop := int(float32(depth) * 0.50)
	tracks = append(tracks, shuffle(m.artistPopularTracks(artist, pop))...)
	// randomly add some unique tracks
	// TODO consider other algorithms
	all := shuffle(m.artistTracks(artist))
	pick := 0
	for len(tracks) < depth && pick < len(all) {
		t := all[pick]
		if !contains(tracks, t) {
			tracks = append(tracks, t)
		}
		pick++
	}
	return shuffle(tracks)
}

func (m *Music) artistDeep(artist Artist, depth int) []Track {
	tracks := m.artistDeepTracks(artist, depth)
	return shuffle(tracks)
}

// func (m *Music) genreRadio(genres []string, depth int) []Track {
// 	var query string
// 	query += fmt.Sprintf("+type:single")
// 	for _, g := range genres {
// 		query += fmt.Sprintf(` +genre:"%s"`, g)
// 	}
// 	tracks := m.Search(query, depth)
// 	return shuffle(tracks)
// }

// func (m *Music) decadeRadio(dstart, dend int, depth int) []Track {
// 	var query string
// 	query += fmt.Sprintf("+type:single")
// 	query += fmt.Sprintf(` +date:>="%d-01-01" +date:<="%d-12-31"`, dstart, dend)
// 	tracks := m.Search(query, depth)
// 	return shuffle(tracks)
// }

// func (m *Music) decadeGenreRadio(dstart, dend int, genres []string, depth int) []Track {
// 	var query string
// 	query += fmt.Sprintf("+type:single")
// 	query += fmt.Sprintf(` +date:>="%d-01-01" +date:<="%d-12-31"`, dstart, dend)
// 	for _, g := range genres {
// 		query += fmt.Sprintf(` +genre:"%s"`, g)
// 	}
// 	tracks := m.Search(query, depth)
// 	return shuffle(tracks)
// }
