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
	"gorm.io/gorm"
	"time"
)

// Artist info from MusicBrainz.
type Artist struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex:idx_artist_name"`
	SortName string
	ARID     string `gorm:"uniqueIndex:idx_artist_arid"`
}

// Release info from MusicBrainz.
type Release struct {
	gorm.Model
	Artist         string `gorm:"uniqueIndex:idx_release"`
	Name           string `gorm:"uniqueIndex:idx_release"`
	RGID           string //`gorm:"uniqueIndex:idx_release"`
	REID           string `gorm:"uniqueIndex:idx_release"`
	Disambiguation string
	Asin           string
	Type           string
	Date           time.Time
	TrackCount     int
	FrontCover     bool
	Media          []Media `gorm:"-"`
}

// Release Media from MusicBrainz.
type Media struct {
	gorm.Model
	REID       string `gorm:"uniqueIndex:idx_media"`
	Name       string `gorm:"uniqueIndex:idx_media"`
	Position   int    `gorm:"uniqueIndex:idx_media"`
	Format     string
	TrackCount int
}

// Popular tracks for an artist from Last.fm.
type Popular struct {
	gorm.Model
	Artist string `gorm:"uniqueIndex:idx_popular"`
	Title  string `gorm:"uniqueIndex:idx_popular"`
	Rank   int
}

func (Popular) TableName() string {
	return "popular" // not populars
}

// Similar artist info from Last.fm
type Similar struct {
	gorm.Model
	Artist string `gorm:"uniqueIndex:idx_similar"`
	ARID   string `gorm:"uniqueIndex:idx_similar"`
	Rank   int
}

func (Similar) TableName() string {
	return "similar" // not similars
}

// Artist tags from MusicBrainz.
type ArtistTag struct {
	gorm.Model
	Artist string `gorm:"uniqueIndex:idx_tag"`
	Tag    string `gorm:"uniqueIndex:idx_tag"`
	Count  int
}

// Tracks from S3 bucket object names. Naming is adjusted based on
// data from MusicBrainz.
type Track struct {
	gorm.Model
	Artist       string `spiff:"creator"`
	Release      string
	Date         string
	TrackNum     int `spiff:"tracknum"`
	DiscNum      int
	Title        string `spiff:"title"`
	Key          string // TODO - unique constraint
	Size         int64
	ETag         string
	LastModified time.Time
	// TODO remove Location, only needed for old spiff
	Location     []string `gorm:"-" spiff:"location"`
	TrackCount   int
	REID         string
	RGID         string
	MediaTitle   string
	ReleaseTitle string `spiff:"album"`
}

type Playlist struct {
	gorm.Model
	User     string `gorm:"uniqueIndex:idx_playlist"`
	Playlist []byte
}

type Station struct {
	gorm.Model
	User     string `gorm:"uniqueIndex:idx_station"`
	Name     string `gorm:"uniqueIndex:idx_station"`
	Ref      string
	Shared   bool
	Type     string
	Playlist []byte `json:"-"`
}
