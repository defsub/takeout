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
	"gorm.io/gorm"
	"time"
)

// Artist info from MusicBrainz.
type Artist struct {
	gorm.Model
	Name           string `gorm:"uniqueIndex:idx_artist_name"`
	SortName       string
	ARID           string `gorm:"uniqueIndex:idx_artist_arid"`
	Disambiguation string
	Country        string
	Area           string
	Date           time.Time
	EndDate        time.Time
	Genre          string
}

// Release info from MusicBrainz.
type Release struct {
	gorm.Model
	Artist         string `gorm:"uniqueIndex:idx_release"`
	Name           string `gorm:"uniqueIndex:idx_release;index:idx_release_name" sql:"collate:nocase"`
	RGID           string `gorm:"index:idx_release_rgid"`
	REID           string `gorm:"uniqueIndex:idx_release;index:idx_release_reid"`
	Disambiguation string
	Asin           string
	Country        string
	Type           string
	Date           time.Time // rg first release
	ReleaseDate    time.Time // re release date
	Status         string
	TrackCount     int
	DiscCount      int
	Artwork        bool
	FrontArtwork   bool
	BackArtwork    bool
	OtherArtwork   string
	GroupArtwork   bool
	Media          []Media `gorm:"-"`
}

func (r Release) official() bool {
	return r.Status == "Official"
}

func (r Release) Cover(size string) string {
	return Cover(r, size)
}

func (r Release) CoverSmall() string {
	return Cover(r, "250")
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
	Artist       string `spiff:"creator" gorm:"index:idx_track_artist"`
	Release      string `gorm:"index:idx_track_release"`
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
	DiscCount    int
	REID         string `gorm:"index:idx_track_reid"`
	RGID         string `gorm:"index:idx_track_rgid"`
	MediaTitle   string
	ReleaseTitle string `spiff:"album"`
	ReleaseDate  time.Time
	Artwork      bool
	FrontArtwork bool
	BackArtwork  bool
	OtherArtwork string
	GroupArtwork bool
}

func (t Track) releaseKey() string {
	return fmt.Sprintf("%s/%s/%d/%d", t.Artist, t.Release, t.TrackCount, t.DiscCount)
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

type ArtistImage struct {
	gorm.Model
	Artist string `gorm:"uniqueIndex:idx_artist_img"`
	URL    string `gorm:"uniqueIndex:idx_artist_img"`
	Source string `gorm:"uniqueIndex:idx_artist_img"`
	Rank   int
}

type ArtistBackground struct {
	gorm.Model
	Artist string `gorm:"uniqueIndex:idx_artist_bg"`
	URL    string `gorm:"uniqueIndex:idx_artist_bg"`
	Source string `gorm:"uniqueIndex:idx_artist_bg"`
	Rank   int
}

// type Scrobble struct {
// 	gorm.Model
// 	User    string
// 	Artist  string
// 	Release string
// 	Title   string
// 	Date    time.Time
// }
