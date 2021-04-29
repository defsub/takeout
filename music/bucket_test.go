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
	"fmt"
	"testing"
	"regexp"
	"strconv"
)

func TestPattern(t *testing.T) {
	var trackRegexp = regexp.MustCompile(`(?:([1-9]+)-)?([\d]+)-(.*)\.(mp3|flac|ogg|m4a)$`)
	var track2Regexp = regexp.MustCompile(`([\d]+)-(.*)\.(mp3|flac|ogg|m4a)$`)

	patterns := []string{
		"Music/Abc/Def/1-Sub.flac",
		"Music/Abc/Def/01-Sub.flac",
		"Music/Abc/Def/2-01-Sub.flac",
		"Music/Abc/Def (2021)/1-Sub.flac",
		"Music/Abc/Def (2021)/01-Sub.flac",
		"Music/Abc/Def (2021)/2-01-Sub.flac",
		"Music/Gorillaz/Gorillaz (2002)/01-Re-Hash.flac",
		"Music/Gorillaz/Gorillaz (2002)/02-5_4.flac", // song is 5/4
		"Music/Gorillaz/Gorillaz (2002)/03-Tomorrow Comes Today.flac",
		"Music/Gorillaz/Gorillaz (2002)/08-Sound Check (Gravity).flac",
		"Music/Gorillaz/Gorillaz (2002)/11-19-2000.flac", // FAIL - song is 19-2000
		"Music/Gorillaz/Gorillaz (2002)/15-M1 A1.flac",
		"Music/Gorillaz/Gorillaz (2002)/18-19-2000 (Soulchild remix).flac", // FAIL
		"Music/The Velvet Underground/Loaded (1997)/1-01-Who Loves the Sun.flac",
		"Music/The Velvet Underground/Loaded (1997)/2-17-Love Makes You Feel Ten Feet Tall (demo).flac",
		"Music/The Beatles/The Beatles in Mono (2009)/8-12-Sgt. Pepper's Lonely Hearts Club Band (reprise).flac",
		"Music/Swervedriver/Raise (1991)/01-Sci-Flyer.mp3",
		"Music/Stereolab/Transient Random-Noise Bursts With Announcements (1993)/10-Lock-Groove Lullaby.flac",
		"Music/Artist/Album (2020)/01-2020.flac",
		"Music/Artist/Album (2020)/1-02-2020.flac",
		"Music/Artist/Album (2020)/01-02-2020.flac",
		"Music/Artist/Album (2020)/1-2-2020.flac",
		"Music/Artist/Album (2020)/11-12 2020.flac",
		"Music/Artist/Album (2020)/11-12 Twenty.flac",
		"Music/Artist/Album (2020)/11-12-2020.flac",
		"Music/Artist/Album (2020)/11-12-Twenty.flac",
		"Music/ZZ Top/XXX (1999)/4-36-22-36.flac", // FAIL
		"Music/Beastie Boys/Paul's Boutique (1989)/07-3-Minute Rule.flac",
		"Music/Beastie Boys/Paul's Boutique (1989)/09-5-Piece Chicken Dinner.flac",
		"Music/Iron Maiden/Nights of the Dead Legacy of the Beast Live in Mexico City (2020)/1-04-2 Minutes to Midnight.flac",
		"Music/N.W.A/The Best of N.W.A - The Strength of Street Knowledge (2006)/06-8-Ball.flac",
	}

	expect := []string{
		"1 / 1 / Sub",
		"1 / 1 / Sub",
		"2 / 1 / Sub",
		"1 / 1 / Sub",
		"1 / 1 / Sub",
		"2 / 1 / Sub",
		"1 / 1 / Re-Hash",
		"1 / 2 / 5_4",
		"1 / 3 / Tomorrow Comes Today",
		"1 / 8 / Sound Check (Gravity)",
		"1 / 11 / 19-2000",
		"1 / 15 / M1 A1",
		"1 / 18 / 19-2000 (Soulchild remix)",
		"1 / 1 / Who Loves the Sun",
		"2 / 17 / Love Makes You Feel Ten Feet Tall (demo)",
		"8 / 12 / Sgt. Pepper's Lonely Hearts Club Band (reprise)",
		"1 / 1 / Sci-Flyer",
		"1 / 10 / Lock-Groove Lullaby",
		"1 / 1 / 2020",
		"1 / 2 / 2020",
		"1 / 1 / 02-2020",
		"1 / 2 / 2020",
		"1 / 11 / 12 2020",
		"1 / 11 / 12 Twenty",
		"11 / 12 / 2020",
		"11 / 12 / Twenty",
		"1 / 4 / 36-22-36",
		"1 / 7 / 3-Minute Rule",
		"1 / 9 / 5-Piece Chicken Dinner",
		"1 / 4 / 2 Minutes to Midnight",
		"1 / 6 / 8-Ball",
	}

	for i, v := range patterns {
		matches := trackRegexp.FindStringSubmatch(v)
		if matches == nil {
			t.Errorf("bummer\n")
			break
		}
		disc, _ := strconv.Atoi(matches[1])
		track, _ := strconv.Atoi(matches[2])
		title := matches[3]
		if disc == 0 {
			disc = 1
		}

		result := fmt.Sprintf("%d / %d / %s", disc, track, title)
		t.Logf("%s\n", result)
		if result != expect[i] {
			t.Logf("Expected: %s\n", expect[i])

			matches := track2Regexp.FindStringSubmatch(v)
			if matches == nil {
				t.Errorf("bummer\n")
				break
			}
			disc = 1
			track, _ = strconv.Atoi(matches[1])
			title = matches[2]

			result = fmt.Sprintf("%d / %d / %s", disc, track, title)
			t.Logf("now: %s\n", result)
		}

	}
}
