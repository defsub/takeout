// Package server Takeout API
//
// This is the service API for the Takeout media server.
//
// Terms Of Service:
//
// TOS is TBD
//
// Schemes: https
// Host: yourhost.com
// BasePath: /api
// Version: 0.0.1
// License: AGPLv3 https://www.gnu.org/licenses/agpl-3.0.en.html
// Contact: defsub@defsub.com
// SecurityDefinitions:
//  Cookie:
//   type: apiKey
//   name: Cookie
//   description: send Cookie Takeout={token}
//   in: header
//  Bearer:
//   type: apiKey
//   name: Authorization
//   description: send Authorization Bearer {token}
//   scheme: bearer
//   in: header
// Security:
//  - Bearer:
//  - Cookie:
// Consumes:
// - application/json
// Produces:
// - application/json
//
// swagger:meta
package server

import (
	"github.com/defsub/takeout/lib/spiff"
	"github.com/defsub/takeout/progress"
	"github.com/defsub/takeout/view"
)

// ---------------------------------------------------------------------------

// swagger:route POST /login Login
// responses:
//  200: StatusResponse
//  401: fail
//  500: ServerError

// swagger:route GET /index Index
// responses:
//  200: IndexResponse

// swagger:route GET /home Home
// responses:
//  200: HomeResponse

// ---------------------------------------------------------------------------

// swagger:route GET /artists ArtistsList
//  List all artists
// responses:
//  200: ArtistsResponse

// swagger:route GET /artists/{id} ArtistGet
//  Get artist details
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: ArtistResponse
//  404: description: artist not found

// swagger:route GET /artists/{id}/playlist.xspf ArtistPlaylistExport
//  Get
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: artist not found

// swagger:route GET /artists/{id}/playlist ArtistPlaylist
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: artist not found

// swagger:route GET /artists/{id}/popular.xspf ArtistPopularExport
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: artist not found

// swagger:route GET /artists/{id}/popular ArtistPopular
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: artist not found

// swagger:route GET /artists/{id}/radio.xspf ArtistRadioExport
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: artist not found

// swagger:route GET /artists/{id}/radio ArtistRadio
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: artist not found

// swagger:route GET /artists/{id}/singles.xspf ArtistSinglesExport
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: artist not found

// swagger:route GET /artists/{id}/singles ArtistSingles
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: artist not found

// ---------------------------------------------------------------------------

// swagger:route GET /movies MoviesList
// Responses:
//  200: MoviesResponse

// swagger:route GET /movies/{id} MovieGet
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: MovieResponse
//  404: description: movie not found

// swagger:route GET /movies/{id}/location MovieLocation
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  302:
//   description: redirect with location
//   headers:
//    "Location":
//     type: string
//     description: URL to movie

// ---------------------------------------------------------------------------

// swagger:route GET /playlist PlaylistGet
// responses:
//  200: PlaylistResponse

// swagger:route PATCH /playlist PlaylistPatch
// responses:
//  200: PlaylistResponse
//  204: description: no change to track entries
//  500: description: server error

// ---------------------------------------------------------------------------

// swagger:route GET /podcasts PodcastsList
// Responses:
//  200: PodcastsResponse

// ---------------------------------------------------------------------------

// swagger:route GET /profiles/{id} ProfileGet
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: ProfileResponse
//  404: description: profile not found

// ---------------------------------------------------------------------------

// swagger:route GET /progress ProgressList
// responses:
//  200: ProgressResponse

// swagger:route POST /progress ProgressUpdate
// responses:
//  204: description: accepted, no response
//  400: description: error
//  500: description: error

// swagger:route GET /progress/{id} ProgressGet
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: ProgressResponse
//  404: description: progress not found

// swagger:route DELETE /progress/{id} ProgressDelete
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  204: description: deleted, no response
//  404: description: id not found
//  500: description: server error

// ---------------------------------------------------------------------------

// swagger:route GET /releases/{id} ReleaseGet
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: ReleaseResponse
//  404: description: release not found

// ---------------------------------------------------------------------------

// swagger:route GET /series/{id} SeriesGet
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: SeriesResponse
//  404: description: series not found

// swagger:route GET /series/{series_id}/episode/{episode_id} SeriesEpisodeGet
// parameters:
//  + in: path
//    name: series_id
//    type: integer
//    required: true
//  + in: path
//    name: episode_id
//    type: integer
//    required: true
// responses:
//  200: SeriesEpisodeResponse
//  404: description: series or episode not found

// swagger:route GET /episode/{id}/location EpisodeLocation
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  301: description: redirect with location

// ---------------------------------------------------------------------------

// swagger:route GET /tracks/{id}/location TrackLocation
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  301: description: redirect with location

// ---------------------------------------------------------------------------

// swagger:route GET /radio/{id} RadioGet
// parameters:
//  + in: path
//    name: id
//    type: integer
//    required: true
// responses:
//  200: PlaylistResponse
//  404: description: not found

// swagger:route GET /radio RadioList
// responses:
//  200: RadioResponse

// swagger:route POST /radio RadioUpdate
// responses:
//  200: RadioResponse


// POST /api/radio < Station{}
// 201: created
// 400: bad request
// 500: error

// ---------------------------------------------------------------------------

// swagger:route GET /search Search
// parameters:
//  + in: query
//    name: q
//    type: string
//    required: true
// responses:
//  200: SearchResponse

// ---------------------------------------------------------------------------


// swagger:parameters Login
type LoginParam struct {
	// in: body
	Body struct {
		login
	}
}

// swagger:response
type StatusResponse struct {
	// in: body
	Body struct {
		status
	}
}

// swagger:response
type ArtistResponse struct {
	// in: body
	Body struct {
		view.Artist
	}
}

// swagger:response
type ArtistsResponse struct {
	// in: body
	Body struct {
		view.Artists
	}
}

// swagger:response
type GenreResponse struct {
	// in: body
	Body struct {
		view.Genre
	}
}

// swagger:response
type HomeResponse struct {
	// in: body
	Body struct {
		view.Home
	}
}

// swagger:response
type IndexResponse struct {
	// in: body
	Body struct {
		view.Index
	}
}

// swagger:response
type KeywordResponse struct {
	// in: body
	Body struct {
		view.Keyword
	}
}

// swagger:response
type MovieResponse struct {
	// in: body
	Body struct {
		view.Movie
	}
}

// swagger:response
type MoviesResponse struct {
	// in: body
	Body struct {
		view.Movies
	}
}

// swagger:response
type OffsetResponse struct {
	// in: body
	Body struct {
		view.Offset
	}
}

// swagger:response
type PlaylistResponse struct {
	// in: body
	Body struct {
		spiff.Playlist
	}
}

// swagger:response
type PodcastsResponse struct {
	// in: body
	Body struct {
		view.Podcasts
	}
}

// swagger:response
type PopularResponse struct {
	// in: body
	Body struct {
		view.Popular
	}
}

// swagger:response
type ProfileResponse struct {
	// in: body
	Body struct {
		view.Profile
	}
}

// swagger:parameters ProgressUpdate
type ProgressParameter struct {
	// in: body
	Body struct {
		progress.Offsets
	}
}

// swagger:response
type ProgressResponse struct {
	// in: body
	Body struct {
		view.Progress
	}
}

// swagger:response
type RadioResponse struct {
	// in: body
	Body struct {
		view.Radio
	}
}

// swagger:response
type ReleaseResponse struct {
	// in: body
	Body struct {
		view.Release
	}
}

// swagger:response
type SearchResponse struct {
	// in: body
	Body struct {
		view.Search
	}
}

// swagger:response
type SeriesResponse struct {
	// in: body
	Body struct {
		view.Series
	}
}

// swagger:response
type SeriesEpisodeResponse struct {
	// in: body
	Body struct {
		view.SeriesEpisode
	}
}

// swagger:response
type SinglesResponse struct {
	// in: body
	Body struct {
		view.Singles
	}
}
