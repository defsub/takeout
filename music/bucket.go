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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/defsub/takeout/config"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

func (m *Music) bucketConfig() config.MusicBucket {
	return m.config.Music.Bucket
}

// Connect to the configured S3 bucket.
// Tested: Wasabi
func (m *Music) openBucket() error {
	bucket := m.bucketConfig()
	creds := credentials.NewStaticCredentials(
		bucket.AccessKeyID, bucket.SecretAccessKey, "")
	s3Config := &aws.Config{
		Credentials:      creds,
		Endpoint:         aws.String(bucket.Endpoint),
		Region:           aws.String(bucket.Region),
		S3ForcePathStyle: aws.Bool(true)}
	session, err := session.NewSession(s3Config)
	m.s3 = s3.New(session)
	return err
}

// Asynchronously obtain all tracks from the bucket.
func (m *Music) syncFromBucket(lastSync time.Time) (trackCh chan *Track, err error) {
	trackCh = make(chan *Track)

	go func() {
		defer close(trackCh)
		bucket := m.bucketConfig()

		var continuationToken *string
		continuationToken = nil
		for {
			req := s3.ListObjectsV2Input{
				Bucket: aws.String(bucket.BucketName),
				Prefix: aws.String(bucket.ObjectPrefix)}
			if continuationToken != nil {
				req.ContinuationToken = continuationToken
			}
			resp, err := m.s3.ListObjectsV2(&req)
			if err != nil {
				break
			}
			for _, obj := range resp.Contents {
				if obj.LastModified != nil &&
					obj.LastModified.After(lastSync) {
					checkObject(obj, trackCh)
				}
			}

			if !*resp.IsTruncated {
				break
			}
			continuationToken = resp.NextContinuationToken
		}
	}()

	return
}

func checkObject(object *s3.Object, trackCh chan *Track) {
	matchPath(*object.Key, trackCh, func(t *Track, trackCh chan *Track) {
		t.Key = *object.Key
		t.ETag = *object.ETag
		t.Size = *object.Size
		t.LastModified = *object.LastModified
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

var trackRegexp = regexp.MustCompile(`(?:([\d]+)-)?([\d]+)-(.*)\.(mp3|flac|ogg|m4a)$`)

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
	return true
}

// Generate a presigned url which expires based on config settings.
func (m *Music) bucketURL(t *Track) *url.URL {
	req, _ := m.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(m.config.Music.Bucket.BucketName),
		Key: aws.String(t.Key)})
	urlStr, _ := req.Presign(m.config.Music.Bucket.URLExpiration)
	url, _ := url.Parse(urlStr)
	return url
}
