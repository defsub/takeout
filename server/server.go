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

	"github.com/bmizerany/pat"
	"github.com/defsub/takeout/activity"
	"github.com/defsub/takeout/auth"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	"github.com/defsub/takeout/lib/hub"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/progress"
)

const (
	SuccessRedirect = "/"
	LinkRedirect    = "/static/link.html"
	LoginRedirect   = "/static/login.html"
)

// doLogin creates a login session for the provided user or returns an error
func doLogin(ctx Context, user, pass string) (auth.Session, error) {
	return ctx.Auth().Login(user, pass)
}

// upgradeContext creates a full context based on user and media configuration.
// This is used for most requests after the user has been authorized.
func upgradeContext(ctx Context, user *auth.User) (RequestContext, error) {
	mediaName, userConfig, err := mediaConfigFor(ctx.Config(), user)
	if err != nil {
		return RequestContext{}, err
	}
	media := makeMedia(mediaName, userConfig)
	return makeContext(ctx, user, userConfig, media), nil
}

// sessionContext creates a minimal context with the provided session.
func sessionContext(ctx Context, session *auth.Session) RequestContext {
	return makeAuthOnlyContext(ctx, session)
}

// imageContext creates a minimal context with the provided client.
func imageContext(ctx Context, client *client.Client) RequestContext {
	return makeImageContext(ctx, client)
}

// loginHandler performs a web based login session and sends back a cookie.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	r.ParseForm()
	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	session, err := doLogin(ctx, user, pass)
	if err != nil {
		authErr(w, ErrUnauthorized)
		return
	}

	cookie := ctx.Auth().NewCookie(&session)
	http.SetCookie(w, &cookie)

	// Use 303 for PRG
	// https://en.wikipedia.org/wiki/Post/Redirect/Get
	http.Redirect(w, r, SuccessRedirect, http.StatusSeeOther)
}

// linkHandler performs a web based login and links to the provided code.
func linkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	r.ParseForm()
	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	value := r.Form.Get("code")
	err := doCodeAuth(ctx, user, pass, value)
	if err == nil {
		// success
		// Use 303 for PRG
		http.Redirect(w, r, SuccessRedirect, http.StatusSeeOther)
		return
	}
	// Use 303 for PRG
	http.Redirect(w, r, LinkRedirect, http.StatusSeeOther)
}

// imageHandler handles unauthenticated image requests.
func imageHandler(ctx RequestContext, handler http.HandlerFunc, client *client.Client) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := imageContext(ctx, client)
		handler.ServeHTTP(w, withContext(r, ctx))
	}
	return http.HandlerFunc(fn)
}

// requestHandler handles unauthenticated requests.
func requestHandler(ctx RequestContext, handler http.HandlerFunc) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, withContext(r, ctx))
	}
	return http.HandlerFunc(fn)
}

// hubHandler handles hub requests.
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

func makeActivity(config *config.Config) (*activity.Activity, error) {
	a := activity.NewActivity(config)
	err := a.Open()
	return a, err
}

func makeProgress(config *config.Config) (*progress.Progress, error) {
	p := progress.NewProgress(config)
	err := p.Open()
	return p, err
}

func makeHub(config *config.Config) (*hub.Hub, error) {
	h := hub.NewHub()
	go h.Run()
	return h, nil
}

// Serve configures and starts the Takeout web, websocket, and API services.
func Serve(config *config.Config) error {
	auth, err := makeAuth(config)
	log.CheckError(err)

	activity, err := makeActivity(config)
	log.CheckError(err)

	progress, err := makeProgress(config)
	log.CheckError(err)

	hub, err := makeHub(config)
	log.CheckError(err)

	schedule(config)

	// base context for all requests
	ctx := RequestContext{
		activity: activity,
		auth:     auth,
		config:   config,
		progress: progress,
		template: getTemplates(config),
	}

	resFileServer := http.FileServer(mountResFS(resStatic))
	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		resFileServer.ServeHTTP(w, r)
	}

	mux := pat.New()
	mux.Get("/static/", http.HandlerFunc(staticHandler))
	mux.Get("/", accessTokenAuthHandler(ctx, viewHandler))
	mux.Get("/v", accessTokenAuthHandler(ctx, viewHandler))

	// cookie auth
	mux.Post("/api/login", requestHandler(ctx, apiLogin))
	mux.Post("/login", requestHandler(ctx, loginHandler))
	mux.Post("/link", requestHandler(ctx, linkHandler))

	// token auth
	mux.Post("/api/token", requestHandler(ctx, apiTokenLogin))
	mux.Get("/api/token", refreshTokenAuthHandler(ctx, apiTokenRefresh))

	// code auth
	mux.Get("/api/code", requestHandler(ctx, apiCodeGet))
	mux.Post("/api/code", codeTokenAuthHandler(ctx, apiCodeCheck))

	// misc
	mux.Get("/api/home", accessTokenAuthHandler(ctx, apiHome))
	mux.Get("/api/index", accessTokenAuthHandler(ctx, apiIndex))
	mux.Get("/api/search", accessTokenAuthHandler(ctx, apiSearch))

	// playlist
	mux.Get("/api/playlist", accessTokenAuthHandler(ctx, apiPlaylistGet))
	mux.Patch("/api/playlist", accessTokenAuthHandler(ctx, apiPlaylistPatch))

	// music
	mux.Get("/api/artists", accessTokenAuthHandler(ctx, apiArtists))
	mux.Get("/api/artists/:id", accessTokenAuthHandler(ctx, apiArtistGet))
	mux.Get("/api/artists/:id/:res", accessTokenAuthHandler(ctx, apiArtistGetResource))
	mux.Get("/api/artists/:id/:res/playlist", accessTokenAuthHandler(ctx, apiArtistGetPlaylist))
	mux.Get("/api/artists/:id/:res/playlist.xspf", accessTokenAuthHandler(ctx, apiArtistGetPlaylist))
	mux.Get("/api/radio", accessTokenAuthHandler(ctx, apiRadioGet))
	mux.Get("/api/radio/:id", accessTokenAuthHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/:id/playlist", accessTokenAuthHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/:id/playlist.xspf", accessTokenAuthHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/stations/:id", accessTokenAuthHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/stations/:id/playlist", accessTokenAuthHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/radio/stations/:id/playlist.xspf", accessTokenAuthHandler(ctx, apiRadioStationGetPlaylist))
	mux.Get("/api/releases/:id", accessTokenAuthHandler(ctx, apiReleaseGet))
	mux.Get("/api/releases/:id/playlist", accessTokenAuthHandler(ctx, apiReleaseGetPlaylist))
	mux.Get("/api/releases/:id/playlist.xspf", accessTokenAuthHandler(ctx, apiReleaseGetPlaylist))

	// video
	mux.Get("/api/movies", accessTokenAuthHandler(ctx, apiMovies))
	mux.Get("/api/movies/:id", accessTokenAuthHandler(ctx, apiMovieGet))
	mux.Get("/api/movies/:id/playlist", accessTokenAuthHandler(ctx, apiMovieGetPlaylist))
	mux.Get("/api/movies/genres/:name", accessTokenAuthHandler(ctx, apiMovieGenreGet))
	mux.Get("/api/movies/keywords/:name", accessTokenAuthHandler(ctx, apiMovieKeywordGet))
	mux.Get("/api/profiles/:id", accessTokenAuthHandler(ctx, apiMovieProfileGet))
	// mux.Get("/api/tv", apiTVShows)
	// mux.Get("/api/tv/:id", apiTVShowGet)
	// mux.Get("/api/tv/:id/episodes/:eid", apiTVShowEpisodeGet)

	// podcast
	mux.Get("/api/podcasts", accessTokenAuthHandler(ctx, apiPodcasts))
	mux.Get("/api/series/:id", accessTokenAuthHandler(ctx, apiPodcastSeriesGet))
	mux.Get("/api/series/:id/playlist", accessTokenAuthHandler(ctx, apiPodcastSeriesGetPlaylist))
	mux.Get("/api/series/:id/playlist.xspf", accessTokenAuthHandler(ctx, apiPodcastSeriesGetPlaylist))
	mux.Get("/api/episodes/:id", accessTokenAuthHandler(ctx, apiPodcastEpisodeGet))
	mux.Get("/api/episodes/:id/playlist", accessTokenAuthHandler(ctx, apiPodcastEpisodeGetPlaylist))
	mux.Get("/api/episodes/:id/playlist.xspf", accessTokenAuthHandler(ctx, apiPodcastEpisodeGetPlaylist))

	// location
	mux.Get("/api/tracks/:uuid/location", mediaTokenAuthHandler(ctx, apiTrackLocation))
	mux.Get("/api/movies/:uuid/location", mediaTokenAuthHandler(ctx, apiMovieLocation))
	mux.Get("/api/episodes/:id/location", mediaTokenAuthHandler(ctx, apiEpisodeLocation))

	// progress
	mux.Get("/api/progress", accessTokenAuthHandler(ctx, apiProgressGet))
	mux.Post("/api/progress", accessTokenAuthHandler(ctx, apiProgressPost))

	// activity
	mux.Get("/api/activity", accessTokenAuthHandler(ctx, apiActivityGet))
	mux.Post("/api/activity", accessTokenAuthHandler(ctx, apiActivityPost))
	mux.Get("/api/activity/tracks", accessTokenAuthHandler(ctx, apiActivityTracksGet))
	mux.Get("/api/activity/tracks/:res", accessTokenAuthHandler(ctx, apiActivityTracksGetResource))
	mux.Get("/api/activity/tracks/:res/playlist", accessTokenAuthHandler(ctx, apiActivityTracksGetPlaylist))
	mux.Get("/api/activity/movies", accessTokenAuthHandler(ctx, apiActivityMoviesGet))
	mux.Get("/api/activity/releases", accessTokenAuthHandler(ctx, apiActivityReleasesGet))
	// /activity/radio - ?

	// Hub
	mux.Get("/live", hubHandler(ctx, hub))

	// Hook
	mux.Post("/hook/", requestHandler(ctx, hookHandler))

	// Images
	client := client.NewClient(&config.Server.ImageClient)
	client.UseOnlyIfCached(true)
	mux.Get("/img/mb/rg/:rgid", imageHandler(ctx, imgReleaseGroupFront, client))
	mux.Get("/img/mb/rg/:rgid/:side", imageHandler(ctx, imgReleaseGroup, client))
	mux.Get("/img/mb/re/:reid", imageHandler(ctx, imgReleaseFront, client))
	mux.Get("/img/mb/re/:reid/:side", imageHandler(ctx, imgRelease, client))
	mux.Get("/img/tm/:size/:path", imageHandler(ctx, imgVideo, client))
	mux.Get("/img/fa/:arid/t/:path", imageHandler(ctx, imgArtistThumb, client))
	mux.Get("/img/fa/:arid/b/:path", imageHandler(ctx, imgArtistBackground, client))

	// pprof
	// mux.Get("/debug/pprof", http.HandlerFunc(pprof.Index))
	// mux.Get("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	// mux.Get("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	// mux.Get("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	// mux.Get("/debug/pprof/heap", pprof.Handler("heap"))
	// mux.Get("/debug/pprof/block", pprof.Handler("block"))
	// mux.Get("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	// mux.Get("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	// // swaggerHandler := func(w http.ResponseWriter, r *http.Request) {
	// // 	http.Redirect(w, r, "/static/swagger.json", 302)
	// // }
	// http.HandleFunc("/swagger.json", swaggerHandler)

	log.Printf("listening on %s\n", config.Server.Listen)
	http.Handle("/", mux)
	return http.ListenAndServe(config.Server.Listen, nil)
}
