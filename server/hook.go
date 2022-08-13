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

	"github.com/defsub/takeout/auth"
	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/actions"
	"github.com/defsub/takeout/lib/token"
	"github.com/defsub/takeout/music"
	"github.com/defsub/takeout/view"
)

const (
	IntentAuth = "TAKEOUT_AUTH"
	IntentPlay = "TAKEOUT_PLAY"
	IntentNew  = "TAKEOUT_NEW"

	UserParamCookie = "cookie"
)

func hookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := contextValue(r)
	w.Header().Set("Content-type", ApplicationJson)

	tokenString := r.Header.Get(actions.GoogleAssistantSignature)
	err := token.ValidateGoogleToken(ctx.Config(), tokenString, ctx.Config().Assistant.ProjectID)
	if err != nil {
		serverErr(w, err)
		return
	}

	hookRequest := actions.NewWebhookRequest(r)
	hookResponse := actions.NewWebhookResponse(hookRequest)
	// fmt.Printf("got intent=%s %+v\n", hookRequest.IntentName(), hookRequest)
	// fmt.Printf("got user %+v\n", *hookRequest.User)
	// if hookRequest.User != nil && hookRequest.User.Params != nil {
	// 	for k, v := range hookRequest.User.Params {
	// 		fmt.Printf("request user[%s]=%s\n", k, v)
	// 	}
	// }
	// if hookRequest.Session != nil && hookRequest.Session.Params != nil {
	// 	for k, v := range hookRequest.Session.Params {
	// 		fmt.Printf("request session[%s]=%s\n", k, v)
	// 	}
	// }

	if !hookRequest.Verified() {
		verificationRequired(ctx, hookRequest, hookResponse)
	} else {
		cookie := hookRequest.UserParam(UserParamCookie)
		if cookie == "" {
			if hookRequest.IntentName() == IntentAuth {
				// try to authenticate
				authNext(ctx, hookRequest, hookResponse)
			} else {
				authRequired(ctx, hookRequest, hookResponse)
			}
		} else {
			user, _ := authCheck(ctx, hookRequest, hookResponse, cookie)
			if user == nil {
				authRequired(ctx, hookRequest, hookResponse)
			} else {


				fulfillIntent(ctx, user, w, hookRequest, hookResponse)
			}
		}
	}

	// fmt.Printf("sending %+v\n", hookResponse)
	// if hookResponse.User != nil && hookResponse.User.Params != nil {
	// 	for k, v := range hookResponse.User.Params {
	// 		fmt.Printf("response user[%s]=%s\n", k, v)
	// 	}
	// }
	// if hookResponse.Session != nil && hookResponse.Session.Params != nil {
	// 	for k, v := range hookResponse.Session.Params {
	// 		fmt.Printf("response session[%s]=%s\n", k, v)
	// 	}
	// }
	hookResponse.Send(w)
}

func fulfillIntent(ctx Context, user *auth.User, resp http.ResponseWriter,
	r *actions.WebhookRequest, w *actions.WebhookResponse) {
	ctx, err := upgradeContext(ctx, user)
	if err != nil {
		serverErr(resp, err)
		return
	}

	switch r.IntentName() {
	case IntentPlay:
		fulfillPlay(ctx, r, w)
	case IntentNew:
		fulfillNew(ctx, r, w)
	default:
		fulfillWelcome(ctx, r, w)
	}
}

func releaseLike(m *music.Music, release string) []music.Track {
	var tracks []music.Track
	releases := m.ReleasesLike("%" + release + "%")
	if len(releases) > 0 {
		r := releases[0]
		tracks = m.ReleaseTracks(r)
	}
	return tracks
}

func fulfillPlay(ctx Context, r *actions.WebhookRequest, w *actions.WebhookResponse) {
	var tracks []music.Track

	m := ctx.Music()
	config := ctx.Config()

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
			v := view.ArtistView(ctx, *a)
			if radio != "" {
				// play [artist] radio
				tracks = v.Radio.Tracks()
			} else if popular != "" {
				// play popular songs by [artist]
				tracks = v.Popular.Tracks()
			} else if latest != "" {
				// play the new [artist]
				// play the latest [artist]
				// play the latest by/from [artist]
				releases := v.Releases
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
		tracks = releaseLike(m, release)
		if len(tracks) == 0 {
			query = fmt.Sprintf(`+release:"%s*"`, release)
		}
	} else if any != "" {
		// play [any]
		a := m.ArtistLike(any)
		if a != nil {
			v := view.ArtistView(ctx, *a)
			tracks = v.Popular.Tracks()
		}
		if len(tracks) == 0 {
			tracks = releaseLike(m, any)
		}
		if len(tracks) == 0 {
			query = fmt.Sprintf(`+title:"%s*"`, any)
		}
	}

	if query != "" {
		tracks = m.Search(query, config.Assistant.TrackLimit)
	}

	if len(tracks) > 0 {
		addSimple(w, config.Assistant.Play)
		for _, t := range tracks {
			name := config.Assistant.MediaObjectName.Execute(t)
			desc := config.Assistant.MediaObjectDesc.Execute(t)
			w.AddMedia(name, desc,
				m.TrackURL(&t).String(),
				m.TrackImage(t).String())
		}
	} else {
		addSimple(w, config.Assistant.Error)
	}
}

func fulfillNew(ctx Context, r *actions.WebhookRequest, w *actions.WebhookResponse) {
	home := view.HomeView(ctx)
	config := ctx.Config()

	speech := config.Assistant.Recent.Speech.Text
	text := config.Assistant.Recent.Text.Text

	for i, rel := range home.AddedReleases {
		if i == config.Assistant.RecentLimit {
			break
		} else if i > 0 {
			speech += " and " // TODO
			text += ", "
		}
		speech += config.Assistant.Release.Speech.Execute(rel)
		text += config.Assistant.Release.Speech.Execute(rel)
	}
	w.AddSimple(speech, text)
}

func fulfillWelcome(ctx Context, r *actions.WebhookRequest, w *actions.WebhookResponse) {
	config := ctx.Config()
	addSimple(w, config.Assistant.Welcome)
	w.AddSuggestions(config.Assistant.SuggestionNew)
}

func addSimple(w *actions.WebhookResponse, m config.AssistantResponse) {
	w.AddSimple(m.Speech.Text, m.Text.Text)
}

func addSimpleTemplate(w *actions.WebhookResponse, m config.AssistantResponse, vars interface{}) {
	w.AddSimple(m.Speech.Execute(vars), m.Text.Execute(vars))
}

func authRequired(ctx Context, r *actions.WebhookRequest, w *actions.WebhookResponse) {
	code := ctx.Auth().GenerateCode()
	config := ctx.Config()
	vars := map[string]string{"Code": code.Value}
	addSimpleTemplate(w, config.Assistant.Link, vars)
	w.AddSuggestions(config.Assistant.SuggestionAuth)
	w.AddSessionParam("code", code.Value)
}

func authNext(ctx Context, r *actions.WebhookRequest, w *actions.WebhookResponse) {
	ok := true
	value := r.SessionParam("code")
	if value == "" {
		ok = false
	}
	code := ctx.Auth().LinkedCode(value)
	if code == nil {
		ok = false
	}
	if !ok {
		authRequired(ctx, r, w)
		return
	}

	w.AddUserParam(UserParamCookie, code.Cookie)
	addSimple(w, ctx.Config().Assistant.Linked)
	w.AddSuggestions("Talk to Takeout")
}

func authCheck(ctx Context, r *actions.WebhookRequest, w *actions.WebhookResponse,
	cookie string) (*auth.User, error) {
	return ctx.Auth().UserAuthValue(cookie)
}

func verificationRequired(ctx Context, r *actions.WebhookRequest, w *actions.WebhookResponse) {
	addSimple(w, ctx.Config().Assistant.Guest)
}
