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
	var logger = log.New("Update:Config:Defaults", log.DEFAULT_LOG_LEVEL)

	if config.ContentPath == "" {
		config.ContentPath = "."
		logger.Infof("Defaulting content path dir to '%s'\n", config.ContentPath)
	}
	if config.TemplateDir == "" {
		config.TemplateDir = "/template"
		logger.Infof("Defaulting template dir to '%s'\n", config.TemplateDir)
	}
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
	}
	if config.WatchlistDir == "" {
		config.WatchlistDir = filepath.Join(config.ContentPath, "web", "watchlists")
		logger.Infof("Defaulting watchlist dir to '%s'\n", config.WatchlistDir)
	}

	if config.BatchSize <= 0 {
		config.BatchSize = 1
		logger.Infof("Defaulting batch size '%d\n", config.BatchSize)
	}

	if config.RunIntervalSeconds <= 0 {
		config.RunIntervalSeconds = 10
		logger.Infof("Defaulting run interval '%d\n", config.RunIntervalSeconds)
	}

	return config
}

//Config for update app
type Config struct {
	//ContentPath all config paths should be relative to the content path.
	ContentPath        string `json:"contentPath"`
	TemplateDir        string `json:"templateDir"`
	DataFileName       string `json:"dataFileName"`
	WatchlistDir       string `json:"watchlistDir"`
	BatchSize          uint64 `json:"batchSize"`
	RunIntervalSeconds uint64 `json:"runIntervalSeconds"`
}
