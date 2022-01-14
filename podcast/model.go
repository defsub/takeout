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

package podcast

import (
	"gorm.io/gorm"
	"time"
)

type Series struct {
	gorm.Model
	SID         string `gorm:"uniqueIndex:idx_series"` // hash of link
	Title       string
	Description string
	Link        string
	Image       string
	Copyright   string
	Date        time.Time // last build date
	TTL         int
}

func (Series) TableName() string {
	return "series" // series is zero plural
}

type Episode struct {
	gorm.Model
	SID         string // series ID
	EID         string `gorm:"uniqueIndex:idx_episode"` // hash of GUID
	Title       string
	Link        string
	Description string
	ContentType string
	Size        int64
	URL         string
	Date        time.Time // publish time
}
