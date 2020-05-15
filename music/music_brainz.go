// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// Takeout is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package music

import (
	"fmt"
	"github.com/defsub/takeout"
	"github.com/michiwend/gomusicbrainz"
	"net/http"
	"net/url"
	"strings"
	"time"
	"encoding/xml"
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

func sleep() {
	time.Sleep(1 * time.Second)
}

func userAgent() string {
	return takeout.AppName + "/" + takeout.Version + " ( " + takeout.Contact + " ) "
}

func newMusicBrainzClient() *gomusicbrainz.WS2Client {
	client, _ := gomusicbrainz.NewWS2Client(
		"https://musicbrainz.org/ws/2",
		takeout.AppName, takeout.Version, takeout.Contact)
	return client
}

func (m *Music) SearchArtistId(mbid string) (a *Artist, tags []ArtistTag) {
	client := newMusicBrainzClient()
	resp, _ := client.SearchArtist(fmt.Sprintf(`arid:"%s"`, mbid), -1, -1)
	artists := resp.ResultsWithScore(100)
	if len(artists) == 0 {
		return
	}
	artist := artists[0]

	a = &Artist{
		Name:     artist.Name,
		SortName: artist.SortName,
		MBID:     string(artist.ID)}

	for _, t := range artist.Tags {
		at := ArtistTag{
			Artist: a.Name,
			Tag:    t.Name,
			Count:  uint(t.Count)}
		tags = append(tags, at)
	}
	return
}

func (m *Music) SearchArtist(name string) (a *Artist, tags []ArtistTag) {
	client := newMusicBrainzClient()
	sleep()

	var query string
	mbid, ok := m.config.Music.UserArtistID(name)
	if ok {
		query = fmt.Sprintf(`arid:"%s"`, mbid)
		fmt.Printf("%s using %s\n", name, mbid)
	} else {
		query = fmt.Sprintf(`artist:"%s"`, name)
	}
	resp, _ := client.SearchArtist(query, -1, -1)
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
		MBID:     string(artist.ID)}

	for _, t := range artist.Tags {
		at := ArtistTag{
			Artist: a.Name,
			Tag:    t.Name,
			Count:  uint(t.Count)}
		tags = append(tags, at)
	}
	return
}


func (m *Music) MusicBrainzRelease(a *Artist, mbid string) (*Release, error) {
	sleep()
	client := newMusicBrainzClient()
	fmt.Printf("lookup %s %s\n", a.Name, mbid)
	r, err := client.LookupRelease(gomusicbrainz.MBID(mbid), "release-groups")
	if err != nil {
		return nil, err
	}
	release := &Release{
		Artist: a.Name,
		Name:   r.Title,
		Date:   r.Date.Time,
		MBID:   string(r.ID),
		Type:   r.ReleaseGroup.PrimaryType,
		Asin:   r.Asin}
	return release, err
}

type releaseGroupListResult struct {
	ReleaseGroupList struct {
		gomusicbrainz.WS2ListResponse
		ReleaseGroups []struct {
			*gomusicbrainz.ReleaseGroup
			Score int `xml:"http://musicbrainz.org/ns/ext#-2.0 score,attr"`
		} `xml:"release-group"`
	} `xml:"release-group-list"`
}

func (m *Music) MusicBrainzReleaseGroups(a *Artist) ([]Release, error) {
	var releases []Release
	limit, offset := 100, 0
	for {
		sleep()
		result, _ := doArtistReleaseGroups(a.MBID, limit, offset)
		//fmt.Printf("got %d %d %d\n", offset, result.ReleaseGroupList.Count, result.ReleaseGroupList.Offset)
		for _, r := range result.ReleaseGroupList.ReleaseGroups {
			//fmt.Printf("%s %s %s\n", r.ID, r.Title, r.FirstReleaseDate)
			releases = append(releases, Release{
				Artist: a.Name,
				Name: r.Title,
				MBID: string(r.ID),
				Type: r.PrimaryType,
				Date: r.FirstReleaseDate.Time})
		}
		offset += len(result.ReleaseGroupList.ReleaseGroups)
		if offset >= result.ReleaseGroupList.Count {
			break
		}
	}

	return releases, nil
}

// gomusicbrainz uses search which doesn't return relevant information
// so this uses browse instead.
func doArtistReleaseGroups(mbid string, limit int, offset int) (*releaseGroupListResult, error) {
	url, _ := url.Parse(
		fmt.Sprintf("http://musicbrainz.org/ws/2/release-group?artist=%s&limit=%d&offset=%d",
			mbid, limit, offset))
	client := &http.Client{}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result releaseGroupListResult
	decoder := xml.NewDecoder(resp.Body)
	if err = decoder.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *Music) MusicBrainzReleases(a *Artist, release string) []Release {
	var releases []Release
	results := doReleaseSearch(fmt.Sprintf(`arid:"%s" AND release:"%s"`, a.MBID, release))
	for _, r := range results {
		releases = append(releases, Release{
			Artist: a.Name,
			Name:   r.Title,
			Date:   r.Date.Time,
			MBID:   string(r.ID),
			Type:   r.ReleaseGroup.PrimaryType,
			Asin:   r.Asin})
	}
	return releases
}

func doReleaseSearch(query string) []*gomusicbrainz.Release {
	var releases []*gomusicbrainz.Release
	client := newMusicBrainzClient()
	limit, offset := 100, 0
	dup := 0
	ids := make(map[string]string)
	for {
		sleep()
		resp, err := client.SearchRelease(query, limit, offset)
		if err != nil {
			fmt.Printf("SearchRelease %s\n", err)
			continue
		}
		//releases = append(releases, resp.Releases...)
		for _, r := range resp.Releases {
			mbid := string(r.ID)
			_, duplicate := ids[mbid]
			if !duplicate {
				ids[mbid] = r.Title
				releases = append(releases, r)
			} else {
				//fmt.Printf("dup %d/%d %s %s -> %s\n", i+offset, len(resp.Releases), mbid, r.Title, x)
				dup = dup + 1
			}
		}
		offset += len(resp.Releases)
		if offset >= resp.Count {
			//fmt.Printf("done %d %d %d\n", offset, len(resp.Releases), resp.Count)
			break
		}
	}

	//fmt.Printf("complete %d %d dups=%d\n", offset, len(releases), dup)
	// for k, v := range ids {
	// 	if strings.HasPrefix(v, "Babylon") || strings.HasPrefix(v, "Archive") || strings.HasPrefix(v, "Asylum") {
	// 		fmt.Printf("%s %s\n", k, v)
	// 	}
	// }

	return releases
}
