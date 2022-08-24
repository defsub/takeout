# Actions

Note - this will stop working June 13, 2023 due to
[Converstational Actions Sunset](https://developers.google.com/assistant/ca-sunset)

Takeout optionally supports a Google Assistant webhook to play music using the
Google Assistant app, Nest Audio, Nest Hub, Google Home, and related products
and services. Details related to how this is designed and implemented are
described below.

## Verified Users

An Assistant verified user is required to use this service due to the fact that
account linking is used to associate a Takeout auth cookie with the Google
user account. Verified users must meet the following Google Account / Assistant
criteria:

* Voice Match - Enabled
* Personal Results - Ennabled
* Web & App Activity - Enabled

An auth code is provided to the Assistant user which they must link within
Takeout using their user, password, and code. Account linking based on a code
is available here: [Takeout Link](https://yourhost.com/link)

More details available here:

* [User Storage](https://developers.google.com/assistant/conversational/storage-user)

## Intents

### TAKEOUT_AUTH

A verified user without a linked account (yet) will receive a code. After
linking the code, the following phases can be used to confirm authentication
using the TAKEOUT_AUTH intent.

> Next

> Auth

> Link

> Authenticate

### TAKEOUT_NEW

A linked user can use the following phrases to send the TAKEOUT_NEW intent.

> What's new

> Tell me what's new

> New

> New albums

> New music

### TAKEOUT_PLAY

A linked user can use the following phrases to play music using the
TAKEOUT_PLAY. This intent supports parameters within the phrases:

* artist - Artist name
* release - Album/Release name
* song - Song title
* radio - Radio mode (similar artist, popular tracks)
* popular - Popular tracks
* latest - Latest release

The actual phrases are defined within the Actions Console. The assumed
supported and example phrases are:

#### Play the latest album release from the specified artist.

> play the new [artist]

> play the latest [artist]

> play the latest by/from [artist]

#### Play artist radio which is popular tracks from the artist and similar artists.

> play [artist] radio

#### Play artist popular songs.

> play popular songs by [artist]

#### Play a specific song (or songs) by name from an artist.

> play [song] by [artist]

#### Play a specific album by name from an artist.

> play album [release] by [artist]

#### Play singles by an artist.

> play [artist] songs

> play songs by [artist]

#### Play a specific song or songs by name.

> play song [song]

#### Play a specific album.

> play album [release]

#### Play any where any can be a release, song or artist.

> play [any]
