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
	"github.com/defsub/takeout/activity"
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	"github.com/defsub/takeout/lib/hub"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/progress"
)

type bits uint8

const (
	AllowCookie bits = 1 << iota
	AllowAccessToken
	AllowMediaToken

	SuccessRedirect = "/"
	LinkRedirect    = "/static/link.html"
	LoginRedirect   = "/static/login.html"

	AuthorizationHeader = "Authorization"
	BearerAuthorization = "Bearer"
)

// doLogin creates a login session for the provided user or returns an error
func doLogin(ctx Context, user, pass string) (auth.Session, error) {
	return ctx.Auth().Login(user, pass)
}

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
		http.Redirect(w, r, SuccessRedirect, http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, LinkRedirect, http.StatusTemporaryRedirect)
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
	mux.Get("/api/podcasts/series/:id", accessTokenAuthHandler(ctx, apiPodcastSeriesGet))
	mux.Get("/api/podcasts/series/:id/playlist", accessTokenAuthHandler(ctx, apiPodcastSeriesGetPlaylist))
	mux.Get("/api/podcasts/series/:id/playlist.xspf", accessTokenAuthHandler(ctx, apiPodcastSeriesGetPlaylist))
	mux.Get("/api/podcasts/series/:id/episodes/:eid", accessTokenAuthHandler(ctx, apiPodcastSeriesEpisodeGet))
	mux.Get("/api/podcasts/series/:id/episodes/:eid/playlist", accessTokenAuthHandler(ctx, apiPodcastSeriesEpisodeGetPlaylist))
	mux.Get("/api/podcasts/series/:id/episodes/:eid/playlist.xspf", accessTokenAuthHandler(ctx, apiPodcastSeriesEpisodeGetPlaylist))
	mux.Get("/api/episodes/:eid", accessTokenAuthHandler(ctx, apiPodcastSeriesEpisodeGet))
	mux.Get("/api/series/:id", accessTokenAuthHandler(ctx, apiPodcastSeriesGet))
	mux.Get("/api/series/:id/playlist", accessTokenAuthHandler(ctx, apiPodcastSeriesGetPlaylist))
	mux.Get("/api/series/:id/playlist.xspf", accessTokenAuthHandler(ctx, apiPodcastSeriesGetPlaylist))

	// location
	mux.Get("/api/tracks/:uuid/location", mediaTokenAuthHandler(ctx, apiTrackLocation))
	mux.Get("/api/movies/:uuid/location", mediaTokenAuthHandler(ctx, apiMovieLocation))
	mux.Get("/api/episodes/:eid/location", mediaTokenAuthHandler(ctx, apiSeriesEpisodeLocation))
	mux.Get("/api/podcasts/:id/episodes/:eid/location", mediaTokenAuthHandler(ctx, apiSeriesEpisodeLocation))

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

	// // swaggerHandler := func(w http.ResponseWriter, r *http.Request) {
	// // 	http.Redirect(w, r, "/static/swagger.json", 302)
	// // }
	// http.HandleFunc("/swagger.json", swaggerHandler)

	log.Printf("listening on %s\n", config.Server.Listen)
	http.Handle("/", mux)
	return http.ListenAndServe(config.Server.Listen, nil)
}
