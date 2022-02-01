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
	"github.com/defsub/takeout/lib/gorm"
	"time"
)

type Movie struct {
	gorm.Model
	TMID             int64 `gorm:"uniqueIndex:idx_movie_tmid"`
	IMID             string
	Title            string
	Date             time.Time
	Rating           string
	Tagline          string
	OriginalTitle    string
	OriginalLanguage string
	Overview         string
	Budget           int64
	Revenue          int64
	Runtime          int
	VoteAverage      float32
	VoteCount        int
	BackdropPath     string
	PosterPath       string
	SortTitle        string
	Key              string
	Size             int64
	ETag             string
	LastModified     time.Time
}

type Collection struct {
	gorm.Model
	Name     string
	SortName string
	TMID     int64
}

type Genre struct {
	gorm.Model
	TMID int64
	Name string
}

type Keyword struct {
	gorm.Model
	TMID int64
	Name string
}

type Person struct {
	gorm.Model
	PEID        int64 `gorm:"uniqueIndex:idx_person_peid"`
	IMID        string
	Name        string
	ProfilePath string
	Bio         string
	Birthplace  string
	Birthday    time.Time
	Deathday    time.Time
}

func (Person) TableName() string {
	return "people" // not peoples
}

type Cast struct {
	gorm.Model
	TMID      int64 `gorm:"index:idx_cast_tmid"`
	PEID      int64 `gorm:"index:idx_cast_peid"`
	Character string
	Rank      int
	Person    Person `gorm:"-"`
}

func (Cast) TableName() string {
	return "cast" // not casts
}

type Crew struct {
	gorm.Model
	TMID       int64
	PEID       int64
	Department string
	Job        string
	Person     Person `gorm:"-"`
}

func (Crew) TableName() string {
	return "crew" // not crews
}

type Recommend struct {
	Name   string
	Movies []Movie
}
