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
	"time"

	"github.com/defsub/takeout/lib/gorm"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
)

// Activity data should be long-lived and w/o internal sync identifiers.  Use
// external globally unique IDs and stable media metadata. It should all be
// meaningful even after a full re-sync.

// Scrobble is a track listening event.
//
// Spec location: https://www.last.fm/api/show/track.scrobble
// not using: context, streamId, chosenByUser
type Scrobble struct {
	Artist      string    `json:""` // Required
	Track       string    `json:""` // Required
	Timestamp   time.Time `json:""` // Required
	Album       string    `json:""` // Optional
	AlbumArtist string    `json:""` // Optional
	TrackNumber int       `json:""` // Optional
	Duration    int       `json:""` // Optional - Length of track in seconds
	MBID        string    `json:""` // Optional - recording (or track?) MBID
}

func (s Scrobble) PreferredArtist() string {
	if s.AlbumArtist != "" {
		return s.AlbumArtist
	}
	return s.Artist
}

// Listen is an album, playlist or podcast listening event.
type Playlist struct {
	gorm.Model
	User    string    `gorm:"index:idx_playlist_user" json:"-"`
	Title   string    // Album title, Station name, etc.
	Creator string    // Artist name, Podcast author
	Image   string    //
	Ref     string    //
	Date    time.Time `gorm:"uniqueIndex:idx_playlist_date"`
}

type Events struct {
	MovieEvents         []MovieEvent
	ReleaseEvents       []ReleaseEvent
	SeriesEpisodeEvents []SeriesEpisodeEvent
	TrackEvents         []TrackEvent
}

type ReleaseEvent struct {
	gorm.Model
	User string    `gorm:"index:idx_release_user" json:"-"`
	Date time.Time `gorm:"uniqueIndex:idx_release_date"`
	RGID string
	REID string
}

type MovieEvent struct {
	gorm.Model
	User string    `gorm:"index:idx_movie_user" json:"-"`
	Date time.Time `gorm:"uniqueIndex:idx_movie_date"`
	TMID string
	IMID string
	ETag string    `gorm:"-"`
}

type TrackEvent struct {
	gorm.Model
	User string    `gorm:"index:idx_track_user" json:"-"`
	Date time.Time `gorm:"uniqueIndex:idx_track_date"`
	RID  string
	RGID string
	ETag string    `gorm:"-"`
}

type SeriesEpisodeEvent struct {
	gorm.Model
	User string    `gorm:"index:idx_series_episode_user" json:"-"`
	Date time.Time `gorm:"uniqueIndex:idx_series_episode_date"`
	EID  string
}

type Release struct {
	Date    time.Time
	Release music.Release
}

type Movie struct {
	Date  time.Time
	Movie video.Movie
}

type Track struct {
	Date  time.Time
	Track music.Track
}

type SeriesEpisode struct {
	Date    time.Time
	Episode podcast.Episode
}

// func (t Track) Valid() bool {
// 	if t.User == "" || t.MBID == "" || t.Date.IsZero() {
// 		return false
// 	}
// 	return true
// }

// func (p Playlist) Valid() bool {
// 	if p.User == "" || p.Ref == "" ||
// 		p.Title == "" || p.Creator == "" || p.Date.IsZero() {
// 		return false
// 	}
// 	return true
// }

// func (m Movie) Valid() bool {
// 	if m.User == "" || m.Date.IsZero() {
// 		return false
// 	}
// 	if m.TMID == "" && m.IMID == "" {
// 		// must have TMID or IMID
// 		return false
// 	}
// 	return true
// }
