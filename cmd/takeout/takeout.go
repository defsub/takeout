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
	"fmt"
	"os"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "takeout",
	Short: "Takeout is a media service",
	Long:  `https://defsub.github.io/`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
	},
}

var configFile string
var configPath string
var configName string

func getConfig() *config.Config {
	if configPath == "" {
		configPath = os.Getenv("TAKEOUT_HOME")
	}
	if configName == "" {
		configName = os.Getenv("TAKEOUT_CONFIG")
	}
	if configFile != "" {
		config.SetConfigFile(configFile)
	} else {
		if configPath == "" {
			configPath = "."
		}
		if configName == "" {
			configName = "takeout"
		}
		config.AddConfigPath(configPath)
		config.SetConfigName(configName)
	}
	config, err := config.GetConfig()
	log.CheckError(err)
	return config
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
