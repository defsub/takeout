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
	"github.com/defsub/takeout/client"
	"github.com/shkh/lastfm-go/lastfm"
	"strconv"
	"sort"
)

// Lastfm is used for:
// * getting popular tracks for each artist
// * getting similar artists for each artist
// * looking up artists not found by MusicBrainz to get their MBID

func (m *Music) lastfmArtistTopTracks(artist *Artist) []Popular {
	client.RateLimit("last.fm")
	api := lastfm.New(m.config.LastFM.Key, m.config.LastFM.Secret)

	result, _ := api.Artist.GetTopTracks(lastfm.P{"mbid": artist.ARID})
	sort.Slice(result.Tracks, func(i, j int) bool {
		a, _ := strconv.Atoi(result.Tracks[i].PlayCount)
		b, _ := strconv.Atoi(result.Tracks[j].PlayCount)
		return a > b
	})

	var popular []Popular
	for _, track := range result.Tracks {
		rank, _ := strconv.Atoi(track.Rank)
		popular = append(popular, Popular{Artist: artist.Name, Title: track.Name, Rank: rank})
	}

	return popular
}

func (m *Music) lastfmSimilarArtists(artist *Artist) []Similar {
	client.RateLimit("last.fm")
	api := lastfm.New(m.config.LastFM.Key, m.config.LastFM.Secret)
	result, _ := api.Artist.GetSimilar(lastfm.P{"mbid": artist.ARID})

	var mbids []string
	rank := make(map[string]float64)
	for _, similar := range result.Similars {
		mbids = append(mbids, similar.Mbid)
		rank[similar.Mbid], _ = strconv.ParseFloat(similar.Match, 64)
	}

	artists := m.artistsByMBID(mbids)
	sort.Slice(artists, func(i, j int) bool {
		return rank[artists[i].ARID] > rank[artists[j].ARID]
	})

	var similar []Similar
	for index, a := range artists {
		similar = append(similar, Similar{Artist: artist.Name, ARID: a.ARID, Rank: index})
	}

	return similar
}

func (m *Music) lastfmArtistSearch(name string) *Artist {
	client.RateLimit("last.fm")
	api := lastfm.New(m.config.LastFM.Key, m.config.LastFM.Secret)
	result, _ := api.Artist.Search(lastfm.P{"artist": name})

	var artist *Artist
	for index, match := range result.ArtistMatches {
		if index == 0 {
			artist = &Artist{Name: match.Name, ARID: match.Mbid}
			break;
		}
	}

	return artist
}
