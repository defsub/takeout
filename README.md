# Takeout

Takeout is media service that indexes organized media files in S3 buckets using
MusicBrainz and Last.fm to make media available for streaming using a web
interface and VLC. Media is browsed using the Takeout server and streamed
directly from S3 using pre-signed time-based URLs.

<p align="center">
<img src="https://github.com/defsub/defsub.github.io/blob/master/takeout/screens/2020-09-07/Screenshot_20200907-082736.png" width="200">
<img src="https://github.com/defsub/defsub.github.io/blob/master/takeout/screens/2020-09-07/Screenshot_20200907-082827.png" width="200">
<img src="https://github.com/defsub/defsub.github.io/blob/master/takeout/screens/2020-09-07/Screenshot_20200907-083006.png" width="200">
<img src="https://github.com/defsub/defsub.github.io/blob/master/takeout/screens/2020-09-07/Screenshot_20200907-083707.png" width="200">
</p>

## Features

* Powerful search and playlists using [Bleve](https://blevesearch.com/). See [SEARCH.md](SEARCH.md).
* Metadata from [MusicBrainz](https://musicbrainz.org) and [Last.fm](https://last.fm/)
* Album covers from the [Cover Art Archive](https://coverartarchive.org/)
* Media streaming directly from S3 using pre-signed time-based URLs
* Recently added and released
* Similar artists and releases
* Popular tracks
* Artist singles
* User-based access control using cookies and [scrypt](https://pkg.go.dev/golang.org/x/crypto/scrypt?tab=doc)
* Server-based playlist (using [jsonpatch](http://jsonpatch.com/))
* Web and json views
* Web playback using HTML5 audio - Chrome, Safari & Firefox tested on desktop & mobile
* [XSPF ("spiff")](https://xspf.org/) and JSPF playlists
* Written in [Go](https://golang.org/), using [SQLite3](https://sqlite.org/index.html) and [Bleve](https://blevesearch.com/)

## Quick Start

* Tag your media with [Picard](https://picard.musicbrainz.org/) (highly recommeded)
* Put your organized media in a S3 bucket ([Wasabi](https://wasabi.com/), [Minio](https://min.io/), [AWS](https://aws.amazon.com/))
* Optionally setup a virtual server ([Linode](https://www.linode.com/), [EC2](https://aws.amazon.com/), [Compute Engine](https://cloud.google.com/compute))
* Optionally setup a TLS front-end ([Nginx](http://nginx.org/), [Let's Encrypt](https://letsencrypt.org/))
* Install [Go](https://golang.org/)
* Build [Takeout](https://github.com/defsub/takeout/)
  * git clone https://github.com/defsub/takeout.git
  * cd takeout/cmd/takeout
  * go build
  * ./takeout
* Create your [takeout.ini](CONFIG.md)
* Sync your data
  * ./takeout sync
* Databases may need ~100MB and Bleve index ~1GB, depending on the media library
* Run the server
  * ./takeout serve
