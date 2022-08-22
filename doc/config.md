# Takeout Configuration

The Takeout server has a server configuration file and media specific
configuration files, one per media sub-directory. The configuration file
formats are parsed using [Viper](https://github.com/spf13/viper) so formats
like JSON, TOML and YAML can be used. YAML is used in most examples unless
stated otherwise.

## Server Configuration

Most of these are defaults and are here as an example only.

```
server:
  listen: :3000

auth:
  maxAge: 24h
  secureCookies: true
  DB:
    Driver: sqlite3
	Source: auth.db
	Logger: default

progress:
  DB:
    Driver: sqlite3
	Source: progress.db
	Logger: default

activity:
  DB:
    Driver: sqlite3
	Source: activity.db
	Logger: default
```

## Database Configuration

* Driver - Use "sqlite3", it's super fast and builtin
* Source - Name of database file
* Logger - Use debug, default or discard

You can try mysql if you'd like. It has worked in the past but it's not tested
much.

* Driver: "mysql"
* Source: "takeout:takeout@tcp(127.0.0.1:3306)/mydb"

## Bucket Configuration

See the included [config.yaml](config.yaml) example configuration. This
configuration file must be included in each media sub-directory. The file
specifies the bucket details along with movie and video metadata configuration.

## Music Configuration

* ArtistFile - Used to help find artists MBID
* ArtistRadioBreadth - How many similar artists to use (default 10)
* ArtistRadioDepth - How many similar artists tracks to include (default 3)
* DeepLimit - How many deep tracks (default 50)
* PopularLimit - How many popular tracks (default 50)
* RadioLimit - How many radio tracks (default 25)
* RadioSearchLimit - How many tracks to search for radio (default 1000)
* RadioStreams - Define Internet radio streams (see below)
* Recent - How recent is recent (default 8760h = 1 year)
* RecentLimit - How many recent (default 50)
* SearchLimit - How many items to return (default 100)
* SimilarArtistsLimit - Many many similar artists (default 10)
* SimilarReleases - Duration for similar (default 8760h = 1 year)
* SimilarRelesesLimit - Many many similar releases (default 10)
* SinglesLimit - How many singles (default 50)
* SyncInterval - How often to automtically resync media from buckets (1h)
* PopularSyncInterval - How oftern to resync popular tracks from Last.fm (24h)
* SimilarSyncInterval - How oftern to resync similar artists from Last.fm (24h)

## Artists File

When you run into trouble matching artist names to MusicBrainz artists, this
file can save the day. You'll need this when there are artists with the same
name and you want a specific one, artists with ambiguous names or for some
reason or another Takeout just can't figure it out. Create an artist file like
this:

```
{
    "Belly" : "c118bc97-11a7-41dc-a55e-48c3bcf22ac2",
    "POW!" : "e00ac97d-3ed1-4f3e-86e1-1b15dd3ad6ad",
    "Phoenix" : "8d455809-96b3-4bb6-8829-ea4beb580d35",
    "Isao Tomita" : "9119e57f-1086-48b2-8a93-57feacb7f6d9",
    "Organisation _ Kraftwerk" : "7b6de1a2-d119-48a6-a17c-5472df12beeb",
    "Kid Rock & The Twisted Brown Trucker" : "ad0ecd8b-805e-406e-82cb-5b00c3a3a29e",
    "R.E.M_" : "ea4dfa26-f633-4da6-a52a-f49ea4897b58",
    "Gary Numan & Ade Fenton" : "6cb79cb2-9087-44d4-828b-5c6fdff2c957",
    "Sonic Youth, I.C.P. & The Ex" : "5cbef01b-cc35-4f52-af7b-d0df0c4f61b9",
    "The Vines" : "4e045c96-538b-46ed-8ea8-7cae20b56574",
    "CHVRCHΞS" : "6a93afbb-257f-4166-b389-9f2a1e5c5df8",
}
```

## Internet Radio

A few examples are included in the builtin configuration. Add your own as follows:

```
music:
  RadioStreams:
	- Creator:  "SomaFM"
	  Title:    "Groove Salad"
	  Image:    "https://somafm.com/img3/groovesalad-400.jpg"
	  Location: "https://somafm.com/groovesalad130.pls"
	- Creator:  "SomaFM"
	  Title:    "Indie Pop Rocks"
	  Image:    "https://somafm.com/img3/indiepop-400.jpg"
	  Location: "https://somafm.com/indiepop130.pls"
```

## Video Configuration

* Recent - How recent is recent (default 8760h = 1 year)
* RecentLimit - How many recent (default 50)
* SearchLimit - How many items to return (default 100)
* SyncInterval - How often to automtically resync media from buckets (1h)

## Video Recommendations

Movies can be recommended based on date patterns and search queries. It's a
convenient way to watch movies without having to think too hard while you sit
on the sofa. There many recommendations included in the builtin
configuration. Add your own as shown below. Note that the Layout field
corresponds to Go's clever yet funky [date parsing layouts](https://pkg.go.dev/time#pkg-constants).

```
video:
  Recommend:
    When:
	  - Name:   "Friday the 13th Movies"
	    Match:  "Fri 13"
	    Layout: "Mon 02"
		Query:  +character:voorhees
      - Name:   "Star Wars Day"
	    Match:  "May 04"
	    Layout: "Jan 02"
		Query:  +title:"star wars"
	  - Name:   "Christmas Movies"
	    Match:  "Dec"
		Layout: "Jan"
		Query:  +keyword:christmas
```

## Podcast Configuration

* RecentLimit - How many recent (default 25)
* SyncInterval - How often to automtically resync from sources (1h)
* Series - List of Podcasts sources. Several are builtin.

Below is an example showning how to add your own Podcast series.

```
podcast:
  Series:
    - "https://feeds.twit.tv/twit.xml"
    - "https://www.pbs.org/newshour/feeds/rss/podcasts/show"
	- "http://feeds.feedburner.com/TEDTalks_audio"
```

## HTTP Client Configuration

* UseCache - Enable or disable http caching (default false)
* MaxAge - Age in seconds to use cached responses (default 30 days)
* CacheDir - Directory to store cached responses (default .httpcache)

## API Keys

* FanArt.ProjectKey - Takeout uses 93ede276ba6208318031727060b697c8
* LastFM.Key - Please obtain your own at [last.fm](https://www.last.fm/api)
* LastFM.Secret - Please obtain your own at [last.fm](https://www.last.fm/api)
* TMDB.Key - Takeout uses 903a776b0638da68e9ade38ff538e1d3
