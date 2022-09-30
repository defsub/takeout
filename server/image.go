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

	"github.com/defsub/takeout/lib/log"
)

const (
	CoverArtArchivePrefix = "http://coverartarchive.org"
	TMDBPrefix            = "https://image.tmdb.org"
	FanArtPrefix          = "https://assets.fanart.tv/fanart"
)

func checkImageCache(w http.ResponseWriter, r *http.Request, url string) {
	ctx := contextValue(r)
	client := ctx.ImageClient()
	header, img, err := client.Get(url)
	if err == nil && len(img) > 0 {
		log.Printf("img using cached image %d for %s\n", len(img), url)
		for k, v := range header {
			switch k {
			case HeaderContentType, HeaderContentLength, HeaderETag,
				HeaderLastModified, HeaderCacheControl:
				w.Header().Set(k, v[0])
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write(img)
	} else {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func imgRelease(w http.ResponseWriter, r *http.Request) {
	reid := r.URL.Query().Get(":reid")
	side := r.URL.Query().Get(":side")
	url := fmt.Sprintf("%s/release/%s/%s-250", CoverArtArchivePrefix, reid, side)
	checkImageCache(w, r, url)
}

func imgReleaseFront(w http.ResponseWriter, r *http.Request) {
	reid := r.URL.Query().Get(":reid")
	url := fmt.Sprintf("%s/release/%s/front-250", CoverArtArchivePrefix, reid)
	checkImageCache(w, r, url)
}

func imgReleaseGroup(w http.ResponseWriter, r *http.Request) {
	rgid := r.URL.Query().Get(":rgid")
	side := r.URL.Query().Get(":side")
	url := fmt.Sprintf("%s/release-group/%s/%s-250", CoverArtArchivePrefix, rgid, side)
	checkImageCache(w, r, url)
}

func imgReleaseGroupFront(w http.ResponseWriter, r *http.Request) {
	rgid := r.URL.Query().Get(":rgid")
	url := fmt.Sprintf("%s/release-group/%s/front-250", CoverArtArchivePrefix, rgid)
	checkImageCache(w, r, url)
}

func imgVideo(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get(":size")
	path := r.URL.Query().Get(":path")
	url := fmt.Sprintf("%s/t/p/%s/%s", TMDBPrefix, size, path)
	checkImageCache(w, r, url)
}

func imgArtistThumb(w http.ResponseWriter, r *http.Request) {
	arid := r.URL.Query().Get(":arid")
	path := r.URL.Query().Get(":path")
	url := fmt.Sprintf("%s/music/%s/artistthumb/%s", FanArtPrefix, arid, path)
	checkImageCache(w, r, url)
}

func imgArtistBackground(w http.ResponseWriter, r *http.Request) {
	arid := r.URL.Query().Get(":arid")
	path := r.URL.Query().Get(":path")
	url := fmt.Sprintf("%s/music/%s/artistbackground/%s", FanArtPrefix, arid, path)
	checkImageCache(w, r, url)
}
