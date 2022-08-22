// Copyright (C) 2022 The Takeout Authors.
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

package server

import (
	"fmt"
	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/log"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/podcast"
	"github.com/defsub/takeout/video"
)

type Media struct {
	config   *config.Config
	music    *music.Music
	video    *video.Video
	podcast  *podcast.Podcast
}

func (m Media) Config() *config.Config {
	return m.config
}

func (m Media) Music() *music.Music {
	return m.music
}

func (m Media) Podcast() *podcast.Podcast {
	return m.podcast
}

func (m Media) Video() *video.Video {
	return m.video
}

func mediaConfigFor(root *config.Config, user *auth.User) (string, *config.Config, error) {
	// only supports one media collection right now
	mediaName := user.FirstMedia()
	if mediaName == "" {
		return "", nil, ErrNoMedia
	}
	config, err := mediaConfig(root, mediaName)
	return mediaName, config, err
}

func mediaConfig(root *config.Config, mediaName string) (*config.Config, error) {
	path := fmt.Sprintf("%s/%s", root.DataDir, mediaName)
	// load relative media configuration
	userConfig, err := config.LoadConfig(path)
	if err != nil {
		return nil, err
	}
	return userConfig, nil
}

var mediaMap map[string]*Media = make(map[string]*Media)

func makeMedia(name string, config *config.Config) *Media {
	media, ok := mediaMap[name]
	if !ok {
		var err error
		media = &Media{}
		media.music, err = media.makeMusic(config)
		log.CheckError(err)
		media.video, err = media.makeVideo(config)
		log.CheckError(err)
		media.podcast, err = media.makePodcast(config)
		log.CheckError(err)
		mediaMap[name] = media
	}
	return media
}

func (Media) makeMusic(config *config.Config) (*music.Music, error) {
	m := music.NewMusic(config)
	err := m.Open()
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (Media) makeVideo(config *config.Config) (*video.Video, error) {
	v := video.NewVideo(config)
	err := v.Open()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (Media) makePodcast(config *config.Config) (*podcast.Podcast, error) {
	p := podcast.NewPodcast(config)
	err := p.Open()
	if err != nil {
		return nil, err
	}
	return p, nil
}
