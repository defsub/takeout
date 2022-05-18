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
	"strings"

	"github.com/defsub/takeout/lib/date"
	"github.com/defsub/takeout/lib/tmdb"
	"github.com/spf13/cobra"
        "gopkg.in/alessio/shellescape.v1"
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
var optDef string
var optExt string

func fixColon(name string) string {
	// change "foo: bar" to "foo - bar"
	colon := regexp.MustCompile(`([A-Za-z0-9])\s*(:)\s`)
	name = colon.ReplaceAllString(name, "${1} - ")
	return name
}

func doit() {
	var query string
	config := getConfig()
	m := tmdb.NewTMDB(config)
	if optFile != "" {
		fileRegexp := regexp.MustCompile(`([^\/]+)_t\d+(\.mkv)$`)
		matches := fileRegexp.FindStringSubmatch(optFile)
		if matches != nil {
			query = matches[1]
			query = strings.Replace(query, "_", " ", -1)
			if optExt == "" {
				optExt = matches[2]
			}
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
			vars := map[string]interface{}{
				"Title":      fixColon(v.Title),
				"Year":       date.ParseDate(v.ReleaseDate).Year(),
				"Definition": optDef,
				"Def":        optDef,
				"Extension":  optExt,
				"Ext":        optExt,
			}
			title := config.TMDB.FileTemplate.Execute(vars)

			vars["Ext"] = ".jpg"
			vars["Extension"] = ".jpg"
			cover := config.TMDB.FileTemplate.Execute(vars)

			fmt.Printf("%s\n", title)
			poster := m.OriginalPoster(v.PosterPath).String()
			fmt.Printf("%s\n", poster)
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
				fmt.Printf("mv %s %s\n", shellescape.Quote(optFile),
					shellescape.Quote(title))
				fmt.Printf("wget -O %s %s\n", shellescape.Quote(cover),
					shellescape.Quote(poster))
			}
			fmt.Printf("\n")
		}
	}
}

func init() {
	searchCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	searchCmd.Flags().StringVarP(&optQuery, "query", "q", "", "search query")
	searchCmd.Flags().StringVarP(&optFile, "file", "f", "", "search file")
	searchCmd.Flags().StringVarP(&optDef, "def", "d", "", "SD, HD, UHD, 4k, etc")
	searchCmd.Flags().StringVarP(&optDef, "ext", "e", "", "file extension w/ dot")
	rootCmd.AddCommand(searchCmd)
}
