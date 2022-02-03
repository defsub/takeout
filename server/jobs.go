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
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/lib/log"
)

func schedule(config *config.Config) {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every("1h").WaitForSchedule().Do(func() {
		a := auth.NewAuth(config)
		err := a.Open()
		if err != nil {
			log.Println(err)
			return
		}
		defer a.Close()
		list := a.AssignedMedia()
		for _, mediaName := range list {
			_, mediaConfig, err := mediaConfig(config, mediaName)
			if err != nil {
				log.Println(err)
				return
			}
			syncPodcasts(mediaConfig)
		}
	})
	scheduler.StartAsync()
}

func syncPodcasts(config *config.Config) {
	log.Printf("sync podcasts\n")
	p := podcast.NewPodcast(config)
	err := p.Open()
	if err != nil {
		log.Println(err)
		return
	}
	defer p.Close()
	err = p.Sync()
	if err != nil {
		log.Println(err)
		return
	}
}
