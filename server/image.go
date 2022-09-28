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

package server

import (
	"fmt"
	"net/http"
)

const (
	CoverArtArchivePrefix = "http://coverartarchive.org"
	TMDBPrefix            = "http://image.tmdb.org"
	FanArtPrefix          = "https://assets.fanart.tv/fanart"
)

func imgRelease(w http.ResponseWriter, r *http.Request) {
	reid := r.URL.Query().Get(":reid")
	side := r.URL.Query().Get(":side")
	url := fmt.Sprintf("%s/release/%s/%s-250", CoverArtArchivePrefix, reid, side)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func imgReleaseFront(w http.ResponseWriter, r *http.Request) {
	reid := r.URL.Query().Get(":reid")
	url := fmt.Sprintf("%s/release/%s/front-250", CoverArtArchivePrefix, reid)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func imgReleaseGroup(w http.ResponseWriter, r *http.Request) {
	rgid := r.URL.Query().Get(":rgid")
	side := r.URL.Query().Get(":side")
	url := fmt.Sprintf("%s/release-group/%s/%s-250", CoverArtArchivePrefix, rgid, side)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func imgReleaseGroupFront(w http.ResponseWriter, r *http.Request) {
	rgid := r.URL.Query().Get(":rgid")
	url := fmt.Sprintf("%s/release-group/%s/front-250", CoverArtArchivePrefix, rgid)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func imgVideo(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get(":size")
	path := r.URL.Query().Get(":path")
	url := fmt.Sprintf("%s/t/p/%s/%s", TMDBPrefix, size, path)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func imgArtistThumb(w http.ResponseWriter, r *http.Request) {
	arid := r.URL.Query().Get(":arid")
	path := r.URL.Query().Get(":path")
	url := fmt.Sprintf("%s/music/%s/artistthumb/%s", FanArtPrefix, arid, path)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func imgArtistBackground(w http.ResponseWriter, r *http.Request) {
	arid := r.URL.Query().Get(":arid")
	path := r.URL.Query().Get(":path")
	url := fmt.Sprintf("%s/music/%s/artistbackground/%s", FanArtPrefix, arid, path)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
