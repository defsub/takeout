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

package music

import (
	"testing"
	"regexp"
	"strconv"
)

func TestPattern(t *testing.T) {
	var trackRegexp = regexp.MustCompile(`(?:([\d]+)-)?([\d]+)-(.*)\.(mp3|flac|ogg|m4a)$`)
	var numericRegexp = regexp.MustCompile(`^[\d]+([\s-])*`)
	var trackRegexp2 = regexp.MustCompile(`([\d]+)-(.*)\.(mp3|flac|ogg|m4a)$`)

	patterns := []string{
		"Music/Gorillaz/Gorillaz (2002)/01-Re-Hash.flac",
		"Music/Gorillaz/Gorillaz (2002)/02-5_4.flac", // song is 5/4
		"Music/Gorillaz/Gorillaz (2002)/03-Tomorrow Comes Today.flac",
		"Music/Gorillaz/Gorillaz (2002)/08-Sound Check (Gravity).flac",
		"Music/Gorillaz/Gorillaz (2002)/11-19-2000.flac", // song is 19-2000
		"Music/Gorillaz/Gorillaz (2002)/15-M1 A1.flac",
		"Music/Gorillaz/Gorillaz (2002)/18-19-2000 (Soulchild remix).flac",
		"Music/The Velvet Underground/Loaded (1997)/1-01-Who Loves the Sun.flac",
		"Music/The Velvet Underground/Loaded (1997)/2-17-Love Makes You Feel Ten Feet Tall (demo).flac",
		"Music/The Beatles/The Beatles in Mono (2009)/8-12-Sgt. Pepper's Lonely Hearts Club Band (reprise).flac",
		"Music/Swervedriver/Raise (1991)/01-Sci-Flyer.mp3",
		"Music/Stereolab/Transient Random-Noise Bursts With Announcements (1993)/10-Lock-Groove Lullaby.flac",
		"Music/Artist/Album (2020)/01-2020.flac",
		"Music/Artist/Album (2020)/1-02-2020.flac",
		"Music/Artist/Album (2020)/01-02-2020.flac",
		"Music/Artist/Album (2020)/1-2-2020.flac",
		"Music/ZZ Top/XXX (1999)/4-36-22-36.flac",
		"Music/Beastie Boys/Paul's Boutique (1989)/07-3-Minute Rule.flac",
		"Music/Beastie Boys/Paul's Boutique (1989)/09-5-Piece Chicken Dinner.flac",
	}

	for _, v := range patterns {
		matches := trackRegexp.FindStringSubmatch(v)
		if matches == nil {
			t.Errorf("bummer\n")
		}
		disc, _ := strconv.Atoi(matches[1])
		track, _ := strconv.Atoi(matches[2])
		title := matches[3]
		if disc == 0 {
			disc = 1
		}
		if numericRegexp.MatchString(title) {
			matches = trackRegexp2.FindStringSubmatch(v)
			disc = 1
			track, _ = strconv.Atoi(matches[1])
			title = matches[2]
		}
		t.Logf("%d / %d / %s\n", disc, track, title)
	}
}
