package music

import (
	"github.com/defsub/takeout/config"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/minio/minio-go/v6"
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
	minioClient, err := minio.New(bucket.Endpoint, bucket.AccessKeyID, bucket.SecretAccessKey, bucket.UseSSL)
	if err == nil {
		m.minio = minioClient
	}
	return err
}

func (m *Music) SyncFromBucket() (trackCh chan *Track, err error) {
	trackCh = make(chan *Track)

	go func() {
		defer close(trackCh)

		doneCh := make(chan struct{})
		defer close(doneCh)

		isRecursive := true
		bucket := m.bucketConfig()
		objectCh := m.minio.ListObjectsV2(bucket.BucketName, bucket.ObjectPrefix, isRecursive, doneCh)
		for object := range objectCh {
			if object.Err != nil {
				break
			}
			checkObject(&object, trackCh)
		}
	}()

	return
}

func checkObject(object *minio.ObjectInfo, trackCh chan *Track) {
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

func (m *Music) objectURL(t Track) *url.URL {
	reqParams := make(url.Values)
	// Generates a presigned url which expires in a day.
	url, _ := m.minio.PresignedGetObject(m.config.Music.Bucket.BucketName, t.Key, time.Second*24*60*60, reqParams)
	return url
}
