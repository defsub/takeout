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

package main

import (
	"fmt"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "takeout stats",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return stats()
	},
}

func stats() error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}
	err = musicStats(cfg)
	if err != nil {
		return err
	}
	err = videoStats(cfg)
	if err != nil {
		return err
	}
	err = podcastStats(cfg)
	if err != nil {
		return err
	}
	return nil
}

func musicStats(cfg *config.Config) error {
	m := music.NewMusic(cfg)
	err := m.Open()
	if err != nil {
		return err
	}
	defer m.Close()
	fmt.Printf("artists %d\n", len(m.Artists()))
	fmt.Printf("tracks %d\n", m.TrackCount())
	fmt.Printf("unmatched tracks %d\n", len(m.UnmatchedTracks()))
	return nil
}

func videoStats(cfg *config.Config) error {
	v := video.NewVideo(cfg)
	err := v.Open()
	if err != nil {
		return err
	}
	defer v.Close()
	fmt.Printf("movies %d\n", v.MovieCount())
	return nil
}

func podcastStats(cfg *config.Config) error {
	p := podcast.NewPodcast(cfg)
	err := p.Open()
	if err != nil {
		return err
	}
	defer p.Close()
	fmt.Printf("series %d\n", p.SeriesCount())
	return nil
}

func init() {
	statsCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	rootCmd.AddCommand(statsCmd)
}
