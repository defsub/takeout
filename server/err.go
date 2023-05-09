// Copyright (C) 2021 The Takeout Authors.
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
	"errors"
	"net/http"

	"github.com/defsub/takeout/lib/log"
)

var (
	ErrNoMedia            = errors.New("media not available")
	ErrInvalidMethod      = errors.New("invalid request method")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCode        = errors.New("invalid code")
	ErrNotFound           = errors.New("not found")
	ErrInvalidOffset      = errors.New("invalid offset")
	ErrAccessDenied       = errors.New("access denied")
	ErrMissingToken       = errors.New("missing token")
	ErrMissingAccessToken = errors.New("missing access token")
	ErrMissingMediaToken  = errors.New("missing media token")
	ErrMissingCcookie     = errors.New("missing cookie")
	ErrInvalidSession     = errors.New("invalid session")
)

func serverErr(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("got err %s\n", err)
		handleErr(w, "bummer", http.StatusInternalServerError)
	}
}

func badRequest(w http.ResponseWriter, err error) {
	if err != nil {
		handleErr(w, err.Error(), http.StatusBadRequest)
	}
}

// client provided no credentials or invalid credentials.
func authErr(w http.ResponseWriter, err error) {
	if err != nil {
		handleErr(w, err.Error(), http.StatusUnauthorized)
	}
}

// client provided credentials but access is not allowed.
func accessDenied(w http.ResponseWriter) {
	handleErr(w, ErrAccessDenied.Error(), http.StatusForbidden)
}

func notFoundErr(w http.ResponseWriter) {
	handleErr(w, ErrNotFound.Error(), http.StatusNotFound)
}

func handleErr(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}
