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
	"io/ioutil"
	"os"
	"path/filepath"
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
	DB                   DatabaseConfig
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
	RadioSeries          []string
	RadioOther           map[string]string
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

type FanartAPIConfig struct {
	ProjectKey  string
	PersonalKey string
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
	UseCache  bool
	MaxAge    int
	CacheDir  string
	UserAgent string
}

type Config struct {
	Auth    AuthConfig
	Bucket  BucketConfig
	Client  ClientConfig
	DataDir string
	Fanart  FanartAPIConfig
	LastFM  LastFMAPIConfig
	Music   MusicConfig
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
	v.SetDefault("Music.ReleaseContries", []string{
		"US", // United States
		"XW", // Worldwide
		"XE", // Europe
	})

	v.SetDefault("Music.DB.Driver", "sqlite3")
	v.SetDefault("Music.DB.Source", "music.db")
	v.SetDefault("Music.DB.LogMode", "false")

	v.SetDefault("Search.BleveDir", ".")

	v.SetDefault("Server.WebDir", "web")
	v.SetDefault("Server.URL", "https://example.com") // w/o trailing slash

	// see https://musicbrainz.org/search (series)
	v.SetDefault("Music.RadioSeries", []string{
		"Billboard Year-End Hot 100 singles of 2019",
		"Billboard Year-End Hot 100 singles of 2020",
		"Indie 88: Top 500 Indie Rock Songs",
		`NME: Greatest "Indie" Anthems Ever: 2007`,
		"Rolling Stone: The 100 Best Songs of the 2010s",
		"The Rolling Stone Magazine's 500 Greatest Songs of All Time",
		"Stereogum: The 200 Best Songs Of The 2010s",
	})

	v.SetDefault("Music.RadioOther", map[string]string{
		"Series Hits":            "+series:*",
		"Top Hits":               "+popularity:1",
		"Top 3 Hits":             "+popularity:<4",
		"Top 5 Hits":             "+popularity:<6",
		"Top 10 Hits":            "+popularity:<11",
		"Epic 10+ Minute Tracks": "+length:>600 -silence",
		"Epic 20+ Minute Tracks": "+length:>1200 -silence",
		"Deep Tracks":            "+track:>10 -silence",
		"4AD: Hits":              "+label:4ad +popularity:<4",
		"Def Jam: Hits":          `+label:"def jam" +popularity:<4`,
		"Sub Pop Records: Hits":  `+label:"sub pop" +popularity:<4`,
		"Beggars Banquet: Hits":  `+label:"beggars banquet" +popularity:<4`,
		"Hits with Violin":       `+violin:* +popularity:<4`,
	})

}

func userAgent() string {
	return takeout.AppName + "/" + takeout.Version + " ( " + takeout.Contact + " ) "
}

func readConfig(v *viper.Viper) (*Config, error) {
	var config Config
	err := v.ReadInConfig()
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
	v.SetConfigFile(filepath.Join(testDir, "test.ini"))
	v.SetDefault("Music.DB.Source", filepath.Join(testDir, "music.db"))
	v.SetDefault("Auth.DB.Source", filepath.Join(testDir, "auth.db"))
	return readConfig(v)
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
	v := viper.New()
	configDefaults(v)
	return readConfig(v)
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	configDefaults(v)
	v.SetConfigFile(path)
	return readConfig(v)
}
