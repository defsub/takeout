# Takeout

Takeout is media service that indexes organized media files in S3 buckets using
MusicBrainz and Last.fm to make media available for streaming using a web
interface and vlc.

Features:

* Powerful search and playlists using [Bleve](https://blevesearch.com/). See SEARCH.md.
* Metadata from [MusicBrainz](https://musicbrainz.org) and [Last.fm](https://last.fm/)
* Album covers from the [Cover Art Archive](https://coverartarchive.org/)
* Recently released
* Recently added
* Similar releases
* Similar artists
* User-based access control using [scrypt](https://pkg.go.dev/golang.org/x/crypto/scrypt?tab=doc)
* Server-based playlist (using [jsonpatch](http://jsonpatch.com/))
* Web and json views
* Web playback using HTML5 audio - Chrome, Safari, Firefox supported on desktop & mobile
* [Spiff](https://xspf.org/) and JSPF playlists
* Written in Go, using [SQLite3](https://sqlite.org/index.html) and [Bleve](https://blevesearch.com/)
