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
	"net/url"
	"regexp"
	"strconv"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/log"
	"github.com/defsub/takeout/spiff"
)

func (m *Music) addTrackEntries(tracks []Track, entries []spiff.Entry) []spiff.Entry {
	for _, t := range tracks {
		e := spiff.Entry{
			Creator:    t.Artist,
			Album:      t.ReleaseTitle,
			Title:      t.Title,
			Image:      m.TrackImage(t).String(),
			Location:   []string{trackLocation(t)},
			Identifier: []string{t.ETag},
			Size:       []int64{t.Size}}
		entries = append(entries, e)
	}
	return entries
}

// Artist Track Refs:
// /music/artists/{id}/singles - artist tracks released as singles
// /music/artists/{id}/popular - artist tracks that are popular (lastfm)
// /music/artists/{id}/tracks - artist tracks
// /music/artists/{id}/similar - artist and similar artist tracks (radio)
// /music/artists/{id}/shuffle - selection of shuffled artist tracks
// /music/artists/{id}/deep - atrist deep tracks
func (m *Music) resolveArtistRef(id, res string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	artist, err := m.lookupArtist(n)
	if err != nil {
		return entries, err
	}
	var tracks []Track
	switch res {
	case "singles":
		tracks = m.artistSingleTracks(artist)
	case "popular":
		tracks = m.artistPopularTracks(artist)
	case "tracks":
		tracks = m.artistTracks(artist)
	case "shuffle":
		tracks = m.artistShuffle(artist, m.config.Music.RadioLimit)
	case "similar":
		tracks = m.artistSimilar(artist,
			m.config.Music.ArtistRadioDepth,
			m.config.Music.ArtistRadioBreadth)
		if len(tracks) > m.config.Music.RadioLimit {
			tracks = tracks[:m.config.Music.RadioLimit]
		}
	case "deep":
		tracks = m.artistDeep(artist, m.config.Music.RadioLimit)
	}
	entries = m.addTrackEntries(tracks, entries)
	return entries, nil
}

// /music/releases/{id}/tracks
func (m *Music) resolveReleaseRef(id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	release, err := m.lookupRelease(n)
	if err != nil {
		return entries, err
	}
	tracks := m.releaseTracks(release)
	entries = m.addTrackEntries(tracks, entries)
	return entries, nil
}

// /music/tracks/{id}
func (m *Music) resolveTrackRef(id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	t, err := m.lookupTrack(n)
	if err != nil {
		return entries, err
	}
	entries = m.addTrackEntries([]Track{t}, entries)
	return entries, nil
}

// /music/search?q={q}[&radio=1]
func (m *Music) resolveSearchRef(uri string, entries []spiff.Entry) ([]spiff.Entry, error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Println(err)
		return entries, err
	}

	q := u.Query().Get("q")
	radio := u.Query().Get("radio")

	var tracks []Track
	if q != "" {
		limit := m.config.Music.SearchLimit
		if radio != "" {
			limit = m.config.Music.RadioSearchLimit
		}
		tracks = m.Search(q, limit)
	}

	if radio != "" {
		tracks = shuffle(tracks)
		if len(tracks) > m.config.Music.RadioLimit {
			tracks = tracks[:m.config.Music.RadioLimit]
		}
	}

	entries = m.addTrackEntries(tracks, entries)
	return entries, nil
}

// /music/radio/{id}
func (m *Music) resolveRadioRef(id string, entries []spiff.Entry, user *auth.User) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	s, err := m.lookupStation(n)
	if err != nil {
		return entries, err
	}
	if !s.visible(user) {
		return entries, err
	}

	m.stationRefresh(&s, user)
	plist, _ := spiff.Unmarshal(s.Playlist)
	entries = append(entries, plist.Spiff.Entries...)

	return entries, nil
}

func (m *Music) Resolve(user *auth.User, plist *spiff.Playlist) (err error) {
	var entries []spiff.Entry

	artistsRegexp := regexp.MustCompile(`/music/artists/([\d]+)/([\w]+)`)
	releasesRegexp := regexp.MustCompile(`/music/releases/([\d]+)/tracks`)
	tracksRegexp := regexp.MustCompile(`/music/tracks/([\d]+)`)
	searchRegexp := regexp.MustCompile(`/music/search.*`)
	radioRegexp := regexp.MustCompile(`/music/radio/([\d]+)`)

	for _, e := range plist.Spiff.Entries {
		if e.Ref == "" {
			entries = append(entries, e)
			continue
		}

		pathRef := e.Ref

		matches := artistsRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = m.resolveArtistRef(matches[1], matches[2], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = releasesRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = m.resolveReleaseRef(matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = tracksRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = m.resolveTrackRef(matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		if searchRegexp.MatchString(pathRef) {
			entries, err = m.resolveSearchRef(pathRef, entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = radioRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = m.resolveRadioRef(matches[1], entries, user)
			if err != nil {
				return err
			}
			continue
		}
	}

	plist.Spiff.Entries = entries

	return nil
}
