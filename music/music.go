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
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/search"
	"github.com/jinzhu/gorm"
	"net/url"
	"regexp"
	"strings"
)

type Music struct {
	config     *config.Config
	db         *gorm.DB
	s3         *s3.S3
	coverCache map[string]string
}

func NewMusic(config *config.Config) *Music {
	return &Music{config: config, coverCache: make(map[string]string)}
}

func (m *Music) Open() (err error) {
	err = m.openDB()
	if err == nil {
		err = m.openBucket()
	}
	return
}

func (m *Music) Close() {
	m.closeDB()
}

// sync steps:
// 1. Sync tracks from bucket based on path name
//    -> Table: tracks
// 2. Sync artists from MusicBrainz (arid)
//    a. Obtain arid for each artist using MusicBrainz
//    b. If none, try last.fm to get arid and use MusicBrainz
//    c. Update track artist name from MusicBrainz
//    -> Table: artists, artist_tags, tracks
// 3. Sync releases for artist from MusicBrainz
//    a. Obtain and store each release group from MusicBrainz (rgid)
//    b. Match each track release with release group
//    c. For tracks w/o matches, search MusicBrainz a release (reid)
//    -> Table: releases, tracks
// 4. Sync top tracks from last.fm for each artist using arid
//    -> Table: popular
// 5. Sync similar artists from last.fm for each artist using arid
//    -> Table: similar
// 6. Sync credits
//    -> Bleve: xxx

func (m *Music) SyncBucketTracks() (err error) {
	m.deleteTracks()
	trackCh, err := m.SyncFromBucket()
	if err != nil {
		return err
	}
	for t := range trackCh {
		fmt.Printf("sync: %s / %s / %s\n", t.Artist, t.Release, t.Title)
		t.Artist = fixName(t.Artist)
		t.Release = fixName(t.Release)
		t.Title = fixName(t.Title)
		// TODO: title may have underscores - picard
		m.createTrack(t)
	}
	err = m.updateTrackCount()
	return
}

// Obtain all releases for each track artist. This will update
// existing releases as well.
func (m *Music) SyncReleases() error {
	artists := m.artists()
	for _, a := range artists {
		fmt.Printf("releases for %s\n", a.Name)
		if a.Name == "Various Artists" {
			// skipping!
			continue
		}

		releases, _ := m.MusicBrainzArtistReleases(&a)
		for _, r := range releases {
			r.Name = fixName(r.Name)
			for i, _ := range r.Media {
				r.Media[i].Name = fixName(r.Media[i].Name)
			}

			curr, _ := m.release(r.REID)
			if curr == nil {
				err := m.createRelease(&r)
				if err != nil {
					fmt.Println(err)
					return err
				}
				for _, d := range r.Media {
					err := m.createMedia(&d)
					if err != nil {
						fmt.Println(err)
						return err
					}
				}
			} else {
				err := m.replaceRelease(curr, &r)
				if err != nil {
					fmt.Println(err)
					return err
				}
				// delete existing release and (re)add new
				m.deleteReleaseMedia(r.REID)
				for _, d := range r.Media {
					err := m.createMedia(&d)
					if err != nil {
						fmt.Println(err)
						return err
					}
				}
			}
		}
	}
	return nil
}

func fuzzyArtist(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9& -]`)
	return re.ReplaceAllString(name, "")
}

func fuzzyName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(name, "")
}

func fixName(name string) string {
	// TODO: use Map?
	name = strings.Replace(name, "–", "-", -1)
	name = strings.Replace(name, "‐", "-", -1)
	name = strings.Replace(name, "’", "'", -1)
	name = strings.Replace(name, "‘", "'", -1)
	name = strings.Replace(name, "“", "\"", -1)
	name = strings.Replace(name, "”", "\"", -1)
	name = strings.Replace(name, "…", "...", -1)
	return name
}

// Assign a track to a specific MusicBrainz REID. This isn't exact and
// instead will pick the first release with the same name with the
// same number of tracks. This way original release dates are
// presented to the user. An attempt is also made to match using
// disambiguations, things like:
//   Weezer (Blue Album)
//   Weezer - Blue Album
//   Weezer Blue Album
//   Weezer [Blue Album]
func (m *Music) AssignTrackReleases() error {
	tracks := m.tracksWithoutAssignedRelease()
	// TODO this could be more efficient
	for _, t := range tracks {
		r := m.trackRelease(&t)
		if r == nil {
			assigned := false
			// try using disambiguation
			releases := m.disambiguate(t.Artist, t.TrackCount)
			for _, r := range releases {
				if r.Disambiguation != "" {
					name1 := fmt.Sprintf("%s (%s)", r.Name, r.Disambiguation)
					name2 := fmt.Sprintf("%s - %s", r.Name, r.Disambiguation)
					name3 := fmt.Sprintf("%s %s", r.Name, r.Disambiguation)
					name4 := fmt.Sprintf("%s [%s]", r.Name, r.Disambiguation)
					if name1 == t.Release || name2 == t.Release ||
						name3 == t.Release || name4 == t.Release {
						err := m.assignTrackRelease(&t, &r)
						if err != nil {
							return err
						}
						assigned = true
						break
					}
				}
			}
			if !assigned {
				fmt.Printf("release not found for %s/%s/%d\n", t.Artist, t.Release, t.TrackCount)
			}
		} else {
			err := m.assignTrackRelease(&t, r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Fix track release names using various pattern matching and name variants.
func (m *Music) FixTrackReleases() error {
	fixReleases := make(map[string]bool)
	var fixTracks []map[string]string
	tracks := m.tracksWithoutReleases()

	for _, t := range tracks {
		artist := m.artist(t.Artist)
		if artist == nil {
			fmt.Printf("artist not found: %s\n", t.Artist)
			continue
		}

		_, ok := fixReleases[t.Release]
		if ok {
			continue
		}

		if artist.Name == "Various Artists" {
			continue
		}

		releases := m.artistReleasesLike(artist, t.Release, t.TrackCount)
		for _, r := range releases {
			fmt.Printf("check %s/%s/%d\n", r.Artist, r.Name, r.TrackCount)
		}

		if len(releases) == 1 {
			fixReleases[t.Release] = true
			fixTracks = append(fixTracks, map[string]string{
				"artist": artist.Name,
				"from":   t.Release,
				"to":     releases[0].Name,
				"count":  itoa(releases[0].TrackCount),
			})
		} else {
			releases = m.releases(artist)
			matched := false
			for _, r := range releases {
				// try fuzzy match
				if fuzzyName(t.Release) == fuzzyName(r.Name) &&
					t.TrackCount == r.TrackCount {
					fixReleases[t.Release] = true
					fixTracks = append(fixTracks, map[string]string{
						"artist": artist.Name,
						"from":   t.Release,
						"to":     r.Name,
						"count":  itoa(r.TrackCount),
					})
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("unmatched '%s' / '%s' with %d tracks\n", t.Artist, t.Release, t.TrackCount)
				fixReleases[t.Release] = false
			}
		}
	}

	for _, v := range fixTracks {
		err := m.updateTrackRelease(v["artist"], v["from"], v["to"], atoi(v["count"]))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Music) FixTrackReleaseTitles() error {
	artists := m.artists()
	for _, a := range artists {
		fmt.Printf("release titles for %s\n", a.Name)
		releases := m.artistReleases(&a)
		for _, r := range releases {
			media := m.releaseMedia(r)
			names := make(map[int]Media)
			for i, _ := range media {
				names[media[i].Position] = media[i]
			}

			tracks := m.releaseTracks(r)
			for i, _ := range tracks {
				name := names[tracks[i].DiscNum].Name
				if name != "" && name != r.Name {
					tracks[i].MediaTitle = name
					// TODO make this format configureable
					tracks[i].ReleaseTitle =
						fmt.Sprintf("%s (%s)", name, r.Name)
				} else {
					tracks[i].MediaTitle = ""
					tracks[i].ReleaseTitle = r.Name
				}
				err := m.updateTrackReleaseTitles(tracks[i])
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Sync popular tracks for each artist from Last.fm.
func (m *Music) SyncPopular() {
	artists := m.artists()
	for _, a := range artists {
		fmt.Printf("popular for %s\n", a.Name)
		popular := m.lastfmArtistTopTracks(&a)
		for _, p := range popular {
			m.createPopular(&p)
		}
	}
}

// Sync similar artists for each artist from Last.fm.
func (m *Music) SyncSimilar() {
	artists := m.artists()
	for _, a := range artists {
		fmt.Printf("similar for %s\n", a.Name)
		similar := m.lastfmSimilarArtists(&a)
		for _, s := range similar {
			m.createSimilar(&s)
		}
	}
}

// Get the artist names from tracks and try to find the correct artist
// from MusicBrainz. This doesn't always work since there can be
// multiple artists with the same name. Last.fm is used to help.
func (m *Music) SyncArtists() error {
	artists := m.trackArtistNames()
	for _, name := range artists {
		var tags []ArtistTag
		artist := m.artist(name)
		if artist == nil {
			artist, tags = m.resolveArtist(name)
			if artist != nil {
				artist.Name = fixName(artist.Name)
				fmt.Printf("creating %s\n", artist.Name)
				m.createArtist(artist)
				for _, t := range tags {
					t.Artist = artist.Name
					m.createArtistTag(&t)
				}
			}
		}

		if artist == nil {
			err := errors.New(fmt.Sprintf("'%s' artist not found", name))
			fmt.Printf("%s\n", err)
			continue
		}

		if name != artist.Name {
			// fix track artist name: AC_DC -> AC/DC
			fmt.Printf("fixing name %s to %s\n", name, artist.Name)
			m.updateTrackArtist(name, artist.Name)
		}
	}
	return nil
}

// Try MusicBrainz and Last.fm to find an artist. Fortunately Last.fm
// will give up the ARID so MusicBrainz can still be used.
func (m *Music) resolveArtist(name string) (artist *Artist, tags []ArtistTag) {
	artist, tags = m.SearchArtist(name)
	if artist == nil {
		// try again
		artist, tags = m.SearchArtist(fuzzyArtist(name))
	}
	if artist == nil {
		// try lastfm
		artist = m.lastfmArtistSearch(name)
		if artist != nil {
			fmt.Printf("try lastfm got %s mbid:'%s'\n", artist.Name, artist.ARID)
			// resolve with mbz
			if artist.ARID != "" {
				artist, tags = m.SearchArtistId(artist.ARID)
			} else {
				artist = nil
			}
		}
	}
	return
}

// Get the URL for the release cover from The Cover Art Archive. Use
// REID front cover if MusicBrainz has that otherwise use the RGID
// cover.
func (m *Music) cover(r Release, s string) string {
	if r.FrontCover && r.REID != "" {
		return fmt.Sprintf("https://coverartarchive.org/release/%s/%s", r.REID, s)
	} else {
		return fmt.Sprintf("https://coverartarchive.org/release-group/%s/%s", r.RGID, s)
	}
}

// Track cover based on assigned release.
func (m *Music) trackCover(t Track, s string) string {
	// TODO should expire the cache
	v, ok := m.coverCache[t.REID]
	if ok {
		return v
	}
	release, _ := m.assignedRelease(&t)
	if release == nil {
		v = ""
	} else {
		v = m.cover(*release, s)
	}
	m.coverCache[t.REID] = v
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
	track, _ := m.lookupETag(etag)
	return track
}

// URL for track cover image.
func (m *Music) TrackImage(t Track) *url.URL {
	url, _ := url.Parse(m.trackCover(t, "front-250"))
	return url
}

func (m *Music) findRelease(rgid string, trackCount int) (string, error) {
	group, err := m.MusicBrainzReleaseGroup(rgid)
	//fmt.Printf("got %+v\n", group)
	if err != nil {
		return "", err
	}
	for _, r := range group.Releases {
		fmt.Printf("find %d vs %d\n", r.totalTracks(), trackCount)
		if r.totalTracks() == trackCount {
			return r.ID, nil
		}
	}
	return "", errors.New("release not found")
}

func (m *Music) releaseIndex(release Release) (search.IndexMap, error) {
	var err error
	tracks := m.releaseTracks(release)

	reid := release.REID
	if reid == "" {
		// is this still needed?
		reid, err = m.findRelease(release.RGID, len(tracks))
		if err != nil {
			return nil, err
		}
	}

	index, err := m.creditsIndex(reid)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`^(\d+)-(\d+)-(.*)$`)
	newIndex := make(search.IndexMap)
	for k, v := range index {
		matches := re.FindStringSubmatch(k)
		if matches == nil {
			fmt.Printf("no re match %s\n", k)
			continue
		}
		discNum := atoi(matches[1])
		trackNum := atoi(matches[2])
		title := matches[3]
		matched := false
		for _, t := range tracks {
			if t.DiscNum == discNum &&
				t.TrackNum == trackNum &&
				t.Title == title {
				// use track key
				newIndex[t.Key] = v
				matched = true
			}
		}
		if !matched {
			//fmt.Printf("no match %s\n", k)
		}
	}

	// update type field with single
	singles := make(map[string]bool)
	for _, t := range m.releaseSingles(release) {
		singles[t.Key] = true
	}
	for k, _ := range singles {
		fields, ok := newIndex[k]
		if ok {
			fields["type"] = "single"
		}
	}

	// use first track release date for tracks index
	for k, v := range newIndex {
		tracks := m.tracksFor([]string{k})
		date := m.trackFirstReleaseDate(&tracks[0])
		s := fmt.Sprintf("%4d-%02d-%02d", date.Year(), date.Month(), date.Day())
		v["date"] = s
	}

	// matched := 0
	// for k, v := range newIndex {
	// 	pos, _ := v["position"].(int)
	// 	track := tracks[pos-1]
	// 	if k == track.Title {
	// 		matched++
	// 		//fmt.Printf("matched %+v\n\n", v)
	// 	}
	// }
	// fmt.Printf("matched %d vs %d\n", matched, len(tracks))

	return newIndex, nil
}

func (m *Music) artistIndex(a *Artist) ([]search.IndexMap, error) {
	var indices []search.IndexMap
	releases := m.artistReleases(a)
	//fmt.Printf("got %d releases\n", len(releases))
	for _, r := range releases {
		//fmt.Printf("%s\n", r.Name)
		index, err := m.releaseIndex(r)
		if err != nil {
			return indices, err
		}
		indices = append(indices, index)
	}
	return indices, nil
}

func (m *Music) SyncIndex() {
	s := search.NewSearch(m.config)
	s.Open("music")
	defer s.Close()

	artists := m.artists()
	for _, a := range artists {
		fmt.Printf("index for %s\n", a.Name)
		if a.Name == "Various Artists" {
			// skipping!
			continue
		}
		index, _ := m.artistIndex(&a)
		for _, idx := range index {
			s.Index(idx)
		}
	}
}

func (m *Music) Search(q string) []Track {
	s := search.NewSearch(m.config)
	s.Open("music")
	defer s.Close()

	keys, err := s.Search(q)
	if err != nil {
		return nil
	}

	return m.tracksFor(keys)
}
