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
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/defsub/takeout/activity"
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/progress"
	"github.com/defsub/takeout/video"
)

type contextKey string

var (
	contextKeyContext = contextKey("context")
)

func withContext(r *http.Request, ctx Context) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKeyContext, ctx))
}

func contextValue(r *http.Request) Context {
	return r.Context().Value(contextKeyContext).(Context)
}

type Context interface {
	Activity() *activity.Activity
	Auth() *auth.Auth
	Config() *config.Config
	Music() *music.Music
	Podcast() *podcast.Podcast
	Progress() *progress.Progress
	Template() *template.Template
	User() *auth.User
	Session() *auth.Session
	Video() *video.Video
	ImageClient() *client.Client

	LocateTrack(music.Track) string
	LocateMovie(video.Movie) string
	LocateEpisode(podcast.Episode) string

	FindArtist(string) (music.Artist, error)
	FindRelease(string) (music.Release, error)
	FindTrack(string) (music.Track, error)
	FindStation(string) (music.Station, error)
	FindMovie(string) (video.Movie, error)
	FindSeries(string) (podcast.Series, error)
	FindEpisode(string) (podcast.Episode, error)

	TrackImage(music.Track) string
	ArtistImage(music.Artist) string
	ArtistBackground(music.Artist) string
	MovieImage(video.Movie) string
	EpisodeImage(podcast.Episode) string
}

type RequestContext struct {
	activity    *activity.Activity
	auth        *auth.Auth
	config      *config.Config
	user        *auth.User
	media       *Media
	progress    *progress.Progress
	session     *auth.Session
	template    *template.Template
	imageClient *client.Client
}

func makeContext(ctx Context, u *auth.User, c *config.Config, m *Media) RequestContext {
	return RequestContext{
		activity: ctx.Activity(),
		config:   c,
		media:    m,
		progress: ctx.Progress(),
		template: ctx.Template(),
		user:     u,
	}
}

func makeAuthOnlyContext(ctx Context, session *auth.Session) RequestContext {
	return RequestContext{
		auth:    ctx.Auth(),
		session: session,
	}
}

func makeImageContext(ctx Context, client *client.Client) RequestContext {
	return RequestContext{
		imageClient: client,
	}
}

func (ctx RequestContext) Activity() *activity.Activity {
	return ctx.activity
}

func (ctx RequestContext) Auth() *auth.Auth {
	return ctx.auth
}

func (ctx RequestContext) Config() *config.Config {
	return ctx.config
}

func (ctx RequestContext) Music() *music.Music {
	return ctx.media.music
}

func (ctx RequestContext) Podcast() *podcast.Podcast {
	return ctx.media.podcast
}

func (ctx RequestContext) Progress() *progress.Progress {
	return ctx.progress
}

func (ctx RequestContext) Template() *template.Template {
	return ctx.template
}

func (ctx RequestContext) User() *auth.User {
	return ctx.user
}

func (ctx RequestContext) Session() *auth.Session {
	return ctx.session
}

func (ctx RequestContext) Video() *video.Video {
	return ctx.media.video
}

func (RequestContext) LocateTrack(t music.Track) string {
	return locateTrack(t)
}

func (RequestContext) LocateMovie(v video.Movie) string {
	return locateMovie(v)
}

func (RequestContext) LocateEpisode(e podcast.Episode) string {
	return locateEpisode(e)
}

func (ctx RequestContext) FindArtist(id string) (music.Artist, error) {
	return ctx.Music().FindArtist(id)
}

func (ctx RequestContext) FindRelease(id string) (music.Release, error) {
	return ctx.Music().FindRelease(id)
}

func (ctx RequestContext) FindTrack(id string) (music.Track, error) {
	return ctx.Music().FindTrack(id)
}

func (ctx RequestContext) FindStation(id string) (music.Station, error) {
	return ctx.Music().FindStation(id)
}

func (ctx RequestContext) FindMovie(id string) (video.Movie, error) {
	return ctx.Video().FindMovie(id)
}

func (ctx RequestContext) FindSeries(id string) (podcast.Series, error) {
	return ctx.Podcast().FindSeries(id)
}

func (ctx RequestContext) FindEpisode(id string) (podcast.Episode, error) {
	return ctx.Podcast().FindEpisode(id)
}

func (ctx RequestContext) TrackImage(t music.Track) string {
	return ctx.Music().TrackImage(t).String()
}

func (ctx RequestContext) ArtistImage(a music.Artist) string {
	return ctx.Music().ArtistImage(&a)
}

func (ctx RequestContext) ArtistBackground(a music.Artist) string {
	return ctx.Music().ArtistBackground(&a)
}

func (ctx RequestContext) MovieImage(m video.Movie) string {
	return ctx.Video().MoviePoster(m)
}

func (ctx RequestContext) EpisodeImage(e podcast.Episode) string {
	return ctx.Podcast().EpisodeImage(e)
}

func (ctx RequestContext) ImageClient() *client.Client {
	return ctx.imageClient
}

func locateTrack(t music.Track) string {
	return fmt.Sprintf("/api/tracks/%s/location", t.UUID)
}

func locateMovie(v video.Movie) string {
	return fmt.Sprintf("/api/movies/%s/location", v.UUID)
}

func locateEpisode(e podcast.Episode) string {
	return fmt.Sprintf("/api/episodes/%d/location", e.ID)
}
