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
	Run: func(cmd *cobra.Command, args []string) {
		sync()
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

func sync() {
	cfg := getConfig()
	if mediaMusic {
		syncMusic(cfg)
	}
	if mediaVideo {
		syncVideo(cfg)
	}
	if mediaPodcast {
		syncPodcast(cfg)
	}
}

func syncMusic(cfg *config.Config) {
	m := music.NewMusic(cfg)
	m.Open()
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
}

func syncVideo(cfg *config.Config) {
	v := video.NewVideo(cfg)
	v.Open()
	defer v.Close()
	v.SyncSince(since(v.LastModified()))
}

func syncPodcast(cfg *config.Config) {
	p := podcast.NewPodcast(cfg)
	p.Open()
	defer p.Close()
	p.Sync()
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
