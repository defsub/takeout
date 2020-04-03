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

package music

type Playlist struct {
	Name     string
	Artists  string
	Releases string
	Titles   string
	Tags     string
	After    string
	Before   string
	Singles  bool
	Popular  bool
	Shuffle  bool
}

type Criteria Playlist

// AllSingles := Criteria{
// 	Name: "All Singles",
// 	Singles: true
// 	Shuffle: true}

func builtin() {

	// alternative/indie
	// dance & electronic
	// hip hop/rap
	// latin
	// metal
	// punk
	// r&b
	// reggae
	// rock
	// soul
	// soundtracks

	// decades 1950s, 1960s ... 2010s, 2020s

	// recently added
	// new releases

	// album radio

	//Playlist{
	//	Name: "Alternative/Indie",
	//	Tags: "altenative,alternative rock,indie,indie rock",
	//	Popular: true,
	//	Shuffle: true}
}
