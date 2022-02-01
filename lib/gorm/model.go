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

package gorm

import (
	g "gorm.io/gorm"
	"time"
)

// gorm model to exclude fields during json seralization
//
// note that gorm uses reflection so fields can be added or removed as needed.
type Model struct {
	ID        uint        `gorm:"primarykey"`
	CreatedAt time.Time   `json:"-"`
	UpdatedAt time.Time   `json:"-"`
	DeletedAt g.DeletedAt `gorm:"index" json:"-"`
}
