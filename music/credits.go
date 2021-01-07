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
	"strings"

	"github.com/defsub/takeout/log"
	"github.com/defsub/takeout/search"
)

const (
	FieldArtist      = "artist"
	FieldAsin        = "asin"
	FieldDate        = "date"
	FieldFirstDate   = "first_date"
	FieldGenre       = "genre"
	FieldLabel       = "label"
	FieldLength      = "length"
	FieldMedia       = "media"
	FieldMediaTitle  = "media_title"
	FieldPopularity  = "popularity"
	FieldRating      = "rating"
	FieldRelease     = "release"
	FieldReleaseDate = "release_date"
	FieldSeries      = "series"
	FieldStatus      = "status"
	FieldTag         = "tag"
	FieldTitle       = "title"
	FieldTrack       = "track"
	FieldType        = "type"

	FieldBass      = "base"
	FieldClarinet  = "clarinet"
	FieldDrums     = "drums"
	FieldFlute     = "flute"
	FieldGuitar    = "guitar"
	FieldPiano     = "piano"
	FieldSaxophone = "saxophone"
	FieldVocals    = "vocals"

	TypePopular = "popular"
	TypeSingle  = "single"
)

func (m *Music) creditsIndex(reid string) (search.IndexMap, error) {
	rel, err := m.MusicBrainzRelease(reid)
	if err != nil {
		return nil, err
	}

	index := make(search.IndexMap)
	fields := make(search.FieldMap)

	// general fields
	if rel.Disambiguation != "" {
		addField(fields, FieldRelease, fmt.Sprintf("%s (%s)", rel.Title, rel.Disambiguation))
	} else {
		addField(fields, FieldRelease, rel.Title)
	}
	addField(fields, FieldAsin, rel.Asin)
	addField(fields, FieldStatus, rel.Status)
	if rel.ReleaseGroup.Rating.Votes > 0 {
		addField(fields, FieldRating, rel.ReleaseGroup.Rating.Value)
	}
	for _, l := range rel.LabelInfo {
		addField(fields, FieldLabel, l.Label.Name)
	}

	// dates
	addField(fields, FieldDate, rel.Date) // refined later
	addField(fields, FieldReleaseDate, rel.Date)
	addField(fields, FieldFirstDate, rel.ReleaseGroup.FirstReleaseDate)

	// genres for artist and release group
	for _, a := range rel.ArtistCredit {
		if a.Name == VariousArtists {
			// this has many genres and tags so don't add
			continue
		}
		for _, g := range a.Artist.Genres {
			if g.Count > 0 {
				addField(fields, FieldGenre, g.Name)
			}
		}
		for _, t := range a.Artist.Tags {
			if t.Count > 0 {
				addField(fields, FieldTag, t.Name)
			}
		}
	}
	for _, g := range rel.ReleaseGroup.Genres {
		if g.Count > 0 {
			addField(fields, FieldGenre, g.Name)
		}
	}
	for _, t := range rel.ReleaseGroup.Tags {
		if t.Count > 0 {
			addField(fields, FieldTag, t.Name)
		}
	}

	relationCredits(fields, rel.Relations)

	for _, m := range rel.Media {
		for _, t := range m.Tracks {
			trackFields := search.CloneFields(fields)
			addField(trackFields, FieldMedia, m.Position)
			if m.Title != "" {
				// include media specific title
				// Eagles / The Long Run (Legacy)
				addField(trackFields, FieldMediaTitle, m.Title)
			}
			addField(trackFields, FieldTrack, t.Position)
			addField(trackFields, FieldTitle, t.Recording.Title)
			addField(trackFields, FieldLength, t.Recording.Length/1000)
			relationCredits(trackFields, t.Recording.Relations)
			for _, a := range t.ArtistCredit {
				addField(trackFields, FieldArtist, a.Name)
			}
			key := fmt.Sprintf("%d-%d-%s", m.Position, t.Position, fixName(t.Recording.Title))
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
		FieldBass,
		FieldClarinet,
		FieldDrums,
		FieldFlute,
		FieldGuitar,
		FieldPiano,
		FieldSaxophone,
		FieldVocals,
	}
	for _, alt := range alternates {
		if strings.Contains(key, alt) {
			keys = append(keys, alt)
			// only match one; order matters
			break
		}
	}

	for _, k := range keys {
		k := strings.Replace(k, " ", "_", -1)
		switch value.(type) {
		case string:
			svalue := value.(string)
			svalue = fixName(svalue)
			if v, ok := c[k]; ok {
				switch v.(type) {
				case string:
					// string becomes array of 2 strings
					c[k] = []string{v.(string), svalue}
				case []string:
					// array of 3+ strings
					s := v.([]string)
					s = append(s, svalue)
					c[k] = s
				default:
					panic("bad field types")
				}
			} else {
				// single string
				c[k] = svalue
			}
		default:
			// numeric, date, etc.
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
					log.Printf("** ignore performance work relation '%s'\n", wr.Type)
				}
			}
		} else if "instrument" == r.Type {
			for _, a := range r.Attributes {
				addField(c, a, r.Artist.Name)
			}
		} else if "part of" == r.Type && "series" == r.TargetType {
			addField(c, FieldSeries, r.Series.Name)
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
