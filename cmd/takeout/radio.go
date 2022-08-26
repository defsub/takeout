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
	"github.com/defsub/takeout/music"
	"github.com/spf13/cobra"
)

var radioCreate bool
var radioClear bool

var radioCmd = &cobra.Command{
	Use:   "radio",
	Short: "radio",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return radio()
	},
}

func radio() error {
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
	if radioCreate {
		m.CreateStations()
	} else if radioClear {
		m.ClearStations()
	}
	return nil
}

func init() {
	radioCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	radioCmd.Flags().BoolVarP(&radioCreate, "create", "n", true, "(re)create radio stations")
	radioCmd.Flags().BoolVarP(&radioClear, "clear", "x", false, "clear cached radio stations")
	rootCmd.AddCommand(radioCmd)
}
