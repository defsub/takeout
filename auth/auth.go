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
	"bytes"
	"crypto/rand"
	"errors"
	"net/http"
	"time"

	"github.com/defsub/takeout"
	"github.com/defsub/takeout/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/scrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	CookieName = takeout.AppName
)

var (
	ErrBadDriver       = errors.New("driver not supported")
	ErrUserNotFound    = errors.New("user not found")
	ErrKeyMismatch     = errors.New("key mismatch")
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrCodeNotFound    = errors.New("code not found")
	ErrCodeExpired     = errors.New("code has expired")
	ErrCodeAlreadyUsed = errors.New("code already authorized")
)

type User struct {
	gorm.Model
	Name  string `gorm:"unique_index:idx_user_name"`
	Key   []byte
	Salt  []byte
	Media string
}

type Session struct {
	gorm.Model
	User    string `gorm:"unique_index:idx_session_user"`
	Cookie  string `gorm:"unique_index:idx_session_cookie"`
	Expires time.Time
}

func (s *Session) Expired() bool {
	now := time.Now()
	return now.After(s.Expires)
}

func (s *Session) Valid() bool {
	return !s.Expired()
}

type Auth struct {
	config *config.Config
	db     *gorm.DB
}

func NewAuth(config *config.Config) *Auth {
	return &Auth{config: config}
}

func (a *Auth) Open() (err error) {
	cfg := a.config.Music.DB.GormConfig()

	if a.config.Auth.DB.Driver == "sqlite3" {
		a.db, err = gorm.Open(sqlite.Open(a.config.Auth.DB.Source), cfg)
	} else {
		err = ErrBadDriver
	}

	if err != nil {
		return
	}

	err = a.db.AutoMigrate(&Code{}, &Session{}, &User{})
	return
}

func (a *Auth) Close() {
	conn, err := a.db.DB()
	if err != nil {
		return
	}
	conn.Close()
}

func (a *Auth) AddUser(email, pass string) error {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}

	key, err := a.key(pass, salt)
	if err != nil {
		return err
	}

	u := User{Name: email, Key: key, Salt: salt}

	return a.createUser(&u)
}

func (a *Auth) Login(email, pass string) (http.Cookie, error) {
	var u User
	err := a.db.Where("name = ?", email).First(&u).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return http.Cookie{}, ErrUserNotFound
	}

	key, err := a.key(pass, u.Salt)
	if err != nil {
		return http.Cookie{}, err
	}

	if !bytes.Equal(u.Key, key) {
		return http.Cookie{}, ErrKeyMismatch
	}

	session := a.session(&u)
	err = a.createSession(session)
	if err != nil {
		return http.Cookie{}, err
	}

	return a.newCookie(session), nil
}

func (a *Auth) ChangePass(email, newpass string) error {
	var u User
	err := a.db.Where("name = ?", email).First(&u).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrUserNotFound
	}

	salt := make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return err
	}

	key, err := a.key(newpass, salt)
	if err != nil {
		return err
	}

	u.Salt = salt
	u.Key = key

	return a.db.Model(u).Update("salt", u.Salt).Update("key", u.Key).Error
}

func (a *Auth) Expire(cookie *http.Cookie) {
	cookie.MaxAge = 0
	cookie.Expires = time.Now().Add(-24 * time.Hour)
}

func (a *Auth) newCookie(session *Session) http.Cookie {
	return http.Cookie{
		Name:     CookieName,
		Value:    session.Cookie,
		MaxAge:   session.maxAge(),
		Path:     "/",
		Secure:   a.config.Auth.SecureCookies,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true}
}

func (a *Auth) AuthenticateCookie(cookie *http.Cookie) *Session {
	if cookie == nil || cookie.Name != CookieName {
		return nil
	}
	return a.findCookieSession(cookie)
}

func (a *Auth) AuthenticateToken(token string) *Session {
	return a.findSession(token)
}

func (a *Auth) CheckToken(token string) bool {
	session := a.AuthenticateToken(token)
	return session != nil && session.Valid()
}

func (a *Auth) SessionUser(session *Session) (*User, error) {
	var u User
	err := a.db.Where("name = ?", session.User).First(&u).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &u, nil
}

func (a *Auth) RefreshCookie(session *Session, cookie *http.Cookie) error {
	err := a.Refresh(session)
	if err != nil {
		return err
	}
	cookie.MaxAge = session.maxAge()
	return nil
}

func (a *Auth) Refresh(session *Session) error {
	if session == nil {
		return ErrSessionNotFound
	}
	a.touch(session)
	return nil
}

func (a *Auth) Logout(session *Session) {
	if session != nil {
		a.db.Delete(session)
	}
}

func (a *Auth) key(pass string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(pass), salt, 32768, 8, 1, 32)
}

func (a *Auth) findCookieSession(cookie *http.Cookie) *Session {
	return a.findSession(cookie.Value)
}

func (a *Auth) findSession(token string) *Session {
	var session Session
	err := a.db.Where("cookie = ?", token).First(&session).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return &session
}

func (a *Auth) session(u *User) *Session {
	cookie := uuid.New().String()
	expires := time.Now().Add(a.maxAge())
	session := &Session{User: u.Name, Cookie: cookie, Expires: expires}
	return session
}

func (a *Auth) touch(s *Session) error {
	expires := time.Now().Add(a.maxAge())
	return a.db.Model(s).Update("expires", expires).Error
}

func (a *Auth) createUser(u *User) (err error) {
	err = a.db.Create(u).Error
	return
}

func (a *Auth) updateUser(u *User) (err error) {
	err = a.db.Save(u).Error
	return
}

func (a *Auth) createSession(s *Session) (err error) {
	err = a.db.Create(s).Error
	return
}

func (a *Auth) maxAge() time.Duration {
	return a.config.Auth.MaxAge
}

func (s *Session) maxAge() int {
	return int(s.Expires.Sub(time.Now()).Seconds())
}
