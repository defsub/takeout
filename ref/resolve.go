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
	"strconv"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/hash"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/lib/spiff"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
)

type Locator interface {
	LocateTrack(t music.Track) string
	LocateMovie(m video.Movie) string
	LocateEpisode(e podcast.Episode) string
}

type Resolver struct {
	config  *config.Config
	music   *music.Music
	video   *video.Video
	podcast *podcast.Podcast
	loc     Locator
}

func NewResolver(c *config.Config,
	m *music.Music, v *video.Video, p *podcast.Podcast,
	l Locator) *Resolver {
	return &Resolver{
		config:  c,
		music:   m,
		video:   v,
		podcast: p,
		loc:     l,
	}
}

func (r *Resolver) addTrackEntries(tracks []music.Track, entries []spiff.Entry) []spiff.Entry {
	for _, t := range tracks {
		e := spiff.Entry{
			Creator:    t.Artist,
			Album:      t.ReleaseTitle,
			Title:      t.Title,
			Image:      r.music.TrackImage(t).String(),
			Location:   []string{r.loc.LocateTrack(t)},
			Identifier: []string{t.ETag},
			Size:       []int64{t.Size}}
		entries = append(entries, e)
	}
	return entries
}

func (r *Resolver) addMovieEntries(movies []video.Movie, entries []spiff.Entry) []spiff.Entry {
	for _, m := range movies {
		e := spiff.Entry{
			Creator:    "Movie", // TODO need better creator
			Album:      m.Title,
			Title:      m.Title,
			Image:      r.video.MoviePoster(m),
			Location:   []string{r.loc.LocateMovie(m)},
			Identifier: []string{m.ETag},
			Size:       []int64{m.Size}}
		entries = append(entries, e)
	}
	return entries
}

func (r *Resolver) addEpisodeEntries(series *podcast.Series, episodes []podcast.Episode,
	entries []spiff.Entry) []spiff.Entry {
	for _, e := range episodes {
		e := spiff.Entry{
			Creator:    "Podcast", // TODO need better creator
			Album:      series.Title,
			Title:      e.Title,
			Image:      r.podcast.EpisodeImage(e),
			Location:   []string{r.loc.LocateEpisode(e)},
			Identifier: []string{hash.MD5Hex(e.URL)}, // TODO hash of episode?
			Size:       []int64{e.Size}}
		entries = append(entries, e)
	}
	return entries
}

// Artist Track Refs:
// /music/artists/{id}/singles - artist tracks released as singles
// /music/artists/{id}/popular - artist tracks that are popular (lastfm)
// /music/artists/{id}/tracks - artist tracks
// /music/artists/{id}/similar - artist and similar artist tracks (radio)
// /music/artists/{id}/shuffle - selection of shuffled artist tracks
// /music/artists/{id}/deep - atrist deep tracks
func (r *Resolver) resolveArtistRef(id, res string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	artist, err := r.music.LookupArtist(n)
	if err != nil {
		return entries, err
	}
	var tracks []music.Track
	switch res {
	case "singles":
		tracks = r.music.ArtistSingleTracks(artist)
	case "popular":
		tracks = r.music.ArtistPopularTracks(artist)
	case "tracks":
		tracks = r.music.ArtistTracks(artist)
	case "shuffle":
		tracks = r.music.ArtistShuffle(artist, r.config.Music.RadioLimit)
	case "similar":
		tracks = r.music.ArtistRadio(artist)
	case "deep":
		tracks = r.music.ArtistDeep(artist, r.config.Music.RadioLimit)
	}
	entries = r.addTrackEntries(tracks, entries)
	return entries, nil
}

// /music/releases/{id}/tracks
func (r *Resolver) resolveReleaseRef(id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	release, err := r.music.LookupRelease(n)
	if err != nil {
		return entries, err
	}
	tracks := r.music.ReleaseTracks(release)
	entries = r.addTrackEntries(tracks, entries)
	return entries, nil
}

// /music/tracks/{id}
func (r *Resolver) resolveTrackRef(id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	t, err := r.music.LookupTrack(n)
	if err != nil {
		return entries, err
	}
	entries = r.addTrackEntries([]music.Track{t}, entries)
	return entries, nil
}

// /movies/{id}
func (r *Resolver) resolveMovieRef(id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	m, err := r.video.LookupMovie(n)
	if err != nil {
		return entries, err
	}
	entries = r.addMovieEntries([]video.Movie{*m}, entries)
	return entries, nil
}

// /series/{id}
func (r *Resolver) resolveSeriesRef(id string, entries []spiff.Entry) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	series, err := r.podcast.LookupSeries(n)
	if err != nil {
		return entries, err
	}
	episodes := r.podcast.Episodes(series)
	if err != nil {
		return entries, err
	}
	entries = r.addEpisodeEntries(series, episodes, entries)
	return entries, nil
}

// /music/search?q={q}[&radio=1]
func (r *Resolver) resolveSearchRef(uri string, entries []spiff.Entry) ([]spiff.Entry, error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Println(err)
		return entries, err
	}

	q := u.Query().Get("q")
	radio := u.Query().Get("radio")

	var tracks []music.Track
	if q != "" {
		limit := r.config.Music.SearchLimit
		if radio != "" {
			limit = r.config.Music.RadioSearchLimit
		}
		tracks = r.music.Search(q, limit)
	}

	if radio != "" {
		tracks = music.Shuffle(tracks)
		if len(tracks) > r.config.Music.RadioLimit {
			tracks = tracks[:r.config.Music.RadioLimit]
		}
	}

	entries = r.addTrackEntries(tracks, entries)
	return entries, nil
}

// /music/radio/{id}
func (r *Resolver) resolveRadioRef(id string, entries []spiff.Entry, user *auth.User) ([]spiff.Entry, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return entries, err
	}
	s, err := r.music.LookupStation(n)
	if err != nil {
		return entries, err
	}
	if !s.Visible(user) {
		return entries, err
	}

	//r.music.StationRefresh(&s, user)
	r.RefreshStation(&s, user)

	plist, _ := spiff.Unmarshal(s.Playlist)
	entries = append(entries, plist.Spiff.Entries...)

	return entries, nil
}

func (r *Resolver) RefreshStation(s *music.Station, user *auth.User) {
	plist := spiff.NewPlaylist()
	// Image
	plist.Spiff.Location = fmt.Sprintf("%s/api/radio/%d", r.config.Server.URL, s.ID)
	plist.Spiff.Title = s.Name
	plist.Spiff.Creator = "Radio"
	plist.Entries = []spiff.Entry{{Ref: s.Ref}}
	r.Resolve(user, plist)
	if plist.Entries == nil {
		plist.Entries = []spiff.Entry{}
	}
	s.Playlist, _ = plist.Marshal()
	r.music.UpdateStation(s)
}

func (r *Resolver) Resolve(user *auth.User, plist *spiff.Playlist) (err error) {
	var entries []spiff.Entry

	artistsRegexp := regexp.MustCompile(`/music/artists/([\d]+)/([\w]+)`)
	releasesRegexp := regexp.MustCompile(`/music/releases/([\d]+)/tracks`)
	tracksRegexp := regexp.MustCompile(`/music/tracks/([\d]+)`)
	searchRegexp := regexp.MustCompile(`/music/search.*`)
	radioRegexp := regexp.MustCompile(`/music/radio/([\d]+)`)
	moviesRegexp := regexp.MustCompile(`/movies/([\d]+)`)
	seriesRegexp := regexp.MustCompile(`/series/([\d]+)`)

	for _, e := range plist.Spiff.Entries {
		if e.Ref == "" {
			entries = append(entries, e)
			continue
		}

		pathRef := e.Ref

		matches := artistsRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = r.resolveArtistRef(matches[1], matches[2], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = releasesRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = r.resolveReleaseRef(matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = tracksRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = r.resolveTrackRef(matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		if searchRegexp.MatchString(pathRef) {
			entries, err = r.resolveSearchRef(pathRef, entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = radioRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = r.resolveRadioRef(matches[1], entries, user)
			if err != nil {
				return err
			}
			continue
		}

		matches = moviesRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = r.resolveMovieRef(matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}

		matches = seriesRegexp.FindStringSubmatch(pathRef)
		if matches != nil {
			entries, err = r.resolveSeriesRef(matches[1], entries)
			if err != nil {
				return err
			}
			continue
		}
	}

	plist.Spiff.Entries = entries

	return nil
}
