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
	"github.com/defsub/takeout"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type MusicBucket struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	ObjectPrefix    string
	UseSSL          bool
	URLExpiration   time.Duration
}

type MusicDB struct {
	Driver  string
	Source  string
	LogMode bool
}

type MusicConfig struct {
	Bucket               MusicBucket
	DB                   MusicDB
	ArtistFile           string
	ReleaseFile          string
	artistMap            map[string]string
	Recent               time.Duration
	RecentLimit          int
	SearchLimit          int
	PopularLimit         int
	SinglesLimit         int
	DeepLimit            int
	SimilarArtistsLimit  int
	SimilarReleases      time.Duration
	SimilarReleasesLimit int
	RadioGenres          []string
	RadioLimit           int
	RadioSearchLimit     int
	ArtistRadioBreadth   int
	ArtistRadioDepth     int
	ReleaseCountries     []string
}

type LastFMAPIConfig struct {
	Key    string
	Secret string
}

type AuthDB struct {
	Driver  string
	Source  string
	LogMode bool
}

type AuthConfig struct {
	DB            AuthDB
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
	UseCache  bool
	MaxAge    int
	CacheDir  string
	UserAgent string
}

type Config struct {
	Auth   AuthConfig
	Music  MusicConfig
	LastFM LastFMAPIConfig
	Search SearchConfig
	Server ServerConfig
	Client ClientConfig
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

func configDefaults() {
	viper.SetDefault("Auth.MaxAge", "24h")
	viper.SetDefault("Auth.SecureCookies", "true")
	viper.SetDefault("Auth.DB.Driver", "sqlite3")
	viper.SetDefault("Auth.DB.Source", "auth.db")
	viper.SetDefault("Auth.DB.LogMode", "false")

	viper.SetDefault("Music.Recent", "8760h") // 1 year
	viper.SetDefault("Music.RecentLimit", "50")
	viper.SetDefault("Music.SearchLimit", "100")
	viper.SetDefault("Music.PopularLimit", "50")
	viper.SetDefault("Music.SinglesLimit", "50")
	viper.SetDefault("Music.DeepLimit", "50")
	viper.SetDefault("Music.SimilarArtistsLimit", "10")
	viper.SetDefault("Music.SimilarReleases", "8760h") // +/- 1 year
	viper.SetDefault("Music.SimilarReleasesLimit", "10")

	viper.SetDefault("Music.RadioLimit", "250")
	viper.SetDefault("Music.RadioSearchLimit", "1000")
	viper.SetDefault("Music.ArtistRadioBreadth", "25")
	viper.SetDefault("Music.ArtistRadioDepth", "10")

	// see https://wiki.musicbrainz.org/Release_Country
	viper.SetDefault("Music.ReleaseContries", []string{
		"US", // United States
		"XW", // Worldwide
		"XE", // Europe
	})

	// see https://musicbrainz.org/genres
	viper.SetDefault("Music.RadioGenres", []string{
		"alternative rock",
		"alternative",
		"ambient",
		"big beat",
		"blues rock",
		"blues",
		"classic rock",
		"contemporary r&b",
		"country rock",
		"dance-pop",
		"disco",
		"dream pop",
		"electronic",
		"gothic rock",
		"grunge",
		"hard rock",
		"heavy metal",
		"hip hop",
		"house",
		"indie pop",
		"indie rock",
		"indie",
		"krautrock",
		"latin",
		"metal",
		"new wave",
		"pop rock",
		"pop",
		"post-punk",
		"post-rock",
		"progressive rock",
		"r&b",
		"rock",
		"shoegaze",
	})

	viper.SetDefault("Music.Bucket.UseSSL", "true")
	viper.SetDefault("Music.Bucket.URLExpiration", "72h")

	viper.SetDefault("Music.DB.Driver", "sqlite3")
	viper.SetDefault("Music.DB.Source", "music.db")
	viper.SetDefault("Music.DB.LogMode", "false")

	viper.SetDefault("LastFM.Key", "77033164cfcda2acc4c58681dcba3cf8")
	viper.SetDefault("LastFM.Secret", "8f43410e8e81c33d4542738ee84dc39b")

	viper.SetDefault("Search.BleveDir", ".")

	viper.SetDefault("Server.WebDir", "web")
	viper.SetDefault("Server.URL", "https://example.com/")

	viper.SetDefault("Client.UseCache", "false")
	viper.SetDefault("Client.MaxAge", 86400*30) // 30 days in seconds
	viper.SetDefault("Client.CacheDir", ".httpcache")
	viper.SetDefault("Client.UserAgent", userAgent())
}

func userAgent() string {
	return takeout.AppName + "/" + takeout.Version + " ( " + takeout.Contact + " ) "
}

func readConfig() (*Config, error) {
	var config Config

	err := viper.ReadInConfig()
	if err == nil {
		err = viper.Unmarshal(&config)
	}
	config.Music.readMaps()
	return &config, err
}

func TestConfig() (*Config, error) {
	testDir := os.Getenv("TEST_CONFIG")
	if testDir == "" {
		return nil, errors.New("missing test config")
	}
	configDefaults()
	viper.SetConfigFile(filepath.Join(testDir, "test.ini"))
	viper.SetDefault("Music.DB.Source", filepath.Join(testDir, "music.db"))
	viper.SetDefault("Auth.DB.Source", filepath.Join(testDir, "auth.db"))
	return readConfig()
}

func SetConfigFile(path string) {
	viper.SetConfigFile(path)
}

func AddConfigPath(path string) {
	viper.AddConfigPath(path)
}

func SetConfigName(name string) {
	viper.SetConfigName(name)
}

func GetConfig() (*Config, error) {
	configDefaults()
	return readConfig()
}
