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

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (p *Progress) openDB() (err error) {
	var glog logger.Interface
	if p.config.Progress.DB.LogMode == false {
		glog = logger.Discard
	} else {
		glog = logger.Default
	}
	cfg := &gorm.Config{
		Logger: glog,
	}

	if p.config.Progress.DB.Driver == "sqlite3" {
		p.db, err = gorm.Open(sqlite.Open(p.config.Progress.DB.Source), cfg)
	} else {
		err = errors.New("driver not supported")
	}

	if err != nil {
		return
	}

	p.db.AutoMigrate(&Offset{})
	return
}

func (p *Progress) closeDB() {
	conn, err := p.db.DB()
	if err != nil {
		return
	}
	conn.Close()
}

func (p *Progress) userOffsets(user string) []Offset {
	var offsets []Offset
	p.db.Where("user = ?", user).
		Order("date desc").Find(&offsets)
	return offsets
}

func (p *Progress) lookupUserOffset(user string, id int) *Offset {
	offset, err := p.lookupOffset(id)
	if err != nil {
		return nil
	}
	if offset.User != user {
		return nil
	}
	return offset
}

func (p *Progress) lookupUserOffsetEtag(user, etag string) *Offset {
	var list []Offset
	p.db.Where("user = ? and e_tag = ?", user, etag).
		Order("date desc").Find(&list)
	if len(list) > 0 {
		return &list[0]
	}
	return nil
}

func (p *Progress) lookupOffset(id int) (*Offset, error) {
	var offset Offset
	err := p.db.First(&offset, id).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("offset not found")
	}
	return &offset, err
}

func (p *Progress) createOffset(o *Offset) error {
	return p.db.Create(o).Error
}

func (p *Progress) updateOffset(o *Offset) error {
	return p.db.Save(o).Error
}

func (p *Progress) deleteOffset(o *Offset) error {
	return p.db.Unscoped().Delete(o).Error
}

// func (p *Progress) deleteByETag(etag string) {
// 	var list []Offset
// 	p.db.Where("e_tag = ?", etag).Find(&list)
// 	for _, o := range list {
// 		p.db.Unscoped().Delete(o)
// 	}
// }
