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

package music

import (
	"encoding/json"
	"github.com/defsub/takeout/spiff"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type location struct {
	ID           uint
	Url          string
	Size         int64
	ETag         string
	LastModified time.Time
}

type login struct {
	User string
	Pass string
}

type status struct {
	Status  int
	Message string
	Cookie  string
}

type reference struct {
	Ref  string
	Name string
}

// POST /api/login < login{} > status{}
// 200: success + cookie
// 401: fail
// 500: error
func (handler *MusicHandler) apiLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	var l login
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &l)
	if err != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	var result status
	cookie, err := handler.doLogin(l.User, l.Pass)
	if err == nil {
		http.SetCookie(w, &cookie)
		result = status{
			Status:  http.StatusOK,
			Message: "ok",
			Cookie:  cookie.Value,
		}
	} else {
		http.Error(w, "bummer", http.StatusUnauthorized)
		result = status{
			Status:  http.StatusUnauthorized,
			Message: "error",
		}
	}

	enc := json.NewEncoder(w)
	enc.Encode(result)
}

// GET /api/music/lists/ > []{ Ref, ID, Artist, Album, Tracks, Size (MB), Minutes }
// GET /api/music/lists/1 > spiff.Playlist{}
// PATCH /api/music/lists/1 < json+path > 204
//
// POST /api/playlists < reference{} > {}
// 204: success, no content
// 500: error
func (handler *MusicHandler) apiPlaylists(w http.ResponseWriter, r *http.Request, music *Music) {
	up := music.lookupPlaylist(handler.user)
	if up == nil {
		data, _ := spiff.NewPlaylist().Marshal()
		up = &UserPlaylist{User: handler.user.Name, Playlist: data}
		music.createPlaylist(up)
	}
}

// GET /api/playlist > spiff.Playlist{}
// 200: success
// 500: error
//
// PATCH /api/playlist < json+patch > spiff.Playlist{}
// 200: success
// 204: no change
// 500: error
func (handler *MusicHandler) apiPlaylist(w http.ResponseWriter, r *http.Request, music *Music) {
	up := music.currentPlaylist(handler.user)
	if up == nil {
		data, _ := spiff.NewPlaylist().Marshal()
		up = &UserPlaylist{User: handler.user.Name, Playlist: data}
		music.createPlaylist(up)
	}

	var plist *spiff.Playlist
	var err error
	dirty := false
	before := up.Playlist

	if r.Method == "PATCH" {
		patch, _ := ioutil.ReadAll(r.Body)
		up.Playlist, err = spiff.Patch(up.Playlist, patch)
		if err != nil {
			http.Error(w, "bummer", http.StatusInternalServerError)
			return
		}
		plist, _ = spiff.Unmarshal(up.Playlist)
		music.Resolve(plist)
		dirty = true
	}

	if dirty {
		if plist.Entries == nil {
			plist.Entries = []spiff.Entry{}
		}
		up.Playlist, _ = plist.Marshal()
		music.updatePlaylist(up)

		v, _ := spiff.Compare(before, up.Playlist)
		if v {
			// entries didn't change, only metadata
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(up.Playlist)
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(up.Playlist)
	}
}

func (handler *MusicHandler) apiView(w http.ResponseWriter, r *http.Request, view interface{}) {
	enc := json.NewEncoder(w)
	enc.Encode(view)
}

// Download playlists, albums, artists
//
// POST /api/playlists < { ref: "/music/releases/1/tracks", name: "name" }
// POST /api/playlists < { ref: "/music/artists/1/[popular,singles,tracks]", name: "name" }
// POST /api/playlists < { ref: "/music/playlist", name: "name" }

// POST /api/login -> see apiLogin
// GET,PATCH /api/playlist -> see apiPlaylist
//
// GET /api/home > HomeView{}
// GET /api/artists > ArtistsView{}
// GET /api/artists/1 > ArtistView{}
// GET /api/releases/1 > ReleaseView{}
// GET /api/playlists
// GET /api/playlists/1 > xxx{}
// GET /api/tracks/1/location -> location{}
// 200: success
// 500: error

func (handler *MusicHandler) apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	/*
		GET /api/playlist
		PATCH /api/playlist

		POST /api/playlists { name: 'xyz', user: cookie } (create & update)
		DELETE /api/playlists/1 - delete
		GET /api/playlists - all playlists, select * from user_playlists where user='boyns'

	*/

	if r.URL.Path == "/api/login" {
		handler.apiLogin(w, r)
	} else {
		if !handler.authorized(w, r) {
			return
		}

		music := handler.NewMusic(w, r)
		if music == nil {
			return
		}
		defer music.Close()

		switch r.URL.Path {
		case "/api/playlist":
			handler.apiPlaylist(w, r, music)
		case "/api/playlists":
			handler.apiPlaylists(w, r, music)
		case "/api/home":
			handler.apiView(w, r, music.HomeView())
		case "/api/artists":
			handler.apiView(w, r, music.ArtistsView())
		default:
			// id sub-resources
			locationRegexp := regexp.MustCompile(`/api/tracks/([0-9]+)/location`)
			matches := locationRegexp.FindStringSubmatch(r.URL.Path)
			if matches != nil {
				v := matches[1]
				id, _ := strconv.Atoi(v)
				track, _ := music.lookupTrack(id)
				url := music.TrackURL(&track)
				handler.apiView(w, r, location{
					ID:           track.ID,
					Url:          url.String(),
					Size:         track.Size,
					ETag:         track.ETag,
					LastModified: track.LastModified,
				})
				return
			}

			// resources with id
			resourceRegexp := regexp.MustCompile(`/api/([a-z]+)/([0-9]+)`)
			matches = resourceRegexp.FindStringSubmatch(r.URL.Path)
			if matches != nil {
				v := matches[1]
				id, _ := strconv.Atoi(matches[2])
				switch v {
				case "artists":
					// /api/artists/1
					artist, _ := music.lookupArtist(id)
					handler.apiView(w, r, music.ArtistView(artist))
				case "releases":
					// /api/releases/1
					release, _ := music.lookupRelease(id)
					handler.apiView(w, r, music.ReleaseView(release))
				case "playlists":
					// /api/playlists/1
					release, _ := music.lookupRelease(id)
					handler.apiView(w, r, music.ReleaseView(release))
				default:
					http.Error(w, "bummer", http.StatusNotFound)
				}
				return
			}

			http.Error(w, "bummer", http.StatusNotFound)
		}
	}
}
