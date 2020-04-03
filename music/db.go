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
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"math/rand"
	"strings"
	"time"
)

func (m *Music) openDB() (err error) {
	m.db, err = gorm.Open(m.config.Music.DB.Driver, m.config.Music.DB.Source)
	if err != nil {
		return
	}
	m.db.LogMode(m.config.Music.DB.LogMode)
	m.db.AutoMigrate(&Artist{}, &ArtistTag{}, &Popular{}, &Release{}, &Track{})
	return
}

func (m *Music) closeDB() {
	m.db.Close()
}

func (m *Music) deleteTracks() {
	m.db.Unscoped().Model(&Track{}).Delete(&Track{})
}

func (m *Music) createTrack(track *Track) (err error) {
	err = m.db.Create(track).Error
	return
}

func (m *Music) updateTrackArtist(oldName, newName string) (err error) {
	var tracks []Track
	m.db.Where("artist = ?", oldName).Find(&tracks)
	for _, t := range tracks {
		err = m.db.Model(t).Update("artist", newName).Error
		if err != nil {
			break
		}
	}
	return
}

func (m *Music) updateTrackRelease(oldName, newName string) (err error) {
	var tracks []Track
	m.db.Where("release = ?", oldName).Find(&tracks)
	for _, t := range tracks {
		err = m.db.Model(t).Update("release", newName).Error
		if err != nil {
			break
		}
	}
	return
}

func (m *Music) tracksWithoutReleases() []Track {
	var tracks []Track
	m.db.Where("not exists" +
		" (select release.name from release where" +
		" release.artist = tracks.artist and release.name = tracks.release)").
		Find(&tracks)
	return tracks
}

func (m *Music) artistNames() []string {
	var tracks []*Track
	m.db.Select("distinct(artist)").Find(&tracks)
	var artists []string
	for _, t := range tracks {
		artists = append(artists, t.Artist)
	}
	return artists
}

func orderBy(db *gorm.DB) *gorm.DB {
	return db.Order("tracks.artist, tracks.release, tracks.disc_num, tracks.track_num")
}

func filterByTags(tags []string, db *gorm.DB) *gorm.DB {
	if len(tags) > 0 {
		db = db.Where("exists (select distinct at.tag from artist_tags at"+
			" where at.artist = tracks.artist and at.tag in (?))", tags)
	}
	return db
}

func filterByArtist(artists []string, db *gorm.DB) *gorm.DB {
	if len(artists) > 0 {
		db = db.Where("artist in (?)", artists)
	}
	return db
}

func filterByRelease(releases []string, db *gorm.DB) *gorm.DB {
	if len(releases) > 0 {
		db = db.Where("release in (?)", releases)
	}
	return db
}

func filterByTitle(titles []string, db *gorm.DB) *gorm.DB {
	if len(titles) > 0 {
		db = db.Where("title in (?)", titles)
	}
	return db
}

func filterBySingles(db *gorm.DB) *gorm.DB {
	return db.Joins("inner join releases on tracks.artist = releases.artist" +
		" and tracks.title = releases.name and releases.type = 'Single'")
}

func filterByPopular(db *gorm.DB) *gorm.DB {
	return db.Joins("inner join popular on tracks.artist = popular.artist" +
		" and tracks.title = popular.title")
}

func parseDate(date string) (t time.Time) {
	if date != "" {
		var err error
		t, err = time.Parse("2006-01-02", date)
		if err != nil {
			t = time.Time{}
		}
	}
	return t
}

func filterByDate(before, after string, db *gorm.DB) *gorm.DB {
	dateRange := NewDateRange(parseDate(before), parseDate(after))
	if !dateRange.IsZero() {
		if dateRange.before.IsZero() {
			const day = 24 * time.Hour
			dateRange.before = time.Now().Add(day)
		}
		db = db.Where("exists (select distinct a.name from releases a"+
			" where a.artist = tracks.artist and a.name = tracks.release"+
			" group by a.artist, a.name having min(a.date) > ? and min(a.date) < ?)",
			dateRange.AfterDate(), dateRange.BeforeDate())
	}
	return db
}

func split(s string) []string {
	if len(s) == 0 {
		// TODO fix this
		return make([]string, 0)
	}
	a := strings.Split(s, ",")
	for i, _ := range a {
		a[i] = strings.Trim(a[i], " ")
	}
	return a
}

func (m *Music) filter(c Criteria) []Track {
	var tracks []Track
	db := orderBy(m.db)
	db = filterByArtist(split(c.Artists), db)
	db = filterByRelease(split(c.Releases), db)
	db = filterByTitle(split(c.Titles), db)
	db = filterByDate(c.Before, c.After, db)
	db = filterByTags(split(c.Tags), db)
	if c.Singles {
		db = filterBySingles(db)
	}
	if c.Popular {
		db = filterByPopular(db)
	}
	db.Group("tracks.artist, tracks.release, tracks.title").Find(&tracks)
	if c.Shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(tracks), func(i, j int) {
			tracks[i], tracks[j] = tracks[j], tracks[i]
		})
	}
	return tracks
}

func (m *Music) tracks(tags string, date *DateRange) []Track {
	c := Criteria{Tags: tags}
	return m.filter(c)
}

func (m *Music) singleTracks(tags string, date *DateRange) []Track {
	c := Criteria{Tags: tags, Singles: true}
	return m.filter(c)
}

func (m *Music) popularTracks(tags string, date *DateRange) []Track {
	c := Criteria{Tags: tags, Popular: true}
	return m.filter(c)
}

func (m *Music) artistTracks(artists string, date *DateRange) []Track {
	c := Criteria{Artists: artists}
	return m.filter(c)
}

func (m *Music) artistSingleTracks(artists string, date *DateRange) []Track {
	c := Criteria{Artists: artists, Singles: true}
	return m.filter(c)
}

func (m *Music) artistPopularTracks(artists string, date *DateRange) []Track {
	c := Criteria{Artists: artists, Popular: true}
	return m.filter(c)
}

func (m *Music) artistReleases(a *Artist) []Release {
	var releases []Release
	m.db.Where("artist = ?", a.Name).Find(&releases)
	return releases
}

func (m *Music) artistReleasesLike(a *Artist, pattern string) []Release {
	var releases []Release
	m.db.Select("distinct(name)").
		Where("artist = ? and name like ?", a.Name, pattern).Find(&releases)
	return releases
}

func (m *Music) artistTags(a *Artist) []ArtistTag {
	var tags []ArtistTag
	m.db.Order("date asc").Where("artist = ?", a.Name).Find(&tags)
	return tags
}

func (m *Music) artist(artist string) (a *Artist) {
	a = &Artist{Name: artist}
	if m.db.Find(a, a).RecordNotFound() {
		return nil
	}
	return a
}

func (m *Music) createArtist(a *Artist) (err error) {
	err = m.db.Create(a).Error
	return
}

func (m *Music) createRelease(a *Release) (err error) {
	err = m.db.Create(a).Error
	return
}

func (m *Music) createPopular(p *Popular) (err error) {
	err = m.db.Create(p).Error
	return
}

func (m *Music) createArtistTag(t *ArtistTag) (err error) {
	err = m.db.Create(t).Error
	return
}

func (m *Music) createPlaylist(p *Playlist) (err error) {
	err = m.db.Create(p).Error
	return
}
