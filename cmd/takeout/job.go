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
	"errors"
	"github.com/defsub/takeout/server"
	"github.com/spf13/cobra"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "takeout job",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return job()
	},
}

var jobName string

func job() error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}
	if jobName == "" {
		return errors.New("no job")
	}
	return server.Job(cfg, jobName)
}

func init() {
	jobCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	jobCmd.Flags().StringVarP(&jobName, "name", "n", "", "name of job")
	rootCmd.AddCommand(jobCmd)
}
