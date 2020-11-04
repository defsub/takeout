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

## Configuration ##

The configuration file _takeout.ini_ is used to store required configuration to
setup and configure Takeout.

### Auth

* Auth.MaxAge - Time until unused cookie expires (default 24h)
* Auth.DB.Driver - Database driver (default sqlite3)
* Auth.DB.Source - Name of database (default auth.db)
* Auth.DB.LogMode - Log SQL commands (default false)

Note: only sqlite3 is supported

### Music Bucket

All of these are required and only a few have defaults.

* Music.Bucket.Endpoint - Fully qualified domain name of the S3 bucket endpoint
* Music.Bucket.Region - Bucket region
* Music.Bucket.AccessKeyID - Access key for S3 bucket
* Music.Bucket.SecretAccessKey - Secret key for S3 bucket
* Music.Bucket.BucketName - Name of the S3 bucket
* Music.Bucket.ObjectPrefix - Prefix of media within the S3 bucket
* Music.Bucket.UseSSL - Use SSL with endpoint (default true)
* Music.Bucket.URLExpiration - Time at with pre-signed URLs expire (default 72h)

Note: client is AWS S3 Go client, only Wasabi has been tested

### Music Database

* Music.DB.Driver - Database driver (default sqlite3)
* Music.DB.Source - Name of database (default music.db)
* Music.DB.LogMore - Log SQL commands (default false)

Note: only sqlite3 is supported

### Music Settings

* Music.Recent - How recent is recent (default 8760h = 1 year)
* Music.RecentLimit - How many recent (default 50)
* Music.SearchLimit - How many items to return (default 50)
* Music.PopularLimit - How many popular (default 50)
* Music.SinglesLimit - How many singles (default 50)
* Music.SimilarArtistsLimit - Many many similar artists (default 10)
* Music.SimilarReleases - Duration for similar (default 8760h = 1 year)
* Music.SimilarRelesesLimit - Many many similar releases (default 10)

### Server

* Server.Listen - Address and port to listen on (default 127.0.0.1:3000)
* Server.WebDir - Directory with web static files and templates

Note: use Nginx or other frontend with TLS (and Let's Encrypt)

### Search Index

* Search.BleveDir - Directory used to store index (default .)

### HTTP Client

* Client.UseCache - Enable or disable http caching (default false)
* Client.MaxAge - Age in seconds to use cached responses (default 30 days)
* Client.CacheDir - Directory to store cached responses (default .httpcache)

Note: this client is currently only used for MusicBrainz API calls

## Examples

### Required fields in ini format

	[music]
	[music.bucket]
	endpoint = "s3.us-west-1.wasabisys.com"
	region = "us-west-1"
	accessKeyID = "<insert access key here>"
	secretAccessKey = "<insert secret key here>"
	bucketName = "MyBucket"
	objectPrefix = "MyMedia"
