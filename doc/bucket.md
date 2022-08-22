# Music Files

Takeout will index all the objects in the S3 bucket to find music files that
start with the configured prefix and end with supported file extensions: mp3, flac,
ogg, m4a. Files are not accessed to inspect metadata tags. Instead, a specific
file path structure is needed to understand and obtain metadata.

The bucket file path structure should be:

	bucket/prefix/artist/album/track

Where album can be:

	album
	album (year)

And track can be:

	1-track.mp3
	01-track.mp3
	1-01-track.mp3

With tracks being:

* 1-track.mp3 - Track #1 on the album
* 01-track.mp3 - Track #1 on the album
* 2-track.mp3 - Track #2 on the album
* 13-track.mp3 - Track #13 on the album
* 1-01-track.mp3 - Track #1 on disc #1 (multi-disc release)
* 2-13-track.mp3 - Track #13 on disc #2 (multi-disc release)

The most important items in the file path are the artist name, release name,
release year, and track/disc number. Takeout will try to find an artist release
in MusicBrainz for each track. Having a release year helps narrow the available
options.

Note that it's a good idea to use underscores (_) in place of special
characters that aren't allowed or preferable in file names.

## Rewrite Rules

Metadata can change and it can be inconvenient to rename files. One way to
address this is using rewrite rules to dynamically change file paths during the
indexing process. The actual file names are not modified. Below is an example
using YAML:

    RewriteRules:
      - Pattern: "^(.+/)Dr. Octagon(/Dr. Octagon, Part II.+/.+)$"
        Replace: "$1Kool Keith$2"
      - Pattern: "^(.+/White Zombie/La Sexorcisto_ Devil Music, )Volume One(.+/.+)$"
        Replace: "$1Vol. 1$2"

Regular expressions are used to match and substitute text. The first example
changes "Dr. Octagon" to "Kool Keith". The second example changes "Volume One"
to "Vol. 1".

## Examples

A single disc release by the Raconteurs in 2019:

	Music/The Raconteurs/Help Us Stranger (2019)/01-Bored and Razed.flac
	Music/The Raconteurs/Help Us Stranger (2019)/02-Help Me Stranger.flac
	Music/The Raconteurs/Help Us Stranger (2019)/03-Only Child.flac
	Music/The Raconteurs/Help Us Stranger (2019)/04-Don't Bother Me.flac
	Music/The Raconteurs/Help Us Stranger (2019)/05-Shine the Light on Me.flac
	Music/The Raconteurs/Help Us Stranger (2019)/06-Somedays (I Don't Feel Like Trying).flac
	Music/The Raconteurs/Help Us Stranger (2019)/07-Hey Gyp (Dig the Slowness).flac
	Music/The Raconteurs/Help Us Stranger (2019)/08-Sunday Driver.flac
	Music/The Raconteurs/Help Us Stranger (2019)/09-Now That You're Gone.flac
	Music/The Raconteurs/Help Us Stranger (2019)/10-Live a Lie.flac
	Music/The Raconteurs/Help Us Stranger (2019)/11-What's Yours Is Mine.flac
	Music/The Raconteurs/Help Us Stranger (2019)/12-Thoughts and Prayers.flac

A multi-disc release by Tubeway Army in 2019:

	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-01-You Are in My Vision (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-02-The Machmen (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-03-Down in the Park (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-04-Do You Need the Service_ (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-05-The Crazies.flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-06-When the Machines Rock (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-07-Me! I Disconnect From You (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-08-Praying to the Aliens (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-09-It Must Have Been Years (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-10-Only a Downstat.flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-11-I Nearly Married a Human 3 (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-12-Replicas (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/1-13-Are 'Friends' Electric_ (early version).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/2-01-Replicas (early version 2).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/2-02-Down in the Park (early version 2).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/2-03-Are 'Friends' Electric_ (early version 2).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/2-04-We Have a Technical.flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/2-05-Replicas (early version 3).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/2-06-Me, I Disconnect From You (BBC Peel session).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/2-07-Down in the Park (BBC Peel session).flac
	Music/Tubeway Army/Replicas - The First Recordings (2019)/2-08-I Nearly Married a Human (BBC Peel session).flac

## Challenges

A file naming scheme isn't perfect and there can be issues. In most cases
Takeout will try to make reasonable assumptions. For example, the Gorillaz
self-titled Gorillaz (2002) album has the track:

    Music/Gorillaz/Gorillaz (2002)/02-5_4.flac

The song name for track #2 is actually "5/4". Slashes aren't typically allowed
in file names so the tagger used underscore (_) for these non-allowed
characters. Takeout will see this as track #2 and use MusicBrainz to
associate the actual title of "5/4".

Another example is track #11 on the same Gorillaz album:

    Music/Gorillaz/Gorillaz (2002)/11-19-2000.flac

The song title for track #11 is "19-2000". As described above, Takeout may
expect this to be track #19 on disc #11 with a title of "2000". There are some
rules in place that allow Takeout will handle this correctly by assuming it's a
single-disc which is more common.

Another example is track #4 on the ZZ Top album XXX:

	Music/ZZ Top/XXX (1999)/4-36-22-36.flac

The song title for track #4 is "36-22-36". Takeout may expect this to be track
#36 on disc #4 with a title of "22-36". Similar to above, Takeout will assume
this is a single-disc album.

Yet another challenge happens with Weezer albums. They have several self-titled
albums called "Weezer" that are commonly referred to as the Red, Green, Teal,
etc., albums. Some of them have the same number of tracks and the same year so
it's not easy to manage this without help from the file path.

This is an option to store the Weezer "Teal" album released in 2019:

	Music/Weezer/Weezer (Teal Album) (2019)/01-Africa.mp3
	Music/Weezer/Weezer (Teal Album) (2019)/02-Everybody Wants to Rule the World.mp3
	Music/Weezer/Weezer (Teal Album) (2019)/03-Sweet Dreams (Are Made of This).mp3
	Music/Weezer/Weezer (Teal Album) (2019)/04-Take on Me.mp3
	Music/Weezer/Weezer (Teal Album) (2019)/05-Happy Together.mp3
	Music/Weezer/Weezer (Teal Album) (2019)/06-Paranoid.mp3
	Music/Weezer/Weezer (Teal Album) (2019)/07-Mr. Blue Sky.mp3
	Music/Weezer/Weezer (Teal Album) (2019)/08-No Scrubs.mp3
	Music/Weezer/Weezer (Teal Album) (2019)/09-Billie Jean.mp3
	Music/Weezer/Weezer (Teal Album) (2019)/10-Stand by Me.mp3

This is an option to store the Weezer "Black" album released in 2019:

	Music/Weezer/Weezer (Black Album) (2019)/01-Can't Knock the Hustle.flac
	Music/Weezer/Weezer (Black Album) (2019)/02-Zombie Bastards.flac
	Music/Weezer/Weezer (Black Album) (2019)/03-High as a Kite.flac
	Music/Weezer/Weezer (Black Album) (2019)/04-Living in L.A..flac
	Music/Weezer/Weezer (Black Album) (2019)/05-Piece of Cake.flac
	Music/Weezer/Weezer (Black Album) (2019)/06-I'm Just Being Honest.flac
	Music/Weezer/Weezer (Black Album) (2019)/07-Too Many Thoughts in My Head.flac
	Music/Weezer/Weezer (Black Album) (2019)/08-The Prince Who Wanted Everything.flac
	Music/Weezer/Weezer (Black Album) (2019)/09-Byzantine.flac
	Music/Weezer/Weezer (Black Album) (2019)/10-California Snow.flac

MusicBrainz includes disambiguation information that can be used to help find
the correct release. Takeout will look for a disambiguation in the release
title.  The following would work for these Weezer examples:

     Weezer (Black Album)
	 Weezer - Black Album
	 Weezer Black Album
     Weezer [Black Album]
	 Black Album

# Movie Files

Takeout will index all the objects in the S3 bucket to find movie files that
start with the configured prefix and end with supported file extensions: mkv,
mp4. Files are not individually opened to inspect tags or headers.  Instead, a
specific file path structure is needed to understand and obtain metadata.

The bucket file path structure should be:

	bucket/prefix/path/movie

Where movie can be:

    Title (year).mkv
    Title (year) - HD.mkv

The path between and prefix and movie file is not used by Takeout. The only
important fields are the title and year, and these are very important since
they are both required to uniquely identify a movie.

An optional dash followed by addition information is allowed and ignored by
Takeout. Use this for indicating whether or not the file is 4k, HD, SD,
director's cut, etc.

It's important to note that since Takeout does not parse file containers,
indexing is not able to determine movie details such as aspect ratio, audio
tracks, closed captions and video quality. This information will only be
available during playback.

Takeout does not support multi-file movies. It's recommended to merge files
like this together into one movie file using mkvmerge or similar.

## Examples

    Movies/Action/Birds of Prey (and the Fantabulous Emancipation of One Harley Quinn) (2020) - HD.mkv
    Movies/Comedy/Monty Python and the Holy Grail (1975).mkv
    Movies/Horror/Evil Dead (2013).mkv
    Movies/Horror/Evil Dead II (1987).mkv
    Movies/Sci-Fi/Total Recall (1990).mkv
    Movies/Sci-Fi/Total Recall (2012) - HD.mkv
    Movies/Sci-Fi/Dune (1984) - HD.mkv
    Movies/Sci-Fi/Dune (2021) - HD.mkv

Note the genre information in the examples is just for organization
purposes. Takeout uses TMDB to obtain metadata and genre information.

# TV Files

Takeout has partial support for indexing TV episodes but nothing is enabled yet.
