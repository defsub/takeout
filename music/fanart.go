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
)

type fanart struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Likes string `json:"likes"`
}

type fanartAlbum struct {
	AlbumCovers []fanart `json:"albumcover"`
	CDArtwork   []fanart `json:"cdart"`
}

type fanartArtist struct {
	Name              string                 `json:"name"`
	MBID              string                 `json:"mbid_id"`
	Albums            map[string]fanartAlbum `json:"albums"`
	ArtistBackgrounds []fanart               `json:"artistbackground"`
	ArtistThumbs      []fanart               `json:"artistthumb"`
	HDMusicLogos      []fanart               `json:"hdmusiclogo"`
	MusicLogos        []fanart               `json:"musiclogo"`
	MusicBanners      []fanart               `json:"musicbanner"`
}

func (m *Music) fanartArtistArt(artist *Artist) *fanartArtist {
	key := m.config.Fanart.PersonalKey
	if key == "" {
		key = m.config.Fanart.ProjectKey
	}
	if key == "" {
		return nil
	}

	url := fmt.Sprintf("http://webservice.fanart.tv/v3/music/%s?api_key=%s",
		artist.ARID, key)

	var result fanartArtist
	m.client.GetJson(url, &result)
	return &result
}
