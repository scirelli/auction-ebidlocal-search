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
	if config.TemplateFile == "" {
		config.TemplateFile = filepath.Join(config.ContentPath, "assets", "templates", "template.html.tmpl")
		logger.Infof("Defaulting template dir to '%s'\n", config.TemplateFile)
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
	if config.ServerUrl == "" {
		config.ServerUrl = "http://localhost:8282"
		logger.Infof("Defaulting ServerUrl to '%s'\n", config.ServerUrl)
	}

	return config
}

//Config for update app
type Config struct {
	//ContentPath all config paths should be relative to the content path.
	ContentPath  string `json:"contentPath"`
	TemplateFile string `json:"templateFile"`
	DataFileName string `json:"dataFileName"`
	WatchlistDir string `json:"watchlistDir"`
	BatchSize    uint64 `json:"batchSize"`
	ServerUrl    string `json:"serverUrl"`

	Debug    bool         `json:"debug"`
	LogLevel log.LogLevel `json:"logLevel"`
}
