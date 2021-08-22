// Copyright (C) 2021 The Takeout Authors.
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
	"regexp"

	"github.com/defsub/takeout/lib/tmdb"
	"github.com/defsub/takeout/lib/date"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		doit()
	},
}

var optQuery string
var optFile string

func fixColon(name string) string {
	// not generalluy good to have colons in filenames so change "foo: bar"
	// to "foo - bar"
	colon := regexp.MustCompile(`([A-Za-z0-9])\s*(:)\s`)
	name = colon.ReplaceAllString(name, "${1} - ")
	return name
}

func doit() {
	var query string
	m := tmdb.NewTMDB(getConfig())
	if optFile != "" {
		fileRegexp := regexp.MustCompile(`([^\/]+)_t\d+\.mkv$`)
		matches := fileRegexp.FindStringSubmatch(optFile)
		if matches != nil {
			query = matches[1]
		}
	} else if optQuery != "" {
		query = optQuery
	}
	if query != "" {
		results, err := m.MovieSearch(query)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		for _, v := range results {
			title := fmt.Sprintf("%s (%d)", v.Title, date.ParseDate(v.ReleaseDate).Year())
			fmt.Printf("%s\n", title)
			fmt.Printf("%s\n", m.MovieOriginalPoster(v.PosterPath).String())
			if len(v.GenreIDs) > 0 {
				for i, id := range v.GenreIDs {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Print(m.MovieGenre(id))
				}
				fmt.Println()
			}
			if optFile != "" {
				fmt.Printf("mv '%s' '%s - HD.mkv'\n", optFile, title)
			}
			fmt.Printf("\n")
		}
	}
}

func init() {
	searchCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	searchCmd.Flags().StringVarP(&optQuery, "query", "q", "", "search query")
	searchCmd.Flags().StringVarP(&optFile, "file", "f", "", "search file")
	rootCmd.AddCommand(searchCmd)
}
