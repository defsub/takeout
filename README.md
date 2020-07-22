# Takeout

Takeout is media service that indexes organized media files in S3 buckets using
MusicBrainz and Last.fm to make media available for streaming using a web
interface and VLC.

## Features

* Powerful search and playlists using [Bleve](https://blevesearch.com/). See [SEARCH.md](README.md).
* Metadata from [MusicBrainz](https://musicbrainz.org) and [Last.fm](https://last.fm/)
* Album covers from the [Cover Art Archive](https://coverartarchive.org/)
* Recently released
* Recently added
* Similar releases
* Similar artists
* Popular tracks
* Artist singles
* User-based access control using cookies and [scrypt](https://pkg.go.dev/golang.org/x/crypto/scrypt?tab=doc)
* Server-based playlist (using [jsonpatch](http://jsonpatch.com/))
* Web and json views
* Web playback using HTML5 audio - Chrome, Safari & Firefox tested on desktop & mobile
* [XSPF ("spiff")](https://xspf.org/) and JSPF playlists
* Written in Go, using [SQLite3](https://sqlite.org/index.html) and [Bleve](https://blevesearch.com/)

## Quick Start

* Tag your media with [Picard](https://picard.musicbrainz.org/) (highly recommeded)
* Put your organized media in a S3 bucket (Wasabi, Minio, AWS)
* Optionally setup a virtual server (Linode, EC2, Compute Engine)
* Optionally setup a TLS front-end (Nginx, [Let's Encrypt](https://letsencrypt.org/))
* Install Go
* Install Takeout
* Create your [takeout.ini](CONFIG.md)
* Sync your data - go run cmd/sync/sync.go
* Run the server - go run cmd/serve/serve.go
