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
	"github.com/defsub/takeout/log"
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

func (handler *MusicHandler) recvStation(w http.ResponseWriter, r *http.Request,
	s *Station, music *Music) error {
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, s)
	if err != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return err
	}
	if s.Name == "" || s.Ref == "" {
		http.Error(w, "bummer", http.StatusBadRequest)
		return err
	}
	s.User = handler.user.Name
	if s.Ref == "/api/playlist" {
		// copy playlist
		p := music.lookupPlaylist(handler.user)
		if p != nil {
			s.Playlist = p.Playlist
		}
	}
	return nil
}

// GET /api/radio > []Station
// 200: success
//
// POST /api/radio < Station{}
// 201: created
// 400: bad request
// 500: error
func (handler *MusicHandler) apiRadio(w http.ResponseWriter, r *http.Request, music *Music) {
	if r.Method == "GET" {
		view := music.RadioView(handler.user)
		enc := json.NewEncoder(w)
		enc.Encode(view)
	} else if r.Method == "POST" {
		var s Station
		err := handler.recvStation(w, r, &s, music)
		if err != nil {
			return
		}
		err = music.createStation(&s)
		if err != nil {
			log.Println(err)
			http.Error(w, "bummer2", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		enc.Encode(s)
	} else {
		http.Error(w, "bummer", http.StatusBadRequest)
	}
}

// GET /api/radio/1 > spiff.Playlist{}
// 200: success
// 404: not found
//
// PUT /api/radio/1 < Station{}
// 204: no content
// 404: not found
// 500: error
//
// PATCH /api/radio/1 < json+patch > 204
// 204: no content
// 404: not found
// 500: error
//
// DELETE /api/radio/1
// 204: success, no content
// 404: not found
// 500: error
func (handler *MusicHandler) apiStation(w http.ResponseWriter, r *http.Request, music *Music, id int) {
	s, err := music.lookupStation(id)
	if err != nil {
		http.Error(w, "bummer", http.StatusNotFound)
		return
	}
	if !s.visible(handler.user) {
		http.Error(w, "bummer", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		music.stationRefresh(&s, handler.user)
		w.WriteHeader(http.StatusOK)
		w.Write(s.Playlist)
	} else if r.Method == "PUT" {
		var up Station
		err := handler.recvStation(w, r, &up, music)
		if err != nil {
			return
		}
		s.Name = up.Name
		s.Ref = up.Ref
		s.Playlist = up.Playlist
		err = music.updateStation(&s)
		if err != nil {
			log.Println(err)
			http.Error(w, "bummer", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else if r.Method == "PATCH" {
		patch, _ := ioutil.ReadAll(r.Body)
		s.Playlist, err = spiff.Patch(s.Playlist, patch)
		if err != nil {
			http.Error(w, "bummer", http.StatusInternalServerError)
			return
		}
		// unmarshal & resovle
		plist, _ := spiff.Unmarshal(s.Playlist)
		music.Resolve(handler.user, plist)
		if plist.Entries == nil {
			plist.Entries = []spiff.Entry{}
		}
		// marshal & persist
		s.Playlist, _ = plist.Marshal()
		music.updateStation(&s)
		w.WriteHeader(http.StatusNoContent)
	} else if r.Method == "DELETE" {
		err = music.deleteStation(&s)
		if err != nil {
			http.Error(w, "bummer", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "bummer", http.StatusBadRequest)
	}
}

// GET /api/playlist > spiff.Playlist{}
// 200: success
// 500: error
//
// PATCH /api/playlist < json+patch > spiff.Playlist{}
// 200: success
// 204: no change to track entries
// 500: error
func (handler *MusicHandler) apiPlaylist(w http.ResponseWriter, r *http.Request, music *Music) {
	p := music.lookupPlaylist(handler.user)
	if p == nil {
		data, _ := spiff.NewPlaylist().Marshal()
		p = &Playlist{User: handler.user.Name, Playlist: data}
		err := music.createPlaylist(p)
		if err != nil {
			http.Error(w, "bummer", http.StatusInternalServerError)
			return
		}
	}

	var plist *spiff.Playlist
	var err error
	dirty := false
	before := p.Playlist

	if r.Method == "PATCH" {
		patch, _ := ioutil.ReadAll(r.Body)
		p.Playlist, err = spiff.Patch(p.Playlist, patch)
		if err != nil {
			http.Error(w, "bummer", http.StatusInternalServerError)
			return
		}
		plist, _ = spiff.Unmarshal(p.Playlist)
		music.Resolve(handler.user, plist)
		dirty = true
	}

	if dirty {
		if plist.Entries == nil {
			plist.Entries = []spiff.Entry{}
		}
		p.Playlist, _ = plist.Marshal()
		music.updatePlaylist(p)

		v, _ := spiff.Compare(before, p.Playlist)
		if v {
			// entries didn't change, only metadata
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(p.Playlist)
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(p.Playlist)
	}
}

func (handler *MusicHandler) apiView(w http.ResponseWriter, r *http.Request, view interface{}) {
	enc := json.NewEncoder(w)
	enc.Encode(view)
}

// POST /api/login -> see apiLogin
// GET,PATCH /api/playlist -> see apiPlaylist
//
// GET /api/home > HomeView{}
// GET /api/artists > ArtistsView{}
// GET /api/artists/1 > ArtistView{}
// GET /api/releases/1 > ReleaseView{}
// GET,POST /api/radio
// GET,PATH,DELETE /api/radio/1 >
// GET /api/tracks/1/location -> location{}
// 200: success
// 500: error
func (handler *MusicHandler) apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

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
		case "/api/radio":
			handler.apiRadio(w, r, music)
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
				case "radio":
					// /api/radio/1
					handler.apiStation(w, r, music, id)
				default:
					http.Error(w, "bummer", http.StatusNotFound)
				}
				return
			}

			http.Error(w, "bummer", http.StatusNotFound)
		}
	}
}
