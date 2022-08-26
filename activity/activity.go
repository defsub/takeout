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

// Package activity manages user activity data.
package activity

import (
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
	"gorm.io/gorm"

	"errors"
	"strconv"
	"time"
)

var (
	ErrInvalidUser = errors.New("invalid user")
)

type Activity struct {
	config *config.Config
	db     *gorm.DB
}

func NewActivity(config *config.Config) *Activity {
	return &Activity{
		config: config,
	}
}

func (a *Activity) Open() (err error) {
	err = a.openDB()
	return
}

func (a *Activity) Close() {
	a.closeDB()
}

//
func (a *Activity) DeleteUserEvents(user *auth.User) error {
	err := a.deleteMovieEvents(user.Name)
	if err != nil {
		log.Println("movie delete error: ", err)
		return err
	}
	err = a.deleteReleaseEvents(user.Name)
	if err != nil {
		log.Println("release delete error: ", err)
		return err
	}
	err = a.deleteSeriesEpisodeEvents(user.Name)
	if err != nil {
		log.Println("series delete error: ", err)
		return err
	}
	err = a.deleteTrackEvents(user.Name)
	if err != nil {
		log.Println("track delete error: ", err)
		return err
	}
	return nil
}

func (a *Activity) resolveMovieEvent(e MovieEvent, v *video.Video) *Movie {
	movie, err := v.FindMovie(e.IMID)
	if err != nil {
		return nil
	}
	return &Movie{Date: e.Date, Movie: movie}
}

func (a *Activity) resolveSeriesEpisodeEvent(e SeriesEpisodeEvent, p *podcast.Podcast) *SeriesEpisode {
	episode, err := p.FindEpisode(e.EID)
	if err != nil {
		return nil
	}
	return &SeriesEpisode{Date: e.Date, Episode: episode}
}

func (a *Activity) resolveReleaseEvent(e ReleaseEvent, m *music.Music) *Release {
	release, err := m.FindRelease(e.REID)
	if err != nil {
		return nil
	}
	return &Release{Date: e.Date, Release: release}
}

func (a *Activity) resolveTrackEvent(e TrackEvent, m *music.Music) *Track {
	track, err := m.FindTrack(e.RID)
	if err != nil {
		return nil
	}
	return &Track{Date: e.Date, Track: track}
}

func (a *Activity) resolveMovieEvents(events []MovieEvent, v *video.Video) []Movie {
	var movies []Movie
	for _, e := range events {
		movie := a.resolveMovieEvent(e, v)
		if movie != nil {
			movies = append(movies, *movie)
		}
	}
	return movies
}

func (a *Activity) resolveSeriesEpisodeEvents(events []SeriesEpisodeEvent, p *podcast.Podcast) []SeriesEpisode {
	var episodes []SeriesEpisode
	for _, e := range events {
		episode := a.resolveSeriesEpisodeEvent(e, p)
		if episode != nil {
			episodes = append(episodes, *episode)
		}
	}
	return episodes
}

func (a *Activity) resolveReleaseEvents(events []ReleaseEvent, m *music.Music) []Release {
	var releases []Release
	for _, e := range events {
		release := a.resolveReleaseEvent(e, m)
		if release != nil {
			releases = append(releases, *release)
		}
	}
	return releases
}

func (a *Activity) resolveTrackEvents(events []TrackEvent, m *music.Music) []Track {
	var tracks []Track
	for _, e := range events {
		track := a.resolveTrackEvent(e, m)
		if track != nil {
			tracks = append(tracks, *track)
		}
	}
	return tracks
}

func (a *Activity) Movies(user *auth.User, v *video.Video, start, end time.Time) []Movie {
	events := a.movieEventsFrom(user.Name, start, end)
	return a.resolveMovieEvents(events, v)
}

func (a *Activity) Tracks(user *auth.User, m *music.Music, start, end time.Time) []Track {
	events := a.trackEventsFrom(user.Name, start, end)
	return a.resolveTrackEvents(events, m)
}

func (a *Activity) Releases(user *auth.User, m *music.Music, start, end time.Time) []Release {
	events := a.releaseEventsFrom(user.Name, start, end)
	return a.resolveReleaseEvents(events, m)
}

//
func (a *Activity) RecentTracks(user *auth.User, m *music.Music) []Track {
	events := a.recentTrackEvents(user.Name, a.config.Activity.RecentTrackLimit)
	if len(events) == 0 {
		a.createTrackEvent(&TrackEvent{
			User: user.Name,
			RID:  "063964f7-35c3-467f-b30c-cb7303d866b9",
			Date: time.Now(),
		})
	}
	return a.resolveTrackEvents(events, m)
}

//
func (a *Activity) RecentMovies(user *auth.User, v *video.Video) []Movie {
	events := a.recentMovieEvents(user.Name, a.config.Activity.RecentMovieLimit)
	if len(events) == 0 {
		a.createMovieEvent(&MovieEvent{
			User: user.Name,
			IMID: "tt0081505",
			Date: time.Now(),
		})
	}
	return a.resolveMovieEvents(events, v)
}

func (a *Activity) RecentReleases(user *auth.User, m *music.Music) []Release {
	events := a.recentReleaseEvents(user.Name, a.config.Activity.RecentReleaseLimit)
	if len(events) == 0 {
		a.createReleaseEvent(&ReleaseEvent{
			User: user.Name,
			REID: "89b14227-9d08-4910-8c15-62755ba3b7bc",
			Date: time.Now(),
		})
	}
	return a.resolveReleaseEvents(events, m)
}

// Add a scrobble with an MBID that should match a track we have
func (a *Activity) UserScrobble(user *auth.User, s Scrobble, music *music.Music) error {
	// ensure there's a valid user
	// if s.User == "" {
	// 	s.User = user.Name
	// } else if s.User != user.Name {
	// 	return ErrInvalidUser
	// }

	// if s.MBID != "" {
	// 	_, err := music.FindTrack(s.MBID)
	// 	if err != nil {
	// 		// no track with that MBID (RID)
	// 		// code below will hopefully find a new one
	// 		s.MBID = ""
	// 	}
	// }
	// if s.MBID == "" {
	// 	tracks := music.SearchTracks(s.Track, s.PreferredArtist(), s.Album)
	// 	if len(tracks) > 0 {
	// 		// use first matching track MBZ recording ID
	// 		s.MBID = tracks[0].RID
	// 	}
	// }

	// // MBID may still be empty but allow anyway for now
	// return a.createScrobble(&s)
	return nil
}

func (a *Activity) CreateEvents(events Events, user *auth.User, m *music.Music, v *video.Video) error {
	for _, e := range events.MovieEvents {
		e.User = user.Name
		if e.ETag != "" {
			// resolve using ETag
			video, err := v.LookupETag(e.ETag)
			if err != nil {
				return err
			}
			e.IMID = video.IMID
			e.TMID = strconv.FormatInt(video.TMID, 10)
		}
		err := a.createMovieEvent(&e)
		if err != nil {
			return err
		}
	}

	for _, e := range events.ReleaseEvents {
		e.User = user.Name
		err := a.createReleaseEvent(&e)
		if err != nil {
			return err
		}
	}

	for _, e := range events.SeriesEpisodeEvents {
		e.User = user.Name
		err := a.createSeriesEpisodeEvent(&e)
		if err != nil {
			return err
		}
	}

	for _, e := range events.TrackEvents {
		e.User = user.Name
		if e.ETag != "" {
			// resolve using ETag
			track, err := m.LookupETag(e.ETag)
			if err != nil {
				return err
			}
			e.RID = track.RID
			e.RGID = track.RGID
		}
		err := a.createTrackEvent(&e)
		if err != nil {
			return err
		}
	}

	return nil
}
