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

package ref

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/defsub/takeout/lib/client"
	"github.com/defsub/takeout/lib/date"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/lib/spiff"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
	"github.com/defsub/takeout/view"
)

type Locator interface {
	LocateTrack(music.Track) string
	LocateMovie(video.Movie) string
	LocateEpisode(podcast.Episode) string

	FindArtist(string) (music.Artist, error)
	FindRelease(string) (music.Release, error)
	FindTrack(string) (music.Track, error)
	FindStation(string) (music.Station, error)
	FindMovie(string) (video.Movie, error)
	FindSeries(string) (podcast.Series, error)

	TrackImage(music.Track) string
	MovieImage(video.Movie) string
	EpisodeImage(podcast.Episode) string
}

type Context interface {
	view.Context
	Locator
}

func trackEntry(ctx Context, t music.Track) spiff.Entry {
	return spiff.Entry{
		Creator:    t.PreferredArtist(),
		Album:      t.ReleaseTitle,
		Title:      t.Title,
		Image:      ctx.TrackImage(t),
		Location:   []string{ctx.LocateTrack(t)},
		Identifier: []string{t.ETag},
		Size:       []int64{t.Size},
		Date:       date.FormatJson(t.ReleaseDate),
	}
}

func movieEntry(ctx Context, m video.Movie) spiff.Entry {
	return spiff.Entry{
		Creator:    "Movie", // TODO need better creator
		Album:      m.Title,
		Title:      m.Title,
		Image:      ctx.MovieImage(m),
		Location:   []string{ctx.LocateMovie(m)},
		Identifier: []string{m.ETag},
		Size:       []int64{m.Size},
		Date:       date.FormatJson(m.Date),
	}
}

func episodeEntry(ctx Context, series podcast.Series, e podcast.Episode) spiff.Entry {
	author := e.Author
	if author == "" {
		author = series.Author
	}
	return spiff.Entry{
		Creator:    author,
		Album:      series.Title,
		Title:      e.Title,
		Image:      ctx.EpisodeImage(e),
		Location:   []string{ctx.LocateEpisode(e)},
		Identifier: []string{e.EID},
		Size:       []int64{e.Size},
		Date:       date.FormatJson(e.Date),
	}
}

func addTrackEntries(ctx Context, tracks []music.Track, entries []spiff.Entry) []spiff.Entry {
	for _, t := range tracks {
		entries = append(entries, trackEntry(ctx, t))
	}
	return entries
}

func addMovieEntries(ctx Context, movies []video.Movie, entries []spiff.Entry) []spiff.Entry {
	for _, m := range movies {
		entries = append(entries, movieEntry(ctx, m))
	}
	return entries
}

func addEpisodeEntries(ctx Context, series podcast.Series, episodes []podcast.Episode,
	entries []spiff.Entry) []spiff.Entry {
	for _, e := range episodes {
		entries = append(entries, episodeEntry(ctx, series, e))
	}
	return entries
}

// /music/artists/{id}/{res}
func resolveArtistRef(ctx Context, id, res string, entries []spiff.Entry) ([]spiff.Entry, error) {
	artist, err := ctx.FindArtist(id)
	if err != nil {
		return entries, err
	}
	v := view.ArtistView(ctx, artist)
	tracks := resolveArtistTrackList(v, res)
	entries = addTrackEntries(ctx, tracks.Tracks(), entries)
	return entries, nil
}

func resolveArtistTrackList(v *view.Artist, res string) view.TrackList {
	var tracks view.TrackList
	switch res {
	case "deep":
		tracks = v.Deep
	case "popular":
		tracks = v.Popular
	case "radio", "similar":
		tracks = v.Radio
	case "shuffle", "playlist":
		tracks = v.Shuffle
	case "singles":
		tracks = v.Singles
	case "tracks":
		tracks = v.Tracks
	}
	return tracks
}

// /music/releases/{id}/tracks
func resolveReleaseRef(ctx Context, id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	release, err := ctx.FindRelease(id)
	if err != nil {
		return entries, err
	}
	rv := view.ReleaseView(ctx, release)
	tracks := rv.Tracks
	entries = addTrackEntries(ctx, tracks, entries)
	return entries, nil
}

// /music/tracks/{id}
func resolveTrackRef(ctx Context, id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	t, err := ctx.FindTrack(id)
	if err != nil {
		return entries, err
	}
	entries = addTrackEntries(ctx, []music.Track{t}, entries)
	return entries, nil
}

// /movies/{id}
func resolveMovieRef(ctx Context, id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	m, err := ctx.FindMovie(id)
	if err != nil {
		return entries, err
	}
	entries = addMovieEntries(ctx, []video.Movie{m}, entries)
	return entries, nil
}

// /series/{id}
func resolveSeriesRef(ctx Context, id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	series, err := ctx.FindSeries(id)
	if err != nil {
		return entries, err
	}
	pv := view.SeriesView(ctx, series)
	episodes := pv.Episodes
	if err != nil {
		return entries, err
	}
	entries = addEpisodeEntries(ctx, series, episodes, entries)
	return entries, nil
}

// /music/search?q={q}[&radio=1]
func resolveSearchRef(ctx Context, uri string, entries []spiff.Entry) ([]spiff.Entry, error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Println(err)
		return entries, err
	}

	q := u.Query().Get("q")
	radio := u.Query().Get("radio") != ""

	var tracks []music.Track
	if q != "" {
		limit := ctx.Config().Music.SearchLimit
		if radio {
			limit = ctx.Config().Music.RadioSearchLimit
		}
		tracks = ctx.Music().Search(q, limit)
	}

	if radio {
		tracks = music.Shuffle(tracks)
		limit := ctx.Config().Music.RadioLimit
		if len(tracks) > limit {
			tracks = tracks[:limit]
		}
	}

	entries = addTrackEntries(ctx, tracks, entries)
	return entries, nil
}

// /music/radio/{id}
func resolveRadioRef(ctx Context, id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	s, err := ctx.FindStation(id)
	if err != nil {
		return entries, err
	}
	if !s.Visible(ctx.User()) {
		return entries, err
	}

	// rerun the station ref to get new tracks
	plist := RefreshStation(ctx, &s)

	entries = append(entries, plist.Spiff.Entries...)

	return entries, nil
}

func resolvePlsRef(ctx Context, url, creator, image string, entries []spiff.Entry) ([]spiff.Entry, error) {
	client := client.NewClient(ctx.Config())
	result, err := client.GetPLS(url)
	if err != nil {
		return entries, err
	}

	for _, v := range result.Entries {
		entry := spiff.Entry{
			Creator:    creator,
			Album:      v.Title,
			Title:      v.Title,
			Image:      image,
			Location:   []string{v.File},
			Identifier: []string{},
			Size:       []int64{int64(v.Length)},
			Date:       date.FormatJson(time.Now()),
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// /activity/tracks
func resolveActivityTracksRef(ctx Context, entries []spiff.Entry) ([]spiff.Entry, error) {
	v := view.ActivityView(ctx)
	var tracks []music.Track
	// TODO only recent is supported for now
	for _, t := range v.RecentTracks {
		tracks = append(tracks, t.Track)
	}
	entries = addTrackEntries(ctx, tracks, entries)
	return entries, nil
}

// /activity/movies
func resolveActivityMoviesRef(ctx Context, entries []spiff.Entry) ([]spiff.Entry, error) {
	v := view.ActivityView(ctx)
	var movies []video.Movie
	// TODO only recent is supported for now
	for _, m := range v.RecentMovies {
		movies = append(movies, m.Movie)
	}
	entries = addMovieEntries(ctx, movies, entries)
	return entries, nil
}

func RefreshStation(ctx Context, s *music.Station) *spiff.Playlist {
	plist := spiff.NewPlaylist(spiff.TypeMusic)
	plist.Spiff.Location = fmt.Sprintf("/api/radio/stations/%d", s.ID)
	plist.Spiff.Title = s.Name
	plist.Spiff.Image = s.Image
	plist.Spiff.Creator = s.Creator
	plist.Spiff.Date = date.FormatJson(time.Now())

	if s.Type == music.TypeStream {
		// internet radio streams
		plist.Type = spiff.TypeStream
		if strings.HasSuffix(s.Ref, ".pls") {
			var entries []spiff.Entry
			entries, err := resolvePlsRef(ctx, s.Ref, s.Creator, s.Image, entries)
			if err != nil {
				log.Printf("pls error %s\n", err)
				return nil
			}
			plist.Spiff.Entries = entries
		} else {
			// TODO add m3u, others?
			log.Printf("unsupported stream %s\n", s.Ref)
		}
	} else {
		plist.Spiff.Entries = []spiff.Entry{{Ref: s.Ref}}
		Resolve(ctx, plist)
		if plist.Spiff.Entries == nil {
			plist.Spiff.Entries = []spiff.Entry{}
		}
	}

	// TODO not saved for now
	//s.Playlist, _ = plist.Marshal()
	//m.UpdateStation(s)

	return plist
}

var (
	artistsRegexp      = regexp.MustCompile(`^/music/artists/([0-9a-zA-Z-]+)/([\w]+)$`)
	releasesRegexp     = regexp.MustCompile(`^/music/releases/([0-9a-zA-Z-]+)/tracks$`)
	tracksRegexp       = regexp.MustCompile(`^/music/tracks/([\d]+)$`)
	searchRegexp       = regexp.MustCompile(`^/music/search.*`)
	radioRegexp        = regexp.MustCompile(`^/music/radio/stations/([\d]+)$`)
	moviesRegexp       = regexp.MustCompile(`^/movies/([\d]+)$`)
	seriesRegexp       = regexp.MustCompile(`^/podcasts/series/([\d]+)$`)
	recentTracksRegexp = regexp.MustCompile(`^/activity/tracks$`)
	recentMoviesRegexp = regexp.MustCompile(`^/activity/movies$`)
)

func Resolve(ctx Context, plist *spiff.Playlist) (err error) {
	var entries []spiff.Entry

	for _, e := range plist.Spiff.Entries {
		if e.Ref == "" {
			entries = append(entries, e)
			continue
		}

		pathRef := e.Ref

		matches := artistsRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = resolveArtistRef(ctx, matches[1], matches[2], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = releasesRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = resolveReleaseRef(ctx, matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = tracksRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = resolveTrackRef(ctx, matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		if searchRegexp.MatchString(pathRef) {
			entries, err = resolveSearchRef(ctx, pathRef, entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = radioRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = resolveRadioRef(ctx, matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = moviesRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = resolveMovieRef(ctx, matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = seriesRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = resolveSeriesRef(ctx, matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = recentTracksRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = resolveActivityTracksRef(ctx, entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = recentMoviesRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = resolveActivityMoviesRef(ctx, entries)
			if err != nil {
				return err
			}
			continue
		}
	}

	plist.Spiff.Entries = entries

	return nil
}

func ResolveArtistPlaylist(ctx Context, v *view.Artist, path, nref string) *spiff.Playlist {
	// /music/artists/{id}/{resource}
	parts := strings.Split(nref, "/")
	res := parts[4]
	trackList := resolveArtistTrackList(v, res)

	plist := spiff.NewPlaylist(spiff.TypeMusic)
	plist.Spiff.Location = path
	plist.Spiff.Creator = v.Artist.Name
	plist.Spiff.Title = trackList.Title
	plist.Spiff.Image = v.Image
	plist.Spiff.Date = date.FormatJson(time.Now())
	if trackList.Tracks != nil {
		plist.Spiff.Entries = addTrackEntries(ctx, trackList.Tracks(), plist.Spiff.Entries)
	}
	return plist
}

func ResolveReleasePlaylist(ctx Context, v *view.Release, path string) *spiff.Playlist {
	// /music/release/{id}
	plist := spiff.NewPlaylist(spiff.TypeMusic)
	plist.Spiff.Location = path
	plist.Spiff.Creator = v.Release.Artist
	plist.Spiff.Title = v.Release.Name
	plist.Spiff.Image = v.Image
	plist.Spiff.Date = date.FormatJson(v.Release.Date)
	plist.Spiff.Entries = addTrackEntries(ctx, v.Tracks, plist.Spiff.Entries)
	return plist
}

func ResolveMoviePlaylist(ctx Context, v *view.Movie, path string) *spiff.Playlist {
	// /movies/{id}
	plist := spiff.NewPlaylist(spiff.TypeVideo)
	plist.Spiff.Location = path
	plist.Spiff.Creator = "Movie"
	plist.Spiff.Title = v.Movie.Title
	plist.Spiff.Image = ctx.MovieImage(v.Movie)
	plist.Spiff.Date = date.FormatJson(v.Movie.Date)
	plist.Spiff.Entries = []spiff.Entry{
		movieEntry(ctx, v.Movie),
	}
	return plist
}

func ResolveSeriesPlaylist(ctx Context, v *view.Series, path string) *spiff.Playlist {
	// /podcasts/series/{id}
	plist := spiff.NewPlaylist(spiff.TypePodcast)
	plist.Spiff.Location = path
	plist.Spiff.Creator = v.Series.Author
	plist.Spiff.Title = v.Series.Title
	plist.Spiff.Image = v.Series.Image
	plist.Spiff.Date = date.FormatJson(v.Series.Date)
	plist.Spiff.Entries = addEpisodeEntries(ctx, v.Series, v.Episodes, plist.Spiff.Entries)
	return plist
}

func ResolveSeriesEpisodePlaylist(ctx Context, series *view.Series,
	v *view.SeriesEpisode, path string) *spiff.Playlist {
	// /podcasts/series/{id}/episode/{eid}
	plist := spiff.NewPlaylist(spiff.TypePodcast)
	plist.Spiff.Location = path
	plist.Spiff.Creator = series.Series.Author
	plist.Spiff.Title = v.Episode.Title
	plist.Spiff.Image = v.EpisodeImage(v.Episode)
	plist.Spiff.Date = date.FormatJson(v.Episode.Date)
	plist.Spiff.Entries = []spiff.Entry{
		episodeEntry(ctx, series.Series, v.Episode),
	}
	return plist
}

func ResolveActivityTracksPlaylist(ctx Context, v *view.ActivityTracks, res, path string) *spiff.Playlist {
	var tracks []music.Track
	artistMap := make(map[string]bool)
	for _, t := range v.Tracks {
		artistMap[t.Track.Artist] = true
		tracks = append(tracks, t.Track)
	}
	var artists []string
	for k := range artistMap {
		artists = append(artists, k)
	}
	sort.Slice(artists, func(i, j int) bool {
		return artists[i] < artists[j]
	})
	creators := strings.Join(artists, " \u2022 ")
	image := ""
	for _, t := range tracks {
		img := ctx.TrackImage(t)
		if img != "" {
			image = img
			break
		}
	}

	title := ""
	switch res {
	case "popular":
		title = ctx.Config().Activity.PopularTracksTitle
	case "recent":
		title = ctx.Config().Activity.RecentTracksTitle
	}

	plist := spiff.NewPlaylist(spiff.TypeMusic)
	plist.Spiff.Location = path
	plist.Spiff.Creator = creators
	plist.Spiff.Title = title
	plist.Spiff.Image = image
	plist.Spiff.Date = date.FormatJson(time.Now())
	plist.Spiff.Entries = addTrackEntries(ctx, tracks, plist.Spiff.Entries)
	return plist
}
