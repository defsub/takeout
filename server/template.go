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

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/date"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
)

//go:embed res/static
var resStatic embed.FS

func mountResFS(resFS embed.FS) http.FileSystem {
	fsys, err := fs.Sub(resFS, "res")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

// func getStaticFS(config *config.Config) http.FileSystem {
// 	// dev := false
// 	// if dev {
// 	// 	return http.FS(os.DirFS(fmt.Sprintf("%s/static", config.Server.WebDir)))
// 	// }
// 	fsys, err := fs.Sub(resStatic, "res")
// 	if err != nil {
// 		panic(err)
// 	}
// 	return http.FS(fsys)
// }

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
		"res/template/video/*.html",
		"res/template/podcast/*.html"))
}

func doFuncMap() template.FuncMap {
	return template.FuncMap{
		"join": strings.Join,
		"ymd":  date.YMD,
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
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
			case podcast.Series:
				link = fmt.Sprintf("/v?series=%d", o.(podcast.Series).ID)
			case podcast.Episode:
				link = fmt.Sprintf("/v?episode=%d", o.(podcast.Episode).ID)
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
			case podcast.Episode:
				loc = locateEpisode(o.(podcast.Episode))
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

func (handler *UserHandler) viewHandler(w http.ResponseWriter, r *http.Request) {
	var view interface{}
	var temp string

	if v := r.URL.Query().Get("release"); v != "" {
		// /v?release={release-id}
		m := handler.music()
		id, _ := strconv.Atoi(v)
		release, _ := m.LookupRelease(id)
		view = handler.releaseView(release)
		temp = "release.html"
	} else if v := r.URL.Query().Get("artist"); v != "" {
		// /v?artist={artist-id}
		m := handler.music()
		id, _ := strconv.Atoi(v)
		artist, _ := m.LookupArtist(id)
		view = handler.artistView(artist)
		temp = "artist.html"
	} else if v := r.URL.Query().Get("artists"); v != "" {
		// /v?artists=x
		view = handler.artistsView()
		temp = "artists.html"
	} else if v := r.URL.Query().Get("popular"); v != "" {
		// /v?popular={artist-id}
		m := handler.music()
		id, _ := strconv.Atoi(v)
		artist, _ := m.LookupArtist(id)
		view = handler.popularView(artist)
		temp = "popular.html"
	} else if v := r.URL.Query().Get("singles"); v != "" {
		// /v?singles={artist-id}
		m := handler.music()
		id, _ := strconv.Atoi(v)
		artist, _ := m.LookupArtist(id)
		view = handler.singlesView(artist)
		temp = "singles.html"
	} else if v := r.URL.Query().Get("home"); v != "" {
		// /v?home=x
		view = handler.homeView()
		temp = "home.html"
	} else if v := r.URL.Query().Get("q"); v != "" {
		// /v?q={pattern}
		view = handler.searchView(strings.TrimSpace(v))
		temp = "search.html"
	} else if v := r.URL.Query().Get("radio"); v != "" {
		// /v?radio=x
		view = handler.radioView(handler.user)
		temp = "radio.html"
	} else if v := r.URL.Query().Get("movies"); v != "" {
		// /v?movies=x
		view = handler.moviesView()
		temp = "movies.html"
	} else if v := r.URL.Query().Get("movie"); v != "" {
		// /v?movie={movie-id}
		vid := handler.video()
		id, _ := strconv.Atoi(v)
		movie, _ := vid.LookupMovie(id)
		view = handler.movieView(movie)
		temp = "movie.html"
	} else if v := r.URL.Query().Get("profile"); v != "" {
		// /v?profile={person-id}
		vid := handler.video()
		id, _ := strconv.Atoi(v)
		person, _ := vid.LookupPerson(id)
		view = handler.profileView(person)
		temp = "profile.html"
	} else if v := r.URL.Query().Get("genre"); v != "" {
		// /v?genre={genre-name}
		name := strings.TrimSpace(v)
		view = handler.genreView(name)
		temp = "genre.html"
	} else if v := r.URL.Query().Get("keyword"); v != "" {
		// /v?keyword={keyword-name}
		name := strings.TrimSpace(v)
		view = handler.keywordView(name)
		temp = "keyword.html"
	} else if v := r.URL.Query().Get("watch"); v != "" {
		// /v?watch={movie-id}
		vid := handler.video()
		id, _ := strconv.Atoi(v)
		movie, _ := vid.LookupMovie(id)
		view = handler.watchView(movie)
		temp = "watch.html"
	} else if v := r.URL.Query().Get("podcasts"); v != "" {
		// /v?podcasts=x
		view = handler.podcastsView()
		temp = "podcasts.html"
	} else if v := r.URL.Query().Get("series"); v != "" {
		// /v?series={series-id}
		p := handler.podcast()
		id, _ := strconv.Atoi(v)
		series, _ := p.LookupSeries(id)
		view = handler.seriesView(series)
		temp = "series.html"
	} else if v := r.URL.Query().Get("episode"); v != "" {
		// /v?episode={episode-id}
		p := handler.podcast()
		id, _ := strconv.Atoi(v)
		episode, _ := p.LookupEpisode(id)
		view = handler.seriesEpisodeView(episode)
		temp = "episode.html"
	} else {
		view = handler.indexView()
		temp = "index.html"
	}

	handler.render(temp, view, w, r)
}

func (handler *UserHandler) render(temp string, view interface{},
	w http.ResponseWriter, r *http.Request) {
	err := handler.template.ExecuteTemplate(w, temp, view)
	if err != nil {
		serverErr(w, err)
	}
}
