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

import ()

type Handler struct {
	Name string `json:"name"`
}

type Param struct {
	Original string `json:"original"`
	Resolved string `json:"resovled"`
}

type Params struct {
	Artist Param `json:"artist"`
	Song   Param `json:"song"`
}

type Intent struct {
	Name   string `json:"name"`
	Params Params `json:"params"`
	Query  string `json:"query"`
}

type Scene struct {
	Name string `json:"name"`
}

type Session struct {
	ID           string            `json:"id"`
	Params       map[string]string `json:"params"`
	LanguageCode string            `json:"languageCode"`
}

type User struct {
	Locale               string `json:"locale"`
	AccountLinkingStatus string `json:"accountLinkingStatus"`
	VerificationStatus   string `json:"verificationStatus"`
	LastSeenTime         string `json:"lastSeenTime"`
}

type TimeZone struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// SPEECH, RICH_RESPONSE, LONG_FORM_AUDIO
type Device struct {
	Capabilities []string `json:"capabilities"`
	TimeZone     TimeZone `json:"timeZone"`
}

type MediaContext struct {
	Index    int    `json:"index"`
	Progress string `json:"progress"`
}

type Context struct {
	Media MediaContext `json:"media"`
}

type Image struct {
	Alt    string `json:"alt"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type MediaImage struct {
	Large Image `json:"large"`
	Icon  Image `json:"icon"`
}

type MediaObject struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	URL         string     `json:"url"`
	Image       MediaImage `json:"image"`
}

type Media struct {
	MediaObjects          []MediaObject `json:"mediaObjects"`
	MediaType             string        `json:"mediaType"`
	OptionalMediaControls []string      `json:"optionalMediaControls"`
	RepeatMode            string        `json:"repeatMode"`
}

type Card struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Text     string `json:"text"`
	Image    Image  `json:"image"`
}

type Content struct {
	Media Media `json:"media"` // media
	Card  Card  `json:"card"`  // card with text and image
	Image Image `json:"image"` // image only card
}

type SimpleResponse struct {
	Speech string `json:"speech"`
	Text   string `json:"text"`
}

type Prompt struct {
	Override    bool           `json:"override"`
	Content     Content        `json:"content"`
	FirstSimple SimpleResponse `json:"firstSimple"`
	LastSimple  SimpleResponse `json:"lastSimple"`
}

type WebhookRequest struct {
	Handler Handler `json:"handler"`
	Indent  Intent  `json:"intent"`
	Scene   Scene   `json:"scene"`
	Session Session `json:"session"`
	User    User    `json:"user"`
	// home?
	Device  Device  `json:"device"`
	Context Context `json:"Context"`
}

type WebhookResponse struct {
	Session Session `json:"session"`
	Prompt  Prompt  `json:"prompt"`
}
