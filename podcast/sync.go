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
	"time"

	"github.com/defsub/takeout/lib/hash"
	"github.com/defsub/takeout/lib/rss"
)

func (p *Podcast) Sync() error {
	return p.SyncSince(time.Time{})
}

func (p *Podcast) SyncSince(lastSync time.Time) error {
	for _, url := range p.config.Podcast.Series {
		err := p.syncPodcast(url)
		if err != nil {
			fmt.Printf("err %s\n", err)
			return err
		}
	}
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
	fmt.Printf("syncing %s %s %s\n", sid, url, channel.Link())
	p.deleteSeriesEpisodes(sid)
	p.deleteSeries(sid)
	series := Series{
		SID:         sid,
		Title:       channel.Title,
		Description: channel.Description,
		Link:        channel.Link(),
		Image:       channel.Image.URL,
		Copyright:   channel.Copyright,
		Date:        channel.LastBuildTime(),
		TTL:         channel.TTL,
	}
	err = p.createSeries(&series)
	if err != nil {
		return err
	}
	for _, i := range channel.Items {
		episode := Episode{
			SID:         sid,
			EID:         hash.MD5Hex(i.GUID),
			Title:       i.ItemTitle(),
			Link:        i.Link,
			Description: i.Description,
			ContentType: i.ContentType(),
			Size:        i.Size(),
			URL:         i.URL(),
			Date:        i.PublishTime(),
		}
		err = p.createEpisode(&episode)
		if err != nil {
			return err
		}
	}
	return nil
}
