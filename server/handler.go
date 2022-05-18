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

	"html/template"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/progress"
	"github.com/defsub/takeout/video"
	"github.com/defsub/takeout/lib/hub"
)

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

type Handler struct {
	config   *config.Config
	auth     *auth.Auth
	template *template.Template
}

type HubHandler struct {
	config   *config.Config
	auth     *auth.Auth
	hub      *hub.Hub
}

type UserHandler struct {
	user     *auth.User
	config   *config.Config
	media    *Media
	template *template.Template
}

func (handler UserHandler) music() *music.Music {
	return handler.media.music
}

func (handler UserHandler) video() *video.Video {
	return handler.media.video
}

func (handler UserHandler) podcast() *podcast.Podcast {
	return handler.media.podcast
}

func (handler UserHandler) progress() *progress.Progress {
	return handler.media.progress
}

func (UserHandler) LocateTrack(t music.Track) string {
	return locateTrack(t)
}

func (UserHandler) LocateMovie(v video.Movie) string {
	return locateMovie(v)
}

func (UserHandler) LocateEpisode(e podcast.Episode) string {
	return locateEpisode(e)
}

func (handler UserHandler) TrackImage(t music.Track) string {
	return handler.music().TrackImage(t).String()
}

func (handler UserHandler) MovieImage(m video.Movie) string {
	return handler.video().MoviePoster(m)
}

func (handler UserHandler) EpisodeImage(e podcast.Episode) string {
	return handler.podcast().EpisodeImage(e)
}

func (handler UserHandler) LookupArtist(id int) (music.Artist, error) {
	return handler.music().LookupArtist(id)
}

func (handler UserHandler) LookupRelease(id int) (music.Release, error) {
	return handler.music().LookupRelease(id)
}

func (handler UserHandler) LookupTrack(id int) (music.Track, error) {
	return handler.music().LookupTrack(id)
}

func (handler UserHandler) LookupStation(id int) (music.Station, error) {
	return handler.music().LookupStation(id)
}

func (handler UserHandler) LookupMovie(id int) (video.Movie, error) {
	return handler.video().LookupMovie(id)
}

func (handler UserHandler) LookupSeries(id int) (podcast.Series, error) {
	return handler.podcast().LookupSeries(id)
}

func (handler UserHandler) ArtistSingleTracks(a music.Artist) []music.Track {
	return handler.music().ArtistSingleTracks(a)
}

func (handler UserHandler) ArtistPopularTracks(a music.Artist) []music.Track {
	return handler.music().ArtistPopularTracks(a)
}

func (handler UserHandler) ArtistTracks(a music.Artist) []music.Track {
	return handler.music().ArtistTracks(a)
}

func (handler UserHandler) ArtistShuffle(a music.Artist) []music.Track {
	return handler.music().ArtistShuffle(a, handler.config.Music.RadioLimit)
}

func (handler UserHandler) ArtistRadio(a music.Artist) []music.Track {
	return handler.music().ArtistRadio(a)
}

func (handler UserHandler) ArtistDeep(a music.Artist) []music.Track {
	return handler.music().ArtistDeep(a, handler.config.Music.RadioLimit)
}

func (handler UserHandler) ReleaseTracks(r music.Release) []music.Track {
	return handler.music().ReleaseTracks(r)
}

func (handler UserHandler) MusicSearch(query string, limit int) []music.Track {
	return handler.music().Search(query, limit)
}

func (handler UserHandler) SeriesEpisodes(series podcast.Series) []podcast.Episode {
	return handler.podcast().Episodes(series)
}

func locateTrack(t music.Track) string {
	return fmt.Sprintf("/api/tracks/%d/location", t.ID)
}

func locateMovie(v video.Movie) string {
	return fmt.Sprintf("/api/movies/%d/location", v.ID)
}

func locateEpisode(e podcast.Episode) string {
	return fmt.Sprintf("/api/episodes/%d/location", e.ID)
}
