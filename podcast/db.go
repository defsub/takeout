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

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (p *Podcast) openDB() (err error) {
	cfg := p.config.Music.DB.GormConfig()

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

func (p *Podcast) LookupSID(sid string) (Series, error) {
	var series Series
	err := p.db.First(&series, "s_id = ?", sid).Error
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

func (p *Podcast) LookupEID(eid string) (Episode, error) {
	var episode Episode
	err := p.db.First(&episode, "e_id = ?", eid).Error
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

func (p *Podcast) retainEpisodes(series *Series, eids []string) ([]string, error) {
	sid := series.SID
	var list []Episode
	var removed []string
	p.db.Where("s_id = ? and e_id not in (?)", sid, eids).Find(&list)
	for _, e := range list {
		removed = append(removed, e.EID)
	}
	err := p.db.Unscoped().Delete(Episode{}, "s_id = ? and e_id not in (?)", sid, eids).Error
	return removed, err
}

func (p *Podcast) search(q string) ([]Series, []Episode) {
	var series []Series
	var episodes []Episode
	query := "%" + q + "%"
	p.db.Where("title like ? or author like ? or description like ?", query, query, query).Find(&series)
	p.db.Where("title like ? or author like ? or description like ?", query, query, query).Find(&episodes)
	return series, episodes
}

func (p *Podcast) episodesFor(keys []string) []Episode {
	var episodes []Episode
	p.db.Where("e_id in (?)", keys).Find(&episodes)
	return episodes
}

func (p *Podcast) seriesFor(keys []string) []Series {
	var series []Series
	p.db.Where("s_id in (?)", keys).Find(&series)
	return series
}
