// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// Takeout is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"log"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/music"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		log.Fatalln(err)
	}

	m := music.NewMusic(config)
	m.Open()
	defer m.Close()
	// m.SyncBucketTracks()
	// m.SyncArtists()
	// m.SyncReleases()
	m.SyncPopular()
	//m.SyncSimilar()
	//m.FixTrackReleases()
}
