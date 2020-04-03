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
	"github.com/michiwend/gomusicbrainz"
	"time"
)

func sleep() {
	time.Sleep(1 * time.Second)
}

func newMusicBrainzClient() *gomusicbrainz.WS2Client {
	client, _ := gomusicbrainz.NewWS2Client(
		"https://musicbrainz.org/ws/2",
		"A GoMusicBrainz example",
		"0.0.1-beta",
		"http://github.com/michiwend/gomusicbrainz")
	return client
}

func (m *Music) SearchArtist(name string) (a *Artist, tags []ArtistTag) {
	client := newMusicBrainzClient()
	sleep()
	resp, _ := client.SearchArtist(fmt.Sprintf(`artist:"%s"`, name), -1, -1)
	artists := resp.ResultsWithScore(100)
	for _, artist := range artists {
		fmt.Printf("ID: %s Name: %-25sScore: %d\n", artist.ID, artist.Name, resp.Scores[artist])
		a = &Artist{Name: artist.Name, MBID: string(artist.ID)}
		for _, t := range artist.Tags {
			at := ArtistTag{Artist: a.Name, Tag: t.Name, Count: uint(t.Count)}
			tags = append(tags, at)
		}
		break
	}
	return
}

func (m *Music) searchReleases(a *Artist) []Release {
	var releases []Release
	results := doReleaseSearch(fmt.Sprintf(`arid:"%s"`, a.MBID))
	for _, r := range results {
		if r.Date.Time.Year() == 0 {
			continue
		}
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

// not used anymore
func doReleaseGroupSearch(query string) []*gomusicbrainz.ReleaseGroup {
	var releaseGroups []*gomusicbrainz.ReleaseGroup
	client := newMusicBrainzClient()
	limit, offset := 100, 0
	for {
		sleep()
		resp, _ := client.SearchReleaseGroup(query, limit, offset)
		releaseGroups = append(releaseGroups, resp.ReleaseGroups...)
		offset += len(resp.ReleaseGroups)
		if offset >= resp.Count {
			break
		}
	}
	return releaseGroups
}

func doReleaseSearch(query string) []*gomusicbrainz.Release {
	var releases []*gomusicbrainz.Release
	client := newMusicBrainzClient()
	limit, offset := 100, 0
	for {
		sleep()
		resp, _ := client.SearchRelease(query, limit, offset)
		releases = append(releases, resp.Releases...)
		offset += len(resp.Releases)
		if offset >= resp.Count {
			break
		}
	}
	return releases
}
