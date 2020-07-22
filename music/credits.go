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
	"github.com/defsub/takeout/search"
	"strings"
)

func (m *Music) creditsIndex(reid string) (search.IndexMap, error) {
	rel, err := m.MusicBrainzReleaseCredits(reid)
	if err != nil {
		return nil, err
	}

	index := make(search.IndexMap)
	fields := make(search.FieldMap)

	// general fields
	if rel.Disambiguation != "" {
		addField(fields, "release", fmt.Sprintf("%s (%s)", rel.Title, rel.Disambiguation))
	} else {
		addField(fields, "release", rel.Title)
	}
	addField(fields, "asin", rel.Asin)
	addField(fields, "status", rel.Status)
	if rel.ReleaseGroup.Rating.Votes > 0 {
		addField(fields, "rating", rel.ReleaseGroup.Rating.Value)
	}
	for _, l := range rel.LabelInfo {
		addField(fields, "label", l.Label.Name)
	}

	// dates
	addField(fields, "date", rel.Date) // refined later
	addField(fields, "release_date", rel.Date)
	addField(fields, "first_date", rel.ReleaseGroup.FirstReleaseDate)

	// genres for artist and release group
	for _, a := range rel.ArtistCredit {
		for _, g := range a.Artist.Genres {
			addField(fields, "genre", g.Name)
		}
		for _, t := range a.Artist.Tags {
			addField(fields, "tag", t.Name)
		}
	}
	for _, g := range rel.ReleaseGroup.Genres {
		addField(fields, "genre", g.Name)
	}
	for _, t := range rel.ReleaseGroup.Tags {
		addField(fields, "tag", t.Name)
	}

	relationCredits(fields, rel.Relations)

	for _, m := range rel.Media {
		for _, t := range m.Tracks {
			trackFields := search.CloneFields(fields)
			addField(trackFields, "media", m.Position)
			if m.Title != "" {
				// include media specific title
				// Eagles / The Long Run (Legacy)
				addField(trackFields, "media_title", m.Title)
			}
			addField(trackFields, "track", t.Position)
			addField(trackFields, "title", t.Recording.Title)
			addField(trackFields, "length", t.Recording.Length / 1000)
			relationCredits(trackFields, t.Recording.Relations)
			for _, a := range t.ArtistCredit {
				addField(trackFields, "artist", a.Name)
			}
			key := fmt.Sprintf("%d-%d-%s", m.Position, t.Position, t.Recording.Title)
			index[key] = trackFields
		}
	}

	return index, nil
}

func addField(c search.FieldMap, key string, value interface{}) search.FieldMap {
	key = strings.ToLower(key)
	keys := []string{key}

	// drums = drums (drum set)
	// guitar = lead guitar, slide guitar, rhythm guitar, acoustic
	// bass = bass guitar, electric bass guitar
	// vocals = lead vocals, backing vocals
	alternates := []string{
		"bass",
		"clarinet",
		"drums",
		"flute",
		"guitar",
		"piano",
		"saxophone",
		"vocals",
	}
	for _, alt := range alternates {
		if strings.Contains(key, alt) {
			keys = append(keys, alt)
			// only match one; order matters
			break;
		}
	}

	for _, k := range keys {
		k := strings.Replace(k, " ", "_", -1)
		switch value.(type) {
		case string:
			if v, ok := c[k]; ok {
				c[k] = v.(string) + ", " + value.(string)
			} else {
				c[k] = value
			}
		default:
			c[k] = value
		}
	}
	return c
}

func relationCredits(c search.FieldMap, relations []mbzRelation) search.FieldMap {
	for _, r := range relations {
		if "performance" == r.Type {
			for _, wr := range r.Work.Relations {
				switch wr.Type {
				case "arranger", "arrangement", "composer",
					"lyricist", "orchestrator", "orchestration",
					"writer":
					addField(c, wr.Type, wr.Artist.Name)
				case "based on", "medley", "misc",
					"instrument arranger", "named after",
					"other version", "revised by",
					"revision of", "parts", "premiere",
					"publishing", "translator", "vocal arranger":
					// ignore these
				default:
					fmt.Printf("** ignore performance work relation '%s'\n", wr.Type)
				}
			}
		} else if "instrument" == r.Type {
			for _, a := range r.Attributes {
				addField(c, a, r.Artist.Name)
			}
		} else {
			if len(r.Attributes) > 0 {
				attr := r.Attributes[0]
				switch attr {
				case "co":
					addField(c, fmt.Sprintf("%s-%s", r.Attributes[0], r.Type), r.Artist.Name)
				case "additional", "assistant":
					addField(c, fmt.Sprintf("%s %s", r.Attributes[0], r.Type), r.Artist.Name)
				case "lead vocals":
					addField(c, attr, r.Artist.Name)
				}
			} else {
				addField(c, r.Type, r.Artist.Name)
			}
		}
	}
	return c
}
