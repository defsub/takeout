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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/defsub/takeout"
	"github.com/spf13/viper"
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
}

type DatabaseConfig struct {
	Driver  string
	Source  string
	LogMode bool
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
	Recent               time.Duration
	RecentLimit          int
	ReleaseCountries     []string
	ReleaseFile          string
	SearchLimit          int
	SimilarArtistsLimit  int
	SimilarReleases      time.Duration
	SimilarReleasesLimit int
	SinglesLimit         int
	artistMap            map[string]string
}

type MovieConfig struct {
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
	Key string
}

type AuthConfig struct {
	DB            DatabaseConfig
	MaxAge        time.Duration
	SecureCookies bool
}

type SearchConfig struct {
	BleveDir string
}

type ServerConfig struct {
	Listen string
	WebDir string
	URL    string
}

type ClientConfig struct {
	CacheDir  string
	MaxAge    int
	UseCache  bool
	UserAgent string
}

type MediaConfig struct {
	MovieTemplate  string
	PosterTemplate string
}

type Config struct {
	Auth    AuthConfig
	Bucket  BucketConfig
	Client  ClientConfig
	DataDir string
	Fanart  FanartAPIConfig
	LastFM  LastFMAPIConfig
	Media   MediaConfig
	Music   MusicConfig
	TMDB    TMDBAPIConfig
	Movie   MovieConfig
	Search  SearchConfig
	Server  ServerConfig
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
	v.SetDefault("Auth.SecureCookies", "true")

	v.SetDefault("Bucket.URLExpiration", "72h")
	v.SetDefault("Bucket.UseSSL", "true")

	v.SetDefault("Client.CacheDir", ".httpcache")
	v.SetDefault("Client.MaxAge", 86400*30) // 30 days in seconds
	v.SetDefault("Client.UseCache", "false")
	v.SetDefault("Client.UserAgent", userAgent())

	v.SetDefault("DataDir", ".")

	v.SetDefault("Fanart.ProjectKey", "93ede276ba6208318031727060b697c8")

	v.SetDefault("LastFM.Key", "77033164cfcda2acc4c58681dcba3cf8")
	v.SetDefault("LastFM.Secret", "8f43410e8e81c33d4542738ee84dc39b")

	v.SetDefault("Music.ArtistRadioBreadth", "10")
	v.SetDefault("Music.ArtistRadioDepth", "10")
	v.SetDefault("Music.DeepLimit", "50")
	v.SetDefault("Music.PopularLimit", "50")
	v.SetDefault("Music.RadioLimit", "25")
	v.SetDefault("Music.RadioSearchLimit", "1000")
	v.SetDefault("Music.Recent", "8760h") // 1 year
	v.SetDefault("Music.RecentLimit", "50")
	v.SetDefault("Music.SearchLimit", "100")
	v.SetDefault("Music.SimilarArtistsLimit", "10")
	v.SetDefault("Music.SimilarReleases", "8760h") // +/- 1 year
	v.SetDefault("Music.SimilarReleasesLimit", "10")
	v.SetDefault("Music.SinglesLimit", "50")

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

	v.SetDefault("Search.BleveDir", ".")

	v.SetDefault("Server.Listen", "127.0.0.1:3000")
	v.SetDefault("Server.WebDir", "web")
	v.SetDefault("Server.URL", "https://example.com") // w/o trailing slash

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
	})

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
