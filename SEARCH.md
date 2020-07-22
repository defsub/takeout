# Search

## General Fields

* artist - Artist name(s) from artist credits within release (album) media tracks
* asin - Amazon Standard Identifcation Number (optional)
* date - First track release date (in any release) or same as release_date
* first_date - First release (album) date
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
* type - Track types including: single

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

## Examples

Tracks with Mogwai in any field:

	mogwai

Tracks by the artist Mogwai:

	artist:mogwai

Tracks with Ringo on lead vocals by The Beatles release between after 1963 and
before 1970:

    +lead_vocals:ringo +artist:beatles +date:>"1963-01-01" +date:<"1969-12-31"

Tracks with Tom Morello that aren't Rage or Audioslave:

	+morello -rage -audioslave

Similar but more specific:

	+guitar:morello -artist:rage -artist:audioslave

And even more specific:

	+guitar:"Tom Morello" -artist:"Rage Against the Machine" -artist:"Audioslave"

80s alternative tracks released as singles:

	+genre:"alternative" +type:single +date:>="1980-01-01" +date:<="1989-01-01"

Tracks with flute and violin played by anyone:

	+flute:* +violin:*

Tracks longer than 15 minutes (60*15=900):

	+length:>900

# Bleve

Please see Bleve documentation for further information on query syntax. Takeout
is using:

	https://blevesearch.com/docs/Query-String-Query/
