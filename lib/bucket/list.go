// Copyright (C) 2021 The Takeout Authors.
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

package bucket

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/defsub/takeout/config"
)

type Bucket struct {
	config *config.BucketConfig
	s3     *s3.S3
}

type Object struct {
	Key          string
	Path         string // Key modified by rewrite rules
	ETag         string
	Size         int64
	LastModified time.Time
}

func OpenAll(buckets []config.BucketConfig) ([]*Bucket, error) {
	var list []*Bucket

	for i := range buckets {
		b, err := Open(buckets[i])
		if err == nil {
			return list, err
		}
		list = append(list, b)
	}

	return list, nil
}

func OpenMedia(buckets []config.BucketConfig, mediaType string) ([]Bucket, error) {
	var list []Bucket

	for i := range buckets {
		if buckets[i].Media != mediaType {
			continue
		}
		b, err := Open(buckets[i])
		if err != nil {
			return list, err
		}
		list = append(list, *b)
	}

	return list, nil
}

// Connect to the configured S3 bucket.
// Tested: Wasabi, Backblaze, Minio
func Open(config config.BucketConfig) (*Bucket, error) {
	creds := credentials.NewStaticCredentials(
		config.AccessKeyID,
		config.SecretAccessKey, "")
	s3Config := &aws.Config{
		Credentials:      creds,
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String(config.Region),
		S3ForcePathStyle: aws.Bool(true)}
	session, err := session.NewSession(s3Config)
	bucket := &Bucket{
		s3:     s3.New(session),
		config: &config,
	}
	return bucket, err
}

func (b *Bucket) List(lastSync time.Time) (objectCh chan *Object, err error) {
	objectCh = make(chan *Object)

	go func() {
		defer close(objectCh)

		var continuationToken *string
		continuationToken = nil
		for {
			req := s3.ListObjectsV2Input{
				Bucket: aws.String(b.config.BucketName),
				Prefix: aws.String(b.config.ObjectPrefix)}
			if continuationToken != nil {
				req.ContinuationToken = continuationToken
			}
			resp, err := b.s3.ListObjectsV2(&req)
			if err != nil {
				break
			}
			for _, obj := range resp.Contents {
				if obj.LastModified != nil &&
					obj.LastModified.After(lastSync) {
					objectCh <- &Object{
						Key:          *obj.Key,
						Path:         b.Rewrite(*obj.Key),
						ETag:         *obj.ETag,
						Size:         *obj.Size,
						LastModified: *obj.LastModified,
					}
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

// Generate a presigned url which expires based on config settings.
func (b *Bucket) Presign(key string) *url.URL {
	req, _ := b.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(b.config.BucketName),
		Key:    aws.String(key)})
	urlStr, _ := req.Presign(b.config.URLExpiration)
	url, _ := url.Parse(urlStr)
	return url
}


func (b *Bucket) Rewrite(path string) string {
	result := path
	for _, rule := range b.config.RewriteRules {
		re := regexp.MustCompile(rule.Pattern)
		matches := re.FindStringSubmatch(result)
		if matches != nil {
			result = rule.Replace
			for i := range matches {
				result = strings.ReplaceAll(result, fmt.Sprintf("$%d", i), matches[i])
			}
		}
	}
	if result != path {
		fmt.Printf("rewrite %s -> %s\n", path, result)
	}
	return result
}
