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

	"time"
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
	"github.com/defsub/takeout/lib/log"
)

type syncFunc func(config *config.Config) error

func schedule(config *config.Config) {
	scheduler := gocron.NewScheduler(time.UTC)

	mediaSync := func(d time.Duration, doit syncFunc) {
		scheduler.Every(d).WaitForSchedule().Do(func() {
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
				doit(mediaConfig)
			}
		})
	}

	mediaSync(config.Music.SyncInterval, syncMusic)
	mediaSync(config.Video.SyncInterval, syncVideo)
	mediaSync(config.Podcast.SyncInterval, syncPodcasts)

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

func syncMusic(config *config.Config) error {
	log.Printf("sync music\n")
	m := music.NewMusic(config)
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

func syncVideo(config *config.Config) error {
	log.Printf("sync video\n")
	v := video.NewVideo(config)
	err := v.Open()
	if err != nil {
		return err
	}
	defer v.Close()
	return v.SyncSince(v.LastModified())
}

func syncPodcasts(config *config.Config) error {
	log.Printf("sync podcasts\n")
	p := podcast.NewPodcast(config)
	err := p.Open()
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Sync()
}
