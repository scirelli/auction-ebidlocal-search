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
	var logger = log.New("Config.Load")

	if config.ContentPath == "" {
		config.ContentPath = "."
		logger.Info.Printf("Defaulting content path dir to '%s'\n", config.ContentPath)
	}
	if config.TemplateDir == "" {
		config.TemplateDir = "/template"
		logger.Info.Printf("Defaulting template dir to '%s'\n", config.TemplateDir)
	}
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
	}
	if config.WatchlistDir == "" {
		config.WatchlistDir = filepath.Join(config.ContentPath, "web", "watchlists")
		logger.Info.Printf("Defaulting watchlist dir to '%s'\n", config.WatchlistDir)
	}
	if config.ScanInterval == 0 {
		config.ScanInterval = 1
		logger.Info.Printf("Defaulting scan interval to '%d'\n", config.ScanInterval)
	}

	return config
}

//Config for scanner app
type Config struct {
	//ContentPath all config paths should be relative to the content path.
	ContentPath   string `json:"contentPath"`
	TemplateDir   string `json:"templateDir"`
	DataFileName  string `json:"dataFileName"`
	WatchlistDir  string `json:"watchlistDir"`
	ScanInterval  int64  `json:"scanIntervalSeconds"`
	AsyncRequests int64  `json:"asyncRequests"`
}
