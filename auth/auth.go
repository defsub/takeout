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
	rando "math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/defsub/takeout"
	"github.com/defsub/takeout/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/scrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	CookieName = takeout.AppName
)

type User struct {
	gorm.Model
	Name  string `gorm:"unique_index:idx_user_name"`
	Key   []byte
	Salt  []byte
	Media string
}

type Code struct {
	gorm.Model
	Value   string `gorm:"unique_index:idx_code_value"`
	Expires time.Time
	Cookie  string
}

func (u *User) MediaList() []string {
	if len(u.Media) == 0 {
		return make([]string, 0)
	}
	list := strings.Split(u.Media, ",")
	for i := range list {
		list[i] = strings.Trim(list[i], " ")
	}
	return list
}

func (u *User) FirstMedia() string {
	list := u.MediaList()
	return list[0]
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
	var glog logger.Interface
	if a.config.Auth.DB.LogMode == false {
		glog = logger.Discard
	} else {
		glog = logger.Default
	}
	cfg := &gorm.Config{
		Logger: glog,
	}

	if a.config.Auth.DB.Driver == "sqlite3" {
		a.db, err = gorm.Open(sqlite.Open(a.config.Auth.DB.Source), cfg)
	} else {
		err = errors.New("driver not supported")
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
		return http.Cookie{}, errors.New("user not found")
	}

	key, err := a.key(pass, u.Salt)
	if err != nil {
		return http.Cookie{}, err
	}

	if !bytes.Equal(u.Key, key) {
		return http.Cookie{}, errors.New("key mismatch")
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
		return errors.New("user not found")
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

func (a *Auth) AssignMedia(email, media string) error {
	var u User
	err := a.db.Where("name = ?", email).First(&u).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user not found")
	}
	u.Media = media
	return a.db.Model(u).Update("media", u.Media).Error
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
		HttpOnly: true}
}

func (a *Auth) Valid(cookie http.Cookie) bool {
	if cookie.Name != CookieName {
		return false
	}
	session := a.findCookieSession(cookie)
	if session == nil {
		return false
	}
	now := time.Now()
	if now.After(session.Expires) {
		return false
	}
	return true
}

func (a *Auth) Refresh(cookie *http.Cookie) error {
	session := a.findCookieSession(*cookie)
	if session == nil {
		return errors.New("session not found")
	}
	a.touch(session)
	cookie.MaxAge = session.maxAge()
	return nil
}

func (a *Auth) UserAuth(cookie http.Cookie) (*User, error) {
	return a.UserAuthValue(cookie.Value)
}

// TODO make private later
func (a *Auth) UserAuthValue(value string) (*User, error) {
	session := a.findSession(value)
	if session == nil {
		return nil, errors.New("session not found")
	}

	// TODO add cache
	var u User
	err := a.db.Where("name = ?", session.User).First(&u).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}

	return &u, nil
}

func (a *Auth) Logout(cookie http.Cookie) {
	session := a.findCookieSession(cookie)
	if session != nil {
		a.db.Delete(session)
	}
}

func (a *Auth) key(pass string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(pass), salt, 32768, 8, 1, 32)
}

func (a *Auth) findCookieSession(cookie http.Cookie) *Session {
	return a.findSession(cookie.Value)
}

func (a *Auth) findSession(value string) *Session {
	var session Session
	err := a.db.Where("cookie = ?", value).First(&session).Error
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

func (a *Auth) codeAge() time.Duration {
	return a.config.Auth.CodeAge
}

func (s *Session) maxAge() int {
	return int(s.Expires.Sub(time.Now()).Seconds())
}

const (
	CodeChars = "123456789ABCDEFGHILKMNPQRSTUVWXYZ"
	CodeSize  = 6
)

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
		return errors.New("code not found")
	}
	if code.expired() {
		return errors.New("code has expired")
	}
	if code.Cookie != "" {
		return errors.New("code already authorized")
	}
	return a.db.Model(code).Update("cookie", cookie).Error
}
