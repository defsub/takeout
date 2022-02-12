// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// Takeout is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for
// more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/defsub/takeout"
	"github.com/spf13/viper"
)

const (
	MediaMusic = "music"
	MediaVideo = "video"
)

type BucketConfig struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	ObjectPrefix    string
	UseSSL          bool
	URLExpiration   time.Duration
	Media           string
}

type DatabaseConfig struct {
	Driver  string
	Source  string
	LogMode bool
}

type Template struct {
	Text  string
	templ *template.Template
}

func (t *Template) Template() *template.Template {
	if t.templ == nil {
		t.templ = template.Must(template.New("t").Parse(t.Text))
	}
	return t.templ
}

func (t *Template) Execute(vars interface{}) string {
	var buf bytes.Buffer
	_ = t.Template().Execute(&buf, vars)
	return buf.String()
}

type AssistantResponse struct {
	Speech Template
	Text   Template
}

type AssistantConfig struct {
	ProjectID       string
	TrackLimit      int
	RecentLimit     int
	Welcome         AssistantResponse
	Play            AssistantResponse
	Error           AssistantResponse
	Link            AssistantResponse
	Linked          AssistantResponse
	Guest           AssistantResponse
	Recent          AssistantResponse
	Release         AssistantResponse
	SuggestionAuth  string
	SuggestionNew   string
	MediaObjectName Template
	MediaObjectDesc Template
}

type RadioStream struct {
	Creator  string
	Title    string
	Image    string
	Location string
}

type MusicConfig struct {
	ArtistFile           string
	ArtistRadioBreadth   int
	ArtistRadioDepth     int
	DB                   DatabaseConfig
	DeepLimit            int
	PopularLimit         int
	RadioGenres          []string
	RadioLimit           int
	RadioOther           map[string]string
	RadioSearchLimit     int
	RadioSeries          []string
	RadioStreams         []RadioStream
	Recent               time.Duration
	RecentLimit          int
	ReleaseCountries     []string
	SearchLimit          int
	SimilarArtistsLimit  int
	SimilarReleases      time.Duration
	SimilarReleasesLimit int
	SinglesLimit         int
	artistMap            map[string]string
	SyncInterval         time.Duration
}

type VideoConfig struct {
	DB               DatabaseConfig
	ReleaseCountries []string
	CastLimit        int
	CrewJobs         []string
	Recent           time.Duration
	RecentLimit      int
	SearchLimit      int
	Recommend        RecommendConfig
	SyncInterval     time.Duration
}

type PodcastConfig struct {
	DB           DatabaseConfig
	Series       []string
	Client       ClientConfig
	RecentLimit  int
	SyncInterval time.Duration
}

type ProgressConfig struct {
	DB DatabaseConfig
}

type RecommendConfig struct {
	When []DateRecommend
}

type DateRecommend struct {
	Name   string
	Layout string
	Match  string
	Query  string
}

type LastFMAPIConfig struct {
	Key    string
	Secret string
}

type FanartAPIConfig struct {
	ProjectKey  string
	PersonalKey string
}

type TMDBAPIConfig struct {
	Key          string
	Language     string
	FileTemplate Template
}

type AuthConfig struct {
	DB            DatabaseConfig
	MaxAge        time.Duration
	CodeAge       time.Duration
	SecureCookies bool
}

type SearchConfig struct {
	BleveDir string
}

type ServerConfig struct {
	Listen string
	URL    string
}

type ClientConfig struct {
	CacheDir  string
	MaxAge    time.Duration
	UseCache  bool
	UserAgent string
}

func (c *ClientConfig) Merge(o ClientConfig) {
	if o.CacheDir != "" {
		c.CacheDir = o.CacheDir
	}
	c.MaxAge = o.MaxAge
	c.UseCache = o.UseCache
	if o.UserAgent != "" {
		c.UserAgent = o.UserAgent
	}
}

type Config struct {
	Auth      AuthConfig
	Buckets   []BucketConfig
	Client    ClientConfig
	DataDir   string
	Fanart    FanartAPIConfig
	LastFM    LastFMAPIConfig
	Music     MusicConfig
	TMDB      TMDBAPIConfig
	Search    SearchConfig
	Server    ServerConfig
	Video     VideoConfig
	Assistant AssistantConfig
	Podcast   PodcastConfig
	Progress  ProgressConfig
}

func (mc *MusicConfig) UserArtistID(name string) (string, bool) {
	mbid, ok := mc.artistMap[name]
	return mbid, ok
}

func readJsonStringMap(file string, m *map[string]string) (err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(data), m)
	return
}

func (mc *MusicConfig) readMaps() {
	if mc.ArtistFile != "" {
		readJsonStringMap(mc.ArtistFile, &mc.artistMap)
	}
}

func configDefaults(v *viper.Viper) {
	v.SetDefault("Auth.DB.Driver", "sqlite3")
	v.SetDefault("Auth.DB.LogMode", "false")
	v.SetDefault("Auth.DB.Source", "auth.db")
	v.SetDefault("Auth.MaxAge", "24h")
	v.SetDefault("Auth.CodeAge", "1h")
	v.SetDefault("Auth.SecureCookies", "true")

	// TODO apply as default
	// v.SetDefault("Bucket.URLExpiration", "72h")
	// v.SetDefault("Bucket.UseSSL", "true")

	v.SetDefault("Client.CacheDir", ".httpcache")
	v.SetDefault("Client.MaxAge", "720h") // 30 days in hours
	v.SetDefault("Client.UseCache", "false")
	v.SetDefault("Client.UserAgent", userAgent())

	v.SetDefault("DataDir", ".")

	v.SetDefault("Fanart.ProjectKey", "93ede276ba6208318031727060b697c8")

	v.SetDefault("LastFM.Key", "77033164cfcda2acc4c58681dcba3cf8")
	v.SetDefault("LastFM.Secret", "8f43410e8e81c33d4542738ee84dc39b")

	v.SetDefault("Music.ArtistRadioBreadth", "10")
	v.SetDefault("Music.ArtistRadioDepth", "3")
	v.SetDefault("Music.DeepLimit", "50")
	v.SetDefault("Music.PopularLimit", "50")
	v.SetDefault("Music.RadioLimit", "25")
	v.SetDefault("Music.RadioSearchLimit", "1000")
	v.SetDefault("Music.RadioStreams", []RadioStream{
		{
			Creator:  "Ted Leibowitz",
			Title:     "BAGeL Radio",
			Image:    "https://cdn-profiles.tunein.com/s187420/images/logod.jpg",
			Location: "https://ais-sa3.cdnstream1.com/2606_128.aac",
		},
		{
			Creator:  "SomaFM",
			Title:     "Groove Salad",
			Image:    "https://somafm.com/img3/groovesalad-400.jpg",
			Location: "https://ice2.somafm.com/groovesalad-128-aac",
		},
	})
	v.SetDefault("Music.Recent", "8760h") // 1 year
	v.SetDefault("Music.RecentLimit", "50")
	v.SetDefault("Music.SearchLimit", "100")
	v.SetDefault("Music.SimilarArtistsLimit", "10")
	v.SetDefault("Music.SimilarReleases", "8760h") // +/- 1 year
	v.SetDefault("Music.SimilarReleasesLimit", "10")
	v.SetDefault("Music.SinglesLimit", "50")
	v.SetDefault("Music.SyncInterval", "1h")

	// see https://wiki.musicbrainz.org/Release_Country
	// v.SetDefault("Music.ReleaseCountries", []string{
	// 	"US", // United States
	// 	"XW", // Worldwide
	// 	"XE", // Europe
	// })

	v.SetDefault("Music.DB.Driver", "sqlite3")
	v.SetDefault("Music.DB.Source", "music.db")
	v.SetDefault("Music.DB.LogMode", "false")

	v.SetDefault("TMDB.Key", "903a776b0638da68e9ade38ff538e1d3")
	v.SetDefault("TMDB.Language", "en-US")
	v.SetDefault("TMDB.FileTemplate.Text",
		"{{.Title}} ({{.Year}}){{if .Definition}} - {{.Definition}}{{end}}{{.Extension}}")

	v.SetDefault("Search.BleveDir", ".")

	v.SetDefault("Server.Listen", "127.0.0.1:3000")
	v.SetDefault("Server.URL", "https://example.com") // w/o trailing slash

	v.SetDefault("Video.DB.Driver", "sqlite3")
	v.SetDefault("Video.DB.Source", "video.db")
	v.SetDefault("Video.DB.LogMode", "true")
	v.SetDefault("Video.ReleaseCountries", []string{
		"US",
	})
	v.SetDefault("Video.CastLimit", "25")
	v.SetDefault("Video.CrewJobs", []string{
		"Director",
		"Executive Producer",
		"Novel",
		"Producer",
		"Screenplay",
		"Story",
	})
	v.SetDefault("Video.Recent", "8760h") // 1 year
	v.SetDefault("Video.RecentLimit", "50")
	v.SetDefault("Video.SearchLimit", "100")
	v.SetDefault("Video.SyncInterval", "1h")
	v.SetDefault("Video.Recommend.When", []DateRecommend{
		// day of week + day of month
		{Match: "Fri 13", Layout: "Mon 02", Name: "Friday 13th Movies", Query: `+character:voorhees`},
		// day of month
		{Match: "Jan 03", Layout: "Jan 02", Name: "Tolkien Movies", Query: `+writing:tolkien`},
		{Match: "Feb 02", Layout: "Jan 02", Name: "Groundhog Day Movies", Query: `+keyword:groundhog`},
		{Match: "Feb 14", Layout: "Jan 02", Name: "Valentine's Day Movies", Query: `+genre:Romance`},
		{Match: "Mar 02", Layout: "Jan 02", Name: "Dr. Seuss Movies", Query: `+writing:seuss`},
		{Match: "Mar 12", Layout: "Jan 02", Name: "Hitchcock Movies", Query: `+directing:hitchcock`},
		{Match: "Mar 17", Layout: "Jan 02", Name: "St. Patrick's Day Movies", Query: `+keyword:leprechaun`},
		{Match: "Mar 27", Layout: "Jan 02", Name: "Tarantino Movies", Query: `+directing:tarantino`},
		{Match: "Apr 01", Layout: "Jan 02", Name: "April Fool's Movies", Query: `+keyword:"april fool's day"`},
		{Match: "Apr 28", Layout: "Jan 02", Name: "Superhero Movies", Query: `+keyword:superhero`},
		{Match: "May 02", Layout: "Jan 02", Name: "Harry Potter Movies", Query: `+title:"harry potter"`},
		{Match: "May 04", Layout: "Jan 02", Name: "Star Wars Movies", Query: `+title:"star wars"`},
		{Match: "May 11", Layout: "Jan 02", Name: "Twilight Zone Movies", Query: `+title:"twilight zone"`},
		{Match: "Jul 04", Layout: "Jan 02", Name: "July 4th Movies", Query: `+keyword:patriotism,patriotic,independence`},
		{Match: "Jul 04", Layout: "Jan 02", Name: "Alice in Wonderland",
			Query: `character:"Alice Kingsleigh" character:"Mad Hatter" character:"Red Queen"`},
		{Match: "Aug 11", Layout: "Jan 02", Name: "Spider-man Movies", Query: `+title:"spider-man"`},
		{Match: "Sep 17", Layout: "Jan 02", Name: "Batman Movies", Query: `+character:batman`},
		{Match: "Sep 22", Layout: "Jan 02", Name: "Hobbit Movies", Query: `+keyword:hobbit`},
		{Match: "Oct 21", Layout: "Jan 02", Name: "Back to the Future Movies", Query: `+title:"back to the future"`},
		{Match: "Dec 23", Layout: "Jan 02", Name: "It's Festivus", Query: `+keyword:festivus`},
		// months
		{Match: "Oct", Layout: "Jan", Name: "Halloween Movies", Query: `+keyword:halloween`},
		{Match: "Dec", Layout: "Jan", Name: "Christmas Movies", Query: `+keyword:christmas`},
	})

	// see https://musicbrainz.org/search (series)
	v.SetDefault("Music.RadioSeries", []string{
		"The Rolling Stone Magazine's 500 Greatest Songs of All Time",
	})

	v.SetDefault("Music.RadioOther", map[string]string{
		"Series Hits": "+series:*",
		"Top Hits":    "+popularity:1",
		"Top 3 Hits":  "+popularity:<4",
		"Top 5 Hits":  "+popularity:<6",
		"Top 10 Hits": "+popularity:<11",
		"Covers":      "+type:cover",
		"Live Hits":   "+type:live +popularity:<3",
	})

	v.SetDefault("Assistant.ProjectID", "undefined")
	v.SetDefault("Assistant.TrackLimit", "10")
	v.SetDefault("Assistant.RecentLimit", "3")
	v.SetDefault("Assistant.Welcome.Speech.Text", "Welcome to Takeout")
	v.SetDefault("Assistant.Welcome.Text.Text", "Welcome to Takeout")
	v.SetDefault("Assistant.Play.Speech.Text", "Enjoy the music")
	v.SetDefault("Assistant.Play.Text.Text", "")
	v.SetDefault("Assistant.Error.Speech.Text", "Please try again")
	v.SetDefault("Assistant.Error.Text.Text", "Please try again")
	v.SetDefault("Assistant.Link.Speech.Text", "Link this device to Takeout using code {{.Code}}")
	v.SetDefault("Assistant.Link.Text.Text", "Link code is: {{.Code}}")
	v.SetDefault("Assistant.Linked.Speech.Text", "Takeout is now linked")
	v.SetDefault("Assistant.Linked.Text.Text", "Takeout is now linked")
	v.SetDefault("Assistant.Guest.Speech.Text", "Guest not supported. A verified user is required.")
	v.SetDefault("Assistant.Guest.Text.Text", "Guest not supported. A verified user is required.")
	v.SetDefault("Assistant.Recent.Speech.Text", "Recently added albums are ")
	v.SetDefault("Assistant.Recent.Text.Text", "Recent Albums: ")
	v.SetDefault("Assistant.Release.Speech.Text", "{{.Name}} by {{.Artist}}")
	v.SetDefault("Assistant.Release.Text.Text", "{{.Artist}} \u2022 {{.Name}}")
	v.SetDefault("Assistant.SuggestionAuth", "Next")
	v.SetDefault("Assistant.SuggestionNew", "What's new")
	v.SetDefault("Assistant.MediaObjectName.Text", "{{.Title}}")
	v.SetDefault("Assistant.MediaObjectDesc.Text", "{{.Artist}} \u2022 {{.Release}}")

	v.SetDefault("Podcast.Client.MaxAge", "15m")
	v.SetDefault("Podcast.Client.UseCache", true)
	v.SetDefault("Podcast.DB.Driver", "sqlite3")
	v.SetDefault("Podcast.DB.Source", "podcast.db")
	v.SetDefault("Podcast.DB.LogMode", "true")
	v.SetDefault("Podcast.RecentLimit", "25")
	v.SetDefault("Podcast.SyncInterval", "1h")
	v.SetDefault("Podcast.Series", []string{
		"https://feeds.twit.tv/twit.xml",
		"https://feeds.twit.tv/sn.xml",
		"https://www.pbs.org/newshour/feeds/rss/podcasts/show",
		"http://feeds.feedburner.com/TEDTalks_audio",
	})

	v.SetDefault("Progress.DB.Driver", "sqlite3")
	v.SetDefault("Progress.DB.Source", "progress.db")
	v.SetDefault("Progress.DB.LogMode", "true")
}

func userAgent() string {
	return takeout.AppName + "/" + takeout.Version + " ( " + takeout.Contact + " ) "
}

func readConfig(v *viper.Viper) (*Config, error) {
	var config Config
	var pathRegexp = regexp.MustCompile(`(file|dir|source)$`)
	err := v.ReadInConfig()
	dir := filepath.Dir(v.ConfigFileUsed())
	for _, k := range v.AllKeys() {
		if pathRegexp.MatchString(k) {
			val := v.Get(k)
			if strings.HasPrefix(val.(string), "/") == false {
				val = fmt.Sprintf("%s/%s", dir, val.(string))
				v.Set(k, val)
			}
		}
	}
	if err == nil {
		err = v.Unmarshal(&config)
		config.Music.readMaps()
	}
	return &config, err
}

func TestConfig() (*Config, error) {
	testDir := os.Getenv("TEST_CONFIG")
	if testDir == "" {
		return nil, errors.New("missing test config")
	}
	v := viper.New()
	configDefaults(v)
	v.SetConfigFile(filepath.Join(testDir, "test.yaml"))
	v.SetDefault("Music.DB.Source", filepath.Join(testDir, "music.db"))
	v.SetDefault("Auth.DB.Source", filepath.Join(testDir, "auth.db"))
	return readConfig(v)
}

var configFile, configPath, configName string

func SetConfigFile(path string) {
	configFile = path
}

func AddConfigPath(path string) {
	configPath = path
}

func SetConfigName(name string) {
	configName = name
}

func GetConfig() (*Config, error) {
	v := viper.New()
	if configFile != "" {
		v.SetConfigFile(configFile)
	}
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	if configName != "" {
		v.SetConfigName(configName)
	}
	configDefaults(v)
	return readConfig(v)
}

func LoadConfig(dir string) (*Config, error) {
	v := viper.New()
	v.AddConfigPath(dir)
	configDefaults(v)
	return readConfig(v)
}
