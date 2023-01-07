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
	"strings"
	"time"

	"github.com/defsub/takeout/lib/hash"
	"github.com/defsub/takeout/lib/rss"
)

func (p *Podcast) Sync() error {
	return p.SyncSince(time.Time{})
}

func (p *Podcast) SyncSince(lastSync time.Time) error {
	// TODO lastSync isn't used yet
	for _, url := range p.config.Podcast.Series {
		err := p.syncPodcast(url)
		if err != nil {
			fmt.Printf("err %s\n", err)
			return err
		}
	}
	// TODO cleanup old series and episodes
	return nil
}

func (p *Podcast) syncPodcast(url string) error {
	rss := rss.NewRSS(p.client)
	channel, err := rss.Fetch(url)
	if err != nil {
		return err
	}
	if channel.Link() == "" {
		return errors.New("empty channel link")
	}
	sid := hash.MD5Hex(channel.Link())

	series := p.findSeries(sid)
	if series == nil {
		series = &Series{
			SID:         sid,
			Title:       channel.Title,
			Author:      channel.Author,
			Description: channel.Description,
			Link:        channel.Link(),
			Image:       channel.Image.URL,
			Copyright:   channel.Copyright,
			Date:        channel.LastBuildTime(),
			TTL:         channel.TTL,
		}
		err = p.createSeries(series)
		if err != nil {
			return err
		}
	} else {
		// TODO update everything for now
		series.Title = channel.Title
		series.Author = channel.Author
		series.Description = channel.Description
		series.Link = channel.Link()
		series.Image = channel.Image.URL
		series.Copyright = channel.Copyright
		series.Date = channel.LastBuildTime()
		if series.Date.IsZero() {
			// no build time so use latest episode publish time
			latest := time.Time{}
			for _, i := range channel.Items {
				t := i.PublishTime()
				if t.After(latest) {
					latest = t
				}
			}
			series.Date = latest
		}
		series.TTL = channel.TTL
		err := p.db.Save(series).Error
		if err != nil {
			return err
		}
	}

	var episodes []string
	for _, i := range channel.Items {
		var eid string
		if strings.Contains(i.GUID, "://") {
			// some GUIDs are URLs, hash them
			eid = hash.MD5Hex(i.GUID)
		} else {
			eid = i.GUID
		}
		episode := p.findEpisode(eid)
		if episode == nil {
			episode = &Episode{
				SID:         sid,
				EID:         eid,
				Title:       i.ItemTitle(),
				Author:      i.Author,
				Link:        i.Link,
				Description: i.Description,
				ContentType: i.ContentType(),
				Size:        i.Size(),
				URL:         i.URL(),
				Date:        i.PublishTime(),
			}
			err = p.createEpisode(episode)
			if err != nil {
				return err
			}
		} else {
			// TODO update everything for now
			episode.Title = i.ItemTitle()
			episode.Author = i.Author
			episode.Link = i.Link
			episode.Description = i.Description
			episode.ContentType = i.ContentType()
			episode.Size = i.Size()
			episode.URL = i.URL()
			episode.Date = i.PublishTime()
			err := p.db.Save(episode).Error
			if err != nil {
				return err
			}
		}
		episodes = append(episodes, eid)
	}
	p.retainEpisodes(series, episodes)
	return nil
}
