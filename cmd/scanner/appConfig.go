package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/notify"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/scanner"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/update"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//LoadConfig a config file.
func LoadConfig(fileName string) (*AppConfig, error) {
	var config AppConfig

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

	if config.LogLevel == "" {
		config.Scanner.LogLevel = log.DEFAULT_LOG_LEVEL
		config.Updater.LogLevel = log.DEFAULT_LOG_LEVEL
		config.Notifier.LogLevel = log.DEFAULT_LOG_LEVEL
	} else {
		config.Scanner.LogLevel = log.GetLevel(config.LogLevel)
		config.Updater.LogLevel = log.GetLevel(config.LogLevel)
		config.Notifier.LogLevel = log.GetLevel(config.LogLevel)
	}

	scanner.Defaults(&config.Scanner)
	update.Defaults(&config.Updater)
	notify.DefaultConfig(&config.Notifier)

	return &config, nil
}

//AppConfig configuration data for entire application.
type AppConfig struct {
	Debug    bool   `json:"debug"`
	LogLevel string `json:"logLevel"`

	Scanner  scanner.Config `json:"scanner"`
	Updater  update.Config  `json:"updater"`
	Notifier notify.Config  `json:"notifier"`
}
