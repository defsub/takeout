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
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// MusicBrainz is used for:
// * getting the MBID for artists
// * correcting artist names
// * correcting release names
// * getting release types (Album, Single, EP, Broadcast, Other)
// * getting original release dates (release groups)
// * getting releases within groups that have different titles
//

type mbzArtistsPage struct {
	Artists []mbzArtist `json:"artists"`
	Offset  int         `json:"offset"`
	Count   int         `json:"count"`
}

// TODO artist detail
// get detail for each artist - lifespan, url rel links, genres
// json api:
// http://musicbrainz.org/ws/2/artist/3798b104-01cb-484c-a3b0-56adc6399b80?inc=genres+url-rels&fmt=json
type mbzLifeSpan struct {
	Ended bool   `json:"ended"`
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// TODO artist detail
type mbzArea struct {
	Name     string `json:"name"`
	SortName string `json:"sort-name"`
}

type mbzArtist struct {
	ID             string        `json:"id"`
	Score          int           `json:"score"`
	Name           string        `json:"name"`
	SortName       string        `json:"sort-name"`
	Country        string        `json:"country"`
	Disambiguation string        `json:"disambiguation"`
	Type           string        `json:"type"`
	Genres         []mbzGenre    `json:"genres"`
	Tags           []mbzTag      `json:"tags"`
	Area           mbzArea       `json:"area"`
	BeginArea      mbzArea       `json:"begin-area"`
	EndArea        mbzArea       `json:"end-area"`
	LifeSpan       mbzLifeSpan   `json:"life-span"`
	Relations      []mbzRelation `json:"relations"`
}

type mbzArtistCredit struct {
	Name   string    `json:name`
	Join   string    `json:joinphrase`
	Artist mbzArtist `json:"artist"`
}

type mbzWork struct {
	Title     string        `json:"title"`
	Relations []mbzRelation `json:"relations"`
}

// TODO artist detail
type mbzURL struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
}

// type="Release group series"
// type="Recording series" (for recording in release)
type mbzSeries struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Disambiguation string `json:"disambiguation"`
	Type           string `json:"type"`
}

// release-group series: type="part of", target-type="series", see series
// release recording series: type="part of", target-type="series", see series
// single: type="single from", target-type="release_group", see release_group
type mbzRelation struct {
	Type        string    `json:"type"`
	TargetType  string    `json:"target-type"`
	Artist      mbzArtist `json:"artist"`
	Attributes  []string  `json:"attributes"`
	Work        mbzWork   `json:"work"`
	URL         mbzURL    `json:"url"`
	Series      mbzSeries `json:"series"`
}

type mbzLabelInfo struct {
	Label         mbzLabel `json:"label"`
	CatalogNumber string   `json:"catalog-number"`
}

type mbzLabel struct {
	Name     string `json:"name"`
	SortName string `json:"sort-name"`
}

type mbzMedia struct {
	Title      string     `json:"title"`
	Format     string     `json:"format"`
	Position   int        `json:"position"`
	TrackCount int        `json:"track-count"`
	Tracks     []mbzTrack `json:"tracks"`
}

func (m mbzMedia) video() bool {
	switch m.Format {
	case "DVD-Video", "Blu-ray", "HD-DVD", "VCD", "SVCD":
		return true
	}
	return false
}

type mbzRecording struct {
	ID           string            `json:"id"`
	Length       int               `json:"length"`
	Title        string            `json:"title"`
	Relations    []mbzRelation     `json:"relations"`
	ArtistCredit []mbzArtistCredit `json:"artist-credit"`
}

type mbzTrack struct {
	Title        string            `json:"title"`
	Position     int               `json:"position"`
	ArtistCredit []mbzArtistCredit `json:"artist-credit"`
	Recording    mbzRecording      `json:"recording"`
}

type mbzReleasesPage struct {
	Releases []mbzRelease `json:"releases"`
	Offset   int          `json:"release-offset"`
	Count    int          `json:"release-count"`
}

type mbzCoverArtArchive struct {
	Count    int  `json:"count"`
	Artwork  bool `json:"artwork"`
	Front    bool `json:"front"`
	Back     bool `json:"back"`
	Darkened bool `json:"darkened"`
}

type mbzRelease struct {
	ID              string             `json:"id"`
	Title           string             `json:"title"`
	Date            string             `json:"date"`
	Disambiguation  string             `json:"disambiguation"`
	Country         string             `json:"country"`
	Status          string             `json:"status"`
	Asin            string             `json:"asin"`
	Relations       []mbzRelation      `json:"relations"`
	LabelInfo       []mbzLabelInfo     `json:"label-info"`
	Media           []mbzMedia         `json:"media"`
	ReleaseGroup    mbzReleaseGroup    `json:"release-group"`
	CoverArtArchive mbzCoverArtArchive `json:"cover-art-archive"`
	ArtistCredit    []mbzArtistCredit  `json:"artist-credit"`
}

func (r mbzRelease) title() string {
	if r.Disambiguation != "" {
		return fmt.Sprintf("%s (%s)", r.Title, r.Disambiguation)
	} else {
		return r.Title
	}
}

func (r mbzRelease) totalTracks() int {
	count := 0
	for _, m := range r.Media {
		if m.video() {
			continue
		}
		count += m.TrackCount
	}
	return count
}

func (r mbzRelease) totalDiscs() int {
	count := 0
	for _, m := range r.Media {
		if m.video() {
			continue
		}
		count++
	}
	return count
}

type mbzTag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type mbzGenre struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type mbzRating struct {
	Votes int     `json:"votes-count"`
	Value float32 `json:"value"`
}

type mbzReleaseGroup struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	Disambiguation   string            `json:"disambiguation"`
	PrimaryType      string            `json:"primary-type"`
	Rating           mbzRating         `json:"rating"`
	FirstReleaseDate string            `json:"first-release-date"`
	Tags             []mbzTag          `json:"tags"`
	Genres           []mbzGenre        `json:"genres"`
	Releases         []mbzRelease      `json:"releases"`
	ArtistCredit     []mbzArtistCredit `json:"artist-credit"`
	Relations        []mbzRelation     `json:"relations"`
}

func (rg mbzReleaseGroup) firstReleaseDate() time.Time {
	return parseDate(rg.FirstReleaseDate)
}

func release(a *Artist, r mbzRelease) Release {
	disambiguation := r.Disambiguation
	if disambiguation == "" {
		disambiguation = r.ReleaseGroup.Disambiguation
	}

	var media []Media
	for _, m := range r.Media {
		media = append(media, Media{
			REID:       string(r.ID),
			Name:       m.Title,
			Position:   m.Position,
			Format:     m.Format,
			TrackCount: m.TrackCount})
	}

	return Release{
		Artist:         a.Name,
		Name:           r.Title,
		Disambiguation: disambiguation,
		REID:           string(r.ID),
		RGID:           string(r.ReleaseGroup.ID),
		Type:           r.ReleaseGroup.PrimaryType,
		Asin:           r.Asin,
		Country:        r.Country,
		TrackCount:     r.totalTracks(),
		DiscCount:      r.totalDiscs(),
		Artwork:        r.CoverArtArchive.Artwork,
		FrontArtwork:   r.CoverArtArchive.Front,
		BackArtwork:    r.CoverArtArchive.Back,
		Media:          media,
		Date:           r.ReleaseGroup.firstReleaseDate()}
}

// Get all releases for an artist from MusicBrainz.
func (m *Music) MusicBrainzArtistReleases(a *Artist) ([]Release, error) {
	var releases []Release
	limit, offset := 100, 0
	for {
		result, _ := m.doArtistReleases(a.ARID, limit, offset)
		for _, r := range result.Releases {
			releases = append(releases, release(a, r))
		}
		offset += len(result.Releases)
		if offset >= result.Count {
			break
		}
	}

	return releases, nil
}

func (m *Music) doArtistReleases(arid string, limit int, offset int) (*mbzReleasesPage, error) {
	var result mbzReleasesPage
	inc := []string{"release-groups", "media"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release?fmt=json&artist=%s&inc=%s&limit=%d&offset=%d",
		arid, strings.Join(inc, "%2B"), limit, offset)
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *Music) MusicBrainzRelease(reid string) (*mbzRelease, error) {
	inc := []string{"aliases", "artist-credits", "labels",
		"discids", "recordings", "artist-rels",
		"release-groups", "genres", "tags", "ratings",
		"recording-level-rels", "series-rels", "work-rels", "work-level-rels"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release/%s?fmt=json&inc=%s",
		reid, strings.Join(inc, "%2B"))
	var result mbzRelease
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *Music) MusicBrainzReleaseGroup(rgid string) (*mbzReleaseGroup, error) {
	inc := []string{"releases", "media", "release-group-rels",
		"genres", "tags", "ratings", "series-rels"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release-group/%s?fmt=json&inc=%s",
		rgid, strings.Join(inc, "%2B"))
	var result mbzReleaseGroup
	err := m.client.GetJson(url, &result)
	for _, r := range result.Releases {
		if r.Title == "" {
			r.Title = result.Title
		}
	}
	return &result, err
}

func (m *Music) MusicBrainzReleases(a *Artist, rgid string) ([]Release, error) {
	var releases []Release
	rg, err := m.MusicBrainzReleaseGroup(rgid)
	if err != nil {
		return releases, err
	}
	for _, r := range rg.Releases {
		r.ReleaseGroup = *rg
		releases = append(releases, release(a, r))
	}
	return releases, nil
}

type mbzSearchResult struct {
	Created       string            `json:"created"`
	Count         int               `json:"count"`
	Offset        int               `json:"offset"`
	ReleaseGroups []mbzReleaseGroup `json:"release-groups"`
}

func (m *Music) MusicBrainzSearchReleaseGroup(arid string, name string) (*mbzSearchResult, error) {
	url := fmt.Sprintf(
		`https://musicbrainz.org/ws/2/release-group/?fmt=json&query=arid:%s+AND+release:"%s"`,
		arid, url.QueryEscape(name))
	var result mbzSearchResult
	err := m.client.GetJson(url, &result)
	return &result, err
}

func doArtist(artist mbzArtist) (a *Artist, tags []ArtistTag) {
	a = &Artist{
		Name:     artist.Name,
		SortName: artist.SortName,
		ARID:     string(artist.ID)}
	for _, t := range artist.Tags {
		at := ArtistTag{
			Artist: a.Name,
			Tag:    t.Name,
			Count:  t.Count}
		tags = append(tags, at)
	}
	return
}

// Obtain artist details using MusicBrainz artist ID.
func (m *Music) MusicBrainzSearchArtistID(arid string) (a *Artist, tags []ArtistTag) {
	query := fmt.Sprintf(`arid:%s`, arid)
	result, _ := m.doArtistSearch(query, 100, 0)
	if len(result.Artists) == 0 {
		return
	}
	a, tags = doArtist(result.Artists[0])
	return
}

// Search for artist by name using MusicBrainz.
func (m *Music) MusicBrainzSearchArtist(name string) (a *Artist, tags []ArtistTag) {
	var artists []mbzArtist
	limit, offset := 100, 0

	// can also add "AND type:group" or "AND type:person"
	query := fmt.Sprintf(`artist:"%s"`, name)
	result, _ := m.doArtistSearch(query, limit, offset)
	for _, r := range result.Artists {
		artists = append(artists, r)
	}

	score := 100 // change to widen matches below
	artists = scoreFilter(artists, score)
	if len(artists) == 0 {
		return
	}

	pick := 0
	if len(artists) > 1 {
		// multiple matches
		for index, artist := range artists {
			fmt.Printf("ID: %s Name: %-25sScore: %d\n",
				artist.ID, artist.Name, artist.Score)
			if strings.EqualFold(name, artist.Name) {
				// try to use a close match
				pick = index
			}
		}
	}
	artist := artists[pick]
	a, tags = doArtist(artist)
	return
}

func scoreFilter(artists []mbzArtist, score int) []mbzArtist {
	result := []mbzArtist{}
	for _, v := range artists {
		if v.Score >= score {
			result = append(result, v)
		}
	}
	return result
}

func (m *Music) doArtistSearch(query string, limit int, offset int) (*mbzArtistsPage, error) {
	var result mbzArtistsPage
	url := fmt.Sprintf(`https://musicbrainz.org/ws/2/artist?fmt=json&query=%s&limit=%d&offset=%d`,
		url.QueryEscape(query), limit, offset)
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *Music) MusicBrainzArtistDetail(a *Artist) (*mbzArtist, error) {
	var result mbzArtist
	url := fmt.Sprintf(`http://musicbrainz.org/ws/2/artist/%s?fmt=json&inc=genres+url-rels`,
		a.ARID)
	err := m.client.GetJson(url, &result)
	return &result, err
}

type coverArtImage struct {
	RawID    json.RawMessage `json:"id"`
	ID       string          `json:"-"`
	Approved bool            `json:"approved"`
	Front    bool            `json:"front"`
	Back     bool            `json:"back"`
	Image    string          `json:"image"`
	// 250, 500, 1200, small (250), large (500)
	Thumbnails map[string]string `json:"thumbnails"`
}

type coverArt struct {
	Release string          `json:"release"`
	Images  []coverArtImage `json:"images"`
}

func (m *Music) coverArtArchive(reid string) (*coverArt, error) {
	var result coverArt
	url := fmt.Sprintf(`https://coverartarchive.org/release/%s`, reid)
	err := m.client.GetJson(url, &result)
	if err != nil {
		return &result, err
	}
	for i, img := range result.Images {
		// api has ID with both int and string types
		// id: 42
		// id: "42"
		result.Images[i].ID = unquote(string(img.RawID))
	}
	return &result, err
}
