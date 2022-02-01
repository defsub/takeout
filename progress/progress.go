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

package progress

import (
	"errors"
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"gorm.io/gorm"
)

var (
	ErrOffsetTooOld = errors.New("offset is old")
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

func (p *Progress) LookupProgress(user *auth.User) []Offset {
	offsets := p.UserOffsets(user.Name)
	return offsets
}

func (p *Progress) Update(user *auth.User, newOffset Offset) error {
	offset := p.LookupUserOffset(user.Name, newOffset.ETag)
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
