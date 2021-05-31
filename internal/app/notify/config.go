package notify

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//Load a config file.
func LoadConfig(fileName string) (*Config, error) {
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
	DefaultConfig(&config)

	return &config, nil
}

func DefaultConfig(config *Config) *Config {
	var logger = log.New("Notifier:Config:Defaults", log.DEFAULT_LOG_LEVEL)

	if config == nil {
		config = &Config{}
	}

	if config.ContentPath == "" {
		config.ContentPath = "."
		logger.Infof("Defaulting content path dir to '%s'\n", config.ContentPath)
	}
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
	}
	if config.UserDir == "" {
		config.UserDir = filepath.Join(config.ContentPath, "web", "user")
		logger.Infof("Defaulting userDir dir to '%s'\n", config.UserDir)
	}
	if config.ServerUrl == "" {
		config.ServerUrl = "http://localhost:8282"
		logger.Infof("Defaulting ServerUrl to '%s'\n", config.ServerUrl)
	}
	if config.Type == "" {
		config.Type = "email"
		logger.Infof("Defaulting Type to '%s'\n", config.Type)
	}
	if config.WatchlistDir == "" {
		config.WatchlistDir = filepath.Join(config.ContentPath, "web", "watchlists")
		logger.Infof("Defaulting watchlist dir to '%s'\n", config.WatchlistDir)
	}

	return config
}

//Config for notifier
type Config struct {
	ContentPath  string `json:"contentPath"`
	DataFileName string `json:"dataFileName"`
	UserDir      string `json:"userDir"`
	ServerUrl    string `json:"serverUrl"`
	Type         string `json:"type"`
	WatchlistDir string `json:"watchlistDir"`
}
