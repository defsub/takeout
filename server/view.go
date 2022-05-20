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
	"time"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/progress"
	"github.com/defsub/takeout/video"
)

type CoverFunc func(interface{}) string

type PosterFunc func(video.Movie) string
type BackdropFunc func(video.Movie) string
type ProfileFunc func(video.Person) string
type SeriesImageFunc func(podcast.Series) string
type EpisodeImageFunc func(podcast.Episode) string

type IndexView struct {
	Time        int64
	HasMusic    bool
	HasMovies   bool
	HasPodcasts bool
}

type HomeView struct {
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

type ArtistsView struct {
	Artists    []music.Artist
	CoverSmall CoverFunc `json:"-"`
}

type ArtistView struct {
	Artist     music.Artist
	Image      string
	Background string
	Releases   []music.Release
	Popular    []music.Track
	Singles    []music.Track
	Similar    []music.Artist
	CoverSmall CoverFunc `json:"-"`
}

type PopularView struct {
	Artist     music.Artist
	Popular    []music.Track
	CoverSmall CoverFunc `json:"-"`
}

type SinglesView struct {
	Artist     music.Artist
	Singles    []music.Track
	CoverSmall CoverFunc `json:"-"`
}

type ReleaseView struct {
	Artist     music.Artist
	Release    music.Release
	Tracks     []music.Track
	Singles    []music.Track
	Popular    []music.Track
	Similar    []music.Release
	CoverSmall CoverFunc `json:"-"`
}

type SearchView struct {
	Artists     []music.Artist
	Releases    []music.Release
	Tracks      []music.Track
	Movies      []video.Movie
	Query       string
	Hits        int
	CoverSmall  CoverFunc  `json:"-"`
	PosterSmall PosterFunc `json:"-"`
}

type RadioView struct {
	Artist     []music.Station
	Genre      []music.Station
	Similar    []music.Station
	Period     []music.Station
	Series     []music.Station
	Other      []music.Station
	Stream     []music.Station
	CoverSmall CoverFunc `json:"-"`
}

type MoviesView struct {
	Movies      []video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
}

type MovieView struct {
	Movie       video.Movie
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

type ProfileView struct {
	Person      video.Person
	Starring    []video.Movie
	Directing   []video.Movie
	Writing     []video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
	Profile     ProfileFunc  `json:"-"`
}

type GenreView struct {
	Name        string
	Movies      []video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
}

type KeywordView struct {
	Name        string
	Movies      []video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
}

type WatchView struct {
	Movie       video.Movie
	PosterSmall PosterFunc   `json:"-"`
	Backdrop    BackdropFunc `json:"-"`
}

type PodcastsView struct {
	Series      []podcast.Series
	SeriesImage SeriesImageFunc `json:"-"`
}

type SeriesView struct {
	Series       podcast.Series
	Episodes     []podcast.Episode
	SeriesImage  SeriesImageFunc  `json:"-"`
	EpisodeImage EpisodeImageFunc `json:"-"`
}

type SeriesEpisodeView struct {
	Episode      podcast.Episode
	EpisodeImage EpisodeImageFunc `json:"-"`
}

type ProgressView struct {
	Offsets []progress.Offset
}

type OffsetView struct {
	Offset progress.Offset
}

func (handler *UserHandler) indexView() *IndexView {
	view := &IndexView{}
	view.Time = time.Now().UnixMilli()
	view.HasMusic = handler.media.music.HasMusic()
	view.HasMovies = handler.media.video.HasMovies()
	view.HasPodcasts = handler.media.podcast.HasPodcasts()
	return view
}

func (handler *UserHandler) homeView() *HomeView {
	view := &HomeView{}
	m := handler.media.music
	v := handler.media.video
	p := handler.media.podcast

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

func (handler *UserHandler) artistsView() *ArtistsView {
	view := &ArtistsView{}
	view.Artists = handler.music().Artists()
	view.CoverSmall = handler.music().CoverSmall
	return view
}

func (handler *UserHandler) artistView(artist music.Artist) *ArtistView {
	m := handler.music()
	view := &ArtistView{}
	view.Artist = artist
	view.Releases = m.ArtistReleases(&artist)
	view.Popular = m.ArtistPopularTracks(artist)
	n := 5
	if len(view.Popular) > n {
		view.Popular = view.Popular[:n]
	}
	view.Singles = m.ArtistSingleTracks(artist)
	if len(view.Singles) > n {
		view.Singles = view.Singles[:n]
	}
	view.Similar = m.SimilarArtists(&artist)
	view.Image = m.ArtistImage(&artist)
	view.Background = m.ArtistBackground(&artist)
	view.CoverSmall = m.CoverSmall
	return view
}

func (handler *UserHandler) popularView(artist music.Artist) *PopularView {
	m := handler.music()
	view := &PopularView{}
	view.Artist = artist
	view.Popular = m.ArtistPopularTracks(artist)
	limit := handler.config.Music.PopularLimit
	if len(view.Popular) > limit {
		view.Popular = view.Popular[:limit]
	}
	view.CoverSmall = m.CoverSmall
	return view
}

func (handler *UserHandler) singlesView(artist music.Artist) *SinglesView {
	m := handler.music()
	view := &SinglesView{}
	view.Artist = artist
	view.Singles = m.ArtistSingleTracks(artist)
	limit := handler.config.Music.SinglesLimit
	if len(view.Singles) > limit {
		view.Singles = view.Singles[:limit]
	}
	view.CoverSmall = m.CoverSmall
	return view
}

func (handler *UserHandler) releaseView(release music.Release) *ReleaseView {
	m := handler.music()
	view := &ReleaseView{}
	view.Release = release
	view.Artist = *m.Artist(release.Artist)
	view.Tracks = m.ReleaseTracks(release)
	view.Singles = m.ReleaseSingles(release)
	view.Popular = m.ReleasePopular(release)
	view.Similar = m.SimilarReleases(&view.Artist, release)
	view.CoverSmall = m.CoverSmall
	return view
}

func (handler *UserHandler) searchView(query string) *SearchView {
	m := handler.media.music
	v := handler.media.video
	view := &SearchView{}
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

func (handler *UserHandler) radioView(user *auth.User) *RadioView {
	view := &RadioView{}
	for _, s := range handler.music().Stations(user) {
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

func (handler *UserHandler) moviesView() *MoviesView {
	v := handler.video()
	view := &MoviesView{}
	view.Movies = v.Movies()
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	return view
}

func (handler *UserHandler) movieView(m video.Movie) *MovieView {
	v := handler.video()
	view := &MovieView{}
	view.Movie = m
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

func (handler *UserHandler) profileView(p video.Person) *ProfileView {
	v := handler.video()
	view := &ProfileView{}
	view.Person = p
	view.Starring = v.Starring(p)
	view.Writing = v.Writing(p)
	view.Directing = v.Directing(p)
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	view.Profile = v.PersonProfile
	return view
}

func (handler *UserHandler) genreView(name string) *GenreView {
	v := handler.video()
	view := &GenreView{}
	view.Name = name
	view.Movies = v.Genre(name)
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	return view
}

func (handler *UserHandler) keywordView(name string) *KeywordView {
	v := handler.video()
	view := &KeywordView{}
	view.Name = name
	view.Movies = v.Keyword(name)
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	return view
}

func (handler *UserHandler) watchView(m video.Movie) *WatchView {
	v := handler.video()
	view := &WatchView{}
	view.Movie = m
	view.PosterSmall = v.MoviePosterSmall
	view.Backdrop = v.MovieBackdrop
	return view
}

func (handler *UserHandler) podcastsView() *PodcastsView {
	p := handler.podcast()
	view := &PodcastsView{}
	view.Series = p.Series()
	view.SeriesImage = p.SeriesImage
	return view
}

func (handler *UserHandler) seriesView(s podcast.Series) *SeriesView {
	p := handler.podcast()
	view := &SeriesView{}
	view.Series = s
	view.Episodes = p.Episodes(s)
	view.SeriesImage = p.SeriesImage
	view.EpisodeImage = p.EpisodeImage
	return view
}

func (handler *UserHandler) seriesEpisodeView(e podcast.Episode) *SeriesEpisodeView {
	p := handler.podcast()
	view := &SeriesEpisodeView{}
	view.Episode = e
	view.EpisodeImage = p.EpisodeImage
	return view
}

func (handler *UserHandler) progressView() *ProgressView {
	view := &ProgressView{}
	view.Offsets = handler.progress().Offsets(handler.user)
	return view
}

func (handler *UserHandler) offsetView(offset progress.Offset) *OffsetView {
	view := &OffsetView{}
	view.Offset = offset
	return view
}
