// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// Takeout is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package spiff

import (
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
)

const (
	ContentType = "application/xspf+xml"
)

type StringTag struct {
	Value string `xml:",chardata"`
}

type IntTag struct {
	Value uint `xml:",chardata"`
}

type TrackTag struct {
	XMLName  xml.Name  `xml:"track"`
	Creator  StringTag `xml:"creator"`
	Album    StringTag `xml:"album"`
	Title    StringTag `xml:"title"`
	TrackNum IntTag    `xml:"trackNum"`
	Location StringTag `xml:"location"`
}

type Encoder struct {
	xmlEncoder *xml.Encoder
	writer io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{writer: w, xmlEncoder: xml.NewEncoder(w)}
	return e
}

func (e *Encoder) Header(title string) {
	fmt.Fprint(e.writer, xml.Header)
	fmt.Fprintf(e.writer, "<playlist version=\"1\" xmlns=\"http://xspf.org/ns/0/\">")
	fmt.Fprintf(e.writer, "<title>%s</title>", title)
	fmt.Fprintf(e.writer, "<trackList>")
}

func (e *Encoder) Encode(track interface{}) (err error) {
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
					trackTag.TrackNum = IntTag{uint(valueField.Uint())}
				case "location":
					trackTag.Location = StringTag{valueField.String()}
				}
			}
		}
	}
	err = e.xmlEncoder.Encode(trackTag)
	return err
}

func (e *Encoder) Footer() {
	fmt.Fprintf(e.writer, "</trackList>")
	fmt.Fprintf(e.writer, "</playlist>\n")
}
