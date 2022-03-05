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

package setlist

import (
	"fmt"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
)

type Setlist struct {
	config *config.Config
	client *client.Client
}

func NewSetlist(config *config.Config, client *client.Client) *Setlist {
	return &Setlist{
		config: config,
		client: client,
	}
}

type setlistArtist struct {
	Mbid           string `json:"mbid"`
	Tmid           int    `json:"tmid"`
	Name           string `json:"name"`
	SorName        string `json:"sortName"`
	Disambiguation string `json:"disambiguation"`
	Url            string `json:"url"`
}

type setlistCountry struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type setlistCity struct {
	Id        string         `json:"id"`
	Name      string         `json:"name"`
	State     string         `json:"state"`
	StateCode string         `json:"stateCode"`
	Country   setlistCountry `json:"country"`
}

type setlistVenue struct {
	Id   string      `json:"id"`
	Name string      `json:"name"`
	Url  string      `json:"url"`
	City setlistCity `json:"city"`
}

type setlistTour struct {
	Name string `json:"name"`
}

type setlistSong struct {
	Name string `json:"name"`
	Info string `json:"info"`
	Tape bool   `json:"tape"`
}

type setlistSet struct {
	Encore int           `json:"encore"`
	Name   string        `json:"name"`
	Songs  []setlistSong `json:"song"`
}

type setlistSets struct {
	Set []setlistSet `json:"set"`
}

type setlist struct {
	Id          string        `json:"id"`
	VersionId   string        `json:"versionId"`
	EventDate   string        `json:"eventDate"`
	LastUpdated string        `json:"lastUpdated"`
	Artist      setlistArtist `json:"artist"`
	Venue       setlistVenue  `json:"venue"`
	Tour        setlistTour   `json:"tour"`
	Sets        setlistSets   `json:"sets"`
	Info        string        `json:"info"`
	Url         string        `json:"url"`
}

type setlistResponse struct {
	Type         string    `json:"type"`
	ItemsPerPage int       `json:"itemsPerPage"`
	Page         int       `json:"page"`
	Total        int       `json:"total"`
	Setlist      []setlist `json:"setlist"`
}

func (s *Setlist) ArtistYear(arid string, year int) []setlist {
	format := fmt.Sprintf("https://api.setlist.fm/rest/1.0/search/setlists?artistMbid=%s&year=%d",
		arid, year)
	return setlistFetch(format + "&page=%d")
}

func (s *Setlist) setlistFetch(format string) []setlist {
	var list []setlist
	total := 0

	for page := 1; ; page++ {
		url := fmt.Sprintf(format, page)
		result := s.setlistPage(url)
		fmt.Printf("page %d, items per page %d, total %d\n",
			result.Page, result.ItemsPerPage, result.Total)
		list = append(list, result.Setlist...)
		total += len(result.Setlist)
		if total >= result.Total {
			break
		}
	}

	return list
}

func (s *Setlist) setlistPage(url string) *setlistResponse {
	fmt.Printf("url %s\n", url)
	headers := map[string]string{
		"Accept":    "application/json",
		"x-api-key": "TGRHYagNV144XhA74sNfNPqoa443ClIjUabD",
	}
	var result setlistResponse
	s.client.GetJsonWith(headers, url, &result)
	return &result
}
