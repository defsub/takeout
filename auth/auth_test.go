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
	"fmt"
	"github.com/defsub/takeout/config"
	"net/http"
	"testing"
	"time"
	"math"
)

func TestAddUser(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	a := NewAuth(config)
	err = a.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	defer a.Close()
	a.AddUser("defsub@defsub.com", "testpass")
}

func TestLogin(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	a := NewAuth(config)
	err = a.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	defer a.Close()

	cookie, err := a.Login("defsub@defsub.com", "testpass")
	if err != nil {
		t.Errorf("login should have worked: %s\n", err)
	}
	if len(cookie.Value) == 0 {
		t.Errorf("no cookie")
	}
	if cookie.Name != CookieName {
		t.Error("bad cookie name")
	}

	cookie, err = a.Login("defsub@defsub.com", "badpass")
	if err == nil {
		t.Errorf("should be incorrect password")
	}

	cookie, err = a.Login("bad@user.com", "testpass")
	if err == nil {
		t.Errorf("should be incorrect user")
	}
}

func TestCookie(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	a := NewAuth(config)
	err = a.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	defer a.Close()

	bad := http.Cookie{Name: CookieName, Value: "foo"}
	if a.Valid(bad) == true {
		t.Errorf("cookie should not exist")
	}

	cookie, err := a.Login("defsub@defsub.com", "testpass")
	if a.Valid(cookie) == false {
		t.Errorf("cookie should be good")
	}
}

func TestLogout(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	a := NewAuth(config)
	err = a.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	defer a.Close()

	cookie, err := a.Login("defsub@defsub.com", "testpass")
	if a.Valid(cookie) == false {
		t.Errorf("cookie should be good")
	}

	a.Logout(cookie)

	if a.Valid(cookie) == true {
		t.Errorf("cookie should fail")
	}
}

func TestMaxAge(t *testing.T) {
	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}
	a := NewAuth(config)
	err = a.Open()
	if err != nil {
		t.Errorf("Open %s\n", err)
	}
	defer a.Close()

	cookie, err := a.Login("defsub@defsub.com", "testpass")
	if a.Valid(cookie) == false {
		t.Errorf("cookie should be good")
	}

	cookie.MaxAge = 0
	now := time.Now()
	err = a.Refresh(&cookie)
	if err != nil {
		t.Errorf("refresh failed")
	}
	if cookie.MaxAge == 0 {
		t.Errorf("no age change")
	}
	d, _ := time.ParseDuration(fmt.Sprintf("%ds", cookie.MaxAge))
	age1 := now.Add(d)
	age2 := now.Add(config.Auth.MaxAge)
	delta := int(math.Abs(age1.Sub(age2).Seconds()))
	if delta > 1 {
		t.Errorf("delta is %d\n", delta)
	}
}
