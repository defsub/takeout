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
	"github.com/defsub/takeout/auth"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "user admin",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doit()
	},
}

var user, pass, media string
var add, change bool

func doit() error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}
	a := auth.NewAuth(cfg)
	err = a.Open()
	if err != nil {
		return err
	}
	defer a.Close()

	if user != "" && pass != "" {
		if add {
			err := a.AddUser(user, pass)
			if err != nil {
				return err
			}
		} else if change {
			err := a.ChangePass(user, pass)
			if err != nil {
				return err
			}
		}
	}

	if user != "" && media != "" {
		err := a.AssignMedia(user, media)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	userCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	userCmd.Flags().StringVarP(&user, "user", "u", "", "user")
	userCmd.Flags().StringVarP(&pass, "pass", "p", "", "pass")
	userCmd.Flags().StringVarP(&media, "media", "m", "", "media")
	userCmd.Flags().BoolVarP(&add, "add", "a", false, "add")
	userCmd.Flags().BoolVarP(&change, "change", "n", false, "change")
	rootCmd.AddCommand(userCmd)
}
