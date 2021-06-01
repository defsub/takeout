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
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/music"
	"github.com/spf13/cobra"
	"time"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync music metadata",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		sync()
	},
}

var syncOptions = music.NewSyncOptions()
var syncAll bool
var syncBack time.Duration

func sync() {
	cfg := getConfig()

	var musicBuckets []config.BucketConfig
	var videoBuckets []config.BucketConfig

	for _, b := range cfg.Buckets {
		if b.Media == config.MediaMusic {
			musicBuckets = append(musicBuckets, b)
		}
		if b.Media == config.MediaVideo {
			videoBuckets = append(videoBuckets, b)
		}
	}

	cfg.Buckets = musicBuckets
	syncMusic(cfg)

	cfg.Buckets = videoBuckets
	///
}

func syncMusic(cfg *config.Config) {
	m := music.NewMusic(cfg)
	m.Open()
	defer m.Close()
	if syncAll {
		syncOptions.Since = time.Time{}
	} else if syncBack > 0 {
		syncOptions.Since = time.Now().Add(-1*syncBack)
	} else {
		syncOptions.Since = m.LastModified()
	}
	m.Sync(syncOptions)
}

func init() {
	syncCmd.Flags().StringVarP(&configFile, "config", "c", "config.ini", "config file")
	syncCmd.Flags().DurationVarP(&syncBack, "back", "b", 0, "Back duration")
	syncCmd.Flags().BoolVarP(&syncOptions.Tracks, "tracks", "t", true, "sync tracks")
	syncCmd.Flags().BoolVarP(&syncOptions.Releases, "releases", "r", true, "sync releases")
	syncCmd.Flags().BoolVarP(&syncOptions.Popular, "popular", "p", true, "sync popular")
	syncCmd.Flags().BoolVarP(&syncOptions.Similar, "similar", "s", true, "sync similar")
	syncCmd.Flags().BoolVarP(&syncOptions.Artwork, "artwork", "z", true, "sync artwork")
	syncCmd.Flags().BoolVarP(&syncOptions.Index, "index", "i", true, "sync index")
	syncCmd.Flags().StringVarP(&syncOptions.Artist, "artist", "a", "", "only sync a specific artist")
	syncCmd.Flags().BoolVarP(&syncAll, "all", "x", false, "(re)sync all tracks instead of modified/new tracks")
	rootCmd.AddCommand(syncCmd)
}
