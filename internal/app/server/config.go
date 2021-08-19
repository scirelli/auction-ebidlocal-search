package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//Load a config file.
func Load(fileName string) (*Config, error) {
	var config Config

	jsonFile, err := os.Open(fileName)
	if err != nil {
		return &config, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return &config, err
	}

	json.Unmarshal(byteValue, &config)

	Defaults(&config)

	return &config, nil
}

func Defaults(config *Config) *Config {
	var logger = log.New("ServerConfig", config.LogLevel)

	if config.ContentPath == "" {
		config.ContentPath = "."
	}
	if config.UserDir == "" {
		config.UserDir = filepath.Join(config.ContentPath, "web", "user")
		logger.Infof("Defaulting UserDir to '%s'\n", config.UserDir)
	}
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
		logger.Infof("Defaulting DataFileName to '%s'\n", config.DataFileName)
	}
	if config.WatchlistDir == "" {
		config.WatchlistDir = filepath.Join(config.ContentPath, "web", "watchlists")
		logger.Infof("Defaulting watchlist dir to '%s'\n", config.WatchlistDir)
	}

	return config
}

type Config struct {
	Port    uint   `json:"port"`
	Address string `json:"address"`

	ContentPath  string `json:"contentPath"`
	UserDir      string `json:"userDir"`
	DataFileName string `json:"dataFileName"`
	WatchlistDir string `json:"watchlistDir"`

	Debug    bool         `json:"debug"`
	LogLevel log.LogLevel `json:"logLevel"`
}
