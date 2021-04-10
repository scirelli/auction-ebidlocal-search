package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/scanner"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/update"
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

	scanner.Defaults(&config.Scanner)
	update.Defaults(&config.Updater)

	return &config, nil
}

//AppConfig configuration data for entire application.
type AppConfig struct {
	Scanner scanner.Config `json:"scanner"`
	Updater update.Config  `json:"updater"`
}
