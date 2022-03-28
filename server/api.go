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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/defsub/takeout/lib/date"
	"github.com/defsub/takeout/lib/encoding/xspf"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/lib/spiff"
	"github.com/defsub/takeout/lib/str"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/progress"
	"github.com/defsub/takeout/ref"
	"github.com/gorilla/websocket"
)

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
func (handler *Handler) apiLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", ApplicationJson)

	if r.Method != "POST" {
		serverErr(w, ErrInvalidMethod)
		return
	}

	var l login
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &l)
	if err != nil {
		serverErr(w, err)
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
		authErr(w, err)
		result = status{
			Status:  http.StatusUnauthorized,
			Message: "error",
		}
	}

	enc := json.NewEncoder(w)
	enc.Encode(result)
}

func (handler *UserHandler) recvStation(w http.ResponseWriter, r *http.Request,
	s *music.Station) error {
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, s)
	if err != nil {
		serverErr(w, err)
		return err
	}
	if s.Name == "" || s.Ref == "" {
		http.Error(w, "bummer", http.StatusBadRequest)
		return err
	}
	s.User = handler.user.Name
	if s.Ref == "/api/playlist" {
		// copy playlist
		p := handler.music().LookupPlaylist(handler.user)
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
func (handler *UserHandler) apiRadio(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		view := handler.radioView(handler.user)
		enc := json.NewEncoder(w)
		enc.Encode(view)
	case "POST":
		var s music.Station
		err := handler.recvStation(w, r, &s)
		if err != nil {
			return
		}
		err = handler.music().CreateStation(&s)
		if err != nil {
			serverErr(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		enc.Encode(s)
	default:
		http.Error(w, "bummer", http.StatusBadRequest)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (handler *UserHandler) apiLive(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "bummer", http.StatusBadRequest)
		return
	}
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("%v %s\n", messageType, p)
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
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
func (handler *UserHandler) apiStation(w http.ResponseWriter, r *http.Request, id int) {
	s, err := handler.music().LookupStation(id)
	if err != nil {
		notFoundErr(w)
		return
	}
	if !s.Visible(handler.user) {
		notFoundErr(w)
		return
	}

	switch r.Method {
	case "GET":
		resolver := ref.NewResolver(handler.config, handler)
		resolver.RefreshStation(&s, handler.user)

		w.WriteHeader(http.StatusOK)
		w.Write(s.Playlist)
	case "PUT":
		var up music.Station
		err := handler.recvStation(w, r, &up)
		if err != nil {
			return
		}
		s.Name = up.Name
		s.Ref = up.Ref
		s.Playlist = up.Playlist
		err = handler.music().UpdateStation(&s)
		if err != nil {
			serverErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case "PATCH":
		patch, _ := ioutil.ReadAll(r.Body)
		s.Playlist, err = spiff.Patch(s.Playlist, patch)
		if err != nil {
			serverErr(w, err)
			return
		}
		// unmarshal & resovle
		plist, _ := spiff.Unmarshal(s.Playlist)
		resolver := ref.NewResolver(handler.config, handler)
		resolver.Resolve(handler.user, plist)
		if plist.Entries == nil {
			plist.Entries = []spiff.Entry{}
		}
		// marshal & persist
		s.Playlist, _ = plist.Marshal()
		handler.music().UpdateStation(&s)
		w.WriteHeader(http.StatusNoContent)
	case "DELETE":
		err = handler.music().DeleteStation(&s)
		if err != nil {
			serverErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "bummer", http.StatusBadRequest)
	}
}

// GET /api/{res}/id/playlist > spiff.Playlist{}
// 200: success
func (handler *UserHandler) apiRefPlaylist(w http.ResponseWriter, r *http.Request,
	listType string, creator, title, image string, spiffDate time.Time, nref string) {
	plist := spiff.NewPlaylist(listType)
	//plist.Spiff.Location = fmt.Sprintf("%s%s", handler.config.Server.URL, r.URL.Path)
	plist.Spiff.Location = r.URL.Path
	plist.Spiff.Creator = creator
	plist.Spiff.Title = title
	plist.Spiff.Image = image
	plist.Spiff.Date = date.FormatJson(spiffDate)
	plist.Entries = []spiff.Entry{{Ref: nref}}
	resolver := ref.NewResolver(handler.config, handler)
	resolver.Resolve(handler.user, plist)
	if plist.Entries == nil {
		plist.Entries = []spiff.Entry{}
	}
	if strings.HasSuffix(r.URL.Path, ".xspf") {
		w.Header().Set("Content-type", xspf.XMLContentType)
		w.WriteHeader(http.StatusOK)
		encoder := xspf.NewXMLEncoder(w)
		encoder.Header(title)
		locationRegexp := regexp.MustCompile(`/api/(movies|tracks|episodes)/([0-9]+)/location`)
		for i := range plist.Entries {
			matches := locationRegexp.FindStringSubmatch(plist.Entries[i].Location[0])
			if matches != nil {
				var url *url.URL
				src := matches[1]
				if src == "tracks" {
					m := handler.music()
					id := str.Atoi(matches[2])
					track, err := m.LookupTrack(id)
					if err != nil {
						continue
					}
					url = m.TrackURL(&track)
					plist.Entries[i].Location = []string{url.String()}
					// } else if src == "movies" {
					// 	// TODO not supported yet
					// 	id := str.Atoi(matches[2])
					// 	movie, err := vid.LookupMovie(id)
					// 	if err != nil {
					// 		continue
					// 	}
					// 	url = vid.MovieURL(movie)
					// 	plist.Entries[i].Location = []string{url.String()}
				}
			}
			encoder.Encode(plist.Entries[i])
		}
		encoder.Footer()
	} else {
		w.WriteHeader(http.StatusOK)
		result, _ := plist.Marshal()
		w.Write(result)
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
func (handler *UserHandler) apiPlaylist(w http.ResponseWriter, r *http.Request) {
	var err error

	m := handler.music()
	p := m.LookupPlaylist(handler.user)
	if p == nil {
		plist := spiff.NewPlaylist(spiff.TypeMusic) // TODO music may not be correct
		//plist.Spiff.Location = fmt.Sprintf("%s/api/playlist", handler.config.Server.URL)
		plist.Spiff.Location = "/api/playlist"
		data, _ := plist.Marshal()
		p = &music.Playlist{User: handler.user.Name, Playlist: data}
		err := m.CreatePlaylist(p)
		if err != nil {
			serverErr(w, err)
			return
		}
	}

	var plist *spiff.Playlist
	dirty := false
	before := p.Playlist

	if r.Method == "PATCH" {
		patch, _ := ioutil.ReadAll(r.Body)
		p.Playlist, err = spiff.Patch(p.Playlist, patch)
		if err != nil {
			serverErr(w, err)
			return
		}
		plist, _ = spiff.Unmarshal(p.Playlist)
		resolver := ref.NewResolver(handler.config, handler)
		resolver.Resolve(handler.user, plist)
		dirty = true
	}

	if dirty {
		if plist.Entries == nil {
			plist.Entries = []spiff.Entry{}
		}
		p.Playlist, _ = plist.Marshal()
		m.UpdatePlaylist(p)

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

// GET /api/progress > []Offset{}
// 200: success
// 500: error
//
// POST /api/progress < Offset{} > no content
// 201: created
// 205: reset content, newer offset exists
// 400: error
// 500: error
//
// DELETE /api/progress/id (see elsewhere)
// 204: accepted no response
// 404: not found
// 500: error
func (handler *UserHandler) apiProgress(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var offset progress.Offset
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("err1 %s\n", err)
			badRequest(w, err)
			return
		}
		err = json.Unmarshal(body, &offset)
		if err != nil {
			log.Printf("err2 %s\n", err)
			badRequest(w, err)
			return
		}
		if len(offset.User) != 0 {
			// post must not have a user
			log.Printf("err3 %s\n", err)
			badRequest(w, err)
			return
		}
		// use authenticated user
		offset.User = handler.user.Name
		if !offset.Valid() {
			log.Printf("err4 %s\n", ErrInvalidOffset)
			badRequest(w, ErrInvalidOffset)
			return
		}
		err = handler.progress().Update(handler.user, offset)
		if err == progress.ErrOffsetTooOld {
			log.Printf("err5 %s\n", err)
			w.WriteHeader(http.StatusResetContent)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case "GET":
		view := handler.progressView()
		handler.apiView(w, r, view)
	default:
		badRequest(w, ErrInvalidMethod)
	}
}

func (handler *UserHandler) apiView(w http.ResponseWriter, r *http.Request, view interface{}) {
	enc := json.NewEncoder(w)
	enc.Encode(view)
}

func (handler *UserHandler) apiSearch(w http.ResponseWriter, r *http.Request) {
	if v := r.URL.Query().Get("q"); v != "" {
		// /api/search?q={pattern}
		view := handler.searchView(strings.TrimSpace(v))
		handler.apiView(w, r, view)
	} else {
		notFoundErr(w)
	}
}

// POST /api/login -> see apiLogin
//
// GET,PATCH /api/playlist -> see apiPlaylist
//
// GET /api/radio > RadioView{}
// GET /api/radio/1 > spiff.Playlist{}
// POST /api/radio
// GET,PATH,DELETE /api/radio/1 >
//
// GET /api/index > IndexView{}
// GET /api/home > HomeView{}
// GET /api/search > SearchView{}
//
// GET /api/tracks/1/location -> Redirect
//
// GET /api/artists > ArtistsView{}
// GET /api/artists/1 > ArtistView{}
// GET /api/artists/1/playlist > spiff.Playlist{}
// GET /api/artists/1/popular > PopularView{}
// GET /api/artists/1/popular/playlist > spiff.Playlist{}
// GET /api/artists/1/singles > SinglesView{}
// GET /api/artists/1/singles/playlist > spiff.Playlist{}
// GET /api/artists/1/radio > spiff.Playlist{}
//
// GET /api/releases/1 > ReleaseView{}
// GET /api/releases/1/playlist > spiff.Playlist{}
//
// POST /api/scrobble <
//
// 200: success
// 500: error
func (handler *UserHandler) apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", ApplicationJson)

	// if r.URL.Path == "/api/live" {
	// 	m := &music.Music{}
	// 	//defer music.Close()
	// 	handler.apiLive(w, r, m)
	// 	return
	// }

	// if r.URL.Path == "/api/watch" {
	// 	v := video.NewVideo(handler.config)
	// 	v.Open()
	// 	movie, _ := v.Movie(11)
	// 	fmt.Printf("%+v\n", movie)
	// 	u := v.MovieURL(movie)
	// 	fmt.Printf("url %s\n", u.String())
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte(u.String()))
	// 	return
	// }

	switch r.URL.Path {
	case "/api/playlist":
		handler.apiPlaylist(w, r)
	case "/api/progress":
		handler.apiProgress(w, r)
	case "/api/radio":
		handler.apiRadio(w, r)
	case "/api/live":
		handler.apiLive(w, r)
	case "/api/index":
		handler.apiView(w, r, handler.indexView())
	case "/api/home":
		handler.apiView(w, r, handler.homeView())
	case "/api/artists":
		handler.apiView(w, r, handler.artistsView())
	case "/api/search":
		handler.apiSearch(w, r)
	case "/api/movies":
		handler.apiView(w, r, handler.moviesView())
	case "/api/podcasts":
		handler.apiView(w, r, handler.podcastsView())
	default:
		// /api/(movies|tracks|episodes)/id/location
		locationRegexp := regexp.MustCompile(`/api/(movies|tracks|episodes)/([0-9]+)/location`)
		matches := locationRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			var url *url.URL
			m := handler.music()
			vid := handler.video()
			p := handler.podcast()
			src := matches[1]
			if src == "tracks" {
				id := str.Atoi(matches[2])
				track, _ := m.LookupTrack(id)
				url = m.TrackURL(&track)
			} else if src == "movies" {
				id := str.Atoi(matches[2])
				movie, _ := vid.LookupMovie(id)
				url = vid.MovieURL(movie)
			} else if src == "episodes" {
				id := str.Atoi(matches[2])
				episode, _ := p.LookupEpisode(id)
				url = p.EpisodeURL(episode)
			}
			// TODO use 307 instead?
			//fmt.Printf("location is %s\n", url.String())
			http.Redirect(w, r, url.String(), http.StatusFound)
			return
		}

		// /api/artists/id/(popular|singles)/playlist
		subPlayistRegexp := regexp.MustCompile(`/api/artists/([0-9]+)/([a-z]+)/playlist(\.xspf)?`)
		matches = subPlayistRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			id := str.Atoi(matches[1])
			m := handler.music()
			artist, _ := m.LookupArtist(id)
			image := m.ArtistImage(&artist)
			sub := matches[2]
			ext := matches[3]
			if ext != "" && ext != ".xspf" {
				notFoundErr(w)
				return
			}
			switch sub {
			case "popular":
				// /api/artists/id/popular/playlist
				handler.apiRefPlaylist(w, r, spiff.TypeMusic,
					artist.Name,
					fmt.Sprintf("%s \u2013 Popular", artist.Name),
					image,
					time.Now(),
					fmt.Sprintf("/music/artists/%d/popular", id))
			case "singles":
				// /api/artists/id/singles/playlist
				handler.apiRefPlaylist(w, r, spiff.TypeMusic,
					artist.Name,
					fmt.Sprintf("%s \u2013 Singles", artist.Name),
					image,
					time.Now(),
					fmt.Sprintf("/music/artists/%d/singles", id))
			default:
				notFoundErr(w)
			}
			return
		}

		// /api/(artists|releases|movies|series|tv)/id/(playlist|popular|radio|singles)
		playlistRegexp := regexp.MustCompile(`/api/([a-z]+)/([0-9]+)/(playlist|popular|singles|radio)(\.xspf)?`)
		matches = playlistRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			v := matches[1]
			id := str.Atoi(matches[2])
			m := handler.music()
			res := matches[3]
			switch v {
			case "artists":
				artist, _ := m.LookupArtist(id)
				image := m.ArtistImage(&artist)
				switch res {
				case "playlist":
					// /api/artists/1/playlist
					handler.apiRefPlaylist(w, r, spiff.TypeMusic,
						artist.Name,
						fmt.Sprintf("%s \u2013 Shuffle", artist.Name),
						image,
						time.Now(),
						fmt.Sprintf("/music/artists/%d/shuffle", id))
				case "radio":
					// /api/artists/1/radio
					handler.apiRefPlaylist(w, r, spiff.TypeMusic,
						"Radio",
						fmt.Sprintf("%s \u2013 Radio", artist.Name),
						image,
						time.Now(),
						fmt.Sprintf("/music/artists/%d/similar", id))
				case "popular":
					// /api/artists/1/popular
					handler.apiView(w, r, handler.popularView(artist))
				case "singles":
					// /api/artists/1/singles
					handler.apiView(w, r, handler.singlesView(artist))
				default:
					notFoundErr(w)
				}
			case "releases":
				// /api/releases/1/playlist
				if res == "playlist" {
					release, _ := m.LookupRelease(id)
					handler.apiRefPlaylist(w, r, spiff.TypeMusic,
						release.Artist,
						release.Name,
						release.Cover("250"),
						release.ReleaseDate,
						fmt.Sprintf("/music/releases/%d/tracks", id))
				} else {
					notFoundErr(w)
				}
			case "movies":
				// /api/movies/1/playlist
				if res == "playlist" {
					movie, _ := handler.video().LookupMovie(id)
					handler.apiRefPlaylist(w, r, spiff.TypeVideo,
						"Movie", // TODO
						movie.Title,
						handler.video().MoviePoster(movie),
						movie.Date,
						fmt.Sprintf("/movies/%d", id))
				} else {
					notFoundErr(w)
				}
			case "series":
				// /api/series/1/playlist
				if res == "playlist" {
					series, _ := handler.podcast().LookupSeries(id)
					handler.apiRefPlaylist(w, r, spiff.TypePodcast,
						series.Author,
						series.Title,
						handler.podcast().SeriesImage(series),
						series.Date,
						fmt.Sprintf("/series/%d", id))
				} else {
					notFoundErr(w)
				}
			case "tv":
				// /api/tv/1/playlist
				if res == "playlist" {
					// tv, _ := handler.video().LookupTV(id)
					// handler.apiRefPlaylist(w, r, spiff.TypeVideo,
					// 	tv.,
					// 	tv.Title,
					// 	handler.video().TVImage(tv),
					// 	tv.Date,
					// 	fmt.Sprintf("/tv/%d", id))
				} else {
					notFoundErr(w)
				}
			default:
				notFoundErr(w)
			}
			return
		}

		// /api/(artists|radio|releases|movies|profiles|progress|series|tv)/id
		resourceRegexp := regexp.MustCompile(`/api/([a-z]+)/([0-9]+)`)
		matches = resourceRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			v := matches[1]
			id := str.Atoi(matches[2])
			switch v {
			case "artists":
				// /api/artists/1
				artist, err := handler.music().LookupArtist(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.artistView(artist))
				}
			case "releases":
				// /api/releases/1
				release, err := handler.music().LookupRelease(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.releaseView(release))
				}
			case "radio":
				// /api/radio/1
				handler.apiStation(w, r, id)
			case "movies":
				// /api/movies/1
				movie, err := handler.video().LookupMovie(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.movieView(movie))
				}
			case "tv":
				// /api/tv/1
				// tv, err := handler.video().LookupTV(id)
			case "profiles":
				// /api/profiles/1
				person, err := handler.video().LookupPerson(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.profileView(person))
				}
			case "series":
				// /api/series/1
				series, err := handler.podcast().LookupSeries(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.seriesView(series))
				}
			case "progress":
				// /api/progress/1
				offset := handler.progress().Offset(handler.user, id)
				if offset == nil {
					notFoundErr(w)
				} else {
					switch r.Method {
					case "GET":
						// GET /api/progress/1
						handler.apiView(w, r, handler.offsetView(*offset))
					case "DELETE":
						// DELETE /api/progress/1
						err := handler.progress().Delete(handler.user, *offset)
						if err != nil {
							serverErr(w, err)
						} else {
							w.WriteHeader(http.StatusNoContent)
						}
					default:
						badRequest(w, ErrInvalidMethod)
					}
				}
			default:
				notFoundErr(w)
			}
			return
		}

		// /api/series/1/episodes/1
		// /api/tv/1/episodes/1
		episodesRegexp := regexp.MustCompile(`/api/([a-z]+)/([0-9]+)/episodes/([0-9]+)`)
		matches = episodesRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			v := matches[1]
			//id := str.Atoi(matches[2])
			episode := str.Atoi(matches[3])
			switch v {
			case "series":
				episode, err := handler.podcast().LookupEpisode(episode)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.seriesEpisodeView(episode))
				}
			case "tv":
			}
			return
		}

		// /api/movies/genres/name
		// allow dash and space in name
		genresRegexp := regexp.MustCompile(`/api/movies/genres/([a-zA-Z -]+)`)
		matches = genresRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			name := strings.TrimSpace(matches[1])
			handler.apiView(w, r, handler.genreView(name))
			return
		}

		// /api/movies/keywords/name
		// allow dash and space in name
		keywordsRegexp := regexp.MustCompile(`/api/movies/keywords/([a-zA-Z -]+)`)
		matches = keywordsRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			name := strings.TrimSpace(matches[1])
			handler.apiView(w, r, handler.keywordView(name))
			return
		}

		notFoundErr(w)
	}
}
