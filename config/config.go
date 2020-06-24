// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// Takeout is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"encoding/json"
	"errors"
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
}

type MusicDB struct {
	Driver  string
	Source  string
	LogMode bool
}

type MusicConfig struct {
	Bucket      MusicBucket
	DB          MusicDB
	ArtistFile  string
	ReleaseFile string
	artistMap   map[string]string
	releaseMap  map[string]string
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
	DB     AuthDB
	MaxAge time.Duration
}

type Config struct {
	Auth        AuthConfig
	Music       MusicConfig
	LastFM      LastFMAPIConfig
	BindAddress string
}

func (mc *MusicConfig) UserArtistID(name string) (string, bool) {
	mbid, ok := mc.artistMap[name]
	return mbid, ok
}

func (mc *MusicConfig) UserReleaseID(name string) (string, bool) {
	mbid, ok := mc.releaseMap[name]
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
	if mc.ReleaseFile != "" {
		readJsonStringMap(mc.ReleaseFile, &mc.releaseMap)
	}
}

func configDefaults() {
	viper.SetDefault("Auth.MaxAge", "24h")
	viper.SetDefault("Auth.DB.Driver", "sqlite3")
	viper.SetDefault("Auth.DB.Source", "auth.db")
	viper.SetDefault("Auth.DB.LogMode", "false")

	viper.SetDefault("Music.Bucket.UseSSL", "true")
	viper.SetDefault("Music.DB.Driver", "sqlite3")
	viper.SetDefault("Music.DB.Source", "music.db")
	viper.SetDefault("Music.DB.LogMode", "false")

	viper.SetDefault("LastFM.Key", "77033164cfcda2acc4c58681dcba3cf8")
	viper.SetDefault("LastFM.Secret", "8f43410e8e81c33d4542738ee84dc39b")

	viper.SetDefault("BindAddress", "127.0.0.1:3000")
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

func GetConfig() (*Config, error) {
	viper.SetConfigName("takeout")
	viper.AddConfigPath(".")
	configDefaults()
	return readConfig()
}
