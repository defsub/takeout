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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/actions"
)

const (
	IntentPlay          = "PLAY"
	MediaTypeAudio      = "AUDIO"
	MediaControlPaused  = "PAUSED"
	MediaControlStopped = "STOPPED"
)

func (handler *UserHandler) hookHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	handler.user = &auth.User{Name: "xxx", Media: "xxx"}
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

	var hookRequest actions.WebhookRequest
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &hookRequest)
	if err != nil {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	fmt.Printf("got %+v\n", hookRequest)

	var hookResponse actions.WebhookResponse
	hookResponse.AddSession(hookRequest.Session.ID)

	if hookRequest.IntentName() == IntentPlay {
		artist := hookRequest.ArtistParam()
		song := hookRequest.SongParam()

		music := handler.NewMusic(w, r)
		defer music.Close()

		query := ""
		if artist != "" && song != "" {
			query = fmt.Sprintf(`+artist:"%s" +title:"%s"`, artist, song)
		} else if artist != "" {
			query = fmt.Sprintf(`+artist:"%s" +type:single`, artist)
		} else if song != "" {
			query = fmt.Sprintf(`+title:"%s"`, song)
		}
		tracks := music.Search(query, 10)
		fmt.Printf("searching for %s\n", query)
		fmt.Printf("got %d\n", len(tracks))

		for _, t := range tracks {
			hookResponse.AddMedia(t.Title,
				fmt.Sprintf("%s \u2022 %s", t.Artist, t.Release),
				music.TrackURL(&t).String(),
				music.TrackImage(t).String())
		}

		speech := ""
		if len(tracks) > 0 {
			speech = "Enjoy the music"
		} else {
			speech = "Sorry try again"
		}
		hookResponse.AddSimple(speech, speech)

	} else {
		hookResponse.AddSimple("Welcome to Takeout", "Welcome to Takeout")
	}

	fmt.Printf("sending %+v\n", hookResponse)

	enc := json.NewEncoder(w)
	enc.Encode(hookResponse)
}
