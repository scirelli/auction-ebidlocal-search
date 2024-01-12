package server

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"time"

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
	if config.VerificationTemplateFile == "" {
		config.VerificationTemplateFile = filepath.Join(config.ContentPath, "assets", "templates", "verification.template.html.tmpl")
		logger.Infof("Defaulting verification template dir to '%s'\n", config.VerificationTemplateFile)
	}
	if config.VerificationWindowMinutes == 0 {
		config.VerificationWindowMinutes = 1
		logger.Infof("Defaulting verification window to '%d' minute(s)\n", config.VerificationWindowMinutes)
	}
	if config.ServerUrl == "" {
		config.ServerUrl = "http://localhost:8282"
		logger.Infof("Defaulting ServerUrl to '%s'\n", config.ServerUrl)
	} else {
		base, err := url.Parse(config.ServerUrl)
		if err != nil {
			logger.Panic(err)
		}
		config.ServerUrl = base.String()
	}
	if config.UiUrl == "" {
		config.UiUrl = "http://localhost"
		logger.Infof("Defaulting UiUrl to '%s'\n", config.UiUrl)
	}
	if config.SearchVersion == "" {
		config.SearchVersion = "v1"
		logger.Infof("Defaulting SearchVersion to '%s'\n", config.SearchVersion)
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

	VerificationTemplateFile  string        `json:"verificationTemplateFile"`
	VerificationWindowMinutes time.Duration `json:"verificationWindowMinutes"`
	ServerUrl                 string        `json:"serverUrl"`
	UiUrl                     string        `json:"uiUrl"`

    SearchVersion string `json:"searchVersion"`

	Debug    bool         `json:"debug"`
	LogLevel log.LogLevel `json:"logLevel"`
}
