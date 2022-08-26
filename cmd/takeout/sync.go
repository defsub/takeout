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

package main

import (
	"time"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync media metadata",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sync()
	},
}

var syncBack time.Duration
var syncAll bool
var mediaMusic bool
var mediaVideo bool
var mediaPodcast bool
var artist string
var resolve bool

func since(lastSync time.Time) time.Time {
	var since time.Time
	if syncAll {
		since = time.Time{}
	} else if syncBack > 0 {
		since = time.Now().Add(-1 * syncBack)
	} else {
		since = lastSync
	}
	return since
}

func sync() error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}
	if mediaMusic {
		err = syncMusic(cfg)
		if err != nil {
			return err
		}
	}
	if mediaVideo {
		err = syncVideo(cfg)
		if err != nil {
			return err
		}
	}
	if mediaPodcast {
		err = syncPodcast(cfg)
		if err != nil {
			return err
		}
	}
	return err
}

func syncMusic(cfg *config.Config) error {
	m := music.NewMusic(cfg)
	err := m.Open()
	if err != nil {
		return err
	}
	defer m.Close()
	syncOptions := music.NewSyncOptions()
	syncOptions.Since = since(m.LastModified())
	if len(artist) > 0 {
		syncOptions.Artist = artist
	}
	if resolve {
		syncOptions.Resolve = true
	}
	m.Sync(syncOptions)
	return nil
}

func syncVideo(cfg *config.Config) error {
	v := video.NewVideo(cfg)
	err := v.Open()
	if err != nil {
		return err
	}
	defer v.Close()
	v.SyncSince(since(v.LastModified()))
	return nil
}

func syncPodcast(cfg *config.Config) error {
	p := podcast.NewPodcast(cfg)
	err := p.Open()
	if err != nil {
		return err
	}
	defer p.Close()
	p.Sync()
	return nil
}

func init() {
	syncCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	syncCmd.Flags().DurationVarP(&syncBack, "back", "b", 0, "Back duration")
	syncCmd.Flags().BoolVarP(&syncAll, "all", "a", false, "Re(sync) all ignoring timestamps")
	syncCmd.Flags().BoolVarP(&mediaMusic, "music", "m", true, "Sync music")
	syncCmd.Flags().BoolVarP(&mediaVideo, "video", "v", true, "Sync video")
	syncCmd.Flags().BoolVarP(&mediaPodcast, "podcast", "p", false, "Sync podcasts")
	syncCmd.Flags().BoolVarP(&resolve, "resolve", "x", false, "Resolve")
	syncCmd.Flags().StringVarP(&artist, "artist", "r", "", "Music artist")
	rootCmd.AddCommand(syncCmd)
}
