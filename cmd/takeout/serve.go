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
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve music metadata",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	music.Serve(getConfig())
}

func init() {
	serveCmd.Flags().StringVarP(&configFile, "config", "c", "takeout.ini", "config file")
	serveCmd.Flags().String("listen", "127.0.0.1:3000", "Address to listen on")
	rootCmd.AddCommand(serveCmd)
	viper.BindPFlag("Listen", serveCmd.Flags().Lookup("listen"))
}
