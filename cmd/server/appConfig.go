package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/ebidlocal"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server"
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

	return &config, nil
}

//AppConfig configuration data for entire application.
type AppConfig struct {
	Server    server.Config    `json:"server"`
	Ebidlocal ebidlocal.Config `json:"ebidlocal"`
}
