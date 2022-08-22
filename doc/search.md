# Takeout Search

## Music Fields

* artist - Artist name(s) from artist credits within release (album) media tracks
* asin - Amazon Standard Identifcation Number (optional)
* date - First release date
* first_date - First track release date
* genre - Genres associated with artist(s) and release (album)
* media - Media (disc) index, 1 for single disc albums
* media_title - Media (disc) specific title (optional)
* label - Labels for the release (album)
* length - Track length in seconds
* rating - Numeric rating (optional)
* release - Name of the release (album)
* release_date - Date release (album) was released
* tag - Tags associated with artists(s) and release (album)
* title - Track title
* track - Track number
* type - Track types including: single, popular, cover, live

Note that date fields are YYYY-MM-DD and leading zeros are required.

## Credit Fields

Credits are relations associated with the track and release from
MusicBrainz. These are all optional and may not be unavailable. Some
generalizations are added to make it easier to search. For example, "lead
guitar", "acoustic guitar", and "slide guitar", are additionally added as
simply "guitar". The same applies for "bass", "clarinet", "drums", "flute",
"piano", "saxophone", and "vocals".  Other credits added include "arranger",
"composer", "engineer", "lyricist", "mix", and "writer". Check MusicBrainz for
more.

## Music Examples

Tracks with Mogwai in any field:

	mogwai

Tracks by the artist Mogwai:

	artist:mogwai

Tracks with Ringo on lead vocals by The Beatles released between 1963 and 1970:

    +lead_vocals:ringo +artist:beatles +first_date:>"1963-01-01" +first_date:<"1969-12-31"

Tracks with Tom Morello that aren't Rage or Audioslave:

	+morello -rage -audioslave

Similar to above but more specific:

	+guitar:morello -artist:rage -artist:audioslave

And even more specific:

	+guitar:"Tom Morello" -artist:"Rage Against the Machine" -artist:"Audioslave"

80s alternative tracks released as singles:

	+genre:"alternative" +type:single +first_date:>="1980-01-01" +first_date:<="1989-12-31"

Tracks with flute and violin played by anyone:

	+flute:* +violin:*

Tracks longer than 15 minutes (60*15=900):

	+length:>900

Tracks produced by Butch Vig:

	producer:"butch vig"

Cover songs:

	+type:cover

Popular cover songs, performed live, that were released as singles:

	+type:cover +type:live +type:single +type:popular

Epic 20+ minute songs:

	+length:>1200 -silence

## Movie Fields

* budget - Budget in USD
* cast - Cast member name
* character - Character name
* collection - Collection name
* crew - Crew member name
* date - Release date
* genre - Movie genre
* rating - Release rating or certification
* runtime - Time in minutes
* tagline - One liner
* title - Movie title
* vote - TMDb vote %
* vote_count - TMDb vote count

Crew jobs to index are configuratable. Defaults are:

* director
* executive_producer
* novel
* producer
* screenplay
* story

Note that date fields are YYYY-MM-DD and leading zeros are required.

## Movie Examples

Directed by Tarantino:

	+director:tarantino

Movies with R rating:

	+rating:R

Animated movies with Liam Neeson:

	+cast:"liam neeson" +genre:Animation

Christmas movies with Bruce Willis:

	+christmas +willis

Yes, Die Hard is a Christmas movie.

Same as above but more specific:

    +keyword:"christmas" +cast:"bruce willis"

Movies with a character named "Yoda":

    +character:"yoda"

Really long movies:

    +runtime:>200

Highly rated movies on IMDb:

    +vote:>80

And those that are not highly rated:

	+vote:>0 +vote:<50

Low budget films (<$250k):

	+budget:>0 +budget:<250000

High budget films (>$250M):

    +budget:>250000000

# Bleve

Takeout uses the Bleve search library for all search capabilities.  Please see
[Bleve](https://blevesearch.com/) documentation for further information on
query syntax. Takeout is using [Query String](https://blevesearch.com/docs/Query-String-Query/).
