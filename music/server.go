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
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/encoding/spiff"
	"log"
	"net/http"
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

type MusicHandler struct {
	config *config.Config
}

func Serve(config *config.Config) {
	handler := &MusicHandler{config: config}
	http.HandleFunc("/tracks", handler.doTracks)
	http.HandleFunc("/singles", handler.doSingles)
	http.HandleFunc("/popular", handler.doPopular)
	log.Printf("running...\n")
	http.ListenAndServe(config.BindAddress, nil)
}
