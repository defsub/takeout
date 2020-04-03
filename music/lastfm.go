package music

import (
	"fmt"
	"github.com/shkh/lastfm-go/lastfm"
	"github.com/defsub/takeout/config"
	"strconv"
	"sort"
)

func (m *Music) popularByArtist(artist *Artist) []Popular {
	sleep()
	api := lastfm.New(m.config.LastFM.Key, m.config.LastFM.Secret)

	result, _ := api.Artist.GetTopTracks(lastfm.P{"mbid": artist.MBID})
	sort.Slice(result.Tracks, func(i, j int) bool {
		a, _ := strconv.Atoi(result.Tracks[i].PlayCount)
		b, _ := strconv.Atoi(result.Tracks[j].PlayCount)
		return a > b
	})

	var popular []Popular
	for _, track := range result.Tracks {
		rank, _ := strconv.Atoi(track.Rank)
		popular = append(popular, Popular{Artist: artist.Name, Title: track.Name, Rank: uint(rank)})
	}

	return popular
}

func Last(config *config.Config) {
	api := lastfm.New(config.LastFM.Key, config.LastFM.Secret)

	result, _ := api.Artist.GetTopTracks(lastfm.P{"artist": "Gary Numan"})
	sort.Slice(result.Tracks, func(i, j int) bool {
		p1, _ := strconv.Atoi(result.Tracks[i].PlayCount)
		p2, _ := strconv.Atoi(result.Tracks[j].PlayCount)
		return p1 > p2
	})
	for _, track := range result.Tracks {
		fmt.Printf("%-20s %-20s %s %s\n", result.Artist, track.Name, track.PlayCount, track.Rank)
	}

	fmt.Println("");

	result2, _ := api.Chart.GetTopArtists(lastfm.P{})
	sort.Slice(result2.Artists, func(i, j int) bool {
		p1, _ := strconv.Atoi(result2.Artists[i].PlayCount)
		p2, _ := strconv.Atoi(result2.Artists[j].PlayCount)
		return p1 > p2
	})
	for _, artist := range result2.Artists {
		fmt.Printf("%-20s %s %s\n", artist.Name, artist.PlayCount, artist.Listeners)
	}

	fmt.Println("");

	result3, _ := api.Chart.GetTopTracks(lastfm.P{})
	sort.Slice(result3.Tracks, func(i, j int) bool {
		p1, _ := strconv.Atoi(result3.Tracks[i].PlayCount)
		p2, _ := strconv.Atoi(result3.Tracks[j].PlayCount)
		return p1 > p2
	})
	for _, track := range result3.Tracks {
		fmt.Printf("%-20s %-20s %s %s\n", track.Artist.Name, track.Name, track.PlayCount, track.Listeners)
	}
}
