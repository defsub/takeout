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

package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	"github.com/defsub/takeout/lib/date"
)

type MusicBrainz struct {
	client *client.Client
}

func NewMusicBrainz(config *config.Config) *MusicBrainz {
	return &MusicBrainz{
		client: client.NewClient(config),
	}
}

// MusicBrainz is used for:
// * getting the MBID for artists
// * correcting artist names
// * correcting release names
// * getting release types (Album, Single, EP, Broadcast, Other)
// * getting original release dates (release groups)
// * getting releases within groups that have different titles

type ArtistsPage struct {
	Artists []Artist `json:"artists"`
	Offset  int      `json:"offset"`
	Count   int      `json:"count"`
}

// TODO artist detail
// get detail for each artist - lifespan, url rel links, genres
// json api:
// http://musicbrainz.org/ws/2/artist/3798b104-01cb-484c-a3b0-56adc6399b80?inc=genres+url-rels&fmt=json
type LifeSpan struct {
	Ended bool   `json:"ended"`
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// TODO artist detail
type Area struct {
	Name     string `json:"name"`
	SortName string `json:"sort-name"`
}

type Artist struct {
	ID             string     `json:"id"`
	Score          int        `json:"score"`
	Name           string     `json:"name"`
	SortName       string     `json:"sort-name"`
	Country        string     `json:"country"`
	Disambiguation string     `json:"disambiguation"`
	Type           string     `json:"type"`
	Genres         []Genre    `json:"genres"`
	Tags           []Tag      `json:"tags"`
	Area           Area       `json:"area"`
	BeginArea      Area       `json:"begin-area"`
	EndArea        Area       `json:"end-area"`
	LifeSpan       LifeSpan   `json:"life-span"`
	Relations      []Relation `json:"relations"`
}

func doSort(genres []Genre) []Genre {
	sort.Slice(genres, func(i, j int) bool {
		return genres[i].Count > genres[j].Count
	})
	return genres
}

func (a Artist) SortedGenres() []Genre {
	return doSort(a.Genres)
}

func (a Artist) PrimaryGenre() string {
	if len(a.Genres) > 0 {
		return doSort(a.Genres)[0].Name
	}
	return "" // TODO default to something?
}

type ArtistCredit struct {
	Name   string `json:"name"`
	Join   string `json:"joinphrase"`
	Artist Artist `json:"artist"`
}

type Work struct {
	Title     string     `json:"title"`
	Relations []Relation `json:"relations"`
}

// TODO artist detail
type URL struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
}

// type="Release group series"
// type="Recording series" (for recording in release)
type Series struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Disambiguation string `json:"disambiguation"`
	Type           string `json:"type"`
}

type AttributeIds struct {
	Cover string `json:"cover"` // 1e8536bd-6eda-3822-8e78-1c0f4d3d2113
	Live  string `json:"live"`  // 70007db6-a8bc-46d7-a770-80e6a0bb551a
}

// release-group series: type="part of", target-type="series", see series
// release recording series: type="part of", target-type="series", see series
// single: type="single from", target-type="release_group", see release_group
type Relation struct {
	Type         string       `json:"type"`
	TargetType   string       `json:"target-type"`
	Artist       Artist       `json:"artist"`
	Attributes   []string     `json:"attributes"`
	AttributeIds AttributeIds `json:"attribute-ids"`
	Work         Work         `json:"work"`
	URL          URL          `json:"url"`
	Series       Series       `json:"series"`
}

type LabelInfo struct {
	Label         Label  `json:"label"`
	CatalogNumber string `json:"catalog-number"`
}

type Label struct {
	Name     string `json:"name"`
	SortName string `json:"sort-name"`
}

type Media struct {
	Title      string  `json:"title"`
	Format     string  `json:"format"`
	Position   int     `json:"position"`
	TrackCount int     `json:"track-count"`
	Tracks     []Track `json:"tracks"`
}

func (m Media) video() bool {
	switch m.Format {
	case "DVD-Video", "Blu-ray", "HD-DVD", "VCD", "SVCD":
		return true
	}
	return false
}

type Recording struct {
	ID               string         `json:"id"`
	Length           int            `json:"length"`
	Title            string         `json:"title"`
	Relations        []Relation     `json:"relations"`
	ArtistCredit     []ArtistCredit `json:"artist-credit"`
	FirstReleaseDate string         `json:"first-release-date"`
}

func (r Recording) FirstReleaseTime() time.Time {
	return date.ParseDate(r.FirstReleaseDate)
}

type Track struct {
	Title        string         `json:"title"`
	Position     int            `json:"position"`
	ArtistCredit []ArtistCredit `json:"artist-credit"`
	Recording    Recording      `json:"recording"`
}

type ReleasesPage struct {
	Releases []Release `json:"releases"`
	Offset   int       `json:"release-offset"`
	Count    int       `json:"release-count"`
}

type CoverArtArchive struct {
	Count    int  `json:"count"`
	Artwork  bool `json:"artwork"`
	Front    bool `json:"front"`
	Back     bool `json:"back"`
	Darkened bool `json:"darkened"`
}

type Release struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Date            string          `json:"date"`
	Disambiguation  string          `json:"disambiguation"`
	Country         string          `json:"country"`
	Status          string          `json:"status"`
	Asin            string          `json:"asin"`
	Relations       []Relation      `json:"relations"`
	LabelInfo       []LabelInfo     `json:"label-info"`
	Media           []Media         `json:"media"`
	ReleaseGroup    ReleaseGroup    `json:"release-group"`
	CoverArtArchive CoverArtArchive `json:"cover-art-archive"`
	ArtistCredit    []ArtistCredit  `json:"artist-credit"`
}

func (r Release) title() string {
	if r.Disambiguation != "" {
		return fmt.Sprintf("%s (%s)", r.Title, r.Disambiguation)
	} else {
		return r.Title
	}
}

func (r Release) TotalTracks() int {
	count := 0
	for _, m := range r.Media {
		if m.video() {
			continue
		}
		count += m.TrackCount
	}
	return count
}

func (r Release) TotalDiscs() int {
	count := 0
	for _, m := range r.Media {
		if m.video() {
			continue
		}
		count++
	}
	return count
}

type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Genre struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Rating struct {
	Votes int     `json:"votes-count"`
	Value float32 `json:"value"`
}

type ReleaseGroup struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	Disambiguation   string         `json:"disambiguation"`
	PrimaryType      string         `json:"primary-type"`
	SecondaryTypes   []string       `json:"secondary-types"`
	Rating           Rating         `json:"rating"`
	FirstReleaseDate string         `json:"first-release-date"`
	Tags             []Tag          `json:"tags"`
	Genres           []Genre        `json:"genres"`
	Releases         []Release      `json:"releases"`
	ArtistCredit     []ArtistCredit `json:"artist-credit"`
	Relations        []Relation     `json:"relations"`
}

func (rg ReleaseGroup) FirstReleaseTime() time.Time {
	return date.ParseDate(rg.FirstReleaseDate)
}

func (rg ReleaseGroup) SortedGenres() []Genre {
	return doSort(rg.Genres)
}

const (
	PrimaryTypeAlbum     = "Album"
	PrimaryTypeSingle    = "Single"
	PrimaryTypeEP        = "EP"
	PrimaryTypeBroadcast = "Broadcast"
	PrimaryTypeOther     = "Other"

	TypeCompilation   = "Compilation"
	TypeSoundtrack    = "Soundtrack"
	TypeSpokenword    = "Spokenword"
	TypeInterview     = "Interview"
	TypeAudiobook     = "Audiobook"
	TypeAudioDrama    = "Audio drama"
	TypeLive          = "Live"
	TypeRemix         = "Remix"
	TypeDJMix         = "DJ-mix"
	TypeMixtapeStreet = "Mixtape/Street"
	TypeNone          = ""
)

var preferredTypes = []string{
	TypeSoundtrack,
	TypeCompilation,
	TypeRemix,
	TypeLive,
}

func (rg ReleaseGroup) SecondaryType() string {
	// none or one
	switch len(rg.SecondaryTypes) {
	case 0:
		return TypeNone
	case 1:
		return rg.SecondaryTypes[0]
	}
	// preferred
	types := make(map[string]bool)
	for _, v := range rg.SecondaryTypes {
		types[v] = true
	}
	for _, v := range preferredTypes {
		_, ok := types[v]
		if ok {
			return v
		}
	}
	// fallback to first
	return rg.SecondaryTypes[0]
}

// func release(artist string, r Release) music.Release {
// 	disambiguation := r.Disambiguation
// 	if disambiguation == "" {
// 		disambiguation = r.ReleaseGroup.Disambiguation
// 	}

// 	var media []music.Media
// 	for _, m := range r.Media {
// 		media = append(media, music.Media{
// 			REID:       string(r.ID),
// 			Name:       m.Title,
// 			Position:   m.Position,
// 			Format:     m.Format,
// 			TrackCount: m.TrackCount})
// 	}

// 	return music.Release{
// 		Artist:         artist,
// 		Name:           r.Title,
// 		Disambiguation: disambiguation,
// 		REID:           string(r.ID),
// 		RGID:           string(r.ReleaseGroup.ID),
// 		Type:           r.ReleaseGroup.PrimaryType,
// 		Asin:           r.Asin,
// 		Country:        r.Country,
// 		TrackCount:     r.totalTracks(),
// 		DiscCount:      r.totalDiscs(),
// 		Artwork:        r.CoverArtArchive.Artwork,
// 		FrontArtwork:   r.CoverArtArchive.Front,
// 		BackArtwork:    r.CoverArtArchive.Back,
// 		Media:          media,
// 		Date:           r.ReleaseGroup.firstReleaseDate(),
// 		ReleaseDate:    date.ParseDate(r.Date),
// 		Status:         r.Status,
// 	}
// }

// Get all releases for an artist from MusicBrainz.
func (m *MusicBrainz) ArtistReleases(artist, arid string) ([]Release, error) {
	var releases []Release
	limit, offset := 100, 0
	for {
		result, _ := m.doArtistReleases(arid, limit, offset)
		for _, r := range result.Releases {
			//releases = append(releases, release(artist, r))
			releases = append(releases, r)
		}
		offset += len(result.Releases)
		if offset >= result.Count {
			break
		}
	}

	return releases, nil
}

func (m *MusicBrainz) doArtistReleases(arid string, limit int, offset int) (*ReleasesPage, error) {
	var result ReleasesPage
	inc := []string{"release-groups", "media"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release?fmt=json&artist=%s&inc=%s&limit=%d&offset=%d",
		arid, strings.Join(inc, "%2B"), limit, offset)
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *MusicBrainz) Release(reid string) (*Release, error) {
	inc := []string{"aliases", "artist-credits", "labels",
		"discids", "recordings", "artist-rels",
		"release-groups", "genres", "tags", "ratings",
		"recording-level-rels", "series-rels", "work-rels", "work-level-rels"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release/%s?fmt=json&inc=%s",
		reid, strings.Join(inc, "%2B"))
	var result Release
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *MusicBrainz) ReleaseGroup(rgid string) (*ReleaseGroup, error) {
	inc := []string{"releases", "media", "release-group-rels",
		"genres", "tags", "ratings", "series-rels"}
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release-group/%s?fmt=json&inc=%s",
		rgid, strings.Join(inc, "%2B"))
	var result ReleaseGroup
	err := m.client.GetJson(url, &result)
	for _, r := range result.Releases {
		if r.Title == "" {
			r.Title = result.Title
		}
	}
	return &result, err
}

func (m *MusicBrainz) Releases(rgid string) ([]Release, error) {
	var releases []Release
	rg, err := m.ReleaseGroup(rgid)
	if err != nil {
		return releases, err
	}
	for _, r := range rg.Releases {
		r.ReleaseGroup = *rg
		//releases = append(releases, release(a, r))
		releases = append(releases, r)
	}
	return releases, nil
}

type SearchResult struct {
	Created       string         `json:"created"`
	Count         int            `json:"count"`
	Offset        int            `json:"offset"`
	ReleaseGroups []ReleaseGroup `json:"release-groups"`
}

func (m *MusicBrainz) SearchReleaseGroup(arid string, name string) (*SearchResult, error) {
	url := fmt.Sprintf(
		`https://musicbrainz.org/ws/2/release-group/?fmt=json&query=arid:%s+AND+release:"%s"`,
		arid, url.QueryEscape(name))
	var result SearchResult
	err := m.client.GetJson(url, &result)
	return &result, err
}

// func doArtist(artist Artist) (a *music.Artist, tags []music.ArtistTag) {
// 	a = &music.Artist{
// 		Name:     artist.Name,
// 		SortName: artist.SortName,
// 		ARID:     string(artist.ID)}
// 	for _, t := range artist.Tags {
// 		at := music.ArtistTag{
// 			Artist: a.Name,
// 			Tag:    t.Name,
// 			Count:  t.Count}
// 		tags = append(tags, at)
// 	}
// 	return
// }

// Obtain artist details using MusicBrainz artist ID.
func (m *MusicBrainz) SearchArtistID(arid string) *Artist /*(a *music.Artist, tags []music.ArtistTag)*/ {
	query := fmt.Sprintf(`arid:%s`, arid)
	result, _ := m.doArtistSearch(query, 100, 0)
	if len(result.Artists) == 0 {
		return nil
	}
	//a, tags = doArtist(result.Artists[0])
	return &result.Artists[0]
}

// Search for artist by name using MusicBrainz.
func (m *MusicBrainz) SearchArtist(name string) *Artist /*(a *music.Artist, tags []music.ArtistTag)*/ {
	var artists []Artist
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
		return nil
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
	//a, tags = doArtist(artist)
	return &artist
}

func scoreFilter(artists []Artist, score int) []Artist {
	result := []Artist{}
	for _, v := range artists {
		//fmt.Printf("%d %s\n", v.Score, v.Name)
		if v.Score >= score {
			result = append(result, v)
		}
	}
	return result
}

func (m *MusicBrainz) doArtistSearch(query string, limit int, offset int) (*ArtistsPage, error) {
	var result ArtistsPage
	url := fmt.Sprintf(`https://musicbrainz.org/ws/2/artist?fmt=json&query=%s&limit=%d&offset=%d`,
		url.QueryEscape(query), limit, offset)
	fmt.Println(url)
	err := m.client.GetJson(url, &result)
	return &result, err
}

func (m *MusicBrainz) ArtistDetail(arid string) (*Artist, error) {
	var result Artist
	url := fmt.Sprintf(`http://musicbrainz.org/ws/2/artist/%s?fmt=json&inc=genres+url-rels`,
		arid)
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
	FromGroup bool
	Release   string          `json:"release"`
	Images    []coverArtImage `json:"images"`
}

var doubleQuotedRegexp = regexp.MustCompile(`"(.*)"`)

func unquote(s string) string {
	matches := doubleQuotedRegexp.FindStringSubmatch(s)
	if matches != nil {
		return matches[1]
	}
	return s
}

func (m *MusicBrainz) CoverArtArchive(reid string, rgid string) (*coverArt, error) {
	var result coverArt
	result.FromGroup = false
	// try release first
	url := fmt.Sprintf(`https://coverartarchive.org/release/%s`, reid)
	err := m.client.GetJson(url, &result)
	if err != nil {
		// can get 404 for direct checks
		// try release-group instead
		url = fmt.Sprintf(`https://coverartarchive.org/release-group/%s`, rgid)
		err = m.client.GetJson(url, &result)
		if err != nil {
			// can get 404 for direct checks
			return &result, err
		}
		result.FromGroup = true
	}
	for i, img := range result.Images {
		// api has ID with both int and string types
		// id: 42
		// id: "42"
		result.Images[i].ID = unquote(string(img.RawID))
	}
	return &result, err
}
