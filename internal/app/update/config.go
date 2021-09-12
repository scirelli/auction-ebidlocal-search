package update

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
	var logger = log.New("Update:Config:Defaults", config.LogLevel)

	if config.ContentPath == "" {
		config.ContentPath = "."
		logger.Infof("Defaulting content path dir to '%s'\n", config.ContentPath)
	}
	if config.WatchlistDir == "" {
		config.WatchlistDir = filepath.Join(config.ContentPath, "web", "watchlists")
		logger.Infof("Defaulting watchlist dir to '%s'\n", config.WatchlistDir)
	}

	return config
}

//Config for update app
type Config struct {
	//ContentPath all config paths should be relative to the content path.
	ContentPath  string `json:"contentPath"`
	WatchlistDir string `json:"watchlistDir"`

	Debug    bool         `json:"debug"`
	LogLevel log.LogLevel `json:"logLevel"`
}
