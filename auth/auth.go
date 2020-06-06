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

package auth

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/defsub/takeout"
	"github.com/defsub/takeout/config"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/crypto/scrypt"
	"net/http"
	"time"
)

const (
	CookieName = takeout.AppName
)

type User struct {
	gorm.Model
	Name string `gorm:"unique_index:idx_user_name"`
	Key  []byte
	Salt []byte
}

type Session struct {
	gorm.Model
	User    string
	Cookie  string
	Expires time.Time
}

type Auth struct {
	config *config.Config
	db     *gorm.DB
}

func NewAuth(config *config.Config) *Auth {
	return &Auth{config: config}
}

func (a *Auth) Open() (err error) {
	a.db, err = gorm.Open(a.config.Auth.DB.Driver, a.config.Auth.DB.Source)
	if err != nil {
		return
	}
	a.db.LogMode(a.config.Auth.DB.LogMode)
	a.db.AutoMigrate(&Session{}, &User{})
	return
}

func (a *Auth) Close() {
	a.db.Close()
}

func (a *Auth) AddUser(email, pass string) error {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		fmt.Printf("rand: %s\n", err)
		return err
	}

	key, err := a.key(pass, salt)
	if err != nil {
		fmt.Printf("key: %s\n", err)
		return err
	}

	u := User{Name: email, Key: key, Salt: salt}

	return a.createUser(&u)
}

func (a *Auth) Login(email, pass string) (http.Cookie, error) {
	u := &User{Name: email}
	if a.db.Find(u, u).RecordNotFound() {
		return http.Cookie{}, errors.New("user not found")
	}

	key, err := a.key(pass, u.Salt)
	if err != nil {
		return http.Cookie{}, err
	}

	if !bytes.Equal(u.Key, key) {
		return http.Cookie{}, errors.New("key mismatch")
	}

	session := a.session(u)
	err = a.createSession(session)
	if err != nil {
		return http.Cookie{}, err
	}

	return session.NewCookie(), nil
}

func (session *Session) NewCookie() http.Cookie {
	return http.Cookie{
		Name: CookieName,
		Value: session.Cookie,
		Secure: true,
		Expires: session.Expires}
}

func (a *Auth) Logout(cookie http.Cookie) error {
	session := &Session{Cookie: cookie.Value}
	if a.db.Find(session, session).RecordNotFound() {
		return nil
	}
	return a.db.Delete(&session).Error
}

func (a *Auth) Valid(cookie http.Cookie) bool {
	if cookie.Name != CookieName {
		return false
	}

	session := &Session{Cookie: cookie.Value}
	if a.db.Find(session, session).RecordNotFound() {
		return false
	}

	now := time.Now()
	if now.After(session.Expires) {
		return false
	}

	return true
}

func (a *Auth) key(pass string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(pass), salt, 32768, 8, 1, 32)
}

func (a *Auth) session(u *User) *Session {
	cookie := uuid.New().String()
	expires := time.Now().Add(time.Hour * 24)
	session := &Session{User: u.Name, Cookie: cookie, Expires: expires}
	return session
}

func (a *Auth) createUser(u *User) (err error) {
	err = a.db.Create(u).Error
	return
}

func (a *Auth) createSession(s *Session) (err error) {
	err = a.db.Create(s).Error
	return
}
