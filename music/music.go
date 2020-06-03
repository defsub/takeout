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
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/defsub/takeout/config"
	"github.com/jinzhu/gorm"
	"net/url"
	"regexp"
	"strings"
)

type Music struct {
	config *config.Config
	db     *gorm.DB
	s3     *s3.S3
}

func NewMusic(config *config.Config) *Music {
	return &Music{config: config}
}

func (m *Music) Open() (err error) {
	err = m.openDB()
	if err == nil {
		err = m.openBucket()
	}
	return
}

func (m *Music) Close() {
	m.closeDB()
}

func (m *Music) SyncBucketTracks() (err error) {
	m.deleteTracks()
	trackCh, err := m.SyncFromBucket()
	if err != nil {
		return err
	}
	for t := range trackCh {
		fmt.Printf("sync: %s / %s / %s\n", t.Artist, t.Release, t.Title)
		t.Artist = fixName(t.Artist)
		t.Release = fixName(t.Release)
		t.Title = fixName(t.Title)
		// TODO: title may have underscores - picard
		m.createTrack(t)
	}
	return
}

func (m *Music) SyncReleases() {
	artists := m.artists()
	for _, a := range artists {
		fmt.Printf("releases for %s\n", a.Name)
		if a.Name == "Various Artists" {
			// skipping!
			continue
		}

		releases, _ := m.MusicBrainzReleaseGroups(&a)
		for _, r := range releases {
			r.Name = fixName(r.Name)
			m.createRelease(&r)
		}

		checked := make(map[string]int)
		tracks := m.artistTracksWithoutReleases(a.Name)
		for _, t := range tracks {
			_, ok := checked[t.Release]
			if ok {
				continue
			}
			releases := m.MusicBrainzReleases(&a, t.Release)
			fmt.Printf("checking %s / %s - found %d\n", t.Artist, t.Release, len(releases))
			for _, r := range releases {
				r.Name = fixName(r.Name)
				m.createRelease(&r)
			}
			checked[t.Release] = len(releases)
		}
	}
}

func fuzzyArtist(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9& -]`)
	return re.ReplaceAllString(name, "")
}

func fuzzyName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(name, "")
}

func fixName(name string) string {
	// TODO: use Map?
	name = strings.Replace(name, "–", "-", -1)
	name = strings.Replace(name, "‐", "-", -1)
	name = strings.Replace(name, "’", "'", -1)
	name = strings.Replace(name, "‘", "'", -1)
	name = strings.Replace(name, "“", "\"", -1)
	name = strings.Replace(name, "”", "\"", -1)
	name = strings.Replace(name, "…", "...", -1)
	return name
}

func (m *Music) FixTrackReleases() error {
	fixReleases := make(map[string]string)
	tracks := m.tracksWithoutReleases()

	for _, t := range tracks {
		artist := m.artist(t.Artist)
		if artist == nil {
			fmt.Printf("artist not found: %s\n", t.Artist)
			continue
		}

		_, ok := fixReleases[t.Release]
		if ok {
			continue
		}

		if artist.Name == "Various Artists" {
			continue
		}

		releases := m.artistReleasesLike(artist, t.Release)
		if len(releases) == 1 {
			fixReleases[t.Release] = releases[0].Name
		} else {
			releases = m.releases(artist)
			matched := false
			for _, r := range releases {
				// try fuzzy match
				if fuzzyName(t.Release) == fuzzyName(r.Name) {
					fixReleases[t.Release] = r.Name
					matched = true
					break
				}
			}
			if !matched {
				// use config
				mbid, ok := m.config.Music.UserReleaseID(t.Release)
				if ok {
					r, _ := m.MusicBrainzRelease(artist, mbid)
					if r != nil {
						r.Name = fixName(r.Name)
						m.createRelease(r)

						fixReleases[t.Release] = r.Name
						matched = true
					}
				}
			}
			if !matched {
				fmt.Printf("unmatched '%s' / '%s'\n", t.Artist, t.Release)
			}
		}
	}

	for oldName, newName := range fixReleases {
		err := m.updateTrackRelease(oldName, newName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Music) SyncPopular() {
	artists := m.artists()
	for _, a := range artists {
		fmt.Printf("popular for %s\n", a.Name)
		popular := m.lastfmArtistTopTracks(&a)
		for _, p := range popular {
			m.createPopular(&p)
		}
	}
}

func (m *Music) SyncSimilar() {
	artists := m.artists()
	for _, a := range artists {
		fmt.Printf("similar for %s\n", a.Name)
		similar := m.lastfmSimilarArtists(&a)
		for _, s := range similar {
			m.createSimilar(&s)
		}
	}
}

func (m *Music) SyncArtists() error {
	artists := m.trackArtistNames()
	for _, name := range artists {
		var tags []ArtistTag
		artist := m.artist(name)
		if artist == nil {
			artist, tags = m.resolveArtist(name)
			if artist != nil {
				artist.Name = fixName(artist.Name)
				fmt.Printf("creating %s\n", artist.Name)
				m.createArtist(artist)
				for _, t := range tags {
					t.Artist = artist.Name
					m.createArtistTag(&t)
				}
			}
		}

		if artist == nil {
			err := errors.New(fmt.Sprintf("'%s' artist not found", name))
			fmt.Printf("%s\n", err)
			continue
		}

		if name != artist.Name {
			// fix track artist name: AC_DC -> AC/DC
			fmt.Printf("fixing name %s to %s\n", name, artist.Name)
			m.updateTrackArtist(name, artist.Name)
		}
	}
	return nil
}

func (m *Music) resolveArtist(name string) (artist *Artist, tags []ArtistTag) {
	artist, tags = m.SearchArtist(name)
	if artist == nil {
		// try again
		artist, tags = m.SearchArtist(fuzzyArtist(name))
	}
	if artist == nil {
		// try lastfm
		artist = m.lastfmArtistSearch(name)
		if artist != nil {
			fmt.Printf("try lastfm got %s mbid:'%s'\n", artist.Name, artist.ARID)
			// resolve with mbz
			if artist.ARID != "" {
				artist, tags = m.SearchArtistId(artist.ARID)
			} else {
				artist = nil
			}
		}
	}
	return
}

func (m *Music) TrackURL(t *Track) *url.URL {
	url := m.bucketURL(t)
	return url
}

func (m *Music) Tracks(tags string, dr *DateRange) []Track {
	return m.tracks(tags, dr)
}

func (m *Music) Singles(tags string, dr *DateRange) []Track {
	return m.singleTracks(tags, dr)
}

func (m *Music) Popular(tags string, dr *DateRange) []Track {
	return m.popularTracks(tags, dr)
}

func (m *Music) ArtistSingles(artists string, dr *DateRange) []Track {
	return m.artistSingleTracks(artists, dr)
}

func (m *Music) ArtistTracks(artists string, dr *DateRange) []Track {
	return m.artistTracks(artists, dr)
}

func (m *Music) ArtistPopular(artists string, dr *DateRange) []Track {
	return m.artistPopularTracks(artists, dr)
}

func (m *Music) SimilarArtists(artist string) []Artist {
	a := m.artist(artist)
	return m.similarArtists(a)
}

func (m *Music) ArtistRelease(artist string, release string) []Track {
	return m.artistReleaseTracks(artist, release)
}
