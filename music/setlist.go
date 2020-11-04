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

package music

import (
	"fmt"
)

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
}

type setlistSongs struct {
	Encore int           `json:"encore"`
	Song   []setlistSong `json:"song"`
}

type setlistSet struct {
	Set []setlistSongs `json:"set"`
}

type setlist struct {
	Id          string        `json:"id"`
	VersionId   string        `json:"versionId"`
	EventDate   string        `json:"eventDate"`
	LastUpdated string        `json:"lastUpdated"`
	Artist      setlistArtist `json:"artist"`
	Venue       setlistVenue  `json:"venue"`
	Tour        setlistTour   `json:"tour"`
	Sets        setlistSet    `json:"sets"`
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

func (m *Music) Setlists(artist *Artist) {
	result := m.setlistPage(artist, 1)
	fmt.Printf("items %d, page %d, total %d (%d)\n",
		result.ItemsPerPage, result.Page, result.Total,
		result.Total / result.ItemsPerPage)
	for _, sl := range result.Setlist {
		fmt.Printf("%s %s @ %s, %s, %s\n", sl.Tour.Name, sl.EventDate,
			sl.Venue.Name, sl.Venue.City.Name, sl.Venue.City.Country.Name)
	}
}

func (m *Music) setlistPage(artist *Artist, page int) *setlistResponse {
	url := fmt.Sprintf("https://api.setlist.fm/rest/1.0/artist/%s/setlists?p=%d",
		artist.ARID, page)

	headers := map[string]string{
		"Accept":    "application/json",
		"x-api-key": "TGRHYagNV144XhA74sNfNPqoa443ClIjUabD",
	}

	var result setlistResponse
	m.client.GetJsonWith(headers, url, &result)
	return &result
}
