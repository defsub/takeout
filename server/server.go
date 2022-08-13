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
	"net/http"
	"strings"

	"github.com/bmizerany/pat"
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/lib/hub"
)

const (
	ApplicationJson = "application/json"

	SuccessRedirect = "/"
	LinkRedirect    = "/static/link.html"
	LoginRedirect   = "/static/login.html"
)

// remove?
// func (handler *UserHandler) tracksHandler(w http.ResponseWriter, r *http.Request, m *music.Music) {
// 	var tracks []music.Track
// 	if v := r.URL.Query().Get("q"); v != "" {
// 		tracks = m.Search(strings.TrimSpace(v))
// 	}

// 	if len(tracks) > 0 {
// 		handler.doSpiff(m, "Takeout", tracks, w, r)
// 	} else {
// 		notFoundErr(w)
// 		return
// 	}
// }

// remove?
// func (handler *UserHandler) doSpiff(m *music.Music, title string, tracks []music.Track,
// 	w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-type", xspf.XMLContentType)

// 	encoder := xspf.NewXMLEncoder(w)
// 	encoder.Header(title)
// 	for _, t := range tracks {
// 		t.Location = []string{m.TrackURL(&t).String()}
// 		encoder.Encode(t)
// 	}
// 	encoder.Footer()
// }

func doLogin(ctx Context, user, pass string) (http.Cookie, error) {
	return ctx.Auth().Login(user, pass)
}

func doCodeAuth(ctx Context, user, pass, value string) error {
	cookie, err := ctx.Auth().Login(user, pass)
	if err != nil {
		return err
	}
	err = ctx.Auth().AuthorizeCode(value, cookie.Value)
	if err != nil {
		return ErrInvalidCode
	}
	return nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	r.ParseForm()
	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	cookie, err := doLogin(ctx, user, pass)
	if err == nil {
		// success
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, SuccessRedirect, http.StatusTemporaryRedirect)
		return
	}
	authErr(w, ErrUnauthorized)
}

func linkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	r.ParseForm()
	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	value := r.Form.Get("code")
	err := doCodeAuth(ctx, user, pass, value)
	if err == nil {
		// success
		http.Redirect(w, r, SuccessRedirect, http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, LinkRedirect, http.StatusTemporaryRedirect)
}

func authorizeBearer(ctx Context, w http.ResponseWriter, r *http.Request) *auth.User {
	value := r.Header.Get("Authorization")
	if value == "" {
		return nil
	}
	result := strings.Split(value, " ")
	var token string
	switch len(result) {
	case 1:
		// Authorization: <token>
		token = result[0]
	case 2:
		// Authorization: Bearer <token>
		if strings.EqualFold(result[0], "Bearer") {
			token = result[1]
		}
	}
	if len(token) == 0 {
		return nil
	}
	user, err := ctx.Auth().TokenUser(token)
	if err != nil {
		return nil
	}
	return user
}

func authorizeCookie(ctx Context, w http.ResponseWriter, r *http.Request) *auth.User {
	a := ctx.Auth()
	cookie, err := r.Cookie(auth.CookieName)
	if err != nil {
		if cookie != nil {
			a.Expire(cookie)
			http.SetCookie(w, cookie)
		}
		http.Redirect(w, r, LoginRedirect, http.StatusTemporaryRedirect)
		return nil
	}

	valid := a.Valid(*cookie)
	if !valid {
		a.Logout(*cookie)
		a.Expire(cookie)
		http.SetCookie(w, cookie)
		authErr(w, ErrUnauthorized)
		return nil
	}

	user, err := a.UserAuth(*cookie)
	if err != nil {
		a.Logout(*cookie)
		authErr(w, ErrUnauthorized)
		a.Expire(cookie)
		http.SetCookie(w, cookie)
		return nil
	}

	a.Refresh(cookie)
	http.SetCookie(w, cookie)

	return user
}

func authorizeUser(ctx Context, w http.ResponseWriter, r *http.Request) *auth.User {
	// TODO JWT
	// check for bearer
	user := authorizeBearer(ctx, w, r)
	if user != nil {
		return user
	}
	// check for cookie
	return authorizeCookie(ctx, w, r)
}

func upgradeContext(ctx Context, user *auth.User) (RequestContext, error) {
	mediaName, userConfig, err := mediaConfigFor(ctx.Config(), user)
	if err != nil {
		return RequestContext{}, err
	}
	media := makeMedia(mediaName, userConfig)
	return makeContext(ctx, user, userConfig, media), nil
}

func authHandler(ctx RequestContext, handler http.HandlerFunc) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := authorizeUser(ctx, w, r)
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

func requestHandler(ctx RequestContext, handler http.HandlerFunc) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r = withContext(r, ctx)
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func hubHandler(ctx RequestContext, h *hub.Hub) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r = withContext(r, ctx)
		h.Handle(ctx.Auth(), w, r)
	}
	return http.HandlerFunc(fn)
}

func makeAuth(config *config.Config) (*auth.Auth, error) {
	a := auth.NewAuth(config)
	err := a.Open()
	return a, err
}

func makeHub(config *config.Config) (*hub.Hub, error) {
	h := hub.NewHub()
	go h.Run()
	return h, nil
}

func Serve(config *config.Config) {
	//schedule(config)

	hub, err := makeHub(config)
	if err != nil {
		log.CheckError(err)
	}

	auth, err := makeAuth(config)
	if err != nil {
		log.CheckError(err)
	}

	ctx := RequestContext{config: config, auth: auth, template: getTemplates(config)}

	resFileServer := http.FileServer(mountResFS(resStatic))
	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		resFileServer.ServeHTTP(w, r)
	}

	mux := pat.New()
	mux.Get("/static/", http.HandlerFunc(staticHandler))
	mux.Get("/", authHandler(ctx, viewHandler))
	mux.Get("/v", authHandler(ctx, viewHandler))

	// authorize
	mux.Post("/api/login", requestHandler(ctx, apiLogin))
	mux.Post("/login", requestHandler(ctx, loginHandler))
	mux.Post("/link", requestHandler(ctx, linkHandler))

	// misc
	mux.Get("/api/home", authHandler(ctx, apiHome))
	mux.Get("/api/index", authHandler(ctx, apiIndex))
	mux.Get("/api/search", authHandler(ctx, apiSearch))

	// playlist
	mux.Get("/api/playlist", authHandler(ctx, apiPlaylistGet))
	mux.Patch("/api/playlist", authHandler(ctx, apiPlaylistPatch))

	// music
	mux.Get("/api/artists", authHandler(ctx, apiArtists))
	mux.Get("/api/artists/:id", authHandler(ctx, apiArtistGet))
	mux.Get("/api/artists/:id/:res", authHandler(ctx, apiArtistGetResource))
	mux.Get("/api/artists/:id/:res/playlist", authHandler(ctx, apiArtistGetPlaylist))
	mux.Get("/api/artists/:id/:res/playlist.xspf", authHandler(ctx, apiArtistGetPlaylist))
	mux.Get("/api/radio", authHandler(ctx, apiRadioGet))
	mux.Get("/api/radio/:id", authHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/:id/playlist", authHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/:id/playlist.xspf", authHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/stations/:id", authHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/stations/:id/playlist", authHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/stations/:id/playlist.xspf", authHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/releases/:id", authHandler(ctx, apiReleaseGet))
	mux.Get("/api/releases/:id/playlist", authHandler(ctx, apiReleaseGetPlaylist))
	mux.Get("/api/releases/:id/playlist.xspf", authHandler(ctx, apiReleaseGetPlaylist))

	// video
	mux.Get("/api/movies", authHandler(ctx, apiMovies))
	mux.Get("/api/movies/:id", authHandler(ctx, apiMovieGet))
	mux.Get("/api/movies/:id/playlist", authHandler(ctx, apiMovieGetPlaylist))
	mux.Get("/api/movies/genres/:name", authHandler(ctx, apiMovieGenreGet))
	mux.Get("/api/movies/keywords/:name", authHandler(ctx, apiMovieKeywordGet))
	mux.Get("/api/profiles/:id", authHandler(ctx, apiMovieProfileGet))
	// mux.Get("/api/tv", apiTVShows)
	// mux.Get("/api/tv/:id", apiTVShowGet)
	// mux.Get("/api/tv/:id/episodes/:eid", apiTVShowEpisodeGet)

	// podcast
	mux.Get("/api/podcasts", authHandler(ctx, apiPodcasts))
	mux.Get("/api/podcasts/series/:id", authHandler(ctx, apiPodcastSeriesGet))
	mux.Get("/api/podcasts/series/:id/playlist", authHandler(ctx, apiPodcastSeriesGetPlaylist))
	mux.Get("/api/podcasts/series/:id/playlist.xspf", authHandler(ctx, apiPodcastSeriesGetPlaylist))
	mux.Get("/api/podcasts/series/:id/episodes/:eid", authHandler(ctx, apiPodcastSeriesEpisodeGet))
	mux.Get("/api/podcasts/series/:id/episodes/:eid/playlist", authHandler(ctx, apiPodcastSeriesEpisodeGetPlaylist))
	mux.Get("/api/podcasts/series/:id/episodes/:eid/playlist.xspf", authHandler(ctx, apiPodcastSeriesEpisodeGetPlaylist))
	mux.Get("/api/episodes/:eid", authHandler(ctx, apiPodcastSeriesEpisodeGet))
	mux.Get("/api/series/:id", authHandler(ctx, apiPodcastSeriesGet))
	mux.Get("/api/series/:id/playlist", authHandler(ctx, apiPodcastSeriesGetPlaylist))
	mux.Get("/api/series/:id/playlist.xspf", authHandler(ctx, apiPodcastSeriesGetPlaylist))

	// location
	mux.Get("/api/tracks/:id/location", authHandler(ctx, apiTrackLocation))
	mux.Get("/api/movies/:id/location", authHandler(ctx, apiMovieLocation))
	mux.Get("/api/episodes/:eid/location", authHandler(ctx, apiSeriesEpisodeLocation))
	mux.Get("/api/podcasts/:id/episodes/:eid/location", authHandler(ctx, apiSeriesEpisodeLocation))

	// progress
	mux.Get("/api/progress", authHandler(ctx, apiProgressGet))
	mux.Post("/api/progress", authHandler(ctx, apiProgressPost))

	// Hub
	mux.Get("/live", hubHandler(ctx, hub))

	// Hook
	mux.Post("/hook/", requestHandler(ctx, hookHandler))

	// // swaggerHandler := func(w http.ResponseWriter, r *http.Request) {
	// // 	http.Redirect(w, r, "/static/swagger.json", 302)
	// // }

	// http.HandleFunc("/static/", staticHandler)
	// http.HandleFunc("/swagger.json", swaggerHandler)
	// http.HandleFunc("/tracks", tracksHandler)
	// http.HandleFunc("/", viewHandler)
	// http.HandleFunc("/v", viewHandler)
	// http.HandleFunc("/login", loginHandler)
	// http.HandleFunc("/link", linkHandler)
	// http.HandleFunc("/api/login", apiLoginHandler)
	// http.HandleFunc("/api/", apiHandler)
	// http.HandleFunc("/hook/", hookHandler)
	// http.HandleFunc("/live", liveHandler)
	log.Printf("listening on %s\n", config.Server.Listen)
	http.Handle("/", mux)
	http.ListenAndServe(config.Server.Listen, nil)
}
