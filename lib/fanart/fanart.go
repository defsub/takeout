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

package fanart

import (
	"fmt"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
)

type Fanart struct {
	config *config.Config
	client *client.Client
}

func NewFanart(config *config.Config) *Fanart {
	return &Fanart{
		config: config,
		client: client.NewClient(config),
	}
}

type Art struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Likes string `json:"likes"`
}

type Album struct {
	AlbumCovers []Art `json:"albumcover"`
	CDArtwork   []Art `json:"cdart"`
}

type Artist struct {
	Name              string           `json:"name"`
	MBID              string           `json:"mbid_id"`
	Albums            map[string]Album `json:"albums"`
	ArtistBackgrounds []Art            `json:"artistbackground"`
	ArtistThumbs      []Art            `json:"artistthumb"`
	HDMusicLogos      []Art            `json:"hdmusiclogo"`
	MusicLogos        []Art            `json:"musiclogo"`
	MusicBanners      []Art            `json:"musicbanner"`
}

func (f *Fanart) ArtistArt(arid string) *Artist {
	key := f.config.Fanart.PersonalKey
	if key == "" {
		key = f.config.Fanart.ProjectKey
	}
	if key == "" {
		return nil
	}

	url := fmt.Sprintf("http://webservice.fanart.tv/v3/music/%s?api_key=%s",
		arid, key)

	var result Artist
	f.client.GetJson(url, &result)
	return &result
}
