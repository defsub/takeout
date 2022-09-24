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

package view

import (
	"fmt"
	"time"

	"github.com/defsub/takeout/activity"
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/progress"
	"github.com/defsub/takeout/video"
)

type Context interface {
	Config() *config.Config
	Activity() *activity.Activity
	Music() *music.Music
	Podcast() *podcast.Podcast
	Progress() *progress.Progress
	User() *auth.User
	Video() *video.Video
	LocateMovie(video.Movie) string
}

type CoverFunc func(interface{}) string
type PosterFunc func(video.Movie) string
type BackdropFunc func(video.Movie) string
type ProfileFunc func(video.Person) string
type SeriesImageFunc func(podcast.Series) string
type EpisodeImageFunc func(podcast.Episode) string
type TracksFunc func() []music.Track

type TrackList struct {
	Title  string
	Tracks TracksFunc
}

type Index struct {
	Time        int64
	HasMusic    bool
	HasMovies   bool
	HasPodcasts bool
}

// swagger:model
type Home struct {
	AddedReleases   []music.Release
	NewReleases     []music.Release
	AddedMovies     []video.Movie
	NewMovies       []video.Movie
	RecommendMovies []video.Recommend
	NewEpisodes     []podcast.Episode
	NewSeries       []podcast.Series
	CoverSmall      CoverFunc        `json:"-"`
	PosterSmall     PosterFunc       `json:"-"`
	SeriesImage     SeriesImageFunc  `json:"-"`
	EpisodeImage    EpisodeImageFunc `json:"-"`
}

// swagger:model
type Artists struct {
	Artists    []music.Artist
	CoverSmall CoverFunc `json:"-"`
}

// swagger:model
type Artist struct {
	Artist     music.Artist
	Image      string
	Background string
	Releases   []music.Release
	Similar    []music.Artist
	CoverSmall CoverFunc `json:"-"`
	Deep       TrackList `json:"-"`
	Popular    TrackList `json:"-"`
	Radio      TrackList `json:"-"`
	Shuffle    TrackList `json:"-"`
	Singles    TrackList `json:"-"`
	Tracks     TrackList `json:"-"`
}

// swagger:model
type Popular struct {
	Artist     music.Artist
	Popular    []music.Track
	CoverSmall CoverFunc `json:"-"`
}

// swagger:model
type Singles struct {
	Artist     music.Artist
	Singles    []music.Track
	CoverSmall CoverFunc `json:"-"`
}

// swagger:model
type Release struct {
	Artist     music.Artist
	Release    music.Release
	Image      string
	Tracks     []music.Track
	Singles    []music.Track
	Popular    []music.Track
	Similar    []music.Release
	CoverSmall CoverFunc `json:"-"`
}

// swagger:model
type Search struct {
	Artists     []music.Artist
	Releases    []music.Release
	Tracks      []music.Track
	Movies      []video.Movie
	Query       string
	Hits        int
	CoverSmall  CoverFunc  `json:"-"`
	PosterSmall PosterFunc `json:"-"`
}

// swagger:model
type Radio struct {
	Artist     []music.Station
	Genre      []music.Station
	Similar    []music.Station
	Period     []music.Station
	Series     []music.Station
	Other      []music.Station
	Stream     []music.Station
	CoverSmall CoverFunc `json:"-"`
}

// swagger:model
type Movies struct {
	Movies      []video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
}

// swagger:model
type Movie struct {
	Movie       video.Movie
	Location    string
	Collection  video.Collection
	Other       []video.Movie
	Cast        []video.Cast
	Crew        []video.Crew
	Starring    []video.Person
	Directing   []video.Person
	Writing     []video.Person
	Genres      []string
	Keywords    []string
	Vote        int
	VoteCount   int
	Poster      PosterFunc   `json:"-"`
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
	Profile     ProfileFunc  `json:"-"`
}

// swagger:model
type Profile struct {
	Person      video.Person
	Starring    []video.Movie
	Directing   []video.Movie
	Writing     []video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
	Profile     ProfileFunc  `json:"-"`
}

// swagger:model
type Genre struct {
	Name        string
	Movies      []video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
}

// swagger:model
type Keyword struct {
	Name        string
	Movies      []video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
}

// swagger:model
type Watch struct {
	Movie       video.Movie
	Location    string
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
}

// swagger:model
type Podcasts struct {
	Series      []podcast.Series
	SeriesImage SeriesImageFunc `json:"-"`
}

// swagger:model
type Series struct {
	Series       podcast.Series
	Episodes     []podcast.Episode
	SeriesImage  SeriesImageFunc  `json:"-"`
	EpisodeImage EpisodeImageFunc `json:"-"`
}

// swagger:model
type SeriesEpisode struct {
	Episode      podcast.Episode
	EpisodeImage EpisodeImageFunc `json:"-"`
}

// swagger:model
type Progress struct {
	Offsets []progress.Offset
}

// swagger:model
type Offset struct {
	Offset progress.Offset
}

// swagger:model
type Activity struct {
	RecentMovies   []activity.Movie
	RecentTracks   []activity.Track
	RecentReleases []activity.Release
}

// swagger:model
type ActivityMovies struct {
	Movies []activity.Movie
}

// swagger:model
type ActivityTracks struct {
	Tracks []activity.Track
}

// swagger:model
type ActivityReleases struct {
	Releases []activity.Release
}

func IndexView(ctx Context) *Index {
	view := &Index{}
	view.Time = time.Now().UnixMilli()
	view.HasMusic = ctx.Music().HasMusic()
	view.HasMovies = ctx.Video().HasMovies()
	view.HasPodcasts = ctx.Podcast().HasPodcasts()
	return view
}

func HomeView(ctx Context) *Home {
	view := &Home{}
	m := ctx.Music()
	v := ctx.Video()
	p := ctx.Podcast()

	view.AddedReleases = m.RecentlyAdded()
	view.NewReleases = m.RecentlyReleased()
	view.AddedMovies = v.RecentlyAdded()
	view.NewMovies = v.RecentlyReleased()
	view.RecommendMovies = v.Recommend()
	view.NewEpisodes = p.RecentEpisodes()
	view.NewSeries = p.RecentSeries()

	view.CoverSmall = m.CoverSmall
	view.PosterSmall = v.MoviePosterSmall
	view.EpisodeImage = p.EpisodeImage
	view.SeriesImage = p.SeriesImage
	return view
}

func ArtistsView(ctx Context) *Artists {
	view := &Artists{}
	view.Artists = ctx.Music().Artists()
	view.CoverSmall = ctx.Music().CoverSmall
	return view
}

func ArtistView(ctx Context, artist music.Artist) *Artist {
	m := ctx.Music()
	view := &Artist{}
	view.Artist = artist
	view.Releases = m.ArtistReleases(&artist)
	view.Similar = m.SimilarArtists(&artist)
	view.Image = m.ArtistImage(&artist)
	view.Background = m.ArtistBackground(&artist)
	view.CoverSmall = m.CoverSmall
	view.Popular = TrackList{
		Title: fmt.Sprintf("%s \u2013 Popular", artist.Name),
		Tracks: func() []music.Track {
			tracks := m.ArtistPopularTracks(artist, ctx.Config().Music.PopularLimit)
			return tracks
		},
	}
	view.Singles = TrackList{
		Title: fmt.Sprintf("%s \u2013 Singles", artist.Name),
		Tracks: func() []music.Track {
			tracks := m.ArtistSingleTracks(artist, ctx.Config().Music.SinglesLimit)
			return tracks
		},
	}
	view.Deep = TrackList{
		Title: fmt.Sprintf("%s \u2013 Deep Tracks", artist.Name),
		Tracks: func() []music.Track {
			return m.ArtistDeep(artist, ctx.Config().Music.RadioLimit)
		},
	}
	view.Radio = TrackList{
		Title: fmt.Sprintf("%s \u2013 Radio", artist.Name),
		Tracks: func() []music.Track {
			return m.ArtistRadio(artist)
		},
	}
	view.Shuffle = TrackList{
		Title: fmt.Sprintf("%s \u2013 Shuffle", artist.Name),
		Tracks: func() []music.Track {
			return m.ArtistShuffle(artist, ctx.Config().Music.RadioLimit)
		},
	}
	view.Tracks = TrackList{
		Title: fmt.Sprintf("%s \u2013 Tracks", artist.Name),
		Tracks: func() []music.Track {
			return m.ArtistTracks(artist)
		},
	}
	return view
}

func PopularView(ctx Context, artist music.Artist) *Popular {
	m := ctx.Music()
	view := &Popular{}
	view.Artist = artist
	view.Popular = m.ArtistPopularTracks(artist)
	limit := ctx.Config().Music.PopularLimit
	if len(view.Popular) > limit {
		view.Popular = view.Popular[:limit]
	}
	view.CoverSmall = m.CoverSmall
	return view
}

func SinglesView(ctx Context, artist music.Artist) *Singles {
	m := ctx.Music()
	view := &Singles{}
	view.Artist = artist
	view.Singles = m.ArtistSingleTracks(artist)
	limit := ctx.Config().Music.SinglesLimit
	if len(view.Singles) > limit {
		view.Singles = view.Singles[:limit]
	}
	view.CoverSmall = m.CoverSmall
	return view
}

func ReleaseView(ctx Context, release music.Release) *Release {
	m := ctx.Music()
	view := &Release{}
	view.Release = release
	view.Artist = *m.Artist(release.Artist)
	view.Tracks = m.ReleaseTracks(release)
	view.Singles = m.ReleaseSingles(release)
	view.Popular = m.ReleasePopular(release)
	view.Similar = m.SimilarReleases(&view.Artist, release)
	view.Image = m.CoverSmall(release)
	view.CoverSmall = m.CoverSmall
	return view
}

func SearchView(ctx Context, query string) *Search {
	m := ctx.Music()
	v := ctx.Video()
	view := &Search{}
	artists, releases, _ := m.Query(query)
	view.Artists = artists
	view.Releases = releases
	view.Query = query
	view.Tracks = m.Search(query)
	view.Movies = v.Search(query)
	view.Hits = len(view.Artists) + len(view.Releases) + len(view.Tracks) + len(view.Movies)
	view.CoverSmall = m.CoverSmall
	view.PosterSmall = v.MoviePosterSmall
	return view
}

func RadioView(ctx Context) *Radio {
	m := ctx.Music()
	view := &Radio{}
	for _, s := range m.Stations(ctx.User()) {
		switch s.Type {
		case music.TypeArtist:
			view.Artist = append(view.Artist, s)
		case music.TypeGenre:
			view.Genre = append(view.Genre, s)
		case music.TypeSimilar:
			view.Similar = append(view.Similar, s)
		case music.TypePeriod:
			view.Period = append(view.Period, s)
		case music.TypeSeries:
			view.Series = append(view.Series, s)
		case music.TypeStream:
			view.Stream = append(view.Stream, s)
		default:
			view.Other = append(view.Other, s)
		}
	}
	return view
}

func MoviesView(ctx Context) *Movies {
	v := ctx.Video()
	view := &Movies{}
	view.Movies = v.Movies()
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	return view
}

func MovieView(ctx Context, m video.Movie) *Movie {
	v := ctx.Video()
	view := &Movie{}
	view.Movie = m
	view.Location = ctx.LocateMovie(m)
	collection := v.MovieCollection(m)
	if collection != nil {
		view.Collection = *collection
		view.Other = v.CollectionMovies(collection)
		if len(view.Other) == 1 && view.Other[0].ID == m.ID {
			// collection is just this movie so remove
			view.Other = view.Other[1:]
		}
	}
	view.Cast = v.Cast(m)
	view.Crew = v.Crew(m)
	for _, c := range view.Crew {
		switch c.Job {
		case video.JobDirector:
			view.Directing = append(view.Directing, c.Person)
		case video.JobNovel, video.JobScreenplay, video.JobStory:
			view.Writing = append(view.Writing, c.Person)
		}
	}
	for i, c := range view.Cast {
		if i == 3 {
			break
		}
		view.Starring = append(view.Starring, c.Person)
	}
	view.Genres = v.Genres(m)
	view.Keywords = v.Keywords(m)
	view.Vote = int(m.VoteAverage * 10)
	view.VoteCount = m.VoteCount
	view.Poster = v.MoviePoster
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	view.Profile = v.PersonProfile
	return view
}

func ProfileView(ctx Context, p video.Person) *Profile {
	v := ctx.Video()
	view := &Profile{}
	view.Person = p
	view.Starring = v.Starring(p)
	view.Writing = v.Writing(p)
	view.Directing = v.Directing(p)
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	view.Profile = v.PersonProfile
	return view
}

func GenreView(ctx Context, name string) *Genre {
	v := ctx.Video()
	view := &Genre{}
	view.Name = name
	view.Movies = v.Genre(name)
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	return view
}

func KeywordView(ctx Context, name string) *Keyword {
	v := ctx.Video()
	view := &Keyword{}
	view.Name = name
	view.Movies = v.Keyword(name)
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	return view
}

func WatchView(ctx Context, m video.Movie) *Watch {
	v := ctx.Video()
	view := &Watch{}
	view.Movie = m
	view.Location = ctx.LocateMovie(m)
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	return view
}

func PodcastsView(ctx Context) *Podcasts {
	p := ctx.Podcast()
	view := &Podcasts{}
	view.Series = p.Series()
	view.SeriesImage = p.SeriesImage
	return view
}

func SeriesView(ctx Context, s podcast.Series) *Series {
	p := ctx.Podcast()
	view := &Series{}
	view.Series = s
	view.Episodes = p.Episodes(s)
	view.SeriesImage = p.SeriesImage
	view.EpisodeImage = p.EpisodeImage
	return view
}

func SeriesEpisodeView(ctx Context, e podcast.Episode) *SeriesEpisode {
	view := &SeriesEpisode{}
	view.Episode = e
	view.EpisodeImage = ctx.Podcast().EpisodeImage
	return view
}

func ProgressView(ctx Context) *Progress {
	view := &Progress{}
	view.Offsets = ctx.Progress().Offsets(ctx.User())
	return view
}

func OffsetView(ctx Context, offset progress.Offset) *Offset {
	view := &Offset{}
	view.Offset = offset
	return view
}

func ActivityView(ctx Context) *Activity {
	view := &Activity{}
	view.RecentTracks = ctx.Activity().RecentTracks(ctx)
	view.RecentMovies = ctx.Activity().RecentMovies(ctx)
	view.RecentReleases = ctx.Activity().RecentReleases(ctx)
	return view
}

func ActivityTracksView(ctx Context, start, end time.Time) *ActivityTracks {
	view := &ActivityTracks{}
	view.Tracks = ctx.Activity().Tracks(ctx, start, end)
	return view
}

func ActivityPopularTracksView(ctx Context, start, end time.Time) *ActivityTracks {
	view := &ActivityTracks{}
	view.Tracks = ctx.Activity().PopularTracks(ctx, start, end)
	return view
}

func ActivityMoviesView(ctx Context, start, end time.Time) *ActivityMovies {
	view := &ActivityMovies{}
	view.Movies = ctx.Activity().Movies(ctx, start, end)
	return view
}

func ActivityReleasesView(ctx Context, start, end time.Time) *ActivityReleases {
	view := &ActivityReleases{}
	view.Releases = ctx.Activity().Releases(ctx, start, end)
	return view
}
