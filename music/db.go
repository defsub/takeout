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
	"errors"
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
	m.db.AutoMigrate(&Artist{}, &ArtistTag{}, &Popular{}, &Similar{}, &Release{}, &Track{})
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
		" (select releases.name from releases where" +
		" releases.artist = tracks.artist and releases.name = tracks.release)").
		Find(&tracks)
	return tracks
}

func (m *Music) artistTracksWithoutReleases(artist string) []Track {
	var tracks []Track
	m.db.Where("artist = ? and not exists"+
		" (select releases.name from releases where"+
		" releases.artist = tracks.artist and releases.name = tracks.release)", artist).
		Find(&tracks)
	return tracks
}

func (m *Music) trackArtistNames() []string {
	var tracks []*Track
	m.db.Select("distinct(artist)").Find(&tracks)
	var artists []string
	for _, t := range tracks {
		artists = append(artists, t.Artist)
	}
	return artists
}

func orderBy(c Criteria, db *gorm.DB) *gorm.DB {
	if c.Popular {
		return db.Order("popular.rank")
	} else {
		return db.Order("tracks.artist, tracks.release, tracks.disc_num, tracks.track_num")
	}
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
		db = db.Where("tracks.artist in (?)", artists)
	}
	return db
}

func filterByRelease(releases []string, db *gorm.DB) *gorm.DB {
	if len(releases) > 0 {
		db = db.Where("tracks.release in (?)", releases)
	}
	return db
}

func filterByTitle(titles []string, db *gorm.DB) *gorm.DB {
	if len(titles) > 0 {
		db = db.Where("tracks.title in (?)", titles)
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
	db := orderBy(c, m.db)
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
		db = db.Group("tracks.artist, tracks.title")
	} else {
		db = db.Group("tracks.artist, tracks.release, tracks.title")
	}
	db.Find(&tracks)
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

func (m *Music) artistReleaseTracks(artist string, release string) []Track {
	c := Criteria{Artists: artist, Releases: release}
	return m.filter(c)
}

func (m *Music) artists() []Artist {
	var artists []Artist
	m.db.Order("sort_name asc").Find(&artists)
	return artists
}

func (m *Music) artistsByMBID(mbids []string) []Artist {
	var artists []Artist
	m.db.Where("ar_id in (?)", mbids).Find(&artists)
	return artists
}

func (m *Music) similarArtistsByTags(a *Artist) []Artist {
	var artists []Artist
	m.db.Order("count(artist_tags.artist) desc").
		Joins("inner join artist_tags on artists.name = artist_tags.artist").
		Where("artist_tags.tag in (select tag from artist_tags where artist = ?)", a.Name).
		Group("artist_tags.artist").Find(&artists)
	return artists
}

func (m *Music) similarArtists(a *Artist) []Artist {
	var artists []Artist
	m.db.Joins("inner join similar on similar.artist = ?", a.Name).
		Where("artists.ar_id = similar.ar_id").
		Order("similar.rank asc").
		Find(&artists)
	return artists
}

func (m *Music) similarReleases(a *Artist, r Release) []Release {
	artists := m.similarArtists(a);
	var names []string
	for _, sa := range artists {
		names = append(names, sa.Name)
	}

	after := r.Date.AddDate(-1, 0, 0);
	before := r.Date.AddDate(1, 0, 0);

	var releases []Release
	m.db.Joins("inner join tracks on tracks.artist in (?)", names).
		Where("releases.name = tracks.release and releases.artist = tracks.artist").
		Having("min(releases.date) >= ? and min(releases.date) <= ?", after, before).
		Group("releases.name").
		Order("releases.date").Find(&releases)
	return releases
}

func (m *Music) artistReleases(a *Artist) []Release {
	var releases []Release
	m.db.Where("releases.artist = ? and releases.name in (select release from tracks where artist = ?)",
		a.Name, a.Name).
		Having("date = min(date)").
		Group("name").
		Order("date").Find(&releases)
	return releases
}

func (m *Music) artistRelease(a *Artist, name string) *Release {
	releases := m.artistReleases(a)
	for _, r := range releases {
		if r.Name == name {
			return &r
		}
	}
	return nil
}

func (m *Music) releases(a *Artist) []Release {
	var releases []Release
	m.db.Where("artist = ?", a.Name).
		Order("date").Find(&releases)
	return releases
}

func (m *Music) releaseID(a *Artist, mbid string) *Release {
	r := &Release{REID: mbid, Artist: a.Name}
	if m.db.Find(r, r).RecordNotFound() {
		return nil
	}
	return r
}

func (m *Music) artistMinReleases(a *Artist, releaseType string) []Release {
	var releases []Release
	m.db.Where("artist = ? and type = ?", a.Name, releaseType).
		Having("date = min(date)").
		Group("name").
		Order("date").Find(&releases)
	return releases
}

func (m *Music) artistAlbums(a *Artist) []Release {
	return m.artistMinReleases(a, "Album")
}

func (m *Music) artistSingles(a *Artist) []Release {
	return m.artistMinReleases(a, "Single")
}

func (m *Music) artistEPs(a *Artist) []Release {
	return m.artistMinReleases(a, "EP")
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

func (m *Music) recentlyAdded() []Release {
	var releases []Release
	limit := 25
	m.db.Joins("inner join tracks on tracks.release = releases.name and tracks.artist = releases.artist").
		Group("releases.name").
		Order("tracks.last_modified desc").
		Limit(limit).
		Find(&releases)
	return releases
}

func (m *Music) recentlyReleased() []Release {
	var releases []Release
	limit := 25
	m.db.Joins("inner join tracks on tracks.release = releases.name and tracks.artist = releases.artist").
		Group("releases.name").
		Order("releases.date desc").
		Limit(limit).
		Find(&releases)
	return releases
}

func (m *Music) lookupRelease(id uint) (Release, error) {
	var release Release
	if m.db.First(&release, id).RecordNotFound() {
		return Release{}, errors.New("release not found")
	}
	return release, nil
}

func (m *Music) releaseTracks(release Release) []Track {
	var tracks []Track
	m.db.Where("artist = ? and release = ?", release.Artist, release.Name).
		Order("tracknum").Find(&tracks)
	return tracks
}

func (m *Music) lookupArtist(id uint) (Artist, error) {
	var artist Artist
	if m.db.First(&artist, id).RecordNotFound() {
		return Artist{}, errors.New("artist not found")
	}
	return artist, nil
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

func (m *Music) createSimilar(s *Similar) (err error) {
	err = m.db.Create(s).Error
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
