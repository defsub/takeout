package config

import (
	"github.com/spf13/viper"
)

type MusicBucket struct {
	Endpoint        string
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
	Bucket MusicBucket
	DB     MusicDB
}

type LastFMAPIConfig struct {
	Key    string
	Secret string
}

type Config struct {
	Music       MusicConfig
	LastFM      LastFMAPIConfig
	BindAddress string
}

func GetConfig() (*Config, error) {
	var config Config
	viper.SetConfigName("takeout")
	viper.AddConfigPath(".")
	viper.SetDefault("Music.Bucket.UseSSL", "true")
	viper.SetDefault("Music.DB.Driver", "sqlite3")
	viper.SetDefault("Music.DB.Source", "music.db")
	viper.SetDefault("Music.DB.LogMode", "false")
	viper.SetDefault("LastFM.Key", "77033164cfcda2acc4c58681dcba3cf8")
	viper.SetDefault("LastFM.Secret", "8f43410e8e81c33d4542738ee84dc39b")
	viper.SetDefault("BindAddress", "127.0.0.1:3000")
	err := viper.ReadInConfig()
	if err == nil {
		err = viper.Unmarshal(&config)
	}
	return &config, err
}
