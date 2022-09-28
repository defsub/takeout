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
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/bucket"
	"github.com/defsub/takeout/lib/client"
	"github.com/defsub/takeout/lib/fanart"
	"github.com/defsub/takeout/lib/lastfm"
	"github.com/defsub/takeout/lib/musicbrainz"
	"github.com/defsub/takeout/lib/search"
	"gorm.io/gorm"
)

const (
	TakeoutUser    = "takeout"
	VariousArtists = "Various Artists"
)

var coverCache map[string]string = make(map[string]string)

type Music struct {
	config  *config.Config
	db      *gorm.DB
	buckets []bucket.Bucket
	client  *client.Client
	lastfm  *lastfm.Lastfm
	fanart  *fanart.Fanart
	mbz     *musicbrainz.MusicBrainz
}

func NewMusic(config *config.Config) *Music {
	return &Music{
		config: config,
		client: client.NewClient(config),
		fanart: fanart.NewFanart(config),
		lastfm: lastfm.NewLastfm(config),
		mbz:    musicbrainz.NewMusicBrainz(config),
	}
}

func (m *Music) Open() (err error) {
	err = m.openDB()
	if err == nil {
		m.buckets, err = bucket.OpenMedia(m.config.Buckets, config.MediaMusic)
	}
	return
}

func (m *Music) Close() {
	m.closeDB()
}

// Get the URL for the release cover from The Cover Art Archive. Use
// REID front cover.
//
// See https://musicbrainz.org/doc/Cover_Art_Archive/API
func Cover(r Release, size string) string {
	var url string
	if r.GroupArtwork {
		url = fmt.Sprintf("/img/mb/rg/%s", r.RGID)
	} else {
		url = fmt.Sprintf("/img/mb/re/%s", r.REID)
	}
	if r.Artwork && r.FrontArtwork {
		// user front-250, front-500, front-1200
		//return fmt.Sprintf("%s/front-%s", url, size)
		return fmt.Sprintf("%s/front", url)
	} else if r.Artwork && r.OtherArtwork != "" {
		// use id-250, id-500, id-1200
		//return fmt.Sprintf("%s/%s-%s", url, r.OtherArtwork, size)
		return url
	} else {
		return "/static/album-white-36dp.svg"
	}
}

func (m *Music) CoverSmall(o interface{}) string {
	switch o.(type) {
	case Release:
		return o.(Release).Cover("250")
	case Track:
		return m.TrackCover(o.(Track), "250")
	}
	return ""
}

// Track cover based on assigned release.
func (m *Music) TrackCover(t Track, size string) string {
	// TODO should expire the cache
	v, ok := coverCache[t.REID]
	if ok {
		return v
	}
	release, _ := m.assignedRelease(&t)
	if release == nil {
		v = ""
	} else {
		v = Cover(*release, size)
	}
	coverCache[t.REID] = v
	return v
}

// URL to stream track from the S3 bucket. This will be signed and
// expired based on config.
func (m *Music) TrackURL(t *Track) *url.URL {
	url := m.bucketURL(t)
	return url
}

// Find track using the etag from the S3 bucket.
func (m *Music) TrackLookup(etag string) *Track {
	track, _ := m.LookupETag(etag)
	return track
}

// URL for track cover image.
func (m *Music) TrackImage(t Track) *url.URL {
	url, _ := url.Parse(m.TrackCover(t, "front-250"))
	return url
}

func (m *Music) FindArtist(identifier string) (Artist, error) {
	id, err := strconv.Atoi(identifier)
	if err != nil {
		return m.LookupARID(identifier)
	} else {
		return m.LookupArtist(id)
	}
}

func (m *Music) FindRelease(identifier string) (Release, error) {
	id, err := strconv.Atoi(identifier)
	if err != nil {
		return m.LookupREID(identifier)
	} else {

		return m.LookupRelease(id)
	}
}

func (m *Music) FindStation(identifier string) (Station, error) {
	id, err := strconv.Atoi(identifier)
	if err != nil {
		return Station{}, errors.New("station not found")
	} else {
		return m.LookupStation(id)
	}
}

func (m *Music) FindTrack(identifier string) (Track, error) {
	id, err := strconv.Atoi(identifier)
	if err != nil {
		if strings.HasPrefix(identifier, "uuid:") {
			return m.LookupUUID(identifier[5:])
		} else if strings.HasPrefix(identifier, "rid:") {
			return m.LookupRID(identifier[4:])
		} else {
			return m.LookupRID(identifier)
		}
	} else {
		return m.LookupTrack(id)
	}
}

func (m *Music) FindTracks(identifiers []string) []Track {
	// TODO support more than RIDs later
	return m.tracksForRIDs(identifiers)
}

func (m *Music) newSearch() *search.Search {
	s := search.NewSearch(m.config)
	s.Keywords = []string{
		FieldGenre,
		FieldStatus,
		FieldTag,
		FieldType,
	}
	s.Open("music")
	return s
}

func (m *Music) Search(q string, limit ...int) []Track {
	s := m.newSearch()
	defer s.Close()

	l := m.config.Music.SearchLimit
	if len(limit) == 1 {
		l = limit[0]
	}

	keys, err := s.Search(q, l)
	if err != nil {
		return nil
	}

	// split potentially large # of result keys into chunks to query
	chunkSize := 100
	var tracks []Track
	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize
		if end > len(keys) {
			end = len(keys)
		}
		chunk := keys[i:end]
		tracks = append(tracks, m.tracksFor(chunk)...)
	}

	return tracks
}

const (
	// rename to Radio* or Station*
	TypeArtist  = "artist"  // Songs by single artist
	TypeGenre   = "genre"   // Songs from one or more genres
	TypeSimilar = "similar" // Songs from similar artists
	TypePeriod  = "period"  // Songs from one or more time periods
	TypeSeries  = "series"  // Songs from one or more series (chart)
	TypeStream  = "stream"  // Internet radio stream
	TypeOther   = "other"
)

func (s *Station) Visible(user *auth.User) bool {
	return s.User == user.Name || s.Shared
}

func (m *Music) ClearStations() {
	m.clearStationPlaylists()
}

func (m *Music) CreateStations() {
	genres := m.config.Music.RadioGenres
	if len(m.config.Music.RadioGenres) == 0 {
		genres = m.artistGenres()
	}
	for _, g := range genres {
		if len(g) == 0 {
			continue
		}
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   TypeGenre,
			Name:   strings.Title(g),
			Creator: "Radio",
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(fmt.Sprintf(`+genre:"%s" +type:single +popularity:<11 -artist:"Various Artists"`, g)))}
		m.CreateStation(&station)
	}

	decades := []int{1960, 1970, 1980, 1990, 2000, 2010, 2020}
	for _, d := range decades {
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   TypePeriod,
			Name:   fmt.Sprintf("%ds Top Tracks", d),
			Creator: "Radio",
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(fmt.Sprintf(
					`+first_date:>="%d-01-01" +first_date:<="%d-12-31" +type:single +popularity:<11`, d, d+9)))}
		m.CreateStation(&station)
	}

	for _, s := range m.config.Music.RadioSeries {
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   TypeSeries,
			Name:   s,
			Creator: "Radio",
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(fmt.Sprintf(`+series:"%s"`, s)))}
		m.CreateStation(&station)
	}

	for k, v := range m.config.Music.RadioOther {
		station := Station{
			User:   TakeoutUser,
			Shared: true,
			Type:   TypeOther,
			Name:   k,
			Creator: "Radio",
			Ref: fmt.Sprintf(`/music/search?q=%s&radio=1`,
				url.QueryEscape(v))}
		m.CreateStation(&station)
	}

	for _, v := range m.config.Music.RadioStreams {
		station := Station{
			User:     TakeoutUser,
			Shared:   true,
			Type:     TypeStream,
			Name:     v.Title,
			Creator:  v.Creator,
			Ref:      v.Location,
			Image:    v.Image,
		}
		m.CreateStation(&station)
	}
}

func (m *Music) ArtistRadio(artist Artist) []Track {
	tracks := m.ArtistSimilar(artist,
		m.config.Music.ArtistRadioDepth,
		m.config.Music.ArtistRadioBreadth)
	if len(tracks) > m.config.Music.RadioLimit {
		tracks = tracks[:m.config.Music.RadioLimit]
	}
	return tracks
}

func (m *Music) ArtistSimilar(artist Artist, depth int, breadth int) []Track {
	var station []Track
	tracks := m.ArtistPopularTracks(artist, depth)
	station = append(station, tracks...)
	artists := m.SimilarArtists(&artist, breadth)
	for _, a := range artists {
		tracks = m.ArtistPopularTracks(a, depth)
		station = append(station, tracks...)
	}
	return Shuffle(station)
}

func (m *Music) ArtistShuffle(artist Artist, depth int) []Track {
	var tracks []Track
	// add 75% popular
	pop := int(float32(depth) * 0.75)
	tracks = append(tracks, Shuffle(m.ArtistPopularTracks(artist, pop))...)
	// randomly add some unique tracks
	// TODO consider other algorithms
	all := Shuffle(m.ArtistTracks(artist))
	pick := 0
	for len(tracks) < depth && pick < len(all) {
		t := all[pick]
		if !contains(tracks, t) {
			tracks = append(tracks, t)
		}
		pick++
	}
	return Shuffle(tracks)
}

func (m *Music) ArtistDeep(artist Artist, depth int) []Track {
	tracks := m.artistDeepTracks(artist, depth)
	return Shuffle(tracks)
}

func contains(tracks []Track, t Track) bool {
	for _, v := range tracks {
		if v.Title == t.Title && v.Artist == t.Artist {
			return true
		}
	}

	return false
}

func Shuffle(tracks []Track) []Track {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tracks), func(i, j int) { tracks[i], tracks[j] = tracks[j], tracks[i] })
	return tracks
}

func (m *Music) HasMusic() bool {
	return m.TrackCount() > 0
}

func (m *Music) SearchTracks(title, artist, album string) []Track {
	return m.searchTracks(title, artist, album)
}

func (m *Music) UnmatchedTracks() []Track {
	return m.tracksWithoutAssignedRelease()
}

func (m *Music) ArtistImage(artist *Artist) string {
	imgs := m.artistImages(artist)
	if len(imgs) == 0 {
		return ""
	}
	// https://assets.fanart.tv/fanart/music/a6c6897a-7415-4f8d-b5a5-3a5e05f3be67/artistthumb/twenty-one-pilots-55362909e8765.jpg
	pattern := fmt.Sprintf("/music/%s/artistthumb", artist.ARID)
	for _, img := range imgs {
		if strings.Contains(img, pattern) {
			parts := strings.Split(img, "/")
			return fmt.Sprintf("/img/fa/%s/t/%s", artist.ARID, parts[len(parts)-1])
		}
	}
	return imgs[0]
}

func (m *Music) ArtistBackground(artist *Artist) string {
	imgs := m.artistBackgrounds(artist)
	if len(imgs) == 0 {
		return ""
	}
	// https://assets.fanart.tv/fanart/music/a6c6897a-7415-4f8d-b5a5-3a5e05f3be67/artistbackground/twenty-one-pilots-538ed3f1068af.jpg
	pattern := fmt.Sprintf("/music/%s/artistbackground", artist.ARID)
	for _, img := range imgs {
		if strings.Contains(img, pattern) {
			parts := strings.Split(img, "/")
			return fmt.Sprintf("/img/fa/%s/b/%s", artist.ARID, parts[len(parts)-1])
		}
	}
	return imgs[0]
}
