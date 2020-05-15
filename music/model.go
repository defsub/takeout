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
	"github.com/jinzhu/gorm"
	"time"
)

type Artist struct {
	gorm.Model
	Name     string `gorm:"unique_index:idx_artist"`
	SortName string
	MBID     string
}

type Release struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_release"`
	Name   string `gorm:"unique_index:idx_release"`
	MBID   string `gorm:"unique_index:idx_release"`
	Asin   string
	Type   string
	Date   time.Time
}

type Popular struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_popular"`
	Title  string `gorm:"unique_index:idx_popular"`
	Rank   uint
}

func (Popular) TableName() string {
	return "popular" // not populars
}

type Similar struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_similar"`
	MBID   string `gorm:"unique_index:idx_similar"`
	Rank   uint
}

func (Similar) TableName() string {
	return "similar" // not similars
}

type ArtistTag struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_tag"`
	Tag    string `gorm:"unique_index:idx_tag"`
	Count  uint
}

type Track struct {
	gorm.Model
	Artist       string `spiff:"creator"`
	Release      string `spiff:"album"`
	TrackNum     uint   `spiff:"tracknum"`
	DiscNum      uint
	Title        string `spiff:"title"`
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	Location     string `gorm:"-" spiff:"location"`
}
