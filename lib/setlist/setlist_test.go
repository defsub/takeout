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

package setlist

import (
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	"testing"
	"fmt"
)

func TestSetlist(t *testing.T) {
	// radiohead
	//arid := "a74b1b7f-71a5-4011-9441-d0b5e4122711"
	// iron maiden
	//arid := "ca891d65-d9b0-4258-89f7-e6ba29d83767"
	// numan
	arid := "6cb79cb2-9087-44d4-828b-5c6fdff2c957"

	config, err := config.TestConfig()
	if err != nil {
		t.Errorf("GetConfig %s\n", err)
	}

	s := NewSetlist(config, client.NewClient(config))
	result := s.ArtistYear(arid, 2001)

	for _, sl := range result {
		fmt.Printf("%s %s @ %s, %s, %s\n", sl.Tour.Name, sl.EventDate,
			sl.Venue.Name, sl.Venue.City.Name, sl.Venue.City.Country.Name)
		for _, v := range sl.Sets.Set {
			if v.Encore == 0 {
				fmt.Printf("set %d\n", len(v.Songs))
			} else {
				fmt.Printf("encore %d\n", len(v.Songs))
			}
		}
	}
}
