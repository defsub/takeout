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
	"net/url"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/defsub/takeout/client"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/search"
	"gorm.io/gorm"
)

const (
	TakeoutUser    = "takeout"
	VariousArtists = "Various Artists"
)

type Music struct {
	config     *config.Config
	db         *gorm.DB
	s3         *s3.S3
	client     *client.Client
	coverCache map[string]string
}

func NewMusic(config *config.Config) *Music {
	return &Music{
		config:     config,
		coverCache: make(map[string]string),
		client:     client.NewClient(config),
	}
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

// Get the URL for the release cover from The Cover Art Archive. Use
// REID front cover.
//
// See https://musicbrainz.org/doc/Cover_Art_Archive/API
func (m *Music) cover(r Release, size string) string {
	var url string
	if r.GroupArtwork {
		url = fmt.Sprintf("https://coverartarchive.org/release-group/%s", r.RGID)
	} else {
		url = fmt.Sprintf("https://coverartarchive.org/release/%s", r.REID)
	}
	if r.Artwork && r.FrontArtwork {
		// user front-250, front-500, front-1200
		return fmt.Sprintf("%s/front-%s", url, size)
	} else if r.Artwork && r.OtherArtwork != "" {
		// use id-250, id-500, id-1200
		return fmt.Sprintf("%s/%s-%s", url, r.OtherArtwork, size)
	} else {
		return "/static/album-white-36dp.svg"
	}
}

// Track cover based on assigned release.
func (m *Music) trackCover(t Track, size string) string {
	// TODO should expire the cache
	v, ok := m.coverCache[t.REID]
	if ok {
		return v
	}
	release, _ := m.assignedRelease(&t)
	if release == nil {
		v = ""
	} else {
		v = m.cover(*release, size)
	}
	m.coverCache[t.REID] = v
	return v
}

// URL to stream track from the S3 bucket. This will be signed and
// expired based on config.
func (m *Music) TrackURL(t *Track) *url.URL {
	url := m.bucketURL(t)
	return url
}

// Find track using the etag from the S3 bucket.
func (m *Music) TrackLookup(etag string) *Track {
	track, _ := m.lookupETag(etag)
	return track
}

// URL for track cover image.
func (m *Music) TrackImage(t Track) *url.URL {
	url, _ := url.Parse(m.trackCover(t, "front-250"))
	return url
}

func (m *Music) newSearch() *search.Search {
	s := search.NewSearch(m.config)
	s.Keywords = []string{
		FieldGenre,
		FieldStatus,
		FieldTag,
		FieldType,
	}
	s.Open("music")
	return s
}

func (m *Music) Search(q string, limit ...int) []Track {
	s := m.newSearch()
	defer s.Close()

	l := m.config.Music.SearchLimit
	if len(limit) == 1 {
		l = limit[0]
	}

	keys, err := s.Search(q, l)
	if err != nil {
		return nil
	}

	// split potentially large # of result keys into chunks to query
	chunkSize := 100
	var tracks []Track
	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize
		if end > len(keys) {
			end = len(keys)
		}
		chunk := keys[i:end]
		tracks = append(tracks, m.tracksFor(chunk)...)
	}

	return tracks
}
