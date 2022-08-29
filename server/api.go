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
	"regexp"
	"strings"
	"time"

	"github.com/defsub/takeout/activity"
	"github.com/defsub/takeout/lib/date"
	"github.com/defsub/takeout/lib/encoding/xspf"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/lib/spiff"
	"github.com/defsub/takeout/lib/str"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/progress"
	"github.com/defsub/takeout/ref"
	"github.com/defsub/takeout/view"
)

const (
	ApplicationJson = "application/json"
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

func apiLogin(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	w.Header().Set("Content-type", ApplicationJson)

	var l login
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &l)
	if err != nil {
		serverErr(w, err)
		return
	}

	var result status
	cookie, err := doLogin(ctx, l.User, l.Pass)
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

var locationRegexp = regexp.MustCompile(`/api/(tracks)/([0-9]+)/location`)

func writePlaylist(w http.ResponseWriter, r *http.Request, plist *spiff.Playlist) {
	if strings.HasSuffix(r.URL.Path, ".xspf") {
		// create XML spiff with tracks fully resolved
		ctx := contextValue(r)
		w.Header().Set("Content-type", xspf.XMLContentType)
		encoder := xspf.NewXMLEncoder(w)
		encoder.Header(plist.Spiff.Title)
		for i := range plist.Spiff.Entries {
			matches := locationRegexp.FindStringSubmatch(plist.Spiff.Entries[i].Location[0])
			if matches != nil {
				src := matches[1]
				if src == "tracks" {
					m := ctx.Music()
					id := str.Atoi(matches[2])
					track, err := m.LookupTrack(id)
					if err != nil {
						continue
					}
					url := m.TrackURL(&track)
					plist.Spiff.Entries[i].Location = []string{url.String()}
				}
			}
			encoder.Encode(plist.Spiff.Entries[i])
		}
		encoder.Footer()

	} else {
		// use json spiff with track location
		w.Header().Set("Content-type", ApplicationJson)
		result, _ := plist.Marshal()
		w.Write(result)
	}
}

// TODO check
func recvStation(w http.ResponseWriter, r *http.Request,
	s *music.Station) error {
	ctx := contextValue(r)
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
	s.User = ctx.User().Name
	if s.Ref == "/api/playlist" {
		// copy playlist
		p := ctx.Music().LookupPlaylist(ctx.User())
		if p != nil {
			s.Playlist = p.Playlist
		}
	}
	return nil
}

func makeEmptyPlaylist(w http.ResponseWriter, r *http.Request) (*music.Playlist, error) {
	ctx := contextValue(r)
	plist := spiff.NewPlaylist(spiff.TypeMusic)
	plist.Spiff.Location = r.URL.Path
	data, _ := plist.Marshal()
	p := music.Playlist{User: ctx.User().Name, Playlist: data}
	err := ctx.Music().CreatePlaylist(&p)
	return &p, err
}

func apiPlaylistGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	p := ctx.Music().LookupPlaylist(ctx.User())
	if p == nil {
		var err error
		p, err = makeEmptyPlaylist(w, r)
		if err != nil {
			serverErr(w, err)
			return
		}
	}
	w.Header().Set("Content-type", ApplicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write(p.Playlist)
}

func apiPlaylistPatch(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := contextValue(r)
	user := ctx.User()
	m := ctx.Music()
	p := m.LookupPlaylist(user)
	if p == nil {
		var err error
		p, err = makeEmptyPlaylist(w, r)
		if err != nil {
			serverErr(w, err)
			return
		}
	}

	before := p.Playlist

	// apply patch
	patch, _ := ioutil.ReadAll(r.Body)
	p.Playlist, err = spiff.Patch(p.Playlist, patch)
	if err != nil {
		serverErr(w, err)
		return
	}
	plist, _ := spiff.Unmarshal(p.Playlist)
	ref.Resolve(ctx, plist)

	// save result
	if plist.Spiff.Entries == nil {
		plist.Spiff.Entries = []spiff.Entry{}
	}
	p.Playlist, _ = plist.Marshal()
	m.UpdatePlaylist(p)

	v, _ := spiff.Compare(before, p.Playlist)
	if v {
		// entries didn't change, only metadata
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.Header().Set("Content-type", ApplicationJson)
		w.WriteHeader(http.StatusOK)
		w.Write(p.Playlist)
	}
}

func apiProgressGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	view := view.ProgressView(ctx)
	apiView(w, r, view)
}

func apiProgressPost(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	user := ctx.User()
	var offsets progress.Offsets
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(w, err)
		return
	}
	err = json.Unmarshal(body, &offsets)
	if err != nil {
		badRequest(w, err)
		return
	}
	for i := range offsets.Offsets {
		// will update array inplace
		o := &offsets.Offsets[i]
		if len(o.User) != 0 {
			// post must not have a user
			badRequest(w, err)
			return
		}
		// use authenticated user
		o.User = user.Name
		if !o.Valid() {
			badRequest(w, ErrInvalidOffset)
			return
		}
	}
	for _, o := range offsets.Offsets {
		// update each offset as needed
		log.Printf("update progress %s %d/%d\n", o.ETag, o.Offset, o.Duration)
		err = ctx.Progress().Update(user, o)
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiView(w http.ResponseWriter, r *http.Request, view interface{}) {
	w.Header().Set("Content-type", ApplicationJson)
	json.NewEncoder(w).Encode(view)
}

func apiHome(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	view := view.HomeView(ctx)
	apiView(w, r, view)
}

func apiIndex(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	view := view.IndexView(ctx)
	apiView(w, r, view)
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	if v := r.URL.Query().Get("q"); v != "" {
		// /api/search?q={pattern}
		view := view.SearchView(ctx, strings.TrimSpace(v))
		apiView(w, r, view)
	} else {
		notFoundErr(w)
	}
}

func apiArtists(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	apiView(w, r, view.ArtistsView(ctx))
}

func apiArtistGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	artist, err := ctx.FindArtist(id)
	if err != nil {
		notFoundErr(w)
	} else {
		apiView(w, r, view.ArtistView(ctx, artist))
	}
}

func apiArtistGetResource(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	res := r.URL.Query().Get(":res")
	artist, err := ctx.FindArtist(id)
	if err != nil {
		notFoundErr(w)
	} else {
		switch res {
		case "popular":
			apiView(w, r, view.PopularView(ctx, artist))
		case "singles":
			apiView(w, r, view.SinglesView(ctx, artist))
		case "playlist":
			apiArtistGetPlaylist(w, r)
		default:
			notFoundErr(w)
		}
	}
}

func apiArtistGetPlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	res := r.URL.Query().Get(":res")
	artist, err := ctx.FindArtist(id)
	if err != nil {
		notFoundErr(w)
	} else {
		// /api/artists/:id/:res/playlist -> /music/artists/:id/:res
		nref := fmt.Sprintf("/music/artists/%s/%s", id, res)
		plist := ref.ResolveArtistPlaylist(ctx,
			view.ArtistView(ctx, artist), r.URL.Path, nref)
		writePlaylist(w, r, plist)
	}
}

func apiRadioGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	apiView(w, r, view.RadioView(ctx))
}

func apiRadioPost(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	var s music.Station
	err := recvStation(w, r, &s)
	if err != nil {
		return
	}
	err = ctx.Music().CreateStation(&s)
	if err != nil {
		serverErr(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	enc.Encode(s)
}

func apiRadioStationGetPlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	station, err := ctx.FindStation(id)
	if err != nil {
		notFoundErr(w)
		return
	}
	if !station.Visible(ctx.User()) {
		notFoundErr(w)
		return
	}
	plist := ref.RefreshStation(ctx, &station)
	writePlaylist(w, r, plist)
}

func apiMovies(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	apiView(w, r, view.MoviesView(ctx))
}

func apiMovieGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	movie, err := ctx.FindMovie(id)
	if err != nil {
		notFoundErr(w)
	} else {
		apiView(w, r, view.MovieView(ctx, movie))
	}
}

func apiMovieGetPlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	movie, err := ctx.FindMovie(id)
	if err != nil {
		notFoundErr(w)
	} else {
		view := view.MovieView(ctx, movie)
		plist := ref.ResolveMoviePlaylist(ctx, view, r.URL.Path)
		writePlaylist(w, r, plist)
	}
}

func apiMovieProfileGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	person, err := ctx.Video().LookupPerson(str.Atoi(id))
	if err != nil {
		notFoundErr(w)
	} else {
		apiView(w, r, view.ProfileView(ctx, person))
	}
}

func apiMovieGenreGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	name := r.URL.Query().Get(":name")
	// TODO sanitize
	apiView(w, r, view.GenreView(ctx, name))
}

func apiMovieKeywordGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	name := r.URL.Query().Get(":name")
	// TODO sanitize
	apiView(w, r, view.KeywordView(ctx, name))
}

func apiPodcasts(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	apiView(w, r, view.PodcastsView(ctx))
}

func apiPodcastSeriesGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	series, err := ctx.Podcast().FindSeries(id)
	if err != nil {
		notFoundErr(w)
	} else {
		apiView(w, r, view.SeriesView(ctx, series))
	}
}

func apiPodcastSeriesGetPlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	series, err := ctx.Podcast().FindSeries(id)
	if err != nil {
		notFoundErr(w)
	} else {
		view := view.SeriesView(ctx, series)
		plist := ref.ResolveSeriesPlaylist(ctx, view, r.URL.Path)
		writePlaylist(w, r, plist)
	}
}

func apiPodcastSeriesEpisodeGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	if id != "" {
		// series is optional for now
		_, err := ctx.Podcast().FindSeries(id)
		if err != nil {
			notFoundErr(w)
			return
		}
	}
	eid := r.URL.Query().Get(":eid")
	episode, err := ctx.Podcast().FindEpisode(eid)
	if err != nil {
		notFoundErr(w)
	} else {
		apiView(w, r, view.SeriesEpisodeView(ctx, episode))
	}
}

func apiPodcastSeriesEpisodeGetPlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	eid := r.URL.Query().Get(":eid")
	series, err := ctx.Podcast().FindSeries(id)
	if err != nil {
		notFoundErr(w)
		return
	}
	episode, err := ctx.Podcast().FindEpisode(eid)
	if err != nil {
		notFoundErr(w)
	} else {
		plist := ref.ResolveSeriesEpisodePlaylist(ctx,
			view.SeriesView(ctx, series),
			view.SeriesEpisodeView(ctx, episode),
			r.URL.Path)
		writePlaylist(w, r, plist)
	}
}

// TODO check
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
func apiStation(w http.ResponseWriter, r *http.Request, id int) {
	ctx := contextValue(r)
	s, err := ctx.Music().LookupStation(id)
	if err != nil {
		notFoundErr(w)
		return
	}
	if !s.Visible(ctx.User()) {
		notFoundErr(w)
		return
	}

	switch r.Method {
	case "GET":
		ref.RefreshStation(ctx, &s)
		w.WriteHeader(http.StatusOK)
		w.Write(s.Playlist)
	case "PUT":
		var up music.Station
		err := recvStation(w, r, &up)
		if err != nil {
			return
		}
		s.Name = up.Name
		s.Ref = up.Ref
		s.Playlist = up.Playlist
		err = ctx.Music().UpdateStation(&s)
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
		ref.Resolve(ctx, plist)
		if plist.Spiff.Entries == nil {
			plist.Spiff.Entries = []spiff.Entry{}
		}
		// marshal & persist
		s.Playlist, _ = plist.Marshal()
		ctx.Music().UpdateStation(&s)
		w.WriteHeader(http.StatusNoContent)
	case "DELETE":
		err = ctx.Music().DeleteStation(&s)
		if err != nil {
			serverErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "bummer", http.StatusBadRequest)
	}
}

func apiReleaseGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	release, err := ctx.FindRelease(id)
	if err != nil {
		notFoundErr(w)
	} else {
		apiView(w, r, view.ReleaseView(ctx, release))
	}
}

func apiReleaseGetPlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	release, err := ctx.FindRelease(id)
	if err != nil {
		notFoundErr(w)
	} else {
		view := view.ReleaseView(ctx, release)
		plist := ref.ResolveReleasePlaylist(ctx, view, r.URL.Path)
		writePlaylist(w, r, plist)
	}
}

func apiTrackLocation(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	track, err := ctx.FindTrack(id)
	if err != nil {
		notFoundErr(w)
	} else {
		url := ctx.Music().TrackURL(&track)
		http.Redirect(w, r, url.String(), http.StatusFound)
	}
}

func apiMovieLocation(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	id := r.URL.Query().Get(":id")
	movie, err := ctx.FindMovie(id)
	if err != nil {
		notFoundErr(w)
	} else {
		url := ctx.Video().MovieURL(movie)
		http.Redirect(w, r, url.String(), http.StatusFound)
	}
}

func apiSeriesEpisodeLocation(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	//id := r.URL.Query().Get(":id")
	eid := r.URL.Query().Get(":eid")
	// series, err := ctx.Podcast().FindSeries(id)
	// if err != nil {
	// 	notFoundErr(w)
	// 	return
	// }
	episode, err := ctx.Podcast().FindEpisode(eid)
	if err != nil {
		notFoundErr(w)
	} else {
		url := ctx.Podcast().EpisodeURL(episode)
		http.Redirect(w, r, url.String(), http.StatusFound)
	}
}

func apiActivityGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	apiView(w, r, view.ActivityView(ctx))
}

func apiActivityPost(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)

	var events activity.Events
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(w, err)
		return
	}
	err = json.Unmarshal(body, &events)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = ctx.Activity().CreateEvents(events, ctx.User(), ctx.Music(), ctx.Video())
	if err != nil {
		serverErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func startEnd(r *http.Request) (time.Time, time.Time) {
	// now until 1 year back, limits will apply
	end := time.Now()
	start := end.AddDate(-1, 0, 0)

	s := r.URL.Query().Get("start")
	if s != "" {
		start = date.ParseDate(s)
	}
	e := r.URL.Query().Get("end")
	if e != "" {
		end = date.ParseDate(e)
	}

	return date.StartOfDay(start), date.EndOfDay(end)
}

func apiActivityTracksGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	start, end := startEnd(r)
	apiView(w, r, view.ActivityTracksView(ctx, start, end))
}

func apiActivityTracksGetResource(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	start, end := startEnd(r)
	res := r.URL.Query().Get(":res")

	switch res {
	case "popular":
		apiView(w, r, view.ActivityPopularTracksView(ctx, start, end))
	case "recent":
		apiView(w, r, view.ActivityTracksView(ctx, start, end))
	case "playlist":
		apiActivityTracksGetPlaylist(w, r)
	default:
		notFoundErr(w)
	}
}

func apiActivityTracksGetPlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	start, end := startEnd(r)
	res := r.URL.Query().Get(":res")
	if res == "playlist" {
		res = "recent"
	}

	var tracks *view.ActivityTracks
	switch res {
	case "popular":
		tracks = view.ActivityPopularTracksView(ctx, start, end)
	case "recent":
		tracks = view.ActivityTracksView(ctx, start, end)
	default:
		notFoundErr(w)
	}

	plist := ref.ResolveActivityTracksPlaylist(ctx, tracks, res, r.URL.Path)
	writePlaylist(w, r, plist)
}

func apiActivityMoviesGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	start, end := startEnd(r)
	apiView(w, r, view.ActivityMoviesView(ctx, start, end))
}

func apiActivityReleasesGet(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	start, end := startEnd(r)
	apiView(w, r, view.ActivityReleasesView(ctx, start, end))
}
