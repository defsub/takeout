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
	"log"
	"net/http"
	"strings"
	"html/template"
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

	handler.doSpiff(title, tracks, w, r)
}

func (handler *MusicHandler) doSpiff(title string, tracks []Track,
	w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", spiff.ContentType)

	encoder := spiff.NewEncoder(w)
	encoder.Header(title)
	for _, t := range tracks {
		log.Printf("spiff: %s / %s / %s\n", t.Artist, t.Release, t.Title)
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
			handler.doSpiff(fmt.Sprintf("%s / %s", artist[0], release[0]), tracks, w, r)
			return
		}
	}
	http.Error(w, "bummer", http.StatusBadRequest)
}

type MusicHandler struct {
	config *config.Config
}

type index struct {
	Name string
	T Track
	Tracks []Track
}

func (handler *MusicHandler) viewHandler(w http.ResponseWriter, r *http.Request) {
	music := NewMusic(handler.config)
	if music.Open() != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}
	defer music.Close()

	data := &index{Name: "mark"}
	data.Tracks = music.Singles("indie", nil)
	data.T = data.Tracks[0]
	log.Printf("got %d\n", len(data.Tracks))

	var templates = template.Must(template.ParseFiles("templates/index.html"))

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/tracks", handler.doTracks)
	http.HandleFunc("/singles", handler.doSingles)
	http.HandleFunc("/popular", handler.doPopular)
	http.HandleFunc("/release", handler.doRelease)
	http.HandleFunc("/view", handler.viewHandler)
	log.Printf("running...\n")
	http.ListenAndServe(config.BindAddress, nil)
}
