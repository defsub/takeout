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

package server

import (
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/date"
	"github.com/defsub/takeout/lib/encoding/xspf"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/video"
)

const (
	ApplicationJson = "application/json"
)

type UserHandler struct {
	config     *config.Config
	template   *template.Template
	user       *auth.User
	userConfig *config.Config
}

func (handler *UserHandler) NewMusic() (*music.Music, error) {
	m := music.NewMusic(handler.userConfig)
	err := m.Open()
	return m, err
}

func (handler *UserHandler) NewVideo() (*video.Video, error) {
	vid := video.NewVideo(handler.userConfig)
	err := vid.Open()
	return vid, err
}

func (handler *UserHandler) NewAuth() (*auth.Auth, error) {
	a := auth.NewAuth(handler.config)
	err := a.Open()
	return a, err
}

func (handler *UserHandler) tracksHandler(w http.ResponseWriter, r *http.Request, m *music.Music) {
	var tracks []music.Track
	if v := r.URL.Query().Get("q"); v != "" {
		tracks = m.Search(strings.TrimSpace(v))
	}

	if len(tracks) > 0 {
		handler.doSpiff(m, "Takeout", tracks, w, r)
	} else {
		notFoundErr(w)
		return
	}
}

func (handler *UserHandler) doSpiff(m *music.Music, title string, tracks []music.Track,
	w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", xspf.XMLContentType)

	encoder := xspf.NewXMLEncoder(w)
	encoder.Header(title)
	for _, t := range tracks {
		t.Location = []string{m.TrackURL(&t).String()}
		encoder.Encode(t)
	}
	encoder.Footer()
}

func doFuncMap() template.FuncMap {
	return template.FuncMap{
		"join": strings.Join,
		"ymd":  date.YMD,
		"link": func(o interface{}) string {
			var link string
			switch o.(type) {
			case music.Release:
				link = fmt.Sprintf("/v?release=%d", o.(music.Release).ID)
			case music.Artist:
				link = fmt.Sprintf("/v?artist=%d", o.(music.Artist).ID)
			case music.Track:
				link = locateTrack(o.(music.Track))
			case video.Movie:
				// m := o.(video.Movie)
				// link = handler.LocateMovie(m)
				link = fmt.Sprintf("/v?movie=%d", o.(video.Movie).ID)
			}
			return link
		},
		"url": func(o interface{}) string {
			var loc string
			switch o.(type) {
			case music.Track:
				loc = locateTrack(o.(music.Track))
			case video.Movie:
				loc = locateMovie(o.(video.Movie))
			}
			return loc
		},
		"popular": func(o interface{}) string {
			var link string
			switch o.(type) {
			case music.Artist:
				link = fmt.Sprintf("/v?popular=%d", o.(music.Artist).ID)
			}
			return link
		},
		"singles": func(o interface{}) string {
			var link string
			switch o.(type) {
			case music.Artist:
				link = fmt.Sprintf("/v?singles=%d", o.(music.Artist).ID)
			}
			return link
		},
		"ref": func(o interface{}, args ...string) string {
			var ref string
			switch o.(type) {
			case music.Release:
				ref = fmt.Sprintf("/music/releases/%d/tracks", o.(music.Release).ID)
			case music.Artist:
				ref = fmt.Sprintf("/music/artists/%d/%s", o.(music.Artist).ID, args[0])
			case music.Track:
				ref = fmt.Sprintf("/music/tracks/%d", o.(music.Track).ID)
			case string:
				ref = fmt.Sprintf("/music/search?q=%s", url.QueryEscape(o.(string)))
			case music.Station:
				ref = fmt.Sprintf("/music/radio/%d", o.(music.Station).ID)
			}
			return ref
		},
		"home": func() string {
			return "/v?home=1"
		},
		"runtime": func(m video.Movie) string {
			hours := m.Runtime / 60
			mins := m.Runtime % 60
			return fmt.Sprintf("%dh %dm", hours, mins)
		},
		"letter": func(a music.Artist) string {
			return a.SortName[0:1]
		},
	}
}

func (handler *UserHandler) render(m *music.Music, vid *video.Video, temp string, view interface{},
	w http.ResponseWriter, r *http.Request) {
	err := handler.template.ExecuteTemplate(w, temp, view)
	if err != nil {
		serverErr(w, err)
	}
}

func (handler *UserHandler) doLogin(user, pass string) (http.Cookie, error) {
	a, err := handler.NewAuth()
	if err != nil {
		return http.Cookie{}, err
	}
	defer a.Close()
	return a.Login(user, pass)
}

func (handler *UserHandler) doCodeAuth(user, pass, value string) error {
	a, err := handler.NewAuth()
	if err != nil {
		return err
	}
	defer a.Close()
	cookie, err := a.Login(user, pass)
	if err != nil {
		return err
	}
	err = a.AuthorizeCode(value, cookie.Value)
	if err != nil {
		return ErrInvalidCode
	}
	return nil
}

func (handler *UserHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		user := r.Form.Get("user")
		pass := r.Form.Get("pass")
		cookie, err := handler.doLogin(user, pass)
		if err == nil {
			// success
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}
	authErr(w, ErrUnauthorized)
}

func (handler *UserHandler) linkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		user := r.Form.Get("user")
		pass := r.Form.Get("pass")
		value := r.Form.Get("code")
		err := handler.doCodeAuth(user, pass, value)
		if err == nil {
			// success
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}
	http.Redirect(w, r, "/static/link.html", http.StatusTemporaryRedirect)
}

func (handler *UserHandler) authorized(w http.ResponseWriter, r *http.Request) bool {
	a, err := handler.NewAuth()
	if err != nil {
		serverErr(w, err)
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
		authErr(w, ErrUnauthorized)
		return false
	}

	handler.user, err = a.UserAuth(*cookie)
	if err != nil {
		a.Logout(*cookie)
		authErr(w, ErrUnauthorized)
		a.Expire(cookie)
		http.SetCookie(w, cookie)
		return false
	}

	a.Refresh(cookie)
	http.SetCookie(w, cookie)

	err = handler.configure(w)
	if err != nil {
		serverErr(w, err)
		return false
	}

	return true
}

// after user authentication, configure available media
func (handler *UserHandler) configure(w http.ResponseWriter) error {
	var err error
	// only supports one media collection right now
	media := handler.user.FirstMedia()
	if media == "" {
		return ErrNoMedia
	}
	path := fmt.Sprintf("%s/%s", handler.config.DataDir, media)
	// load relative media configuration
	handler.userConfig, err = config.LoadConfig(path)
	if err != nil {
		return err
	}
	handler.userConfig.Server.URL = handler.config.Server.URL // TODO FIXME
	return nil
}

func (handler *UserHandler) viewHandler(w http.ResponseWriter, r *http.Request, m *music.Music, vid *video.Video) {
	var view interface{}
	var temp string

	if v := r.URL.Query().Get("release"); v != "" {
		// /v?release={release-id}
		id, _ := strconv.Atoi(v)
		release, _ := m.LookupRelease(id)
		view = handler.releaseView(m, release)
		temp = "release.html"
	} else if v := r.URL.Query().Get("artist"); v != "" {
		// /v?artist={artist-id}
		id, _ := strconv.Atoi(v)
		artist, _ := m.LookupArtist(id)
		view = handler.artistView(m, artist)
		temp = "artist.html"
	} else if v := r.URL.Query().Get("artists"); v != "" {
		// /v?artists=x
		view = handler.artistsView(m)
		temp = "artists.html"
	} else if v := r.URL.Query().Get("popular"); v != "" {
		// /v?popular={artist-id}
		id, _ := strconv.Atoi(v)
		artist, _ := m.LookupArtist(id)
		view = handler.popularView(m, artist)
		temp = "popular.html"
	} else if v := r.URL.Query().Get("singles"); v != "" {
		// /v?singles={artist-id}
		id, _ := strconv.Atoi(v)
		artist, _ := m.LookupArtist(id)
		view = handler.singlesView(m, artist)
		temp = "singles.html"
	} else if v := r.URL.Query().Get("home"); v != "" {
		// /v?home=x
		view = handler.homeView(m, vid)
		temp = "home.html"
	} else if v := r.URL.Query().Get("q"); v != "" {
		// /v?q={pattern}
		view = handler.searchView(m, vid, strings.TrimSpace(v))
		temp = "search.html"
	} else if v := r.URL.Query().Get("radio"); v != "" {
		// /v?radio=x
		view = handler.radioView(m, handler.user)
		temp = "radio.html"
	} else if v := r.URL.Query().Get("movies"); v != "" {
		// /v?movies=x
		view = handler.moviesView(vid)
		temp = "movies.html"
	} else if v := r.URL.Query().Get("movie"); v != "" {
		// /v?movie={movie-id}
		id, _ := strconv.Atoi(v)
		movie, _ := vid.LookupMovie(id)
		view = handler.movieView(vid, movie)
		temp = "movie.html"
	} else if v := r.URL.Query().Get("profile"); v != "" {
		// /v?profile={person-id}
		id, _ := strconv.Atoi(v)
		person, _ := vid.LookupPerson(id)
		view = handler.profileView(vid, person)
		temp = "profile.html"
	} else if v := r.URL.Query().Get("genre"); v != "" {
		// /v?genre={genre-name}
		name := strings.TrimSpace(v)
		view = handler.genreView(vid, name)
		temp = "genre.html"
	} else if v := r.URL.Query().Get("keyword"); v != "" {
		// /v?keyword={keyword-name}
		name := strings.TrimSpace(v)
		view = handler.keywordView(vid, name)
		temp = "keyword.html"
	} else if v := r.URL.Query().Get("watch"); v != "" {
		// /v?watch={movie-id}
		id, _ := strconv.Atoi(v)
		movie, _ := vid.LookupMovie(id)
		view = handler.watchView(vid, movie)
		temp = "watch.html"
	} else {
		view = handler.indexView(m, vid)
		temp = "index.html"
	}

	handler.render(m, vid, temp, view, w, r)
}

//go:embed res/static
var resStatic embed.FS

func getStaticFS(config *config.Config) http.FileSystem {
	// dev := false
	// if dev {
	// 	return http.FS(os.DirFS(fmt.Sprintf("%s/static", config.Server.WebDir)))
	// }
	fsys, err := fs.Sub(resStatic, "res")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

//go:embed res/template
var resTemplates embed.FS

func getTemplateFS(config *config.Config) fs.FS {
	// dev := false
	// if dev {
	// 	return os.DirFS(fmt.Sprintf("%s/template", config.Server.WebDir))
	// }
	return resTemplates
}

func getTemplates(config *config.Config) *template.Template {
	return template.Must(template.New("").Funcs(doFuncMap()).ParseFS(getTemplateFS(config),
		"res/template/*.html",
		"res/template/music/*.html",
		"res/template/video/*.html"))
}

func Serve(config *config.Config) {
	template := getTemplates(config)

	makeHandler := func() *UserHandler {
		return &UserHandler{
			config:   config,
			template: template,
		}
	}

	loginHandler := func(w http.ResponseWriter, r *http.Request) {
		makeHandler().loginHandler(w, r)
	}

	linkHandler := func(w http.ResponseWriter, r *http.Request) {
		makeHandler().linkHandler(w, r)
	}

	tracksHandler := func(w http.ResponseWriter, r *http.Request) {
		// TODO keep this? auth?
		handler := makeHandler()
		m, err := handler.NewMusic()
		if err != nil {
			serverErr(w, err)
			return
		}
		defer m.Close()
		handler.tracksHandler(w, r, m)
	}

	viewHandler := func(w http.ResponseWriter, r *http.Request) {
		handler := makeHandler()
		if !handler.authorized(w, r) {
			return
		}
		m, err := handler.NewMusic()
		if err != nil {
			serverErr(w, err)
			return
		}
		defer m.Close()
		v, err := handler.NewVideo()
		if err != nil {
			serverErr(w, err)
			return
		}
		defer v.Close()
		handler.viewHandler(w, r, m, v)
	}

	apiLoginHandler := func(w http.ResponseWriter, r *http.Request) {
		makeHandler().apiLogin(w, r)
	}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		handler := makeHandler()
		if !handler.authorized(w, r) {
			return
		}
		m, err := handler.NewMusic()
		if err != nil {
			serverErr(w, err)
			return
		}
		defer m.Close()
		v, err := handler.NewVideo()
		if err != nil {
			serverErr(w, err)
			return
		}
		defer v.Close()
		handler.apiHandler(w, r, m, v)
	}

	hookHandler := func(w http.ResponseWriter, r *http.Request) {
		makeHandler().hookHandler(w, r)
	}

	http.Handle("/static/", http.FileServer(getStaticFS(config)))
	http.HandleFunc("/tracks", tracksHandler)
	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/v", viewHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/link", linkHandler)
	http.HandleFunc("/api/login", apiLoginHandler)
	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/hook/", hookHandler)
	log.Printf("listening on %s\n", config.Server.Listen)
	http.ListenAndServe(config.Server.Listen, nil)
}
