// Copyright (C) 2022 The Takeout Authors.
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

// Package progress manages user progress data which contains media offset and
// duration to allow incremental watch/listen progress to be saved and
// retrieved to/from the server based on etag.
package progress

import (
	"errors"
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"gorm.io/gorm"
)

var (
	ErrOffsetTooOld = errors.New("offset is old")
	ErrAccessDenied = errors.New("access denied")
)

type Progress struct {
	config *config.Config
	db     *gorm.DB
}

func NewProgress(config *config.Config) *Progress {
	return &Progress{
		config: config,
	}
}

func (p *Progress) Open() (err error) {
	err = p.openDB()
	return
}

func (p *Progress) Close() {
	p.closeDB()
}

// Offsets gets all the offets for the user.
func (p *Progress) Offsets(user *auth.User) []Offset {
	offsets := p.userOffsets(user.Name)
	return offsets
}

// Offset gets the user offset based on the internal id.
func (p *Progress) Offset(user *auth.User, id int) *Offset {
	return p.lookupUserOffset(user.Name, id)
}

// Update will create or update an offset for the provided user using the etag
// as the primary key.
func (p *Progress) Update(user *auth.User, newOffset Offset) error {
	offset := p.lookupUserOffsetEtag(user.Name, newOffset.ETag)
	if offset != nil {
		if newOffset.Date.Before(offset.Date) {
			return ErrOffsetTooOld
		}
		offset.Offset = newOffset.Offset
		offset.Date = newOffset.Date
		if newOffset.Duration > 0 {
			offset.Duration = newOffset.Duration
		}
		err := p.updateOffset(offset)
		if err != nil {
			return err
		}
	} else {
		err := p.createOffset(&newOffset)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete will delete the provided user & offset, ensuring the offset belongs
// to the user.
func (p *Progress) Delete(user *auth.User, offset Offset) error {
	if user.Name != offset.User {
		return ErrAccessDenied
	}
	return p.deleteOffset(&offset)
}
