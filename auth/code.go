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
	rando "math/rand"
	"time"

	"gorm.io/gorm"
)

const (
	CodeChars = "123456789ABCDEFGHILKMNPQRSTUVWXYZ"
	CodeSize  = 6
)

type Code struct {
	gorm.Model
	Value   string `gorm:"unique_index:idx_code_value"`
	Expires time.Time
	Cookie  string
}

func randomCode() string {
	var code string
	rando.Seed(time.Now().UnixNano())
	for i := 0; i < CodeSize; i++ {
		n := rando.Intn(len(CodeChars))
		code += string(CodeChars[n])
	}
	return code
}

func (a *Auth) createCode(c *Code) (err error) {
	err = a.db.Create(c).Error
	return
}

func (a *Auth) findCode(value string) *Code {
	var code Code
	err := a.db.Where("value = ?", value).First(&code).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return &code
}

func (a *Auth) GenerateCode() *Code {
	value := randomCode()
	expires := time.Now().Add(a.codeAge())
	c := &Code{Value: value, Expires: expires}
	a.db.Create(c)
	return c
}

func (c *Code) expired() bool {
	now := time.Now()
	return now.After(c.Expires)
}

func (a *Auth) LinkedCode(value string) *Code {
	code := a.findCode(value)
	if code == nil || code.Cookie == "" || code.expired() {
		return nil
	}
	return code
}

// This assumes cookie is valid
func (a *Auth) AuthorizeCode(value, cookie string) error {
	code := a.findCode(value)
	if code == nil {
		return ErrCodeNotFound
	}
	if code.expired() {
		return ErrCodeExpired
	}
	if code.Cookie != "" {
		return ErrCodeAlreadyUsed
	}
	return a.db.Model(code).Update("cookie", cookie).Error
}

func (a *Auth) codeAge() time.Duration {
	return a.config.Auth.CodeAge
}
