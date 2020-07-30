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
	"github.com/defsub/takeout/auth"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"strconv"
	"time"
)

func (m *Music) openDB() (err error) {
	m.db, err = gorm.Open(m.config.Music.DB.Driver, m.config.Music.DB.Source)
	if err != nil {
		return
	}
	m.db.LogMode(m.config.Music.DB.LogMode)
	m.db.AutoMigrate(&Artist{}, &ArtistTag{}, &Media{}, &Popular{},
		&Similar{}, &Release{}, &Track{}, &UserPlaylist{})
	return
}

func (m *Music) closeDB() {
	m.db.Close()
}

func (m *Music) lastModified() time.Time {
	var tracks []Track
	m.db.Order("last_modified desc").Limit(1).Find(&tracks)
	if len(tracks) == 1 {
		return tracks[0].LastModified
	} else {
		return time.Time{}
	}
}

func (m *Music) deleteTracks() {
	m.db.Unscoped().Model(&Track{}).Delete(&Track{})
}

func (m *Music) createTrack(track *Track) error {
	return m.db.Create(track).Error
}

// Find an artist by name.
func (m *Music) artist(artist string) (a *Artist) {
	a = &Artist{Name: artist}
	if m.db.Find(a, a).RecordNotFound() {
		return nil
	}
	return a
}

// Compute and update TrackCount for each track with total number of
// tracks in the associated release/album. This helps to match up
// MusicBrainz releases with tracks, especially with non-exact
// matches.
func (m *Music) updateTrackCount() error {
	rows, err := m.db.Table("tracks").
		Select("artist, release, date, count(title)").
		Group("artist, release, date").
		Order("artist, release").Rows()
	if err != nil {
		return err
	}
	var results []map[string]string
	for rows.Next() {
		var artist, release, date string
		var count int
		rows.Scan(&artist, &release, &date, &count)
		results = append(results, map[string]string{
			"artist":  artist,
			"release": release,
			"date":    date,
			"count":   strconv.Itoa(count)})
	}
	rows.Close()

	for _, v := range results {
		err = m.db.Table("tracks").
			Where("artist = ? and release = ? and date = ?", v["artist"], v["release"], v["date"]).
			Update("track_count", v["count"]).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// Tracks may have artist names that are modified to meet
// file/directory naming limitations. Update track entries with these
// modified names to the actual artist name.
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

// Tracks may have release names that are modified to meet
// file/directory naming limitations. Update the track entries with
// these modified names to the actual release name.
func (m *Music) updateTrackRelease(artist, oldName, newName string,
	trackCount int) (err error) {
	var tracks []Track
	m.db.Where("artist = ? and release = ? and track_count = ?",
		artist, oldName, trackCount).Find(&tracks)
	for _, t := range tracks {
		err = m.db.Model(t).Update("release", newName).Error
		if err != nil {
			break
		}
	}
	return
}

func (m *Music) updateTrackReleaseTitles(t Track) error {
	return m.db.Model(t).
		Update("media_title", t.MediaTitle).
		Update("release_title", t.ReleaseTitle).Error
}

func (m *Music) tracksAddedSince(t time.Time) []Track {
	var tracks []Track
	m.db.Where("last_modified > ?", t).Find(&tracks)
	return tracks
}

// Find all tracks without a corresponding release - same artist,
// release name and track count. These likely have the wrong artist or
// release name.
func (m *Music) tracksWithoutReleases() []Track {
	var tracks []Track
	m.db.Where("not exists" +
		" (select releases.name from releases where" +
		" releases.artist = tracks.artist and releases.name = tracks.release" +
		" and releases.track_count = tracks.track_count)").
		Find(&tracks)
	return tracks
}

// Try to pattern match releases names. This can help if the track
// release name contains underscores which match nicely with 'like'.
func (m *Music) artistReleasesLike(a *Artist, pattern string, trackCount int) []Release {
	var releases []Release
	m.db.Where("artist = ? and name like ? and track_count = ?",
		a.Name, pattern, trackCount).Find(&releases)
	return releases
}

// Find the tracks that haven't been assigned a REID or RGID.
func (m *Music) tracksWithoutAssignedRelease() []Track {
	var tracks []Track
	m.db.Where("ifnull(re_id, '') = '' or ifnull(rg_id, '') = ''").
		Find(&tracks)
	return tracks
}

// Assign a track to a specific MusicBrainz release. Since the
// original data is just file names, the release is selected
// automatically.
func (m *Music) assignTrackRelease(t *Track, r *Release) error {
	err := m.db.Model(t).Update("re_id", r.REID).Error
	if err != nil {
		return err
	}
	err = m.db.Model(t).Update("rg_id", r.RGID).Error
	if err != nil {
		return err
	}
	return nil
}

// Replace an existing release with a potentially new one. This allows
// for a re-sync from MusicBrainz, preserving timestamps and also the
// track assignment using the RGID & REID.
func (m *Music) replaceRelease(curr *Release, with *Release) error {
	with.ID = curr.ID
	with.CreatedAt = curr.CreatedAt
	with.UpdatedAt = curr.UpdatedAt
	with.DeletedAt = curr.DeletedAt
	return m.db.Save(with).Error
}

func (m *Music) deleteReleaseMedia(reid string) {
	var media []Media
	m.db.Where("re_id = ?", reid).Find(&media)
	for _, d := range media {
		m.db.Unscoped().Delete(d)
	}
}

// Part of the sync process to find releases that match the track. The
// preferred release will be the first one so dates corresponding to
// original release dates.
func (m *Music) trackReleases(t *Track) []Release {
	var releases []Release
	m.db.Where("artist = ? and name = ? and track_count = ?",
		t.Artist, t.Release, t.TrackCount).
		Having("date = min(date)").
		Group("name").
		Order("date").Find(&releases)
	return releases
}

// During sync try to find a single release to match a track.
func (m *Music) trackRelease(t *Track) *Release {
	releases := m.trackReleases(t)
	if len(releases) == 0 {
		return nil
	}
	return &releases[0]
}

// Find the first release date for the release(s) with` this track, including
// an media specific release from a multi-disc set like: Eagles/Legacy or The
// Beatles/The Beatles in Mono. These each have media with titles that
// themselves were previous releases so check them too.
func (m *Music) trackFirstReleaseDate(t *Track) time.Time {
	result := time.Time{}
	var releases []Release
	names := []string{t.Release}
	if t.MediaTitle != "" {
		// names would be "Legacy" and "Hotel California"
		names = append(names, t.MediaTitle)
	}
	m.db.Where("artist = ? and name in (?)",
		t.Artist, names).
		Having("date = min(date)").
		Group("name").
		Order("date").Find(&releases)
	if len(releases) > 0 {
		result = releases[0].Date
	} else {
		// could be disambiguation like "Weezer (Blue Album)" so just
		// use release date for now
		r, err := m.assignedRelease(t)
		if err == nil {
			result = r.Date
		}
	}
	return result
}

// At this point a release couldn't be found easily. Like Weezer has
// multiple albums called Weezer with the same number of tracks. Use
// MusicBrainz disambiguate for look further. This returns all artist
// releases with a specific track count that have a disambiguate.
func (m *Music) disambiguate(artist string, trackCount int) []Release {
	var releases []Release
	m.db.Where("releases.artist = ? and releases.track_count = ? and releases.disambiguation != ''",
		artist, trackCount).
		Order("date desc").Find(&releases)
	return releases
}

// Find all releases for an artist. This is used during sync to match
// tracks again releases when there aren't exact matches.
func (m *Music) releases(a *Artist) []Release {
	var releases []Release
	m.db.Where("artist = ?", a.Name).
		Order("date").Find(&releases)
	return releases
}

// Unique list of artist names based on tracks. This is useful to find
// artist names that may need to be modified to actual names.  Use
// artists() otherwise since it uses sortName from MusicBrainz.
func (m *Music) trackArtistNames() []string {
	var tracks []*Track
	m.db.Select("distinct(artist)").Find(&tracks)
	var artists []string
	for _, t := range tracks {
		artists = append(artists, t.Artist)
	}
	return artists
}

func (m *Music) artistSingleTracks(a Artist) []Track {
	var tracks []Track
	m.db.Where("tracks.artist = ?", a.Name).
		Joins("inner join releases on tracks.artist = releases.artist" +
		" and tracks.title = releases.name and releases.type = 'Single'").
		Order("releases.date").
		Group("tracks.artist, tracks.title").
		Find(&tracks)
	return tracks
}

func (m *Music) artistPopularTracks(a Artist) []Track {
	var tracks []Track
	m.db.Where("tracks.artist = ?", a.Name).
		Joins("inner join popular on tracks.artist = popular.artist" +
 		" and tracks.title = popular.title").
		Order("popular.rank").
		Group("tracks.artist, tracks.title").
		Find(&tracks)
	return tracks
}

// All artist names ordered by sortName from MusicBrainz.
func (m *Music) artists() []Artist {
	var artists []Artist
	m.db.Order("sort_name asc").Find(&artists)
	return artists
}

// All artists with corresponding MusicBrainz artist IDs.
func (m *Music) artistsByMBID(arids []string) []Artist {
	var artists []Artist
	m.db.Where("ar_id in (?)", arids).Find(&artists)
	return artists
}

// not used
func (m *Music) similarArtistsByTags(a *Artist) []Artist {
	var artists []Artist
	m.db.Order("count(artist_tags.artist) desc").
		Joins("inner join artist_tags on artists.name = artist_tags.artist").
		Where("artist_tags.tag in (select tag from artist_tags where artist = ?)", a.Name).
		Group("artist_tags.artist").Find(&artists)
	return artists
}

// Similar artists based on similarity rank from Last.fm.
func (m *Music) similarArtists(a *Artist) []Artist {
	var artists []Artist
	m.db.Joins("inner join similar on similar.artist = ?", a.Name).
		Where("artists.ar_id = similar.ar_id").
		Order("similar.rank asc").
		Limit(m.config.Music.SimilarArtistsLimit).
		Find(&artists)
	return artists
}


// Similar releases based on releases from similar artists in the
// previous and following year.
func (m *Music) similarReleases(a *Artist, r Release) []Release {
	artists := m.similarArtists(a)
	var names []string
	for _, sa := range artists {
		names = append(names, sa.Name)
	}

	after := r.Date.Add(m.config.Music.SimilarReleases * -1)
	before := r.Date.Add(m.config.Music.SimilarReleases)

	var releases []Release
	m.db.Joins("inner join tracks on tracks.artist in (?)", names).
		Where("releases.re_id = tracks.re_id").
		Group("releases.date").
		Having("releases.date >= ? and releases.date <= ?", after, before).
		Limit(m.config.Music.SimilarReleasesLimit).
		Order("releases.date").Find(&releases)
	return releases
}

// All releases for an artist that have corresponing tracks.
func (m *Music) artistReleases(a *Artist) []Release {
	var releases []Release
	m.db.Where("releases.re_id in (select distinct re_id from tracks where artist = ?)",
		a.Name).Order("date asc").Find(&releases)
	return releases
}

// Recently added releases are ordered by LastModified which comes
// from the track object in the S3 bucket.  Use config Recent and
// RecentLimit to tune the result count.
func (m *Music) recentlyAdded() []Release {
	var releases []Release
	m.db.Joins("inner join tracks on tracks.re_id = releases.re_id").
		Group("releases.name").
		Having("tracks.last_modified >= ?", time.Now().Add(m.config.Music.Recent * -1)).
		Order("tracks.last_modified desc").
		Limit(m.config.Music.RecentLimit).
		Find(&releases)
	return releases
}

// Recently released releases are ordered by the MusicBrainz first
// release date of the release, newest first.  Use config Recent and
// RecentLimit to tune the result count.
func (m *Music) recentlyReleased() []Release {
	var releases []Release
	m.db.Joins("inner join tracks on tracks.re_id = releases.re_id").
		Group("releases.name").
		Having("releases.date >= ?", time.Now().Add(m.config.Music.Recent * -1)).
		Order("releases.date desc").
		Limit(m.config.Music.RecentLimit).
		Find(&releases)
	return releases
}

// Obtain the specfic release for this track based on the assigned
// REID or RGID from MusicBrainz. This is useful for covers.
func (m *Music) assignedRelease(t *Track) (*Release, error) {
	var release Release
	if m.db.Where("re_id = ?", t.REID).First(&release).RecordNotFound() {
		if m.db.Where("rg_id = ?", t.RGID).First(&release).RecordNotFound() {
			return nil, errors.New("release not found")
		}
	}
	return &release, nil
}

// func (m *Music) releaseGroup(rgid string) (*Release, error) {
// 	var release Release
// 	if m.db.Where("rg_id = ?", rgid).First(&release).RecordNotFound() {
// 		return nil, errors.New("release group not found")
// 	}
// 	return &release, nil
// }

// Obtain a release using MusicBrainz REID.
func (m *Music) release(reid string) (*Release, error) {
	var release Release
	if m.db.Where("re_id = ?", reid).First(&release).RecordNotFound() {
		return nil, errors.New("release group not found")
	}
	return &release, nil
}

// Obtain all the tracks for this release, ordered by disc and track
// number.
func (m *Music) releaseTracks(release Release) []Track {
	var tracks []Track
	m.db.Where("re_id = ?", release.REID).Order("disc_num, track_num").Find(&tracks)
	return tracks
}

func (m *Music) releaseMedia(release Release) []Media {
	var media []Media
	m.db.Where("re_id = ?", release.REID).Find(&media)
	return media
}

func (m *Music) releaseSingles(release Release) []Track {
	var tracks []Track
	m.db.Where("re_id = ? and" +
		" exists (select name from releases where type = 'Single' and name = title)",
		release.REID).
		Order("disc_num, track_num").Find(&tracks)
	return tracks
}

// Lookup a release given the internal record ID.
func (m *Music) lookupRelease(id int) (Release, error) {
	var release Release
	if m.db.First(&release, id).RecordNotFound() {
		return Release{}, errors.New("release not found")
	}
	return release, nil
}

// Lookup an artist given the internal record ID.
func (m *Music) lookupArtist(id int) (Artist, error) {
	var artist Artist
	if m.db.First(&artist, id).RecordNotFound() {
		return Artist{}, errors.New("artist not found")
	}
	return artist, nil
}

// Lookup a track given the internal record ID.
func (m *Music) lookupTrack(id int) (Track, error) {
	var track Track
	if m.db.First(&track, id).RecordNotFound() {
		return Track{}, errors.New("track not found")
	}
	return track, nil
}

// Lookup a track given the internal record ID.
func (m *Music) tracksFor(keys []string) []Track {
	var tracks []Track
	m.db.Where("key in (?)", keys).
		Limit(m.config.Music.SearchLimit).
		Find(&tracks)
	return tracks
}


// Lookup a track given the etag from the S3 bucket object. Etag can
// be used as a good external identifier (for playlists) since the
// interal record ID can change.
func (m *Music) lookupETag(etag string) (*Track, error) {
	track := Track{ETag: etag}
	if m.db.First(&track, &track).RecordNotFound() {
		return nil, nil
	}
	return &track, nil
}

// Simple sql search for artists, releases and tracks. Use config
// SearchLimit to change the result count.
func (m *Music) search(query string) ([]Artist, []Release, []Track) {
	var artists []Artist
	var releases []Release
	var tracks []Track

	query = "%" + query + "%"

	m.db.Where("name like ?", query).
		Order("sort_name asc").
		Limit(m.config.Music.SearchLimit).Find(&artists)

	m.db.Joins("inner join tracks on"+
		" tracks.re_id = releases.re_id and tracks.release like ?", query).
		Group("releases.name, releases.date").
		Order("releases.name").Limit(m.config.Music.SearchLimit).
		Find(&releases)

	m.db.Where("title like ?", query).
		Order("title").Limit(m.config.Music.SearchLimit).Find(&tracks)

	return artists, releases, tracks
}

func (m *Music) lookupPlaylist(user *auth.User) *UserPlaylist {
	up := &UserPlaylist{User: user.Name}
	if m.db.Find(up, up).RecordNotFound() {
		return nil
	}
	return up
}

func (m *Music) updatePlaylist(up *UserPlaylist) error {
	return m.db.Save(up).Error
}

func (m *Music) createArtist(a *Artist) error {
	return m.db.Create(a).Error
}

func (m *Music) createRelease(a *Release) error {
	return m.db.Create(a).Error
}

func (m *Music) createMedia(a *Media) error {
	return m.db.Create(a).Error
}

func (m *Music) createPopular(p *Popular) error {
	return m.db.Create(p).Error
}

func (m *Music) createSimilar(s *Similar) error {
	return m.db.Create(s).Error
}

func (m *Music) createArtistTag(t *ArtistTag) error {
	return m.db.Create(t).Error
}

func (m *Music) createPlaylist(up *UserPlaylist) error {
	return m.db.Create(up).Error
}
