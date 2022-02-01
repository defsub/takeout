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

package podcast

import (
	"errors"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (p *Podcast) openDB() (err error) {
	var glog logger.Interface
	if p.config.Podcast.DB.LogMode == false {
		glog = logger.Discard
	} else {
		glog = logger.Default
	}
	cfg := &gorm.Config{
		Logger: glog,
	}

	if p.config.Podcast.DB.Driver == "sqlite3" {
		p.db, err = gorm.Open(sqlite.Open(p.config.Podcast.DB.Source), cfg)
	} else {
		err = errors.New("driver not supported")
	}

	if err != nil {
		return
	}

	p.db.AutoMigrate(&Series{}, &Episode{})
	return
}

func (p *Podcast) closeDB() {
	conn, err := p.db.DB()
	if err != nil {
		return
	}
	conn.Close()

}

func (p *Podcast) Series() []Series {
	var series []Series
	p.db.Order("date desc").Find(&series)
	return series
}

func (p *Podcast) Episodes(series Series) []Episode {
	var episodes []Episode
	p.db.Where(`episodes.s_id = ?`, series.SID).
		Order("date desc").Find(&episodes)
	return episodes
}

func (p *Podcast) RecentEpisodes() []Episode {
	var episodes []Episode
	p.db.Order("date desc").
		Limit(p.config.Podcast.RecentLimit).
		Find(&episodes)
	return episodes
}

func (p *Podcast) RecentSeries() []Series {
	var series []Series
	p.db.Order("date desc").
		Limit(p.config.Podcast.RecentLimit).
		Find(&series)
	return series
}

func (p *Podcast) deleteSeries(sid string) {
	var list []Series
	p.db.Where("s_id = ?", sid).Find(&list)
	for _, o := range list {
		p.db.Unscoped().Delete(o)
	}
}

func (p *Podcast) deleteSeriesEpisodes(sid string) {
	var list []Episode
	p.db.Where("s_id = ?", sid).Find(&list)
	for _, o := range list {
		p.db.Unscoped().Delete(o)
	}
}

func (p *Podcast) deleteEpisode(eid string) {
	var list []Episode
	p.db.Where("e_id = ?", eid).Find(&list)
	for _, o := range list {
		p.db.Unscoped().Delete(o)
	}
}

func (p *Podcast) createSeries(s *Series) error {
	return p.db.Create(s).Error
}

func (p *Podcast) createEpisode(e *Episode) error {
	return p.db.Create(e).Error
}

func (p *Podcast) findSeries(sid string) *Series {
	var list []Series
	p.db.Where("s_id = ?", sid).Find(&list)
	if len(list) > 0 {
		return &list[0]
	}
	return nil
}

func (p *Podcast) findEpisode(eid string) *Episode {
	var list []Episode
	p.db.Where("e_id = ?", eid).Find(&list)
	if len(list) > 0 {
		return &list[0]
	}
	return nil
}

func (p *Podcast) LookupSeries(id int) (Series, error) {
	var series Series
	err := p.db.First(&series, id).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return Series{}, errors.New("series not found")
	}
	return series, err
}

func (p *Podcast) LookupEpisode(id int) (Episode, error) {
	var episode Episode
	err := p.db.First(&episode, id).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return Episode{}, errors.New("episode not found")
	}
	return episode, err
}

func (p *Podcast) SeriesCount() int64 {
	var count int64
	p.db.Model(&Series{}).Count(&count)
	return count
}

func (p *Podcast) retainEpisodes(series *Series, eids []string) error {
	sid := series.SID
	var list []Episode
	p.db.Where("s_id = ? and e_id not in (?)", sid, eids).Find(&list)
	fmt.Printf("will delete %d epsidoes\n", len(list))
	for _, e := range list {
		fmt.Printf("deleting %s : %s\n", e.EID, e.Title)
	}
	return p.db.Unscoped().Delete(Episode{}, "s_id = ? and e_id not in (?)", sid, eids).Error
}
