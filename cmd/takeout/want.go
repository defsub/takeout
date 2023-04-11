// Copyright (C) 2023 The Takeout Authors.
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
	"github.com/defsub/takeout/music"
	"github.com/spf13/cobra"
)

var wantCmd = &cobra.Command{
	Use:   "want",
	Short: "want",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return want()
	},
}

func want() error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}
	m := music.NewMusic(cfg)
	err = m.Open()
	if err != nil {
		return err
	}
	defer m.Close()
	list := m.WantList()
	fmt.Println("# Takeout: Wantlist")
	var prevArtist string
	for _, r := range list {
		var disamb = ""
		if r.Disambiguation != "" {
			disamb = fmt.Sprintf(" (%s)", r.Disambiguation)
		}
		if r.Artist != prevArtist {
			fmt.Printf("## %s\n", r.Artist)
		}
		fmt.Printf("- %04d [%s%s](https://musicbrainz.org/release-group/%s)\n", r.Date.Year(), r.Name, disamb, r.REID)
		prevArtist = r.Artist
	}
	return nil
}

func init() {
	wantCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	rootCmd.AddCommand(wantCmd)
}
