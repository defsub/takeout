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

package video

import (
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/bucket"
	"github.com/defsub/takeout/lib/client"
	"github.com/defsub/takeout/lib/tmdb"
	"gorm.io/gorm"
	"net/url"
)

type Video struct {
	config  *config.Config
	db      *gorm.DB
	client  *client.Client
	buckets []*bucket.Bucket
}

func NewVideo(config *config.Config) *Video {
	return &Video{
		config: config,
		client: client.NewClient(config),
	}
}

func (v *Video) Open() (err error) {
	err = v.openDB()
	if err == nil {
		v.buckets, err = bucket.OpenMedia(v.config.Buckets, config.MediaVideo)
	}
	return
}

func (v *Video) Close() {
	v.closeDB()
}

func (v *Video) List() {
	v.Sync()
}

func (v *Video) MovieURL(m *Movie) *url.URL {
	// FIXME assume first bucket!!!
	return v.buckets[0].Presign(m.Key)
}

func (v *Video) MoviePoster(m Movie) *url.URL {
	client := tmdb.NewTMDB(v.config)
	return client.MoviePoster(m.PosterPath, tmdb.Poster342)
}

func (v *Video) MoviePosterSmall(m Movie) *url.URL {
	client := tmdb.NewTMDB(v.config)
	return client.MoviePoster(m.PosterPath, tmdb.Poster154)
}

func (v *Video) MovieBackdrop(m Movie) *url.URL {
	client := tmdb.NewTMDB(v.config)
	return client.MovieBackdrop(m.BackdropPath, tmdb.Backdrop1280)
}

func (v *Video) PersonProfile(p Person) *url.URL {
	client := tmdb.NewTMDB(v.config)
	return client.PersonProfile(p.ProfilePath, tmdb.Profile185)
}
