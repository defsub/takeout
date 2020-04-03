package music

import (
	"errors"
	"fmt"
	"github.com/defsub/takeout/config"
	"github.com/jinzhu/gorm"
	"github.com/minio/minio-go/v6"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Music struct {
	config *config.Config
	db     *gorm.DB
	minio  *minio.Client
}

type Artist struct {
	gorm.Model
	Name string `gorm:"unique_index:idx_artist"`
	MBID string
}

type Release struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_release"`
	Name   string `gorm:"unique_index:idx_release"`
	MBID   string `gorm:"unique_index:idx_release"`
	Asin   string
	Type   string
	Date   time.Time
}

type Popular struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_popular"`
	Title  string `gorm:"unique_index:idx_popular"`
	Rank   uint
}

type ArtistTag struct {
	gorm.Model
	Artist string `gorm:"unique_index:idx_tag"`
	Tag    string `gorm:"unique_index:idx_tag"`
	Count  uint
}

type Track struct {
	gorm.Model
	Artist       string `spiff:"creator"`
	Release      string `spiff:"album"`
	TrackNum     uint   `spiff:"tracknum"`
	DiscNum      uint
	Title        string `spiff:"title"`
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	Location     string `gorm:"-" spiff:"location"`
}

func NewMusic(config *config.Config) *Music {
	return &Music{config: config}
}

func (Popular) TableName() string {
	return "popular" // not populars
}

func (m *Music) Open() (err error) {
	err = m.openDB()
	if err == nil {
		err = m.openBucket()
	}
	return
}

func (m *Music) Close() {
	m.closeDB()
}

func (m *Music) Sync() (err error) {
	m.deleteTracks()
	trackCh, err := m.SyncFromBucket()
	if err != nil {
		return err
	}
	for t := range trackCh {
		fmt.Printf("sync: %s / %s / %s\n", t.Artist, t.Release, t.Title)
		// TODO: title may have underscores - picard
		m.createTrack(t)
	}
	return
}

func (m *Music) Releases(artist string) {
	a := m.artist(artist)
	releases := m.artistReleases(a)
	for _, v := range releases {
		fmt.Printf("releases: %s / %s / %s / %s / %d\n",
			v.MBID, v.Artist, v.Name, v.Type, v.Date.Year())
	}
}

func (m *Music) SyncReleases() {
	artists := m.artistNames()
	for _, a := range artists {
		fmt.Printf("releases for %s\n", a)
		artist := m.artist(a)
		if artist == nil {
			continue
		}

		if artist.Name == "Various Artists" {
			// skipping!
			continue
		}

		releases := m.artistReleases(artist)
		if len(releases) > 0 {
			fmt.Printf("skipping %s have %d releases\n", artist.Name, len(releases))
			continue
		}

		releases = m.searchReleases(artist)
		sort.Slice(releases, func(i, j int) bool {
			return releases[i].Date.Before(releases[j].Date)
		})

		for _, r := range releases {
			r.Name = fixName(r.Name)
			if r.Date.Year() == 1 {
				// skip those w/o year
				continue
			}
			m.createRelease(&r)
		}
	}
}

func fuzzyArtist(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9& -]`)
	return re.ReplaceAllString(name, "")
}

func fuzzyName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(name, "")
}

func fixName(name string) string {
	// TODO: use Map?
	name = strings.Replace(name, "–", "-", -1)
	name = strings.Replace(name, "‐", "-", -1)
	name = strings.Replace(name, "’", "'", -1)
	name = strings.Replace(name, "‘", "'", -1)
	name = strings.Replace(name, "“", "\"", -1)
	name = strings.Replace(name, "”", "\"", -1)
	return name
}

func (m *Music) FixTrackReleases() error {
	tracks := m.tracksWithoutReleases()

	fixReleases := make(map[string]string)

	for _, t := range tracks {
		artist := m.artist(t.Artist)
		if artist == nil {
			fmt.Printf("artist not found: %s\n", t.Artist)
			continue
		}
		releases := m.artistReleasesLike(artist, t.Release)
		if len(releases) == 1 {
			fixReleases[t.Release] = releases[0].Name
		} else {
			releases = m.artistReleases(artist)
			matched := false
			for _, r := range releases {
				if fuzzyName(t.Release) == fuzzyName(r.Name) {
					fixReleases[t.Release] = r.Name
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("unmatched '%s' / '%s'\n", t.Artist, t.Release)
			}
		}
	}

	for oldName, newName := range fixReleases {
		err := m.updateTrackRelease(oldName, newName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Music) SyncPopular() {
	artists := m.artistNames()
	for _, a := range artists {
		fmt.Printf("popular for %s\n", a)
		artist := m.artist(a)
		if artist == nil {
			continue
		}
		popular := m.popularByArtist(artist)
		for _, p := range popular {
			m.createPopular(&p)
		}
	}
}

func (m *Music) MetaSync() error {
	artists := m.artistNames()
	for _, a := range artists {
		artist := m.artist(a)
		if artist == nil {
			fmt.Printf("not found %s\n", a)
			var tags []ArtistTag
			artist, tags = m.SearchArtist(a)
			if artist == nil {
				fmt.Printf("try with %s\n", fuzzyArtist(a))
				artist, tags = m.SearchArtist(fuzzyArtist(a))
			}
			if artist != nil {
				artist.Name = fixName(artist.Name)
				fmt.Printf("creating %s\n", artist.Name)
				m.createArtist(artist)
				for _, t := range tags {
					t.Artist = artist.Name
					m.createArtistTag(&t)
				}
			}
		}

		if artist == nil {
			err := errors.New("artist not found")
			fmt.Printf("MetaSync %s\n", err)
			continue
		}

		if a != artist.Name {
			// fix track artist name: AC_DC -> AC/DC
			fmt.Printf("fixing name %s to %s\n", a, artist.Name)
			m.updateTrackArtist(a, artist.Name)
		}
	}
	return nil
}

func (m *Music) doTracks(f func() []Track) []Track {
	tracks := f()
	for i, _ := range tracks {
		tracks[i].Location = m.objectURL(tracks[i]).String()
	}
	return tracks
}

func (m *Music) Tracks(tags string, dr *DateRange) []Track {
	return m.doTracks(func() []Track { return m.tracks(tags, dr) })
}

func (m *Music) Singles(tags string, dr *DateRange) []Track {
	return m.doTracks(func() []Track { return m.singleTracks(tags, dr) })
}

func (m *Music) Popular(tags string, dr *DateRange) []Track {
	return m.doTracks(func() []Track { return m.popularTracks(tags, dr) })
}

func (m *Music) ArtistSingles(artists string, dr *DateRange) []Track {
	return m.doTracks(func() []Track { return m.artistSingleTracks(artists, dr) })
}

func (m *Music) ArtistTracks(artists string, dr *DateRange) []Track {
	return m.doTracks(func() []Track { return m.artistTracks(artists, dr) })
}

func (m *Music) ArtistPopular(artists string, dr *DateRange) []Track {
	return m.doTracks(func() []Track { return m.artistPopularTracks(artists, dr) })
}
