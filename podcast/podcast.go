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
	"net/url"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	"gorm.io/gorm"
)

type Podcast struct {
	config *config.Config
	db     *gorm.DB
	client *client.Client
}

func NewPodcast(config *config.Config) *Podcast {
	return &Podcast{
		config: config,
		client: client.NewClient(mergeClientConfig(config)),
	}
}

func mergeClientConfig(cfg *config.Config) *config.Config {
	var merged config.Config
	merged.Client = cfg.Client
	merged.Client.Merge(cfg.Podcast.Client)
	return &merged
}

func (p *Podcast) Open() (err error) {
	err = p.openDB()
	return
}

func (p *Podcast) Close() {
	p.closeDB()
}

func (p *Podcast) SeriesImage(series Series) string {
	return series.Image
}

// TODO expire cache
var seriesImageCache = make(map[string]string)

func (p *Podcast) EpisodeImage(episode Episode) string {
	if v, ok := seriesImageCache[episode.SID]; ok {
		return v
	}
	series := p.findSeries(episode.SID)
	img := ""
	if series != nil {
		img = p.SeriesImage(*series)
	}
	seriesImageCache[episode.SID] = img
	return img
}

func (p *Podcast) HasPodcasts() bool {
	return p.SeriesCount() > 0
}

func (p *Podcast) EpisodeURL(e Episode) *url.URL {
	u, err := url.Parse(e.URL)
	if err != nil {
		// TODO
		return nil
	}
	return u
}
