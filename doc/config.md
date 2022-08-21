# Takeout Configuration

## Configuration ##

The configuration file _takeout.ini_ is used to store required configuration to
setup and configure Takeout. YAML and other formats supported by
[Viper](https://github.com/spf13/viper) can be used as well.


### Auth

* Auth.DB.Driver - Database driver (default sqlite3)
* Auth.DB.Logger - "default", "debug", or "discard"
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
