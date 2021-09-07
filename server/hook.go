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

package server

import (
	"fmt"
	"net/http"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/actions"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/video"
)

const (
	IntentPlay = "TAKEOUT_PLAY"
	IntentNew  = "TAKEOUT_NEW"
)

func (handler *UserHandler) hookHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-type", ApplicationJson)

	if r.Method != "POST" {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	a := handler.NewAuth()
	if a == nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}
	defer a.Close()

	handler.user, err = a.UserAuthValue("6c796f84-2267-406c-942b-388e248d1b92")
	if err != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	media := handler.user.FirstMedia()
	if media == "" {
		http.Error(w, "bummer", http.StatusServiceUnavailable)
		return
	}
	path := fmt.Sprintf("%s/%s", handler.config.DataDir, media)

	handler.userConfig, err = config.LoadConfig(path)
	if err != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	mus := handler.NewMusic(w)
	if mus == nil {
		return
	}
	defer mus.Close()

	vid := handler.NewVideo(w)
	if vid == nil {
		return
	}
	defer vid.Close()

	hookRequest := actions.NewWebhookRequest(r)
	hookResponse := actions.NewWebhookResponse(hookRequest)
	fmt.Printf("got %+v\n", hookRequest)

	switch hookRequest.IntentName() {
	case IntentPlay:
		handler.fulfillPlay(hookRequest, hookResponse, mus, vid)
	case IntentNew:
		handler.fulfillNew(hookRequest, hookResponse, mus, vid)
	default:
		handler.fulfillWelcome(hookRequest, hookResponse, mus, vid)
	}

	fmt.Printf("sending %+v\n", hookResponse)
	hookResponse.Send(w)
}

func (handler *UserHandler) fulfillPlay(r *actions.WebhookRequest, w *actions.WebhookResponse,
	m *music.Music, v *video.Video) {
	song := r.SongParam()
	artist := r.ArtistParam()
	release := r.ReleaseParam()

	query := ""
	if artist != "" && song != "" {
		// play [song] by [artist]
		query = fmt.Sprintf(`+artist:"%s" +title:"%s*"`, artist, song)
	} else if artist != "" && release != "" {
		// play album [release] by [artist]
		query = fmt.Sprintf(`+artist:"%s" +release:"%s"`, artist, release)
	} else if artist != "" {
		// play [artist] songs
		// play songs by [artist]
		query = fmt.Sprintf(`+artist:"%s" +type:"single"`, artist)
	} else if song != "" {
		// play song [song]
		// play [song]
		query = fmt.Sprintf(`+title:"%s*"`, song)
	} else if release != "" {
		// play album [release]
		query = fmt.Sprintf(`+release:"%s"`, release)
	}

	var tracks []music.Track
	if query != "" {
		tracks = m.Search(query, 10)
	}
	fmt.Printf("search for %s -> %d tracks\n", query, len(tracks))

	for _, t := range tracks {
		w.AddMedia(t.Title,
			fmt.Sprintf("%s \u2022 %s", t.Artist, t.Release),
			m.TrackURL(&t).String(),
			m.TrackImage(t).String())
	}

	if len(tracks) > 0 {
		addSimple(w, handler.config.Assistant.Play)
	} else {
		addSimple(w, handler.config.Assistant.Error)
	}
}

func (handler *UserHandler) fulfillNew(r *actions.WebhookRequest, w *actions.WebhookResponse,
	m *music.Music, v *video.Video) {
	home := handler.homeView(m, v)
	speech := "Recent additions are "
	text := ""
	for i, rel := range home.AddedReleases {
		if i == 3 {
			break
		} else if i > 0 {
			speech += " and "
			text += ", "
		}
		speech += fmt.Sprintf("%s by %s", rel.Name, rel.Artist)
		text += fmt.Sprintf("%s \u2022 %s", rel.Artist, rel.Name)
	}
	w.AddSimple(speech, text)
}

func (handler *UserHandler) fulfillWelcome(r *actions.WebhookRequest, w *actions.WebhookResponse,
	m *music.Music, v *video.Video) {
	suggest := []string{}
	home := handler.homeView(m, v)
	if len(home.AddedReleases) > 0 {
		release := home.AddedReleases[0]
		suggest = append(suggest, fmt.Sprintf("%s by %s", release.Name, release.Artist))
	}
	addSimple(w, handler.config.Assistant.Welcome)
	w.AddSuggestions(suggest...)
}

func addSimple(w *actions.WebhookResponse, m config.AssistantResponse) {
	w.AddSimple(m.Speech, m.Text)
}
