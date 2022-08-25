# Takeout

Takeout is a copyleft media service that indexes organized media files in S3
buckets using MusicBrainz, Last.fm, Fanart.tv, and The Movie Database to make
media available for streaming using a [mobile app](https://github.com/defsub/takeout_app),
TV app, web interface and VLC. Media is browsed using the Takeout server and
streamed directly from S3 using pre-signed time-based URLs.

Takeout is primarily a music player that also supports radio, movies and
podcasts. It's not intended to be a replacement for more feature-rich systems
like [Plex](https://plex.tv), [Jellyfin](https://jellyfin.org) and
[Kodi](https://kodi.tv/). Instead, Takeout takes inspiration from these systems
and attempts to be a small yet capable system designed around media being
stored in the cloud. You can take your personal media collection with you, on
your own terms, create your own personal streaming service, and enjoy your
media with free and open source software.

## Features

* Music metadata from [MusicBrainz](https://musicbrainz.org/) and [Last.fm](https://last.fm/)
* Album covers from the [Cover Art Archive](https://coverartarchive.org/)
* Artist artwork from [Fanart.tv](https://fanart.tv/)
* Powerful search and playlists. See [search.md](doc/search.md)
* Movie metadata and artwork from [The Movie Database (TMDb)](https://www.themoviedb.org/)
* Podcasts with series and episode metadata using [RSS 2.0](https://www.rssboard.org/rss-specification)
* Internet radio stations (pls)
* Support for [Google Assistant](https://assistant.google.com/) until June
  13, 2023. See [actions.md](doc/actions.md) for more details.
* Media streaming directly from S3 using pre-signed time-based URLs
* User-based access control using cookies, tokens and
  [scrypt](https://pkg.go.dev/golang.org/x/crypto/scrypt?tab=doc)
* Server-based playlist API (using [jsonpatch](http://jsonpatch.com/))
* Web and JSON views
* Web playback using HTML5 audio - Chrome, Safari & Firefox tested on desktop & mobile
* [Flutter app](https://github.com/defsub/takeout_app) available for Android (and iOS)
* [XSPF ("spiff")](https://xspf.org/) and JSPF playlists
* Written in [Go](https://go.dev/), with [SQLite3](https://sqlite.org/index.html) and [Bleve](https://blevesearch.com/)
* Supports [caching](https://github.com/gregjones/httpcache) of metadata API
  data for faster (re)syncing
* REST APIs are available to build custom interfaces
* Free and open source with AGPLv3 license

The [privacy policy](doc/privacy.md), [setup documentation](doc/setup.md), and
more details on how to manage media in the [S3 bucket](doc/bucket.md), can be
found in the doc directory.

Please see how you can [contribute](doc/contribute.md) to Takeout and related
projects and services.
