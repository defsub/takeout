// Copyright (C) 2023 The Takeout Authors.
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

package server

import (
	"net/http"
	"strings"

	"github.com/defsub/takeout/auth"
)

type bits uint8

const (
	AllowCookie bits = 1 << iota
	AllowAccessToken
	AllowMediaToken

	AuthorizationHeader = "Authorization"
	BearerAuthorization = "Bearer"
)

// doCodeAuth creates a login session and binds to the provided code value.
func doCodeAuth(ctx Context, user, pass, value string) error {
	session, err := doLogin(ctx, user, pass)
	if err != nil {
		return err
	}
	err = ctx.Auth().AuthorizeCode(value, session.Token)
	if err != nil {
		return ErrInvalidCode
	}
	return nil
}

// getAuthToken returns the bearer token from the request, if any.
func getAuthToken(r *http.Request) string {
	value := r.Header.Get(AuthorizationHeader)
	if value == "" {
		return ""
	}
	result := strings.Split(value, " ")
	var token string
	switch len(result) {
	case 1:
		// Authorization: <token>
		token = result[0]
	case 2:
		// Authorization: Bearer <token>
		if strings.EqualFold(result[0], BearerAuthorization) {
			token = result[1]
		}
	}
	return token
}

// authorizeAccessToken validates the provided JWT access token for API access.
func authorizeAccessToken(ctx Context, w http.ResponseWriter, r *http.Request) (*auth.User, error) {
	token := getAuthToken(r)
	if token == "" {
		return nil, nil
	}
	// token should be a JWT
	user, err := ctx.Auth().CheckAccessTokenUser(token)
	if err != nil {
		authErr(w, err)
		return nil, err
	}
	return &user, nil
}

// authorizeMediaToken validates the provided JWT media token for API access.
func authorizeMediaToken(ctx Context, w http.ResponseWriter, r *http.Request) (*auth.User, error) {
	token := getAuthToken(r)
	if token == "" {
		return nil, nil
	}
	// token should be a JWT
	user, err := ctx.Auth().CheckMediaTokenUser(token)
	if err != nil {
		authErr(w, err)
		return nil, err
	}
	return &user, nil
}

// authorizeCodeToken validates the provided JWT code token for code auth access.
func authorizeCodeToken(ctx Context, w http.ResponseWriter, r *http.Request) error {
	token := getAuthToken(r)
	if token == "" {
		err := ErrMissingToken
		authErr(w, err)
		return err
	}
	// token should be a JWT with valid code in the subject
	err := ctx.Auth().CheckCodeToken(token)
	if err != nil {
		authErr(w, err)
		return err
	}

	return err
}

// authorizeCookie validates the provided cookie for API or web view access.
func authorizeCookie(ctx Context, w http.ResponseWriter, r *http.Request) (*auth.User, error) {
	a := ctx.Auth()
	cookie, err := r.Cookie(auth.CookieName)
	if err != nil {
		if err != http.ErrNoCookie {
			http.SetCookie(w, auth.ExpireCookie(cookie)) // what cookie is this?
		}
		http.Redirect(w, r, LoginRedirect, http.StatusTemporaryRedirect)
		return nil, err
	}

	session := a.CookieSession(cookie)
	if session == nil {
		http.SetCookie(w, auth.ExpireCookie(cookie))
		http.Redirect(w, r, LoginRedirect, http.StatusTemporaryRedirect)
		return nil, ErrAccessDenied
	} else if session.Expired() {
		err = ErrAccessDenied
		a.DeleteSession(*session)
		http.SetCookie(w, auth.ExpireCookie(cookie))
		http.Redirect(w, r, LoginRedirect, http.StatusTemporaryRedirect)
		return nil, ErrAccessDenied
	}

	user, err := a.SessionUser(session)
	if err != nil {
		// session with no user?
		a.DeleteSession(*session)
		http.SetCookie(w, auth.ExpireCookie(cookie))
		http.Redirect(w, r, LoginRedirect, http.StatusTemporaryRedirect)
		return nil, err
	}

	// send back an updated cookie
	auth.UpdateCookie(session, cookie)
	http.SetCookie(w, cookie)

	return user, nil
}

// authorizeRefreshToken validates the provided refresh token for API access.
func authorizeRefreshToken(ctx Context, w http.ResponseWriter, r *http.Request) *auth.Session {
	token := getAuthToken(r)
	if token == "" {
		authErr(w, ErrUnauthorized)
		return nil
	}
	// token should be a refresh token not JWT
	a := ctx.Auth()
	session := a.TokenSession(token)
	if session == nil {
		// no session for token
		authErr(w, ErrUnauthorized)
		return nil
	} else if session.Expired() {
		// session expired
		a.DeleteSession(*session)
		authErr(w, ErrUnauthorized)
		return nil
	} else if session.Duration() < ctx.Config().Auth.AccessToken.Age {
		// session will expire before token
		authErr(w, ErrUnauthorized)
		return nil
	}
	// session still valid
	return session
}

// authorizeRequest authorizes the request with one or more of the allowed
// authorization methods.
func authorizeRequest(ctx Context, w http.ResponseWriter, r *http.Request, auth bits) *auth.User {
	if auth&AllowAccessToken != 0 {
		user, err := authorizeAccessToken(ctx, w, r)
		if user != nil {
			return user
		}
		if err != nil {
			return nil
		}
	}

	if auth&AllowMediaToken != 0 {
		user, err := authorizeMediaToken(ctx, w, r)
		if user != nil {
			return user
		}
		if err != nil {
			return nil
		}
	}

	if auth&AllowCookie != 0 {
		user, err := authorizeCookie(ctx, w, r)
		if user != nil {
			return user
		}
		if err != nil {
			return nil
		}
	}

	return nil
}

// refreshTokenAuthHandler handles requests intended to refresh and access token.
func refreshTokenAuthHandler(ctx RequestContext, handler http.HandlerFunc) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		session := authorizeRefreshToken(ctx, w, r)
		if session != nil {
			ctx := sessionContext(ctx, session)
			handler.ServeHTTP(w, withContext(r, ctx))
		}
	}
	return http.HandlerFunc(fn)
}

// authHandler authorizes and handles all (except refresh) requests based on
// allowed auth methods.
func authHandler(ctx RequestContext, handler http.HandlerFunc, auth bits) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := authorizeRequest(ctx, w, r, auth)
		if user != nil {
			ctx, err := upgradeContext(ctx, user)
			if err != nil {
				serverErr(w, err)
				return
			}
			handler.ServeHTTP(w, withContext(r, ctx))
		}
	}
	return http.HandlerFunc(fn)
}

// mediaTokenAuthHandler handles media access requests using the media token (or cookie).
func mediaTokenAuthHandler(ctx RequestContext, handler http.HandlerFunc) http.Handler {
	return authHandler(ctx, handler, AllowMediaToken|AllowCookie)
}

// accessTokenAuthHandler handles non-media requests using the access token (or cookie).
func accessTokenAuthHandler(ctx RequestContext, handler http.HandlerFunc) http.Handler {
	return authHandler(ctx, handler, AllowAccessToken|AllowCookie)
}

func codeTokenAuthHandler(ctx RequestContext, handler http.HandlerFunc) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		err := authorizeCodeToken(ctx, w, r)
		if err == nil {
			handler.ServeHTTP(w, withContext(r, ctx))
		}
	}
	return http.HandlerFunc(fn)
}
