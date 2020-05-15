# Takeout #

Takeout indexes S3 buckets to make media available for download and
streaming.

Currently supported media:

  * Music - files ending in mp3, flac, ogg and m4a


## Music ##

### Music Naming ###

Takeout will index all the objects in the S3 bucket to find music
files that end with supported file extensions. Since the actual files
are not individually opened to inspect tags, a specific file structure
is needed. MusicBrainz Picard is _highly_ recommended to manage and tag
music files.

The bucket structure should be:

> bucket/prefix/artist/album/track

where album can be:

> album
> album (year)

and track can be:

> 1-track.mp3
> 01-track.mp3
> 1-01-track.mp3

Examples:

A single disc release by the Raconteurs in 2019:

> The Raconteurs / Help Us Stranger (2019) / 01-Bored and Razed.flac

A multi disc release by Tubeway Army in 2019, track 1 from disc 1 and 2:

> Tubeway Army / Replicas - The First Recordings (2019) / 1-01-You Are in My Vision (early version).flac
> Tubeway Army / Replicas - The First Recordings (2019) / 2-01-Replicas (early version 2).flac

### Music Metadata ###

MusicBrainz APIs are used to obtain further information about artists
and albums found in the bucket. This is used to try to correct artist
and album names, and also to determine which tracks are singles or EPs
and their first release date.

Last.fm APIs are used to obtain popular track information for each
artist.


## Configuration ##

Create a takeout.ini file with:

  * endpoint - Fully qualified domain name of the S3 bucket endpoint
  * accessKeyID - Access key for S3 bucket
  * secretAccessKey - Secret key for S3 bucket
  * bucketName - Name of the S3 bucket
  * objectPrefix - Prefix of media within the S3 bucket

Example:

	[music]
	[music.bucket]
	endpoint = "s3.us-west-1.wasabisys.com"
	accessKeyID = "<insert access key here>"
	secretAccessKey = "<insert secret key here>"
	bucketName = "MyBucket"
	objectPrefix = "MyMedia"

## Tools ##

  * [MusicBrainz Picard][https://picard.musicbrainz.org/]
  * [Rclone][https://rclone.org/]
  * [Minio][https://min.io/]
  * [Wasabi][https://wasabi.com/]
