// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// Takeout is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package music

import (
	"fmt"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/encoding/spiff"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"net/http"
	"strconv"
	"strings"
)

type trackType string

const (
	singleTrack  trackType = "singles"
	popularTrack trackType = "popular"
	anyTrack     trackType = "tracks"
)

func (handler *MusicHandler) doit(t trackType, w http.ResponseWriter, r *http.Request) {
	music := NewMusic(handler.config)
	if music.Open() != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}
	defer music.Close()

	var title string
	switch t {
	case popularTrack:
		title = "Popular"
	case singleTrack:
		title = "Singles"
	default:
		title = "Tracks"
	}

	dateRange := &DateRange{}

	var tracks []Track

	artists, ok := r.URL.Query()["artist"]
	if ok {
		a := strings.Join(artists, ",")
		switch t {
		case popularTrack:
			tracks = music.ArtistPopular(a, dateRange)
		case singleTrack:
			tracks = music.ArtistSingles(a, dateRange)
		default:
			tracks = music.ArtistTracks(a, dateRange)
		}
	} else {
		tags, ok := r.URL.Query()["tag"]
		if !ok {
			tags = []string{}
		}
		a := strings.Join(tags, ",")
		switch t {
		case popularTrack:
			tracks = music.Popular(a, dateRange)
		case singleTrack:
			tracks = music.Singles(a, dateRange)
		default:
			tracks = music.Tracks(a, dateRange)
		}
	}

	handler.doSpiff(music, title, tracks, w, r)
}

func (handler *MusicHandler) doSpiff(music *Music, title string, tracks []Track,
	w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", spiff.ContentType)

	encoder := spiff.NewEncoder(w)
	encoder.Header(title)
	for _, t := range tracks {
		log.Printf("spiff: %s / %s / %s\n", t.Artist, t.Release, t.Title)
		t.Location = music.TrackURL(&t).String()
		encoder.Encode(t)
	}
	encoder.Footer()
}

func (handler *MusicHandler) doTracks(w http.ResponseWriter, r *http.Request) {
	handler.doit(anyTrack, w, r)
}

func (handler *MusicHandler) doSingles(w http.ResponseWriter, r *http.Request) {
	handler.doit(singleTrack, w, r)
}

func (handler *MusicHandler) doPopular(w http.ResponseWriter, r *http.Request) {
	handler.doit(popularTrack, w, r)
}

func (handler *MusicHandler) doRelease(w http.ResponseWriter, r *http.Request) {
	music := NewMusic(handler.config)
	if music.Open() != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}
	defer music.Close()
	artist, ok := r.URL.Query()["artist"]
	if ok {
		release, ok := r.URL.Query()["name"]
		if !ok {
			release, ok = r.URL.Query()["release"]
		}
		if ok {
			tracks := music.ArtistRelease(artist[0], release[0])
			handler.doSpiff(music, fmt.Sprintf("%s / %s", artist[0], release[0]), tracks, w, r)
			return
		}
	}
	http.Error(w, "bummer", http.StatusBadRequest)
}

type MusicHandler struct {
	config *config.Config
}

func cover(r Release, s string) string {
	if r.REID != "" {
		return fmt.Sprintf("https://coverartarchive.org/release/%s/%s", r.REID, s)
	} else {
		return fmt.Sprintf("https://coverartarchive.org/release-group/%s/%s", r.RGID, s)
	}
}

func trackCover(music *Music, t Track, s string) string {
	artist := music.artist(t.Artist)
	release := music.artistRelease(artist, t.Release)
	if release == nil {
		return ""
	}
	return cover(*release, s)
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

	if err != nil {
		fmt.Printf("%s\n", err)
	}

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
				link = music.TrackURL(&t).String()
			}
			return link
		},
		"coverSmall": func(o interface{}) string {
			switch o.(type) {
			case Release:
				return cover(o.(Release), "front-250")
			case Track:
				return trackCover(music, o.(Track), "front-250")
			}
			return ""
		},
		"coverLarge": func(r Release) string {
			return cover(r, "front-500")
		},
		"coverExtraLarge": func(r Release) string {
			return cover(r, "front-1200")
		},
		"letter": func(a Artist) string {
			return a.SortName[0:1]
		},

	}

	var templates = parseTemplates(template.New("").Funcs(funcMap), "web/template")

	err := templates.ExecuteTemplate(w, temp, view)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *MusicHandler) viewHandler(w http.ResponseWriter, r *http.Request) {
	music := NewMusic(handler.config)
	if music.Open() != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}
	defer music.Close()

	var view interface{}
	var temp string

	if v := r.URL.Query().Get("release"); v != "" {
		id, _ := strconv.Atoi(v)
		release, _ := music.lookupRelease(uint(id))
		view = music.ReleaseView(release)
		temp = "release.html"
	} else if v := r.URL.Query().Get("artist"); v != "" {
		id, _ := strconv.Atoi(v)
		artist, _ := music.lookupArtist(uint(id))
		view = music.ArtistView(artist)
		temp = "artist.html"
	} else if v := r.URL.Query().Get("music"); v != "" {
		view = music.HomeView()
		temp = "music.html"
	} else if v := r.URL.Query().Get("q"); v != "" {
		view = music.SearchView(v)
		temp = "search.html"
	} else {
		temp = "index.html"
	}

	handler.render(music, temp, view, w, r)
}

// /artists/
// /artists/name
// /artists/name/releases
// /artists/name/releases/name
// /artists/name/tracks/name
// /artists/name/singles
// /artists/name/popular
// /tracks/name
//
// play
// playlist
// artists
// artist, album

func Serve(config *config.Config) {
	handler := &MusicHandler{config: config}
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/tracks", handler.doTracks)
	http.HandleFunc("/singles", handler.doSingles)
	http.HandleFunc("/popular", handler.doPopular)
	//http.HandleFunc("/release", handler.doRelease)
	http.HandleFunc("/", handler.viewHandler)
	http.HandleFunc("/v", handler.viewHandler)
	log.Printf("running...\n")
	http.ListenAndServe(config.BindAddress, nil)
}
