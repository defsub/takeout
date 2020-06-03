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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/defsub/takeout/config"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

func (m *Music) bucketConfig() config.MusicBucket {
	return m.config.Music.Bucket
}

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

func (m *Music) SyncFromBucket() (trackCh chan *Track, err error) {
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
				checkObject(obj, trackCh)
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
		t.Release = matchRelease(matches[2])
		if matchTrack(matches[3], &t) {
			doMatch(&t, trackCh)
		}
	}
}

var releaseRegexp = regexp.MustCompile(`(.+)\s+\(([\d]+)\)\s*$`)

func matchRelease(release string) string {
	matches := releaseRegexp.FindStringSubmatch(release)
	if matches != nil {
		release = matches[1]
	}
	return release
}

var trackRegexp = regexp.MustCompile(`(?:([\d]+)-)?([\d]+)-(.*)\.(mp3|flac|ogg|m4a)$`)

func matchTrack(file string, t *Track) bool {
	matches := trackRegexp.FindStringSubmatch(file)
	if matches == nil {
		return false
	}
	disc, _ := strconv.Atoi(matches[1])
	track, _ := strconv.Atoi(matches[2])
	t.DiscNum = uint(disc)
	t.TrackNum = uint(track)
	t.Title = matches[3]
	if t.DiscNum == 0 {
		t.DiscNum = 1
	}
	return true
}

func (m *Music) bucketURL(t *Track) *url.URL {
	// Generates a presigned url which expires in a day.
	req, _ := m.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(m.config.Music.Bucket.BucketName),
		Key: aws.String(t.Key)})
	urlStr, _ := req.Presign(24 * time.Hour)
	url, _ := url.Parse(urlStr)
	return url
}
