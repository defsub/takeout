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

package auth

import (
	"errors"
	"gorm.io/gorm"
	"strings"
)

func mediaList(media string) []string {
	list := strings.Split(media, ",")
	for i := range list {
		list[i] = strings.Trim(list[i], " ")
	}
	return list
}

func (u *User) MediaList() []string {
	if len(u.Media) == 0 {
		return make([]string, 0)
	}
	return mediaList(u.Media)
}

func (u *User) FirstMedia() string {
	list := u.MediaList()
	return list[0]
}

func (a *Auth) AssignMedia(email, media string) error {
	var u User
	err := a.db.Where("name = ?", email).First(&u).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}
	u.Media = media
	return a.db.Model(u).Update("media", u.Media).Error
}

func (a *Auth) AssignedMedia() []string {
	var list []string
	rows, err := a.db.Table("users").
		Select("distinct(media)").Rows()
	if err != nil {
		return list
	}
	uniqueMedia := make(map[string]bool)
	for rows.Next() {
		var v string
		rows.Scan(&v)
		for _, media := range mediaList(v) {
			uniqueMedia[media] = true
		}
	}
	rows.Close()
	for k := range uniqueMedia {
		list = append(list, k)
	}
	return list
}
