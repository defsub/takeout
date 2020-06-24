package xspf

import (
	"testing"
	"os"
)

type Track struct {
	Artist       string `spiff:"creator"`
	Release      string `spiff:"album"`
	TrackNum     uint   `spiff:"tracknum"`
	Title        string `spiff:"title"`
	Location     string `spiff:"location"`
}

func TestXml(t *testing.T) {
	e := NewXMLEncoder(os.Stdout)
	e.Header("test title")
	var track Track
	track.Artist = "My Artist"
	track.Release = "My Release"
	track.TrackNum = 1
	track.Title = "My Title"
	track.Location = "https://a/b/c"
	e.Encode(track)
	track.TrackNum = 2
	e.Encode(track)
	track.TrackNum = 3
	e.Encode(track)
	e.Footer()
}

func TestJson(t *testing.T) {
	e := NewJsonEncoder(os.Stdout)
	e.Header("test title")
	var track Track
	track.Artist = "My Artist"
	track.Release = "My Release"
	track.TrackNum = 1
	track.Title = "My Title"
	track.Location = "https://a/b/c"
	e.Encode(track)
	track.TrackNum = 2
	e.Encode(track)
	track.TrackNum = 3
	e.Encode(track)
	e.Footer()
}
