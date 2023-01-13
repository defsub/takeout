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
	"strconv"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	"github.com/defsub/takeout/lib/search"
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

func (p *Podcast) newSearch() (*search.Search, error) {
	s := search.NewSearch(p.config)
	s.Keywords = []string{
		FieldAuthor,
		FieldDescription,
		FieldTitle,
	}
	err := s.Open("podcast")
	if err != nil {
		return nil, err
	}
	return s, nil
}

func mergeClientConfig(cfg *config.Config) *config.ClientConfig {
	var merged config.ClientConfig
	merged = cfg.Client
	merged.Merge(cfg.Podcast.Client)
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

func (p *Podcast) FindSeries(identifier string) (Series, error) {
	id, err := strconv.Atoi(identifier)
	if err != nil {
		return p.LookupSID(identifier)
	} else {
		return p.LookupSeries(id)
	}
}

func (p *Podcast) FindEpisode(identifier string) (Episode, error) {
	id, err := strconv.Atoi(identifier)
	if err != nil {
		return p.LookupEID(identifier)
	} else {
		return p.LookupEpisode(id)
	}
}

func (p *Podcast) Search(q string, limit ...int) (series []Series, episodes []Episode) {
	s, err := p.newSearch()
	if err != nil {
		return
	}
	defer s.Close()

	l := p.config.Podcast.SearchLimit
	if len(limit) == 1 {
		l = limit[0]
	}

	keys, err := s.Search(q, l)
	if err != nil {
		return
	}

	seriesMap := make(map[string]bool)

	// split potentially large # of result keys into chunks to query
	chunkSize := 100
	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize
		if end > len(keys) {
			end = len(keys)
		}
		chunk := keys[i:end]
		episodes = append(episodes, p.episodesFor(chunk)...)
		for _, e := range episodes {
			seriesMap[e.SID] = true
		}
	}

	// include unique series for episode results
	seriesKeys := make([]string, 0, len(seriesMap))
	for k := range seriesMap {
		seriesKeys = append(seriesKeys, k)
	}
	series = p.seriesFor(seriesKeys)

	return series, episodes
}
