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
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/actions"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/video"
)

const (
	IntentAuth = "TAKEOUT_AUTH"
	IntentPlay = "TAKEOUT_PLAY"
	IntentNew  = "TAKEOUT_NEW"

	UserParamCookie = "cookie"
)

func (handler *UserHandler) hookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", ApplicationJson)

	if r.Method != "POST" {
		http.Error(w, "bummer", http.StatusInternalServerError)
		return
	}

	hookRequest := actions.NewWebhookRequest(r)
	hookResponse := actions.NewWebhookResponse(hookRequest)
	fmt.Printf("got intent=%s %+v\n",  hookRequest.IntentName(), hookRequest)
	fmt.Printf("got user %+v\n", *hookRequest.User)
	if hookRequest.User != nil && hookRequest.User.Params != nil {
		for k, v := range hookRequest.User.Params {
			fmt.Printf("request user[%s]=%s\n", k, v)
		}
	}
	if hookRequest.Session != nil && hookRequest.Session.Params != nil {
		for k, v := range hookRequest.Session.Params {
			fmt.Printf("request session[%s]=%s\n", k, v)
		}
	}

	if !hookRequest.Verified() {
		handler.verificationRequired(hookRequest, hookResponse)
	} else {
		a := handler.NewAuth()
		if a == nil {
			http.Error(w, "bummer", http.StatusInternalServerError)
			return
		}
		defer a.Close()

		cookie := hookRequest.UserParam(UserParamCookie)
		if cookie == "" {
			if hookRequest.IntentName() == IntentAuth {
				// try to authenticate
				handler.authNext(hookRequest, hookResponse, a)
			} else {
				handler.authRequired(hookRequest, hookResponse, a)
			}
		} else if !handler.authCheck(hookRequest, hookResponse, a, cookie) {
			handler.authRequired(hookRequest, hookResponse, a)
		} else {
			handler.fulfillIntent(w, hookRequest, hookResponse)
		}
	}

	fmt.Printf("sending %+v\n", hookResponse)
	if hookResponse.User != nil && hookResponse.User.Params != nil {
		for k, v := range hookResponse.User.Params {
			fmt.Printf("response user[%s]=%s\n", k, v)
		}
	}
	if hookResponse.Session != nil && hookResponse.Session.Params != nil {
		for k, v := range hookResponse.Session.Params {
			fmt.Printf("response session[%s]=%s\n", k, v)
		}
	}
	hookResponse.Send(w)
}

func (handler *UserHandler) fulfillIntent(resp http.ResponseWriter,
	r *actions.WebhookRequest, w *actions.WebhookResponse) {
	var err error
	media := handler.user.FirstMedia()
	if media == "" {
		http.Error(resp, "bummer", http.StatusServiceUnavailable)
		return
	}
	path := fmt.Sprintf("%s/%s", handler.config.DataDir, media)

	handler.userConfig, err = config.LoadConfig(path)
	if err != nil {
		http.Error(resp, "bummer", http.StatusInternalServerError)
		return
	}

	mus := handler.NewMusic(resp)
	if mus == nil {
		return
	}
	defer mus.Close()

	vid := handler.NewVideo(resp)
	if vid == nil {
		return
	}
	defer vid.Close()

	switch r.IntentName() {
	case IntentPlay:
		handler.fulfillPlay(r, w, mus, vid)
	case IntentNew:
		handler.fulfillNew(r, w, mus, vid)
	default:
		handler.fulfillWelcome(r, w, mus, vid)
	}
}

func (handler *UserHandler) artistPopular(m *music.Music, a *music.Artist) []music.Track {
	return m.ArtistPopularTracks(*a, handler.config.Assistant.TrackLimit)
}

func (handler *UserHandler) releaseLike(m *music.Music, release string) []music.Track {
	var tracks []music.Track
	releases := m.ReleasesLike("%"+release+"%")
	if len(releases) > 0 {
		r := releases[0]
		tracks = m.ReleaseTracks(r)
	}
	return tracks
}

func (handler *UserHandler) fulfillPlay(r *actions.WebhookRequest, w *actions.WebhookResponse,
	m *music.Music, v *video.Video) {
	var tracks []music.Track

	song := r.SongParam()
	artist := r.ArtistParam()
	release := r.ReleaseParam()
	radio := r.RadioParam()
	popular := r.PopularParam()
	latest := r.LatestParam()
	any := r.AnyParam()
	query := ""

	if artist != "" {
		a := m.ArtistLike(artist)
		if a != nil {
			if radio != "" {
				// play [artist] radio
				tracks = m.ArtistRadio(*a)
			} else if popular != "" {
				// play popular songs by [artist]
				tracks = handler.artistPopular(m, a)
			} else if latest != "" {
				// play the new [artist]
				// play the latest [artist]
				// play the latest by/from [artist]
				releases := m.ArtistReleases(a)
				if len(releases) > 0 {
					r := releases[len(releases)-1]
					tracks = m.ReleaseTracks(r)
				}
			} else if song != "" {
				// play [song] by [artist]
				query = fmt.Sprintf(`+artist:"%s" +title:"%s*"`, artist, song)
			} else if release != "" {
				// play album [release] by [artist]

				// first try to find a fuzzy match
				releases := m.ArtistReleases(a)
				for _, r := range releases {
					if music.FuzzyName(r.Name) == music.FuzzyName(release) {
						tracks = m.ReleaseTracks(r)
						break
					}
				}
				// fallback to search
				if len(tracks) == 0 {
					query = fmt.Sprintf(`+artist:"%s" +release:"%s"`, artist, release)
				}
			} else {
				// play [artist] songs
				// play songs by [artist]
				query = fmt.Sprintf(`+artist:"%s" +type:"single"`, artist)
			}
		}
	} else if song != "" {
		// play song [song]
		query = fmt.Sprintf(`+title:"%s*"`, song)
	} else if release != "" {
		// play album [release]
		tracks = handler.releaseLike(m, release)
		if len(tracks) == 0 {
			query = fmt.Sprintf(`+release:"%s*"`, release)
		}
	} else if any != "" {
		// play [any]
		a := m.ArtistLike(any)
		if a != nil {
			tracks = handler.artistPopular(m, a)
		}
		if len(tracks) == 0 {
			tracks = handler.releaseLike(m, any)
		}
		if len(tracks) == 0 {
			query = fmt.Sprintf(`+title:"%s*"`, any)
		}
	}

	if query != "" {
		tracks = m.Search(query, handler.config.Assistant.TrackLimit)
		fmt.Printf("search for %s -> %d tracks\n", query, len(tracks))
	}

	if len(tracks) > 0 {
		addSimple(w, handler.config.Assistant.Play)
		for _, t := range tracks {
			fmt.Printf("%d. %s/%s/%s\n",
				t.TrackNum, t.Artist, t.Release, t.Title)
			name := execute(handler.config.Assistant.MediaObjectNameTemplate(), t)
			desc := execute(handler.config.Assistant.MediaObjectDescTemplate(), t)
			w.AddMedia(name, desc,
				m.TrackURL(&t).String(),
				m.TrackImage(t).String())
		}
	} else {
		addSimple(w, handler.config.Assistant.Error)
	}
}

func (handler *UserHandler) fulfillNew(r *actions.WebhookRequest, w *actions.WebhookResponse,
	m *music.Music, v *video.Video) {
	home := handler.homeView(m, v)

	speech := handler.config.Assistant.Recent.Speech
	text := handler.config.Assistant.Recent.Text

	for i, rel := range home.AddedReleases {
		if i == handler.config.Assistant.RecentLimit {
			break
		} else if i > 0 {
			speech += " and " // TODO
			text += ", "
		}
		speech += execute(handler.config.Assistant.Release.SpeechTemplate(), rel)
		text += execute(handler.config.Assistant.Release.TextTemplate(), rel)
	}
	w.AddSimple(speech, text)
}

func (handler *UserHandler) fulfillWelcome(r *actions.WebhookRequest, w *actions.WebhookResponse,
	m *music.Music, v *video.Video) {
	addSimple(w, handler.config.Assistant.Welcome)
	w.AddSuggestions(handler.config.Assistant.SuggestionNew)
}

func addSimple(w *actions.WebhookResponse, m config.AssistantResponse) {
	w.AddSimple(m.Speech, m.Text)
}

func addSimpleTemplate(w *actions.WebhookResponse, m config.AssistantResponse, vars interface{}) {
	w.AddSimple(execute(m.SpeechTemplate(), vars), execute(m.TextTemplate(), vars))
}

func execute(t *template.Template, vars interface{}) string {
	var buf bytes.Buffer
	_ = t.Execute(&buf, vars)
	return buf.String()
}

func (handler *UserHandler) authRequired(r *actions.WebhookRequest, w *actions.WebhookResponse, a *auth.Auth) {
	code := a.GenerateCode()
	vars := map[string]string{"Code": code.Value}
	addSimpleTemplate(w, handler.config.Assistant.Link, vars)
	w.AddSuggestions(handler.config.Assistant.SuggestionAuth)
	w.AddSessionParam("code", code.Value)
}

func (handler *UserHandler) authNext(r *actions.WebhookRequest, w *actions.WebhookResponse,
	a *auth.Auth) {
	ok := true
	value := r.SessionParam("code")
	if value == "" {
		ok = false
	}
	code := a.LinkedCode(value)
	if code == nil {
		ok = false
	}
	if !ok {
		handler.authRequired(r, w, a)
		return
	}

	w.AddUserParam(UserParamCookie, code.Cookie)
	addSimple(w, handler.config.Assistant.Linked)
	w.AddSuggestions("Talk to Takeout")
}

func (handler *UserHandler) authCheck(r *actions.WebhookRequest, w *actions.WebhookResponse,
	a *auth.Auth, cookie string) bool {
	var err error
	handler.user, err = a.UserAuthValue(cookie)
	return err == nil
}

func (handler *UserHandler) verificationRequired(r *actions.WebhookRequest, w *actions.WebhookResponse) {
	addSimple(w, handler.config.Assistant.Guest)
}
