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

package server

import (
	"github.com/go-co-op/gocron"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
	"time"
)

type syncFunc func(config *config.Config, mediaConfig *config.Config) error

func schedule(config *config.Config) {
	scheduler := gocron.NewScheduler(time.UTC)

	mediaSync := func(d time.Duration, doit syncFunc, startImmediately bool) {
		sched := scheduler.Every(d)
		if startImmediately {
			sched = sched.StartImmediately()
		} else {
			sched = sched.WaitForSchedule()
		}
		sched.Do(func() {
			list, err := assignedMedia(config)
			if err != nil {
				log.Println(err)
				return
			}
			for _, mediaName := range list {
				mediaConfig, err := mediaConfig(config, mediaName)
				if err != nil {
					log.Println(err)
					return
				}
				doit(config, mediaConfig)
			}
		})
	}

	// music
	mediaSync(config.Music.SyncInterval, syncMusic, false)
	mediaSync(config.Music.PopularSyncInterval, syncMusicPopular, false)
	mediaSync(config.Music.SimilarSyncInterval, syncMusicSimilar, false)
	mediaSync(config.Music.CoverSyncInterval, syncMusicCovers, false)

	// podcasts
	mediaSync(config.Podcast.SyncInterval, syncPodcasts, false)

	// video
	mediaSync(config.Video.SyncInterval, syncVideo, false)
	mediaSync(config.Video.PosterSyncInterval, syncVideoPosters, false)
	mediaSync(config.Video.BackdropSyncInterval, syncVideoBackdrops, false)

	scheduler.StartAsync()
}

func assignedMedia(config *config.Config) ([]string, error) {
	a := auth.NewAuth(config)
	err := a.Open()
	if err != nil {
		return []string{}, err
	}
	defer a.Close()
	return a.AssignedMedia(), nil
}

func syncMusic(config *config.Config, mediaConfig *config.Config) error {
	m := music.NewMusic(mediaConfig)
	err := m.Open()
	if err != nil {
		return err
	}
	defer m.Close()
	syncOptions := music.NewSyncOptions()
	syncOptions.Since = m.LastModified()
	m.Sync(syncOptions)
	return nil
}

func syncWithOptions(mediaConfig *config.Config, syncOptions music.SyncOptions) error {
	m := music.NewMusic(mediaConfig)
	err := m.Open()
	if err != nil {
		return err
	}
	defer m.Close()
	m.Sync(syncOptions)
	return nil
}

func syncMusicPopular(config *config.Config, mediaConfig *config.Config) error {
	return syncWithOptions(mediaConfig, music.NewSyncPopular())
}

func syncMusicSimilar(config *config.Config, mediaConfig *config.Config) error {
	return syncWithOptions(mediaConfig, music.NewSyncSimilar())
}

func syncMusicCovers(config *config.Config, mediaConfig *config.Config) error {
	m := music.NewMusic(mediaConfig)
	err := m.Open()
	if err != nil {
		return err
	}
	defer m.Close()
	m.SyncCovers(config.Server.ImageClient)
	return nil
}

func syncMusicFanArt(config *config.Config, mediaConfig *config.Config) error {
	m := music.NewMusic(mediaConfig)
	err := m.Open()
	if err != nil {
		return err
	}
	defer m.Close()
	m.SyncFanArt(config.Server.ImageClient)
	return nil
}

func syncVideo(config *config.Config, mediaConfig *config.Config) error {
	v := video.NewVideo(mediaConfig)
	err := v.Open()
	if err != nil {
		return err
	}
	defer v.Close()
	return v.SyncSince(v.LastModified())
}

func syncVideoPosters(config *config.Config, mediaConfig *config.Config) error {
	v := video.NewVideo(mediaConfig)
	err := v.Open()
	if err != nil {
		return err
	}
	defer v.Close()
	v.SyncPosters(config.Server.ImageClient)
	return nil
}

func syncVideoBackdrops(config *config.Config, mediaConfig *config.Config) error {
	v := video.NewVideo(mediaConfig)
	err := v.Open()
	if err != nil {
		return err
	}
	defer v.Close()
	v.SyncBackdrops(config.Server.ImageClient)
	return nil
}

func syncVideoProfileImages(config *config.Config, mediaConfig *config.Config) error {
	v := video.NewVideo(mediaConfig)
	err := v.Open()
	if err != nil {
		return err
	}
	defer v.Close()
	v.SyncProfileImages(config.Server.ImageClient)
	return nil
}

func syncPodcasts(config *config.Config, mediaConfig *config.Config) error {
	p := podcast.NewPodcast(mediaConfig)
	err := p.Open()
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Sync()
}

func Job(config *config.Config, name string) error {
	list, err := assignedMedia(config)
	if err != nil {
		return err
	}
	for _, mediaName := range list {
		mediaConfig, err := mediaConfig(config, mediaName)
		if err != nil {
			return err
		}
		switch name {
		case "backdrops":
			syncVideoBackdrops(config, mediaConfig)
		case "covers":
			syncMusicCovers(config, mediaConfig)
		case "fanart":
			syncMusicFanArt(config, mediaConfig)
		case "lastfm":
			syncMusicPopular(config, mediaConfig)
			syncMusicSimilar(config, mediaConfig)
		case "media":
			syncMusic(config, mediaConfig)
			syncVideo(config, mediaConfig)
			syncPodcasts(config, mediaConfig)
		case "music":
			syncMusic(config, mediaConfig)
		case "popular":
			syncMusicPopular(config, mediaConfig)
		case "podcasts":
			syncPodcasts(config, mediaConfig)
		case "posters":
			syncVideoPosters(config, mediaConfig)
		case "profiles":
			syncVideoProfileImages(config, mediaConfig)
		case "similar":
			syncMusicSimilar(config, mediaConfig)
		case "video":
			syncVideo(config, mediaConfig)
		}
	}
	return nil
}
