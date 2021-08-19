package scanner

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
	var logger = log.New("Scanner:Config:Defaults", config.LogLevel)

	if config.ContentPath == "" {
		config.ContentPath = "."
		logger.Infof("Defaulting content path dir to '%s'\n", config.ContentPath)
	}
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
	}
	if config.WatchlistDir == "" {
		config.WatchlistDir = filepath.Join(config.ContentPath, "web", "watchlists")
		logger.Infof("Defaulting watchlist dir to '%s'\n", config.WatchlistDir)
	}
	if config.ScanInterval == 0 {
		config.ScanInterval = 10
		logger.Infof("Defaulting scan interval to '%d'\n", config.ScanInterval)
	}

	return config
}

//Config for scanner app
type Config struct {
	//ContentPath all config paths should be relative to the content path.
	ContentPath  string `json:"contentPath"`
	DataFileName string `json:"dataFileName"`
	WatchlistDir string `json:"watchlistDir"`
	ScanInterval int64  `json:"scanIntervalSeconds"`

	Debug    bool         `json:"debug"`
	LogLevel log.LogLevel `json:"logLevel"`
}
