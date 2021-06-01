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

package video

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/defsub/takeout/lib/bucket"
	"github.com/defsub/takeout/lib/date"
	"github.com/defsub/takeout/lib/str"
	"github.com/defsub/takeout/lib/tmdb"
)

func (v *Video) Sync() error {
	for _, bucket := range v.buckets {
		err := v.syncBucket(bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *Video) syncBucket(bucket *bucket.Bucket) error {
	objectCh, err := bucket.List(time.Time{})
	if err != nil {
		return err
	}

	client := tmdb.NewTMDB(v.config)

	// Movies/Thriller/Zero Dark Thirty (2012).mkv
	// Movies/Thriller/Zero Dark Thirty (2012) - HD.mkv
	movieRegexp := regexp.MustCompile(`.*/(.+?)\s*\(([\d]+)\)(\s-\s(.+))?\.(mkv|mp4)$`)

	for o := range objectCh {
		matches := movieRegexp.FindStringSubmatch(o.Key)
		if matches == nil {
			//fmt.Printf("no match -- %s\n", o.Key)
			continue
		}
		title := matches[1]
		year := matches[2]
		opt := ""
		if len(matches) > 3 {
			opt = matches[4]
		}
		fmt.Printf("%s (%s) - %s\n", title, year, opt)

		results, err := client.MovieSearch(title)
		if err != nil {
			fmt.Printf("err is %s\n", err)
			continue
		}

		for _, r := range results {
			if fuzzyName(title) == fuzzyName(r.Title) &&
				strings.Contains(r.ReleaseDate, year) {
				fmt.Printf("--> matched: %s (%s)\n", r.Title, r.ReleaseDate)
				v.syncMovie(client, r.ID,
					o.Key, o.Size, o.ETag, o.LastModified)
			}
		}
	}
	return nil
}

func fuzzyName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(name, "")
}

func (v *Video) syncMovie(client *tmdb.TMDB, tmid int,
	key string, size int64, etag string, lastModified time.Time) error {
	v.deleteMovie(tmid)
	v.deleteCast(tmid)
	v.deleteCollections(tmid)
	v.deleteCrew(tmid)
	v.deleteGenres(tmid)

	detail, err := client.MovieDetail(tmid)
	if err != nil {
		return err
	}

	m := Movie{
		TMID:             int64(detail.ID),
		IMID:             detail.IMDB_ID,
		Title:            detail.Title,
		SortTitle:        str.SortTitle(detail.Title),
		OriginalTitle:    detail.OriginalTitle,
		OriginalLanguage: detail.OriginalLanguage,
		BackdropPath:     detail.BackdropPath,
		PosterPath:       detail.PosterPath,
		Budget:           detail.Budget,
		Revenue:          detail.Revenue,
		Overview:         detail.Overview,
		Tagline:          detail.Tagline,
		Runtime:          detail.Runtime,
		Date:             date.ParseDate(detail.ReleaseDate), // 2013-02-06
		Key:              key,
		Size:             size,
		ETag:             etag,
		LastModified:     lastModified,
	}
	// rating / certification
	for _, country := range v.config.Video.ReleaseCountries {
		release, err := v.certification(client, tmid, country)
		if err != nil {
			return err
		}
		if release != nil {
			m.Rating = release.Certification
			break
		}
	}
	//fmt.Printf("%s create\n", m.Title)
	err = v.createMovie(&m)
	if err != nil {
		return nil
	}

	// collections
	if detail.Collection.Name != "" {
		c := Collection{
			TMID:     m.TMID,
			Name:     detail.Collection.Name,
			SortName: str.SortTitle(detail.Collection.Name),
		}
		//fmt.Printf("%s collection %s\n", m.Title, c.Name)
		err = v.createCollection(&c)
		if err != nil {
			return err
		}
	}

	// genres
	for _, o := range detail.Genres {
		g := Genre{
			Name: o.Name,
			TMID: m.TMID,
		}
		//fmt.Printf("%s genre %s\n", m.Title, g.Name)
		err = v.createGenre(&g)
		if err != nil {
			return err
		}
	}

	// credits
	credits, err := client.MovieCredits(tmid)
	if err != nil {
		return err
	}
	// cast
	sort.Slice(credits.Cast, func(i, j int) bool {
		// sort by order
		return credits.Cast[i].Order < credits.Cast[j].Order
	})
	for i, o := range credits.Cast {
		if i > v.config.Video.CastLimit {
			break
		}
		p, err := v.Person(o.ID)
		if p == nil {
			// person detail
			p, err = personDetail(client, o.ID)
			if err != nil {
				return err
			}
			//fmt.Printf("%s cast person %s -> %s\n", m.Title, p.Name, o.Character)
			err = v.createPerson(p)
			if err != nil {
				return err
			}
		}
		c := Cast{
			TMID:      m.TMID,
			PEID:      p.PEID,
			Character: o.Character,
			Rank:      o.Order,
		}
		err = v.createCast(&c)
		if err != nil {
			return err
		}
	}

	// DEPARTMENT - JOB
	// Production - Producer, Executive Producer, Casting, Production Coordinator...
	// Directing - Director, Script Supervisor
	// Writing - Story, Screenplay, Novel, Characters
	//deptRegexp := regexp.MustCompile(`^(Production|Directing|Writing)$`)
	jobRegexp := regexp.MustCompile("^(" + strings.Join(v.config.Video.CrewJobs, "|") + ")$")
	for _, o := range credits.Crew {
		matches := jobRegexp.FindStringSubmatch(o.Job)
		if matches == nil {
			// ignore other jobs
			continue
		}
		p, err := v.Person(o.ID)
		if p == nil {
			// person detail
			p, err = personDetail(client, o.ID)
			if err != nil {
				return err
			}
			//fmt.Printf("%s crew person %s -> %s\n", m.Title, p.Name, o.Job)
			err = v.createPerson(p)
			if err != nil {
				return err
			}
		}
		c := Crew{
			TMID:       m.TMID,
			PEID:       p.PEID,
			Department: o.Department,
			Job:        o.Job,
		}
		err = v.createCrew(&c)
		if err != nil {
			return err
		}
	}

	return nil
}

func personDetail(client *tmdb.TMDB, peid int) (*Person, error) {
	detail, err := client.PersonDetail(peid)
	if err != nil {
		return nil, err
	}
	p := Person{
		PEID:        int64(peid),
		IMID:        detail.IMDB_ID,
		Name:        detail.Name,
		ProfilePath: detail.ProfilePath,
		Bio:         detail.Biography,
		Birthplace:  detail.Birthplace,
		Birthday:    date.ParseDate(detail.Birthday),
		Deathday:    date.ParseDate(detail.Deathday),
	}
	return &p, nil
}

func (v *Video) certification(client *tmdb.TMDB, tmid int, country string) (*tmdb.Release, error) {
	types := []int{tmdb.TypeTheatrical, tmdb.TypeDigital}
	for _, t := range types {
		release, err := client.MovieReleaseType(tmid, country, t)
		if err != nil {
			return nil, err
		}
		if release != nil {
			return release, err
		}
	}
	return nil, nil
}