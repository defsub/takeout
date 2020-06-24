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

import (
	"github.com/defsub/takeout/spiff"
	"regexp"
	"strconv"
)

func (m *Music) addTrackEntries(tracks []Track, entries []spiff.Entry) []spiff.Entry {
	for _, t := range tracks {
		e := spiff.Entry{
			Creator:    t.Artist,
			Album:      t.Release,
			Title:      t.Title,
			Image:      m.TrackImage(&t).String(),
			Location:   []string{m.TrackURL(&t).String()},
			Identifier: []string{t.ETag}}
		entries = append(entries, e)
	}
	return entries
}

// /music/artists/{id}/{singles}
// /music/artists/{id}/{popular}
func (m *Music) resolveArtistRef(id, res string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	artist, err := m.lookupArtist(uint(n))
	if err != nil {
		return entries, err
	}
	var tracks []Track
	switch res {
	case "singles":
		tracks = m.artistSingleTracks(artist.Name, nil)
	case "popular":
		tracks = m.artistPopularTracks(artist.Name, nil)
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
	release, err := m.lookupRelease(uint(n))
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
	t, err := m.lookupTrack(uint(n))
	if err != nil {
		return entries, err
	}
	entries = m.addTrackEntries([]Track{t}, entries)
	return entries, nil
}

func (m *Music) Resolve(plist *spiff.Playlist) (err error) {
	var entries []spiff.Entry

	artistsRegexp := regexp.MustCompile(`/music/artists/([\d]+)/([\w]+)`)
	releasesRegexp := regexp.MustCompile(`/music/releases/([\d]+)/tracks`)
	tracksRegexp := regexp.MustCompile(`/music/tracks/([\d]+)`)

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
	}

	plist.Spiff.Entries = entries

	return nil
}
