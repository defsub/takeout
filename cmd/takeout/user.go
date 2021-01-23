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
	"github.com/defsub/takeout/log"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "user admin",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		doit()
	},
}

var user, pass, buckets string
var add, change bool

func doit() {
	a := auth.NewAuth(getConfig())
	err := a.Open()
	log.CheckError(err)
	defer a.Close()

	if user != "" && pass != "" {
		if add {
			err := a.AddUser(user, pass)
			log.CheckError(err)
		} else if change {
			err := a.ChangePass(user, pass)
			log.CheckError(err)
		}
	}

	if user != "" && buckets != "" {
		err := a.AssignBuckets(user, buckets)
		log.CheckError(err)
	}
}

func init() {
	userCmd.Flags().StringVarP(&configFile, "config", "c", "takeout.ini", "config file")
	userCmd.Flags().StringVarP(&user, "user", "u", "", "user")
	userCmd.Flags().StringVarP(&pass, "pass", "p", "", "pass")
	userCmd.Flags().StringVarP(&buckets, "buckets", "b", "", "music")
	userCmd.Flags().BoolVarP(&add, "add", "a", false, "add")
	userCmd.Flags().BoolVarP(&change, "change", "n", false, "change")
	rootCmd.AddCommand(userCmd)
}
