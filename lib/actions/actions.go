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

package actions

// This package works with Google Assistant Actions SDK and provides bindings
// to support json encoding and decoding for some data types. Enough
// functionality to get a webhook and fulfillment working with Takeout is
// supported.
//
// https://developers.google.com/assistant/conversational/overview
// https://developers.google.com/assistant/conversational/prompts

const (
	MediaTypeAudio = "AUDIO"

	MediaControlPaused  = "PAUSED"
	MediaControlStopped = "STOPPED"

	CapabilitySpeech            = "SPEECH"
	CapabilityRichResponse      = "RICH_RESPONSE"
	CapabilityLongFormAudio     = "LONG_FORM_AUDIO"
	CapabilityInteractiveCanvas = "LONG_FORM_AUDIO"
	CapabilityWebLink           = "WEB_LINK"
	CapabilityHomeStorage       = "HOME_STORAGE"
)

type Handler struct {
	Name string `json:"name,omitempty"`
}

type Param struct {
	Original string `json:"original,omitempty"`
	Resolved string `json:"resolved,omitempty"`
}

type Params struct {
	Artist *Param `json:"artist"`
	Song   *Param `json:"song"`
}

type Intent struct {
	Name   string  `json:"name,omitempty"`
	Params *Params `json:"params"`
	Query  string  `json:"query,omitempty"`
}

type Scene struct {
	Name string `json:"name,omitempty"`
}

type Session struct {
	ID           string            `json:"id,omitempty"`
	Params       map[string]string `json:"params"`
	LanguageCode string            `json:"languageCode,omitempty"`
}

type User struct {
	Params               map[string]string `json:"params"`
	Locale               string            `json:"locale,omitempty"`
	AccountLinkingStatus string            `json:"accountLinkingStatus,omitempty"`
	VerificationStatus   string            `json:"verificationStatus,omitempty"`
	LastSeenTime         string            `json:"lastSeenTime,omitempty"`
}

type TimeZone struct {
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
}

type Device struct {
	Capabilities []string  `json:"capabilities,omitempty"`
	TimeZone     *TimeZone `json:"timeZone"`
}

type MediaContext struct {
	Index    int    `json:"index,omitempty"`
	Progress string `json:"progress,omitempty"`
}

type Context struct {
	Media *MediaContext `json:"media"`
}

type Image struct {
	Alt    string `json:"alt,omitempty"`
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type MediaImage struct {
	Large *Image `json:"large"`
	Icon  *Image `json:"icon"`
}

type MediaObject struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	URL         string      `json:"url,omitempty"`
	Image       *MediaImage `json:"image"`
}

type Media struct {
	MediaObjects          []*MediaObject `json:"mediaObjects,omitempty"`
	MediaType             string         `json:"mediaType,omitempty"`
	OptionalMediaControls []string       `json:"optionalMediaControls,omitempty"`
	RepeatMode            string         `json:"repeatMode,omitempty"`
}

type Card struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Text     string `json:"text,omitempty"`
	Image    *Image `json:"image,omitempty"`
}

type Content struct {
	Card  *Card  `json:"card"`  // basic card
	Image *Image `json:"image"` // image card
	Media *Media `json:"media"` // media
}

type SimpleResponse struct {
	Speech string `json:"speech,omitempty"`
	Text   string `json:"text,omitempty"`
}

type Suggestion struct {
	Title string `json:"title,omitempty"`
}

type Prompt struct {
	Override    bool            `json:"override"`
	Content     *Content        `json:"content"`
	FirstSimple *SimpleResponse `json:"firstSimple"`
	LastSimple  *SimpleResponse `json:"lastSimple"`
	Suggestions []*Suggestion   `json:"suggestions,omitempty"`
}

type Home struct {
	Params map[string]string `json:"params"`
}

type WebhookRequest struct {
	Handler *Handler `json:"handler"`
	Intent  *Intent  `json:"intent"`
	Scene   *Scene   `json:"scene"`
	Session *Session `json:"session"`
	User    *User    `json:"user"`
	Home    *Home    `json:"home"`
	Device  *Device  `json:"device"`
	Context *Context `json:"Context"`
}

type WebhookResponse struct {
	Session *Session `json:"session"`
	Prompt  *Prompt  `json:"prompt"`
	Home    *Home    `json:"home"`
}

func (r WebhookRequest) IntentName() string {
	if r.Intent == nil {
		return ""
	}
	return r.Intent.Name
}

func (r WebhookRequest) ArtistParam() string {
	if r.Intent == nil || r.Intent.Params == nil || r.Intent.Params.Artist == nil {
		return ""
	}
	return r.Intent.Params.Artist.Resolved
}

func (r WebhookRequest) SongParam() string {
	if r.Intent == nil || r.Intent.Params == nil || r.Intent.Params.Song == nil {
		return ""
	}
	return r.Intent.Params.Song.Resolved
}

func (r WebhookRequest) SupportsRichResponse() bool {
	if r.Device == nil {
		return false
	}
	for _, s := range r.Device.Capabilities {
		if s == CapabilityRichResponse {
			return true
		}
	}
	return false
}

func (r WebhookRequest) SupportsMedia() bool {
	if r.Device == nil {
		return false
	}
	hasRichResponse := false
	hasLongFormAudio := false
	for _, s := range r.Device.Capabilities {
		switch s {
		case CapabilityRichResponse:
			hasRichResponse = true
		case CapabilityLongFormAudio:
			hasLongFormAudio = true
		}
	}
	return hasRichResponse && hasLongFormAudio
}

func (r *WebhookResponse) AddSession(id string) {
	if r.Session == nil {
		r.Session = &Session{}
	}
	r.Session.ID = id
}


func (r *WebhookResponse) AddSimple(speech, text string) {
	if r.Prompt == nil {
		r.Prompt = &Prompt{}
	}
	r.Prompt.FirstSimple = &SimpleResponse{
		Speech: speech,
		Text:   text,
	}
}

func (r *WebhookResponse) AddSuggestions(suggestions ...string) {
	if r.Prompt == nil {
		r.Prompt = &Prompt{}
	}
	for _, s := range suggestions {
		r.Prompt.Suggestions = append(r.Prompt.Suggestions, &Suggestion{Title: s})
	}
}

func (r *WebhookResponse) AddMedia(name, desc, url, image string) {
	if r.Prompt == nil {
		r.Prompt = &Prompt{}
	}
	if r.Prompt.Content == nil {
		r.Prompt.Content = &Content{}
	}
	if r.Prompt.Content.Media == nil {
		r.Prompt.Content.Media = &Media{
			MediaType: MediaTypeAudio,
			MediaObjects: []*MediaObject{},
		}
	}
	object := &MediaObject{
		Name: name,
		Description: desc,
		URL: url,
		Image: &MediaImage{
			Large: &Image{
				URL: image,
 				Alt: "album cover",
			},
		},
	}
	r.Prompt.Content.Media.MediaObjects = append(r.Prompt.Content.Media.MediaObjects, object)
}
