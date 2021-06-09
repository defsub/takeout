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

package xspf

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
)

const (
	XMLContentType  = "application/xspf+xml"
	JsonContentType = "application/xspf+json"
)

type StringTag struct {
	Value string `xml:",chardata" json:"-"`
}

func (tag *StringTag) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", tag.Value)), nil
}

type IntTag struct {
	Value int `xml:",chardata"`
}

func (tag *IntTag) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", tag.Value)), nil
}

type TrackTag struct {
	XMLName    xml.Name  `xml:"track" json:"-"`
	Creator    StringTag `xml:"creator" json:"creator"`
	Album      StringTag `xml:"album" json:"album"`
	Title      StringTag `xml:"title" json:"title"`
	TrackNum   IntTag    `xml:"trackNum" json:"trackNum"`
	Location   StringTag `xml:"location" json:"location"`
	Image      StringTag `xml:"image" json:"image"`
	Identifier StringTag `xml:"identifier" json:"identifier"`
}

type SpiffEncoder interface {
	Header(title string)
	Encode(e interface{}) error
	Footer()
}

type Encoder interface {
	Encode(e interface{}) error
}

type xmlEncoder struct {
	writer  io.Writer
	encoder *xml.Encoder
}

func (e xmlEncoder) Header(title string) {
	fmt.Fprint(e.writer, xml.Header)
	fmt.Fprintf(e.writer, "<playlist version=\"1\" xmlns=\"http://xspf.org/ns/0/\">")
	fmt.Fprintf(e.writer, "<title>")
	xml.Escape(e.writer, []byte(title))
	fmt.Fprintf(e.writer, "</title>")
	fmt.Fprintf(e.writer, "<trackList>")
}

func (e xmlEncoder) Encode(v interface{}) error {
	return encode(e.encoder, v)
}

func (e xmlEncoder) Footer() {
	fmt.Fprintf(e.writer, "</trackList>")
	fmt.Fprintf(e.writer, "</playlist>\n")
}

type jsonEncoder struct {
	writer  io.Writer
	encoder *json.Encoder
	count   *int
}

func (e jsonEncoder) Header(title string) {
	fmt.Fprintf(e.writer, "{\"playlist\":{\"title\":\"%s\",\"track\":[", title)
}

func (e jsonEncoder) Encode(v interface{}) error {
	if *e.count > 0 {
		fmt.Fprintf(e.writer, ",")
	}
	*e.count++
	return encode(e.encoder, v)
}

func (e jsonEncoder) Footer() {
	fmt.Fprintf(e.writer, "]}}")
}

func NewXMLEncoder(w io.Writer) SpiffEncoder {
	e := xmlEncoder{w, xml.NewEncoder(w)}
	return e
}

func NewJsonEncoder(w io.Writer) SpiffEncoder {
	count := 0
	e := jsonEncoder{w, json.NewEncoder(w), &count}
	return e
}

func encode(e Encoder, track interface{}) error {
	trackTag := &TrackTag{}

	t := reflect.TypeOf(track)
	v := reflect.ValueOf(track)

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			typeField := t.Field(i)
			valueField := v.Field(i)
			if tag, ok := typeField.Tag.Lookup("spiff"); ok {
				switch tag {
				case "creator":
					trackTag.Creator = StringTag{valueField.String()}
				case "album":
					trackTag.Album = StringTag{valueField.String()}
				case "title":
					trackTag.Title = StringTag{valueField.String()}
				case "tracknum":
					trackTag.TrackNum = IntTag{int(valueField.Int())}
				case "location":
					trackTag.Location = StringTag{valueField.Index(0).String()}
				case "image":
					trackTag.Image = StringTag{valueField.String()}
				// case "identifier":
				// 	trackTag.Identifier = StringTag{valueField.Index(0).String()}
				}
			}
		}
	}
	return e.Encode(trackTag)
}
