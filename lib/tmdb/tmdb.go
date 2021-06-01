// Copyright (C) 2021 The Takeout Authors.
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

package tmdb

import (
	"fmt"
	"net/url"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
)

type TMDB struct {
	config      *config.Config
	client      *client.Client
	configCache *apiConfig
	genreCache  Genres
}

func NewTMDB(config *config.Config) *TMDB {
	return &TMDB{
		config: config,
		client: client.NewClient(config),
	}
}

type Collection struct {
	ID           int    `json:"id"` // unique collection ID
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`
	// Parts []MovieResult
}

type Movie struct {
	ID               int        `json:"id"` // unique movie ID
	IMDB_ID          string     `json:"imdb_id"`
	Adult            bool       `json:"adult"`
	BackdropPath     string     `json:"backdrop_path"`
	Collection       Collection `json:"belongs_to_collection"`
	Genres           []Genre    `json:"genres"`
	OriginalLanguage string     `json:"original_language"`
	OriginalTitle    string     `json:"original_title"`
	Overview         string     `json:"overview"`
	Popularity       float32    `json:"populartity"`
	PosterPath       string     `json:"poster_path"`
	ReleaseDate      string     `json:"release_date"`
	Tagline          string     `json:"tagline"`
	Title            string     `json:"title"`
	Video            bool       `json:"video"`
	VoteAverage      float32    `json:"vote_average"`
	VoteCount        int        `json:"vote_count"`
	Budget           int64      `json:"budget"`
	Revenue          int64      `json:"revenue"`
	Runtime          int        `json:"runtime"`
}

type Cast struct {
	ID           int    `json:"id"` // unique person ID
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	ProfilePath  string `json:"profile_path"`
	Character    string `json:"character"`
	Order        int    `json:"order"`
}

type Crew struct {
	ID           int    `json:"id"` // unique person ID
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	ProfilePath  string `json:"profile_path"`
	Department   string `json:"department"`
	Job          string `json:"job"`
}

type Credits struct {
	ID   int    `json:"id"` // unique movie ID
	Cast []Cast `json:"cast"`
	Crew []Crew `json:"crew"`
}

type Person struct {
	ID          int    `json:"id"` // unique person ID
	IMDB_ID     string `json:"imdb_id"`
	Name        string `json:"name"`
	ProfilePath string `json:"profile_path"`
	Birthday    string `json:"birthday"`
	Deathday    string `json:"deathday"`
	Biography   string `json:"biography"`
	Birthplace  string `json:"place_of_birth"`
}

// https://developers.themoviedb.org/3/movies/get-movie-release-dates
const (
	TypePremiere = iota + 1
	TypeTheatricalLimited
	TypeTheatrical
	TypeDigital
	TypePhysical
	TypeTV
)

type Release struct {
	Certification string `json:"certification"`
	Date          string `json:"release_date"`
	Note          string `json:"note"`
	Type          int    `json:"type"`
}

type ReleaseCountry struct {
	CountryCode string    `json:"iso_3166_1"`
	Releases    []Release `json:"release_dates"`
}

type Releases struct {
	ID      int              `json:"id"`
	Results []ReleaseCountry `json:"results"`
}

type MovieResult struct {
	ID               int     `json:"id"`
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float32 `json:"populartity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float32 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

type moviePage struct {
	Page         int           `json:"page"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
	Results      []MovieResult `json:"results"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type genreList struct {
	Genres []Genre `json:"genres"`
}

type Genres map[int]string

type imagesConfig struct {
	BaseURL       string   `json:"base_url"`
	SecureBaseURL string   `json:"secure_base_url"`
	BackdropSizes []string `json:"backdrop_sizes"`
	LogoSizes     []string `json:"logo_sizes"`
	PosterSizes   []string `json:"poster_sizes"`
	ProfileSizes  []string `json:"profile_sizes"`
}

type apiConfig struct {
	Images    imagesConfig `json:"images"`
	ChangeKey []string     `json:"change_keys"`
}

const (
	endpoint = "api.themoviedb.org"
)

func (m *TMDB) moviePage(q string, page int) (*moviePage, error) {
	url := fmt.Sprintf(
		"https://%s/3/search/movie?api_key=%s&language=en-US&query=%s&page=%d",
		endpoint, m.config.TMDB.Key, url.QueryEscape(q), page)
	var result moviePage
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *TMDB) MovieSearch(q string) ([]MovieResult, error) {
	// TODO only supports one page right now
	page, err := m.moviePage(q, 1)
	return page.Results, err
}

func (m *TMDB) MovieDetail(tmid int) (*Movie, error) {
	url := fmt.Sprintf(
		"https://%s/3/movie/%d?api_key=%s",
		endpoint, tmid, m.config.TMDB.Key)
	var result Movie
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *TMDB) MovieCredits(tmid int) (*Credits, error) {
	url := fmt.Sprintf(
		"https://%s/3/movie/%d/credits?api_key=%s",
		endpoint, tmid, m.config.TMDB.Key)
	var result Credits
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *TMDB) MovieReleases(tmid int) (map[string][]Release, error) {
	url := fmt.Sprintf(
		"https://%s/3/movie/%d/release_dates?api_key=%s",
		endpoint, tmid, m.config.TMDB.Key)
	var result Releases
	var countryMap map[string][]Release
	err := m.client.GetJson(url, &result)
	if err == nil {
		countryMap = make(map[string][]Release)
		for _, rc := range result.Results {
			countryMap[rc.CountryCode] = rc.Releases
		}
	}
	return countryMap, err
}

func (m *TMDB) MovieReleaseType(tmid int, country string, releaseType int) (*Release, error) {
	url := fmt.Sprintf(
		"https://%s/3/movie/%d/release_dates?api_key=%s",
		endpoint, tmid, m.config.TMDB.Key)
	var result Releases
	err := m.client.GetJson(url, &result)
	if err == nil {
		for _, rc := range result.Results {
			if rc.CountryCode == country {
				for _, r := range rc.Releases {
					if r.Type == releaseType {
						return &r, nil
					}
				}
			}
		}
	}
	return nil, err
}

func (m *TMDB) PersonDetail(peid int) (*Person, error) {
	url := fmt.Sprintf(
		"https://%s/3/person/%d?api_key=%s",
		endpoint, peid, m.config.TMDB.Key)
	var result Person
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *TMDB) MovieGenres() (Genres, error) {
	genres := make(Genres)
	url := fmt.Sprintf(
		"https://%s/3/genre/movie/list?api_key=%s", endpoint, m.config.TMDB.Key)
	var result genreList
	err := m.client.GetJson(url, &result)
	if err == nil {
		for _, g := range result.Genres {
			genres[g.ID] = g.Name
		}
	}
	return genres, err
}

func (m *TMDB) MovieGenre(id int) string {
	if m.genreCache == nil {
		m.genreCache, _ = m.MovieGenres()
	}
	return m.genreCache[id]
}

func (m *TMDB) MoveGenreNames() []string {
	if m.genreCache == nil {
		m.genreCache, _ = m.MovieGenres()
	}
	var result []string
	for _, v := range m.genreCache {
		result = append(result, v)
	}
	return result
}

func (m *TMDB) configuration() (*apiConfig, error) {
	url := fmt.Sprintf(
		"https://%s/3/configuration?api_key=%s", endpoint, m.config.TMDB.Key)
	var result apiConfig
	err := m.client.GetJson(url, &result)
	return &result, err
}

/*
   https://developers.themoviedb.org/3/configuration/get-api-configuration

   To build an image URL, you will need 3 pieces of data. The base_url, size and
   file_path. Simply combine them all and you will have a fully qualified
   URL. Here’s an example URL:

   https://image.tmdb.org/t/p/w500/8uO0gUM8aNqYLs1OsTBQiXu0fEv.jpg
*/

func moviePoster(c *apiConfig, size string, r *MovieResult) string {
	url := fmt.Sprintf("%s%s%s", c.Images.SecureBaseURL, size, r.PosterPath)
	return url
}

func movieBackdrop(c *apiConfig, size string, r *MovieResult) string {
	url := fmt.Sprintf("%s%s%s", c.Images.SecureBaseURL, size, r.BackdropPath)
	return url
}

func (m *TMDB) MovieOriginalPoster(r *MovieResult) *url.URL {
	var err error
	if m.configCache == nil {
		m.configCache, err = m.configuration()
	}
	if err != nil {
		return nil
	}
	v := moviePoster(m.configCache, "original", r)
	url, err := url.Parse(v)
	if err != nil {
		return nil
	}
	return url
}
