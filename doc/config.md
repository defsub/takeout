# Takeout Configuration

## Music Files

Takeout will index all the objects in the S3 bucket to find music
files that end with supported file extensions. Since the actual files
are not individually opened to inspect tags, a specific file structure
is needed. MusicBrainz Picard is _highly_ recommended to manage and tag
music files.

The bucket structure should be:

	bucket/prefix/artist/album/track

where album can be:

	album
	album (year)

and track can be:

	1-track.mp3
	01-track.mp3
	1-01-track.mp3

### Examples

A single disc release by the Raconteurs in 2019:

	The Raconteurs/Help Us Stranger (2019)/01-Bored and Razed.flac

A multi disc release by Tubeway Army in 2019, track 1 from disc 1 and 2:

	Tubeway Army/Replicas - The First Recordings (2019)/1-01-You Are in My Vision (early version).flac
	Tubeway Army/Replicas - The First Recordings (2019)/2-01-Replicas (early version 2).flac

## Music Metadata

MusicBrainz APIs are used to obtain further information about artists
and albums found in the bucket. This is used to try to correct artist
and album names, and also to determine which tracks are singles or EPs
and their first release date.

Last.fm APIs are used to obtain popular track information for each
artist.

Fanart.tv APIs are used to obtain artist images.

Cover Art Archive APIs are used to obtain links to album covers.

## Movie Files

Takeout will index all the objects in the S3 bucket to find movie
files that end with supported file extensions. The actual files
are not individually opened to inspect tags or headers.

The bucket structure should be:

	bucket/prefix/path/movie

where movie can be:

    Title (year).mkv
    Title (year) - HD.mkv

### Examples

    Joker (2019) - HD.mkv
	Get Out (2017).mp4

## Movie Metadata

The Movie Database (TMDb) APIs are used to obtain all movie, cast and crew information.

## Configuration ##

The configuration file _takeout.ini_ is used to store required configuration to
setup and configure Takeout. Yaml and other formats supported by
[Viper](https://github.com/spf13/viper) can be used as well.


### Auth

* Auth.DB.Driver - Database driver (default sqlite3)
* Auth.DB.LogMode - Log SQL commands (default false)
* Auth.DB.Source - Name of database (default auth.db)
* Auth.MaxAge - Time until unused cookie expires (default 24h)
* Auth.SecureCookies - Use secure cookies (default true)

Note: only sqlite3 is supported

### Bucket

All of these are required and only a few have defaults.

* Bucket.Endpoint - Fully qualified domain name of the S3 bucket endpoint
* Bucket.Region - Bucket region
* Bucket.AccessKeyID - Access key for S3 bucket
* Bucket.SecretAccessKey - Secret key for S3 bucket
* Bucket.BucketName - Name of the S3 bucket
* Bucket.ObjectPrefix - Prefix of media within the S3 bucket
* Bucket.UseSSL - Use SSL with endpoint (default true)
* Bucket.URLExpiration - Time at with pre-signed URLs expire (default 72h)

Note: client is the AWS S3 Go client. Wasabi, Backblaze and Minio have been tested.

### Music Database

* Music.DB.Driver - Database driver (default sqlite3)
* Music.DB.Source - Name of database (default music.db)
* Music.DB.LogMore - Log SQL commands (default false)

Note: only sqlite3 is supported

### Music Settings

* Music.ArtistRadioBreadth - How many similar artists to use (default 10)
* Music.ArtistRadioDepth - How many similar artists tracks to include (default 3)
* Music.DeepLimit - How many deep tracks (default 50)
* Music.PopularLimit - How many popular tracks (default 50)
* Music.RadioLimit - How many radio tracks (default 25)
* Music.RadioSearchLimit - How many tracks to search for radio (default 1000)
* Music.Recent - How recent is recent (default 8760h = 1 year)
* Music.RecentLimit - How many recent (default 50)
* Music.SearchLimit - How many items to return (default 100)
* Music.SimilarArtistsLimit - Many many similar artists (default 10)
* Music.SimilarReleases - Duration for similar (default 8760h = 1 year)
* Music.SimilarRelesesLimit - Many many similar releases (default 10)
* Music.SinglesLimit - How many singles (default 50)

### Server

* Server.Listen - Address and port to listen on (default 127.0.0.1:3000)

Note: use Nginx or other frontend with TLS (and Let's Encrypt)

### Search Index

* Search.BleveDir - Directory used to store index (default .)

### HTTP Client

* Client.UseCache - Enable or disable http caching (default false)
* Client.MaxAge - Age in seconds to use cached responses (default 30 days)
* Client.CacheDir - Directory to store cached responses (default .httpcache)

Note: this client is used for all API calls except Last.fm

## Examples

### Required fields in ini format

	[bucket]
	endpoint = "s3.us-west-1.wasabisys.com"
	region = "us-west-1"
	accessKeyID = "<insert access key here>"
	secretAccessKey = "<insert secret key here>"
	bucketName = "MyBucket"
	objectPrefix = "MyMedia"
