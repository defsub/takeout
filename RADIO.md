# Radio

The radio feature is one of the primary reasons Takeout was created. Radio is
intended to be a way to avoid manual creation of playlists and instead
dynamically find and (re)discover songs in a music collection.

Radio stations can be created from anything in the vast search index of
metadata. The [search index](SEARCH.md) contains the usual artist, album and
track names but also goes deep into metadata:

* Genres
* Labels
* Artist credits including performers and instruments
* First release dates
* Singles
* Popularity
* Cover songs
* Live songs

Radio stations are search queries. You can create your own by adding them to
the [config file](CONFIG.md). Test queries using the search interface and see
what you can find. The defaults are:

* Series Hits:  +series:*
* Top Hits:    +popularity:1
* Top 3 Hits:  +popularity:<4
* Top 5 Hits:  +popularity:<6
* Top 10 Hits: +popularity:<11
* Covers:      +type:cover
* Live Hits:   +type:live +popularity:<3

Tracks with series have been included in a MusicBrainz series. These are
generally things like top 100 lists. A specific series name can also be used.

Popularity comes from Last.fm with 1 being the most popular.

## Genre Stations

Genre based stations are also created using search queries. The query used is:

	+genre:"alternative" +type:single +popularity:<11 -artist:"Various Artists"

This finds all alternative tracks, released as singles, with popularity 1-10,
and excludes various artists.

## Decade Stations

Decade based stations are created using search queries. The query used similar to:

	+first_date:>="1980-01-01" +first_date:<="1980-12-31" +type:single +popularity:<11

This finds all tracks first released in the 1980s, released as a singles, with popularity 1-10.

## Go Crazy

* Epic 20+ minute tracks: +length:>1200
* Def Jam singles:        +label:"def jam" +type:single
* Deep tracks:            +track:>10
* Popular violin:         +violin:* +popularity:<4
* Ozzy covers:            +ozzy +type:cover
* Popular live covers:    +cover +popular +live +single
