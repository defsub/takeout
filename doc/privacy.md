# Takeout Project Privacy Policy

The Takeout project is a copyleft personal media system that may be entirely
managed by the end user, a third-party, or a combination of the two. Privacy as
it relates to the original intended design of the Takeout project, independent
to how it is managed, is described within this document. The Takeout project
refers to the Takeout server, mobile apps, TV apps, and the web interface. If
you have questions regarding this privacy policy, please contact
privacy@defsub.com.

## Design Considerations

The Takeout project is designed around a server and APIs for clients/apps to
browse and consume media. Clients are not required to store any information but
may do so to improve performance. State may be stored on the server such that
activity and progress can be resumed or shared across multiple client
devices. The Takeout project does not store or manage any media and instead
media must be stored in an S3 bucket that is available for the server to index
and for clients to access directly using time-based pre-signed URLs.

## Personal Information

The Takeout server requires a username and password for each user to access
related media and services. The usernames and passwords are stored in the
server _auth_ database. The username is stored in the clear and the password is
stored with [Scrypt](https://en.wikipedia.org/wiki/Scrypt). No other personal
information is requested or stored by Takeout.

The Takeout server may temporarily store access logs that contain client
request information and IP addresses. This information, if used, is only used
for debugging or development purposes.

The Takeout server is recommended to be configured with TLS to ensure all
communication is encrypted to avoid unintended disclosure of usernames and
passwords.

## Cookies

Cookies are small tokens or files stored on your device as part of the user
login process to uniquely identify you later without requiring you to provide
your username and password again. Each cookie is a UUID comprised of 122 random
bits, stored within the Takeout server _auth_ database, and within the client
app or web browser. Cookies are valid for a limited time (based on server
configuration) and when expired, you will be required to login again.

The Takeout server is recommended to be configured with TLS to ensure all
communication is encrypted to avoid unintended disclosure of cookies.

## Media

The Takeout server requires access to your S3 bucket to obtain a listing of
media stored within the S3 bucket. The bucket object file names are used to
obtain further metadata related to music and video files. These object names
are stored in the corresponding _music_, _video_ and _search_ databases to
enable media streaming or downloading directly from your S3 bucket using
time-based pre-signed URLs.

The Takeout server does not access your media, it does not parse your media
containers, and it does not parse any embedded tags or related information in
your media. All related metadata is obtained using third-party services based
on file naming conventions.  The Takeout server is not a source or provider of
any music, video, or movie media.

Media stored in your S3 bucket can potentially be visible to the S3 bucket
service provider. Contact your service provider to obtain further information
regarding the S3 bucket privacy policy. Personal S3 bucket hosting options,
such as [Minio](https://min.io/), are available.

## Metadata

The Takeout server uses the following services to discover your media metadata:

- [Cover Art Archive](https://coverartarchive.org/) - obtain links to cover images
- [Fanart.tv](https://fanart.tv/) - obtain links to Artist images (requires API key)
- [MusicBrainz](https://musicbrainz.org/) - obtain music metadata
- [Last.fm](https://www.last.fm/) - obtain popular tracks, artist name resolving (requires API key)
- [The Movie Database](https://www.themoviedb.org/) - obtain movie metadata (requires API key)

The Takeout server uses the respective service APIs to query and store related
metadata based on your S3 bucket object file names. Requests to the service
APIs will include an API key (where required), media information (such as
artist or movie name), and the Takeout server IP address. Third-party services
can infer information about your music or movies that are being indexed and
potentially relate the media to a unique IP address. No other information is
directly provided to these third-party services.

Metadata related to your S3 bucket object file names is stored in the
respective _music_, _video_, and _search_ databases to improve performance and
reduce the overall impact on third-party services. Similarly, API responses can
also be cached to avoid repeated or duplicate requests for the same
information.

Metadata includes links or URLs to images such as covers, posters, and
profiles. The Takeout server includes such URLs or information to construct
URLs in responses to API clients. The URLs are used by clients to render
associated media images in their UI. Third-party services can infer information
about your media from image requests and relate to a unique IP address. No
other information is directly provided to these third-party services.

Podcast URLs that have been added to the Takeout server configuration are
periodically queried and metadata is stored in the _podcast_ database. Requests
to the podcast providers can infer information about your interests and
potentially relate the podcasts to a unique IP address. No other information is
directly provided to the podcast providers.

Radio stream URLs that have been added to the Takeout server configuration are
used to create a radio station in the _music_ database. The URLs for these
streams are sent directly to API clients and clients can use them to directly
stream media. Requests to radio stream providers can infer information about
your interests and potentially relate streams to a unique IP address. No other
information is directly provided to the radio providers.

## Progress

The Takeout server provides APIs for clients to store media watch/listen
progress which is intended to allow playback to be conveniently resumed on the
same or other devices at a later time. The Takeout server stores progress using
the ETag (or entity tag) of the media and an offset in the media stream. An
ETag is generally an MD5 digest of the media content obtained from the S3
bucket. This design of ETag based progress provides a layer of indirection such
that user media consumption is not readily available. It's possible, with
access to the _progress_ database and the S3 bucket, to reconstruct media
consumption information by mapping progress ETags to S3 bucket media.

## Activity

The Takeout server provides APIs for clients to store activity events which are
intended to allow the user to easily access recently consumed media on the same
or other devices. Activity events can relate to music, video, and podcasts.
Activity events are stored in the _activity_ database and each event uses
third-party identifiers (MBIDs, GUIDs, IMDB IDs) such that activity data is
stable. It's possible, with access to the _activity_ database, to reconstruct
media consumption information by mapping third-party identifiers to actual
metadata.

Activity APIs are still a working in progress.

## Information Disclosure

The Takeout server does not directly disclose any information to any outside
parties beyond what is needed to obtain metadata.

## Children’s Online Privacy Protection Act Compliance

The Takeout project is directed at people that are 13 years old or older. If
the Takeout server is in the USA, and you are under age of 13, per the
requirements of COPPA (Children’s Online Privacy Protection Act), do not use
the Takeout server.

## Google Assistant

The Takeout server can optionally be used with Google Assistant enabled devices
and apps. A cookie is used to link your Google Assistant user to your Takeout
user.  The cookie is stored in your Google Assistant [user
storage](https://developers.google.com/assistant/conversational/storage-user).
This process requires you to enable voice match, personal results, and web &
app activity. See the corresponding Google privacy policy for information
regarding these settings.

Phrases you use with Google Assistant to access the Takeout server are processed
by the Assistant, sent to Google services, and finally to the Takeout server
webhook where they are processed as text strings contained within intents and
parameters. The webhook will respond to Google services with media metadata and
time-based pre-signed URLs for the Assistant to access your media.

The Takeout server does not store or have access to your voice data. The
resulting translated text queries, intents, and parameters are not stored by
Takeout however they may be used for debugging purposes. Only voice matched
authenticated requests from the Google Assistant are allowed to access your
media and metadata.

## Consent

By using the Takeout project, you consent to this privacy policy.

## Changes

Any changes made to this privacy policy will be made available in this file at
the [Takeout server github repository](https://github.com/defsub/takeout).
