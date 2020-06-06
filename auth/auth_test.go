package auth

import (
	"github.com/defsub/takeout/config"
	"testing"
	"net/http"
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

	err = a.Logout(cookie)
	if err != nil {
		t.Errorf("Logout %s\n", err)
	}

	if a.Valid(cookie) == true {
		t.Errorf("cookie should fail")
	}
}
