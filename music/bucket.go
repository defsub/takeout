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
	"time"

	"github.com/defsub/takeout/lib/bucket"
)

// Asynchronously obtain all tracks from the bucket.
func (m *Music) syncFromBucket(bucket bucket.Bucket, lastSync time.Time) (trackCh chan *Track, err error) {
	trackCh = make(chan *Track)

	go func() {
		defer close(trackCh)
		objectCh, err := bucket.List(lastSync)
		if err != nil {
			return
		}
		for o := range objectCh {
			checkObject(o, trackCh)
		}
	}()

	return
}

func checkObject(object *bucket.Object, trackCh chan *Track) {
	matchPath(object.Key, trackCh, func(t *Track, trackCh chan *Track) {
		t.Key = object.Key
		t.ETag = object.ETag
		t.Size = object.Size
		t.LastModified = object.LastModified
		trackCh <- t
	})
}

// Examples:
// The Raconteurs / Help Us Stranger (2019) / 01-Bored and Razed.flac
// Tubeway Army / Replicas - The First Recordings (2019) / 1-01-You Are in My Vision (early version).flac
// Tubeway Army / Replicas - The First Recordings (2019) / 2-01-Replicas (early version 2).flac
var coverRegexp = regexp.MustCompile(`cover\.(png|jpg)$`)

var pathRegexp = regexp.MustCompile(`([^\/]+)\/([^\/]+)\/([^\/]+)$`)

func matchPath(path string, trackCh chan *Track, doMatch func(t *Track, music chan *Track)) {
	matches := pathRegexp.FindStringSubmatch(path)
	if matches != nil {
		var t Track
		t.Artist = matches[1]
		release, date := matchRelease(matches[2])
		if release != "" && date != "" {
			t.Release = release
			t.Date = date
		} else {
			t.Release = release
		}
		if matchTrack(matches[3], &t) {
			doMatch(&t, trackCh)
		}
	}
}

var releaseRegexp = regexp.MustCompile(`(.+?)\s*(\(([\d]+)\))?\s*$`)
// 1|1|Airlane|Music/Gary Numan/The Pleasure Principle (1998)/01-Airlane.flac
// 1|1|Airlane|Music/Gary Numan/The Pleasure Principle (2009)/1-01-Airlane.flac
//
// The Pleasure Principle
// 1: The Pleasure Principle
//
// The Pleasure Principle (2000)
// 1: The Pleasure Principle
// 2: (2000)
// 3: 2000
//
// The Pleasure Principle (Live)
// 1: The Pleasure Principle (Live)
//
// The Pleasure Principle (Live) (2000)
// 1: The Pleasure Principle (Live)
// 2: (2000)
// 3: 2000
func matchRelease(release string) (string, string) {
	var name, date string
	matches := releaseRegexp.FindStringSubmatch(release)
	if matches != nil {
		if len(matches) == 2 {
			name = matches[1]
		} else if len(matches) == 4 {
			name = matches[1]
			date = matches[3]
		}
	}
	return name, date
}

var trackRegexp = regexp.MustCompile(`(?:([1-9]+)-)?([\d]+)-(.*)\.(mp3|flac|ogg|m4a)$`)
var singleDiscRegexp = regexp.MustCompile(`([\d]+)-(.*)\.(mp3|flac|ogg|m4a)$`)
var numericRegexp = regexp.MustCompile(`^[\d]+([\s-])*`)

func matchTrack(file string, t *Track) bool {
	matches := trackRegexp.FindStringSubmatch(file)
	if matches == nil {
		return false
	}
	disc, _ := strconv.Atoi(matches[1])
	track, _ := strconv.Atoi(matches[2])
	t.DiscNum = disc
	t.TrackNum = track
	t.Title = matches[3]
	if t.DiscNum == 0 {
		t.DiscNum = 1
	}

	// potentially not multi-disc so assume single disc if too many
	// TODO make this configurable?
	// eg: 18-19-2000 (Soulchild remix).flac
	// Beatles in Mono - 13 discs
	// Eagles Legacy - 12 discs
	// Kraftwerk The Catalogue - 8 discs
	if t.DiscNum > 13 {
		matches := singleDiscRegexp.FindStringSubmatch(file)
		if matches == nil {
			return false
		}
		t.DiscNum = 1
		t.TrackNum, _ = strconv.Atoi(matches[1])
		t.Title = matches[2]
	}

	// all numeric assume is single disc since most are single
	// eg: 11-19-2000.flac
	// eg: 4-36-22-36.flac
	if numericRegexp.MatchString(t.Title) {
		matches := singleDiscRegexp.FindStringSubmatch(file)
		if matches == nil {
			return false
		}
		t.DiscNum = 1
		t.TrackNum, _ = strconv.Atoi(matches[1])
		t.Title = matches[2]
	}

	return true
}

// Generate a presigned url which expires based on config settings.
func (m *Music) bucketURL(t *Track) *url.URL {
	// TODO FIXME assume first bucket!!!
	return m.buckets[0].Presign(t.Key)
}
