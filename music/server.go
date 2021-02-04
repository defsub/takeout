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
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/encoding/xspf"
	"github.com/defsub/takeout/log"
)

type MusicHandler struct {
	user        *auth.User
	config      *config.Config
	musicConfig *config.Config
}

func (handler *MusicHandler) NewMusic(w http.ResponseWriter, r *http.Request) *Music {
	music := NewMusic(handler.musicConfig)
	if music.Open() != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return nil
	}
	return music
}

func (handler *MusicHandler) NewAuth() *auth.Auth {
	a := auth.NewAuth(handler.config)
	err := a.Open()
	log.CheckError(err)
	if err != nil {
		return nil
	}
	return a
}

func (handler *MusicHandler) doit(w http.ResponseWriter, r *http.Request) {
	music := handler.NewMusic(w, r)
	if music == nil {
		return
	}
	defer music.Close()

	var tracks []Track
	if v := r.URL.Query().Get("q"); v != "" {
		tracks = music.Search(strings.TrimSpace(v))
	}

	if len(tracks) > 0 {
		handler.doSpiff(music, "Takeout", tracks, w, r)
	} else {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}
}

func (handler *MusicHandler) doSpiff(music *Music, title string, tracks []Track,
	w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", xspf.XMLContentType)

	encoder := xspf.NewXMLEncoder(w)
	encoder.Header(title)
	for _, t := range tracks {
		t.Location = []string{music.TrackURL(&t).String()}
		encoder.Encode(t)
	}
	encoder.Footer()
}

func (handler *MusicHandler) doTracks(w http.ResponseWriter, r *http.Request) {
	handler.doit(w, r)
}

func parseTemplates(templ *template.Template, dir string) *template.Template {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = templ.ParseFiles(path)
			if err != nil {
				return err
			}
		}
		return err
	})

	log.CheckError(err)

	return templ
}

func (handler *MusicHandler) render(music *Music, temp string, view interface{},
	w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"link": func(o interface{}) string {
			var link string
			switch o.(type) {
			case Release:
				link = fmt.Sprintf("/v?release=%d", o.(Release).ID)
			case Artist:
				link = fmt.Sprintf("/v?artist=%d", o.(Artist).ID)
			case Track:
				t := o.(Track)
				link = trackLocation(t)
			}
			return link
		},
		"popular": func(o interface{}) string {
			var link string
			switch o.(type) {
			case Artist:
				link = fmt.Sprintf("/v?popular=%d", o.(Artist).ID)
			}
			return link
		},
		"singles": func(o interface{}) string {
			var link string
			switch o.(type) {
			case Artist:
				link = fmt.Sprintf("/v?singles=%d", o.(Artist).ID)
			}
			return link
		},
		"ref": func(o interface{}, args ...string) string {
			var ref string
			switch o.(type) {
			case Release:
				ref = fmt.Sprintf("/music/releases/%d/tracks", o.(Release).ID)
			case Artist:
				ref = fmt.Sprintf("/music/artists/%d/%s", o.(Artist).ID, args[0])
			case Track:
				ref = fmt.Sprintf("/music/tracks/%d", o.(Track).ID)
			case string:
				ref = fmt.Sprintf("/music/search?q=%s", url.QueryEscape(o.(string)))
			case Station:
				ref = fmt.Sprintf("/music/radio/%d", o.(Station).ID)
			}
			return ref
		},
		"home": func() string {
			return "/v?music=1"
		},
		"coverSmall": func(o interface{}) string {
			switch o.(type) {
			case Release:
				return music.cover(o.(Release), "250")
			case Track:
				return music.trackCover(o.(Track), "250")
			}
			return ""
		},
		"coverLarge": func(o interface{}) string {
			switch o.(type) {
			case Release:
				return music.cover(o.(Release), "500")
			case Track:
				return music.trackCover(o.(Track), "500")
			}
			return ""
		},
		"coverExtraLarge": func(o interface{}) string {
			switch o.(type) {
			case Release:
				return music.cover(o.(Release), "1200")
			case Track:
				return music.trackCover(o.(Track), "1200")
			}
			return ""
		},
		"letter": func(a Artist) string {
			return a.SortName[0:1]
		},
	}

	var templates = parseTemplates(template.New("").Funcs(funcMap),
		fmt.Sprintf("%s/template", handler.config.Server.WebDir))

	err := templates.ExecuteTemplate(w, temp, view)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *MusicHandler) doLogin(user, pass string) (http.Cookie, error) {
	a := handler.NewAuth()
	if a == nil {
		return http.Cookie{}, errors.New("noauth")
	}
	defer a.Close()
	return a.Login(user, pass)
}

func (handler *MusicHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		user := r.Form.Get("user")
		pass := r.Form.Get("pass")
		cookie, err := handler.doLogin(user, pass)
		if err == nil {
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}

	http.Error(w, "bummer", http.StatusUnauthorized)
}

func (handler *MusicHandler) authorized(w http.ResponseWriter, r *http.Request) bool {
	a := handler.NewAuth()
	if a == nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return false
	}
	defer a.Close()

	cookie, err := r.Cookie(auth.CookieName)
	if err != nil {
		if cookie != nil {
			a.Expire(cookie)
			http.SetCookie(w, cookie)
		}
		http.Redirect(w, r, "/static/login.html", http.StatusTemporaryRedirect)
		return false
	}

	valid := a.Valid(*cookie)
	if !valid {
		a.Logout(*cookie)
		a.Expire(cookie)
		http.SetCookie(w, cookie)
		http.Error(w, "bummer", http.StatusUnauthorized)
		return false
	}

	handler.user, err = a.User(*cookie)
	if err != nil {
		a.Logout(*cookie)
		http.Error(w, "bummer", http.StatusUnauthorized)
		a.Expire(cookie)
		http.SetCookie(w, cookie)
		return false
	}

	a.Refresh(cookie)
	http.SetCookie(w, cookie)

	bucketName := handler.user.Bucket()
	if bucketName == "" {
		http.Error(w, "bummer", http.StatusServiceUnavailable)
		return false
	}
	path := fmt.Sprintf("%s/%s", handler.config.DataDir, bucketName)

	handler.musicConfig, err = config.LoadConfig(path)
	if err != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return false
	}
	handler.musicConfig.Server.URL = handler.config.Server.URL // TODO FIXME

	return true
}

func (handler *MusicHandler) viewHandler(w http.ResponseWriter, r *http.Request) {
	if !handler.authorized(w, r) {
		return
	}

	music := handler.NewMusic(w, r)
	if music == nil {
		return
	}
	defer music.Close()

	var view interface{}
	var temp string

	if v := r.URL.Query().Get("release"); v != "" {
		// /v?release={release-id}
		id, _ := strconv.Atoi(v)
		release, _ := music.lookupRelease(id)
		view = music.ReleaseView(release)
		temp = "release.html"
	} else if v := r.URL.Query().Get("artist"); v != "" {
		// /v?artist={artist-id}
		id, _ := strconv.Atoi(v)
		artist, _ := music.lookupArtist(id)
		view = music.ArtistView(artist)
		temp = "artist.html"
	} else if v := r.URL.Query().Get("artists"); v != "" {
		// /v?artists=x
		view = music.ArtistsView()
		temp = "artists.html"
	} else if v := r.URL.Query().Get("popular"); v != "" {
		// /v?popular={artist-id}
		id, _ := strconv.Atoi(v)
		artist, _ := music.lookupArtist(id)
		view = music.PopularView(artist)
		temp = "popular.html"
	} else if v := r.URL.Query().Get("singles"); v != "" {
		// /v?singles={artist-id}
		id, _ := strconv.Atoi(v)
		artist, _ := music.lookupArtist(id)
		view = music.SinglesView(artist)
		temp = "singles.html"
	} else if v := r.URL.Query().Get("music"); v != "" {
		// /v?music=x
		view = music.HomeView()
		temp = "music.html"
	} else if v := r.URL.Query().Get("q"); v != "" {
		// /v?q={pattern}
		view = music.SearchView(strings.TrimSpace(v))
		temp = "search.html"
	} else if v := r.URL.Query().Get("radio"); v != "" {
		// /v?radio=x
		view = music.RadioView(handler.user)
		temp = "radio.html"
	} else {
		view = time.Now().Unix()
		temp = "index.html"
	}

	handler.render(music, temp, view, w, r)
}

func Serve(config *config.Config) {
	handler := &MusicHandler{config: config}
	fs := http.FileServer(http.Dir(fmt.Sprintf("%s/static", config.Server.WebDir)))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/tracks", handler.doTracks)
	http.HandleFunc("/", handler.viewHandler)
	http.HandleFunc("/v", handler.viewHandler)
	http.HandleFunc("/login", handler.loginHandler)
	http.HandleFunc("/api/", handler.apiHandler)
	log.Printf("listening on %s\n", config.Server.Listen)
	http.ListenAndServe(config.Server.Listen, nil)
}
