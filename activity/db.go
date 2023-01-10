// Copyright (C) 2022 The Takeout Authors.
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

package activity

import (
	"errors"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (a *Activity) openDB() (err error) {
	cfg := a.config.Music.DB.GormConfig()

	if a.config.Activity.DB.Driver == "sqlite3" {
		a.db, err = gorm.Open(sqlite.Open(a.config.Activity.DB.Source), cfg)
	} else {
		err = errors.New("driver not supported")
	}

	if err != nil {
		return
	}

	a.db.AutoMigrate(&MovieEvent{}, &ReleaseEvent{}, &EpisodeEvent{}, &TrackEvent{})
	return
}

func (a *Activity) closeDB() {
	conn, err := a.db.DB()
	if err != nil {
		return
	}
	conn.Close()
}

func (a *Activity) trackEventsFrom(user string, start, end time.Time, limit int) []TrackEvent {
	var events []TrackEvent
	a.db.Where("user = ? and date >= ? and date <= ?", user, start, end).
		Order("date desc").Limit(limit).Find(&events)
	return events
}

func (a *Activity) popularTrackEventsFrom(user string, start, end time.Time, limit int) []TrackEvent {
	var events []TrackEvent
	a.db.Where("user = ? and date >= ? and date <= ?", user, start, end).
		Group("r_id").
		Order("count(r_id) desc").Limit(limit).Find(&events)
	return events
}

func (a *Activity) movieEventsFrom(user string, start, end time.Time, limit int) []MovieEvent {
	var events []MovieEvent
	a.db.Where("user = ? and date >= ? and date <= ?", user, start, end).
		Order("date desc").Limit(limit).Find(&events)
	return events
}

func (a *Activity) releaseEventsFrom(user string, start, end time.Time, limit int) []ReleaseEvent {
	var events []ReleaseEvent
	a.db.Where("user = ? and date >= ? and date <= ?", user, start, end).
		Order("date desc").Limit(limit).Find(&events)
	return events
}

func (a *Activity) episodeEventsFrom(user string, start, end time.Time, limit int) []EpisodeEvent {
	var events []EpisodeEvent
	a.db.Where("user = ? and date >= ? and date <= ?", user, start, end).
		Order("date desc").Limit(limit).Find(&events)
	return events
}

func (a *Activity) movieEvents(user string) []MovieEvent {
	var movies []MovieEvent
	a.db.Where("user = ?", user).
		Order("date desc").Find(&movies)
	return movies
}

func (a *Activity) episodeEvents(user string) []EpisodeEvent {
	var events []EpisodeEvent
	a.db.Where("user = ?", user).
		Order("date desc").Find(&events)
	return events
}

func (a *Activity) releaseEvents(user string) []ReleaseEvent {
	var releases []ReleaseEvent
	a.db.Where("user = ?", user).
		Order("date desc").Find(&releases)
	return releases
}

func (a *Activity) trackEvents(user string) []TrackEvent {
	var tracks []TrackEvent
	a.db.Where("user = ?", user).
		Order("date desc").Find(&tracks)
	return tracks
}

func (a *Activity) recentMovieEvents(user string, limit int) []MovieEvent {
	var movies []MovieEvent
	a.db.Where("user = ?", user).
		Order("date desc").Limit(limit).Find(&movies)
	return movies
}

func (a *Activity) recentReleaseEvents(user string, limit int) []ReleaseEvent {
	var releases []ReleaseEvent
	a.db.Where("user = ?", user).
		Order("date desc").Limit(limit).Find(&releases)
	return releases
}

func (a *Activity) recentEpisodeEvents(user string, limit int) []EpisodeEvent {
	var events []EpisodeEvent
	a.db.Where("user = ?", user).
		Order("date desc").Limit(limit).Find(&events)
	return events
}

func (a *Activity) recentTrackEvents(user string, limit int) []TrackEvent {
	var tracks []TrackEvent
	a.db.Where("user = ?", user).
		Order("date desc").Limit(limit).Find(&tracks)
	return tracks
}

func (a *Activity) deleteTrackEvents(user string) error {
	return a.db.Unscoped().Where("user = ?", user).Delete(TrackEvent{}).Error
}

func (a *Activity) deleteMovieEvents(user string) error {
	return a.db.Unscoped().Where("user = ?", user).Delete(MovieEvent{}).Error
}

func (a *Activity) deleteReleaseEvents(user string) error {
	return a.db.Unscoped().Where("user = ?", user).Delete(ReleaseEvent{}).Error
}

func (a *Activity) deleteEpisodeEvents(user string) error {
	return a.db.Unscoped().Where("user = ?", user).Delete(EpisodeEvent{}).Error
}

func (a *Activity) createMovieEvent(m *MovieEvent) error {
	return a.db.Create(m).Error
}

func (a *Activity) createReleaseEvent(m *ReleaseEvent) error {
	return a.db.Create(m).Error
}

func (a *Activity) createTrackEvent(t *TrackEvent) error {
	return a.db.Create(t).Error
}

func (a *Activity) createEpisodeEvent(m *EpisodeEvent) error {
	return a.db.Create(m).Error
}

func (a *Activity) updateMovieEvent(m *MovieEvent) error {
	return a.db.Save(m).Error
}

func (a *Activity) updateEpisodeEvent(m *EpisodeEvent) error {
	return a.db.Save(m).Error
}

func (a *Activity) updateReleaseEvent(m *ReleaseEvent) error {
	return a.db.Save(m).Error
}

func (a *Activity) updateTrackEvent(t *TrackEvent) error {
	return a.db.Save(t).Error
}

func (a *Activity) deleteTrackEvent(t *TrackEvent) error {
	return a.db.Unscoped().Delete(t).Error
}

func (a *Activity) deleteMovieEvent(m *MovieEvent) error {
	return a.db.Unscoped().Delete(m).Error
}

func (a *Activity) deleteReleaseEvent(m *ReleaseEvent) error {
	return a.db.Unscoped().Delete(m).Error
}

func (a *Activity) deleteEpisodeEvent(m *EpisodeEvent) error {
	return a.db.Unscoped().Delete(m).Error
}
