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
	"github.com/defsub/takeout"
	"github.com/defsub/takeout/client"
	"github.com/defsub/takeout/log"
	"github.com/michiwend/gomusicbrainz"
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
// MusicBrainz is *not* used for:
// * getting the exact MBID for a release
// * correcting track titles

func newMusicBrainzClient() *gomusicbrainz.WS2Client {
	mbz, _ := gomusicbrainz.NewWS2Client(
		"https://musicbrainz.org/ws/2",
		takeout.AppName, takeout.Version, takeout.Contact)
	return mbz
}

// Obtain artist details using MusicBrainz artist ID.
//
// TODO replace with internal client
func (m *Music) SearchArtistId(arid string) (a *Artist, tags []ArtistTag) {
	mbz := newMusicBrainzClient()
	resp, _ := mbz.SearchArtist(fmt.Sprintf(`arid:"%s"`, arid), -1, -1)
	artists := resp.ResultsWithScore(100)
	if len(artists) == 0 {
		return
	}
	artist := artists[0]

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

// Search MusicBrainz for an artist by name. There can be duplicates
// or unintended matches. For those cases it's best to override using
// config with the specific MusicBrainz artist ID that is needed.
//
// TODO replace with internal client
func (m *Music) SearchArtist(name string) (a *Artist, tags []ArtistTag) {
	client.RateLimit()
	mbz := newMusicBrainzClient()

	var query string
	mbid, ok := m.config.Music.UserArtistID(name)
	if ok {
		query = fmt.Sprintf(`arid:"%s"`, mbid)
		log.Printf("%s using %s\n", name, mbid)
	} else {
		query = fmt.Sprintf(`artist:"%s"`, name)
	}
	resp, _ := mbz.SearchArtist(query, -1, -1)
	score := 100 // change to widen matches below
	artists := resp.ResultsWithScore(score)
	if len(artists) == 0 {
		return
	}

	pick := 0
	if len(artists) > 1 {
		// multiple matches within results
		for index, artist := range artists {
			fmt.Printf("ID: %s Name: %-25sScore: %d\n",
				artist.ID, artist.Name, resp.Scores[artist])
			if strings.EqualFold(name, artist.Name) {
				// try to use a close match
				pick = index
			}
		}
	}
	artist := artists[pick]

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

type mbzArtist struct {
	Name           string     `json:"name"`
	SortName       string     `json:"sort-name"`
	Disambiguation string     `json:"disambiguation"`
	Type           string     `json:"type"`
	Genres         []mbzGenre `json:"genres"`
	Tags           []mbzTag   `json:"tags"`
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

type mbzRelation struct {
	Type       string    `json:"type"`
	Artist     mbzArtist `json:"artist"`
	Attributes []string  `json:"attributes"`
	Work       mbzWork   `json:"work"`
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
	Front    bool `json:"front"`
	Artwork  bool `json:"artwork"`
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
		TrackCount:     r.totalTracks(),
		FrontCover:     r.CoverArtArchive.Front,
		Media:          media,
		Date:           r.ReleaseGroup.firstReleaseDate()}
}

// Get all releases for an artist from MusicBrainz.
func (m *Music) MusicBrainzArtistReleases(a *Artist) ([]Release, error) {
	var releases []Release
	limit, offset := 100, 0
	for {
		result, _ := doArtistReleases(a.ARID, limit, offset)
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

func doArtistReleases(arid string, limit int, offset int) (*mbzReleasesPage, error) {
	var result mbzReleasesPage
	inc := []string{"release-groups", "media"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release?fmt=json&artist=%s&inc=%s&limit=%d&offset=%d",
		arid, strings.Join(inc, "%2B"), limit, offset)
	client.RateLimit()
	err := client.GetJson(url, &result)
	return &result, err
}

func (m *Music) MusicBrainzReleaseCredits(reid string) (*mbzRelease, error) {
	inc := []string{"aliases", "artist-credits", "labels",
		"discids", "recordings", "artist-rels",
		"release-groups", "genres", "tags", "ratings",
		"recording-level-rels", "work-rels", "work-level-rels"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release/%s?fmt=json&inc=%s",
		reid, strings.Join(inc, "%2B"))
	var result mbzRelease
	client.RateLimit()
	err := client.GetJson(url, &result)
	return &result, err
}

func (m *Music) MusicBrainzReleaseGroup(rgid string) (*mbzReleaseGroup, error) {
	inc := []string{"releases", "media", "release-group-rels",
		"genres", "tags", "ratings"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release-group/%s?fmt=json&inc=%s",
		rgid, strings.Join(inc, "%2B"))
	var result mbzReleaseGroup
	client.RateLimit()
	err := client.GetJson(url, &result)
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
		arid, strings.Replace(name, " ", "+", -1))
	var result mbzSearchResult
	client.RateLimit()
	err := client.GetJson(url, &result)
	return &result, err
}
