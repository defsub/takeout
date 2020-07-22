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
	"github.com/jinzhu/gorm"
	"time"
)

// Artist info from MusicBrainz.
type Artist struct {
	gorm.Model
	Name     string `gorm:"unique_index:idx_artist_name"`
	SortName string
	ARID     string `gorm:"unique_index:idx_artist_arid"`
}

// Release info from MusicBrainz.
type Release struct {
	gorm.Model
	Artist         string `gorm:"unique_index:idx_release"`
	Name           string `gorm:"unique_index:idx_release"`
	RGID           string //`gorm:"unique_index:idx_release"`
	REID           string `gorm:"unique_index:idx_release"`
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
	REID       string `gorm:"unique_index:idx_media"`
	Name       string `gorm:"unique_index:idx_media"`
	Position   int    `gorm:"unique_index:idx_media"`
	Format     string
	TrackCount int
}

// Popular tracks for an artist from Last.fm.
type Popular struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_popular"`
	Title  string `gorm:"unique_index:idx_popular"`
	Rank   int
}

func (Popular) TableName() string {
	return "popular" // not populars
}

// Similar artist info from Last.fm
type Similar struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_similar"`
	ARID   string `gorm:"unique_index:idx_similar"`
	Rank   int
}

func (Similar) TableName() string {
	return "similar" // not similars
}

// Artist tags from MusicBrainz.
type ArtistTag struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_tag"`
	Tag    string `gorm:"unique_index:idx_tag"`
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
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	Location     []string `gorm:"-" spiff:"location"`
	TrackCount   int
	REID         string
	RGID         string
	MediaTitle   string
	ReleaseTitle string `spiff:"album"`
}

type UserPlaylist struct {
	gorm.Model
	User     string
	Playlist []byte
}

// type Criteria struct {
// 	Name     string
// 	Artists  string
// 	Releases string
// 	Titles   string
// 	Tags     string
// 	After    string
// 	Before   string
// 	Singles  bool
// 	Popular  bool
// 	Shuffle  bool
// }