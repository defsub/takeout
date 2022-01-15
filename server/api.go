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

	"github.com/defsub/takeout/lib/encoding/xspf"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/lib/spiff"
	"github.com/defsub/takeout/lib/str"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/ref"
	"github.com/defsub/takeout/video"
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
func (handler *UserHandler) apiLogin(w http.ResponseWriter, r *http.Request) {
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
	s *music.Station, m *music.Music) error {
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
		p := m.LookupPlaylist(handler.user)
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
func (handler *UserHandler) apiRadio(w http.ResponseWriter, r *http.Request, m *music.Music) {
	if r.Method == "GET" {
		view := handler.radioView(m, handler.user)
		enc := json.NewEncoder(w)
		enc.Encode(view)
	} else if r.Method == "POST" {
		var s music.Station
		err := handler.recvStation(w, r, &s, m)
		if err != nil {
			return
		}
		err = m.CreateStation(&s)
		if err != nil {
			serverErr(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		enc.Encode(s)
	} else {
		http.Error(w, "bummer", http.StatusBadRequest)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (handler *UserHandler) apiLive(w http.ResponseWriter, r *http.Request, m *music.Music) {
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
func (handler *UserHandler) apiStation(w http.ResponseWriter, r *http.Request,
	m *music.Music, vid *video.Video, p *podcast.Podcast, id int) {
	s, err := m.LookupStation(id)
	if err != nil {
		notFoundErr(w)
		return
	}
	if !s.Visible(handler.user) {
		notFoundErr(w)
		return
	}

	if r.Method == "GET" {
		resolver := ref.NewResolver(handler.config, m, vid, p, handler)
		resolver.RefreshStation(&s, handler.user)

		w.WriteHeader(http.StatusOK)
		w.Write(s.Playlist)
	} else if r.Method == "PUT" {
		var up music.Station
		err := handler.recvStation(w, r, &up, m)
		if err != nil {
			return
		}
		s.Name = up.Name
		s.Ref = up.Ref
		s.Playlist = up.Playlist
		err = m.UpdateStation(&s)
		if err != nil {
			serverErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else if r.Method == "PATCH" {
		patch, _ := ioutil.ReadAll(r.Body)
		s.Playlist, err = spiff.Patch(s.Playlist, patch)
		if err != nil {
			serverErr(w, err)
			return
		}
		// unmarshal & resovle
		plist, _ := spiff.Unmarshal(s.Playlist)
		resolver := ref.NewResolver(handler.config, m, vid, p, handler)
		resolver.Resolve(handler.user, plist)
		if plist.Entries == nil {
			plist.Entries = []spiff.Entry{}
		}
		// marshal & persist
		s.Playlist, _ = plist.Marshal()
		m.UpdateStation(&s)
		w.WriteHeader(http.StatusNoContent)
	} else if r.Method == "DELETE" {
		err = m.DeleteStation(&s)
		if err != nil {
			serverErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "bummer", http.StatusBadRequest)
	}
}

// GET /api/{res}/id/playlist > spiff.Playlist{}
// 200: success
func (handler *UserHandler) apiRefPlaylist(w http.ResponseWriter, r *http.Request,
	m *music.Music, vid *video.Video, p *podcast.Podcast, creator, title, image, nref string) {
	plist := spiff.NewPlaylist()
	plist.Spiff.Location = fmt.Sprintf("%s%s", handler.config.Server.URL, r.URL.Path)
	plist.Spiff.Creator = creator
	plist.Spiff.Title = title
	plist.Spiff.Image = image
	plist.Entries = []spiff.Entry{{Ref: nref}}
	resolver := ref.NewResolver(handler.config, m, vid, p, handler)
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
func (handler *UserHandler) apiPlaylist(w http.ResponseWriter, r *http.Request,
	m *music.Music, vid *video.Video, pd *podcast.Podcast) {
	p := m.LookupPlaylist(handler.user)
	if p == nil {
		plist := spiff.NewPlaylist()
		plist.Spiff.Location = fmt.Sprintf("%s/api/playlist", handler.config.Server.URL)
		data, _ := plist.Marshal()
		p = &music.Playlist{User: handler.user.Name, Playlist: data}
		err := m.CreatePlaylist(p)
		if err != nil {
			serverErr(w, err)
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
			serverErr(w, err)
			return
		}
		plist, _ = spiff.Unmarshal(p.Playlist)
		resolver := ref.NewResolver(handler.config, m, vid, pd, handler)
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

func (handler *UserHandler) apiView(w http.ResponseWriter, r *http.Request, view interface{}) {
	enc := json.NewEncoder(w)
	enc.Encode(view)
}

func (handler *UserHandler) apiSearch(w http.ResponseWriter, r *http.Request, m *music.Music, vid *video.Video) {
	if v := r.URL.Query().Get("q"); v != "" {
		// /api/search?q={pattern}
		view := handler.searchView(m, vid, strings.TrimSpace(v))
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
func (handler *UserHandler) apiHandler(w http.ResponseWriter, r *http.Request,
	m *music.Music, vid *video.Video, p *podcast.Podcast) {
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
		handler.apiPlaylist(w, r, m, vid, p)
	case "/api/radio":
		handler.apiRadio(w, r, m)
	case "/api/live":
		handler.apiLive(w, r, m)
	case "/api/home":
		handler.apiView(w, r, handler.homeView(m, vid, p))
	case "/api/artists":
		handler.apiView(w, r, handler.artistsView(m))
	case "/api/search":
		handler.apiSearch(w, r, m, vid)
	case "/api/movies":
		handler.apiView(w, r, handler.moviesView(vid))
	case "/api/podcasts":
		handler.apiView(w, r, handler.podcastsView(p))
	default:
		// /api/(movies|tracks|episodes)/id/location
		locationRegexp := regexp.MustCompile(`/api/(movies|tracks|episodes)/([0-9]+)/location`)
		matches := locationRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			var url *url.URL
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
				handler.apiRefPlaylist(w, r, m, vid, p,
					artist.Name,
					fmt.Sprintf("%s \u2013 Popular", artist.Name),
					image,
					fmt.Sprintf("/music/artists/%d/popular", id))
			case "singles":
				// /api/artists/id/singles/playlist
				handler.apiRefPlaylist(w, r, m, vid, p,
					artist.Name,
					fmt.Sprintf("%s \u2013 Singles", artist.Name),
					image,
					fmt.Sprintf("/music/artists/%d/singles", id))
			default:
				notFoundErr(w)
			}
			return
		}

		// /api/(artists|releases|movies|series)/id/(playlist|popular|radio|singles)
		playlistRegexp := regexp.MustCompile(`/api/([a-z]+)/([0-9]+)/(playlist|popular|singles|radio)(\.xspf)?`)
		matches = playlistRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			v := matches[1]
			id := str.Atoi(matches[2])
			res := matches[3]
			switch v {
			case "artists":
				artist, _ := m.LookupArtist(id)
				image := m.ArtistImage(&artist)
				if res == "playlist" {
					// /api/artists/1/playlist
					handler.apiRefPlaylist(w, r, m, vid, p,
						artist.Name,
						fmt.Sprintf("%s \u2013 Shuffle", artist.Name),
						image,
						fmt.Sprintf("/music/artists/%d/shuffle", id))
				} else if res == "radio" {
					// /api/artists/1/radio
					handler.apiRefPlaylist(w, r, m, vid, p,
						"Radio",
						fmt.Sprintf("%s \u2013 Radio", artist.Name),
						image,
						fmt.Sprintf("/music/artists/%d/similar", id))
				} else if res == "popular" {
					// /api/artists/1/popular
					handler.apiView(w, r, handler.popularView(m, artist))
				} else if res == "singles" {
					// /api/artists/1/singles
					handler.apiView(w, r, handler.singlesView(m, artist))
				} else {
					notFoundErr(w)
				}
			case "releases":
				// /api/releases/1/playlist
				if res == "playlist" {
					release, _ := m.LookupRelease(id)
					handler.apiRefPlaylist(w, r, m, vid, p,
						release.Artist,
						release.Name,
						release.Cover("250"),
						fmt.Sprintf("/music/releases/%d/tracks", id))
				} else {
					notFoundErr(w)
				}
			case "movies":
				// /api/movies/1/playlist
				if res == "playlist" {
					movie, _ := vid.LookupMovie(id)
					handler.apiRefPlaylist(w, r, m, vid, p,
						"Movie", // TODO
						movie.Title,
						vid.MoviePoster(*movie),
						fmt.Sprintf("/movies/%d", id))
				} else {
					notFoundErr(w)
				}
			case "series":
				// /api/series/1/playlist
				if res == "playlist" {
					series, _ := p.LookupSeries(id)
					handler.apiRefPlaylist(w, r, m, vid, p,
						"Podcast", // TODO
						series.Title,
						p.SeriesImage(*series),
						fmt.Sprintf("/series/%d", id))
				} else {
					notFoundErr(w)
				}
			default:
				notFoundErr(w)
			}
			return
		}

		// /api/(artists|radio|releases|movies|profiles|series|episodes)/id
		resourceRegexp := regexp.MustCompile(`/api/([a-z]+)/([0-9]+)`)
		matches = resourceRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			v := matches[1]
			id := str.Atoi(matches[2])
			switch v {
			case "artists":
				// /api/artists/1
				artist, err := m.LookupArtist(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.artistView(m, artist))
				}
			case "releases":
				// /api/releases/1
				release, err := m.LookupRelease(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.releaseView(m, release))
				}
			case "radio":
				// /api/radio/1
				handler.apiStation(w, r, m, vid, p, id)
			case "movies":
				// /api/movies/1
				movie, err := vid.LookupMovie(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.movieView(vid, movie))
				}
			case "profiles":
				// /api/profiles/1
				person, err := vid.LookupPerson(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.profileView(vid, person))
				}
			case "series":
				// /api/series/1
				series, err := p.LookupSeries(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.seriesView(p, series))
				}
			case "episodes":
				// /api/episodes/1
				episode, err := p.LookupEpisode(id)
				if err != nil {
					notFoundErr(w)
				} else {
					handler.apiView(w, r, handler.episodeView(p, episode))
				}
			default:
				notFoundErr(w)
			}
			return
		}

		// /api/movies/genres/name
		// allow dash and space in name
		genresRegexp := regexp.MustCompile(`/api/movies/genres/([a-zA-Z -]+)`)
		matches = genresRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			name := strings.TrimSpace(matches[1])
			handler.apiView(w, r, handler.genreView(vid, name))
			return
		}

		// /api/movies/keywords/name
		// allow dash and space in name
		keywordsRegexp := regexp.MustCompile(`/api/movies/keywords/([a-zA-Z -]+)`)
		matches = keywordsRegexp.FindStringSubmatch(r.URL.Path)
		if matches != nil {
			name := strings.TrimSpace(matches[1])
			handler.apiView(w, r, handler.keywordView(vid, name))
			return
		}

		notFoundErr(w)
	}
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

func (UserHandler) LocateTrack(t music.Track) string {
	return locateTrack(t)
}

func (UserHandler) LocateMovie(v video.Movie) string {
	return locateMovie(v)
}

func (UserHandler) LocateEpisode(e podcast.Episode) string {
	return locateEpisode(e)
}
