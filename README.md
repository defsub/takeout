# Takeout

Takeout is a copyleft media service that indexes organized media files in S3
buckets using MusicBrainz, Last.fm, Fanart.tv, and The Movie Database to make
media available for streaming using a [Flutter app](https://github.com/defsub/takeout_app),
TV app, web interface and VLC. Media is browsed using the Takeout server and
streamed directly from S3 using pre-signed time-based URLs. REST APIs are available
to build custom interfaces.

## Features

* Powerful search and playlists using [Bleve](https://blevesearch.com/). See [search.md](doc/search.md)
* Music metadata from [MusicBrainz](https://musicbrainz.org/) and [Last.fm](https://last.fm/)
* Album covers from the [Cover Art Archive](https://coverartarchive.org/)
* Artist artwork from [Fanart.tv](https://fanart.tv/)
* Movie metadata and artwork from [The Movie Database (TMDb)](https://www.themoviedb.org/)
* Podcasts with series and episode metadata using [RSS 2.0](https://www.rssboard.org/rss-specification)
* Support for [Google Assistant](https://assistant.google.com/). See [actions.md](doc/actions.md) for more details.
* Media streaming directly from S3 using pre-signed time-based URLs
* Recently added and released
* Similar artists and releases
* Popular tracks
* Artist singles
* Radio stations based on genres, mixes, singles, and anything searchable
* User-based access control using cookies and [scrypt](https://pkg.go.dev/golang.org/x/crypto/scrypt?tab=doc)
* Server-based playlist (using [jsonpatch](http://jsonpatch.com/))
* Web and json views
* Web playback using HTML5 audio - Chrome, Safari & Firefox tested on desktop & mobile
* [Flutter app](https://github.com/defsub/takeout_app) available for Android and iOS
* [XSPF ("spiff")](https://xspf.org/) and JSPF playlists
* Written in [Go](https://golang.org/), using [SQLite3](https://sqlite.org/index.html) and [Bleve](https://blevesearch.com/)
* Supports [caching](https://github.com/gregjones/httpcache) of API data for faster syncing

## Quick Start

* Tag your media with [Picard](https://picard.musicbrainz.org/) (highly recommeded)
* Put your organized media in an S3 bucket ([Wasabi](https://wasabi.com/),
  [Minio](https://min.io/), [AWS](https://aws.amazon.com/))
* Optionally setup a virtual server ([Linode](https://www.linode.com/),
  [EC2](https://aws.amazon.com/), [Compute Engine](https://cloud.google.com/compute))
* Optionally setup a TLS front-end ([Nginx](http://nginx.org/), [Let's Encrypt](https://letsencrypt.org/))
* Install [Go](https://golang.org/)
* Build [Takeout](https://github.com/defsub/takeout/)
  * git clone https://github.com/defsub/takeout.git
  * cd takeout/cmd/takeout
  * go build
  * go install
* Create your [takeout.ini](docs/config.md)
* Sync your data
  * ./takeout sync
* Databases may need ~100MB and Bleve index ~100MB, depending on the media library
* Run the server
  * ./takeout serve
