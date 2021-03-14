package server

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
	var logger = log.New("ServerConfig")

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

	if config.ContentPath == "" {
		config.ContentPath = "."
	}
	if config.UserDir == "" {
		config.UserDir = filepath.Join(config.ContentPath, "web", "user")
		logger.Info.Printf("Defaulting UserDir to '%s'\n", config.UserDir)
	}
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
		logger.Info.Printf("Defaulting DataFileName to '%s'\n", config.DataFileName)
	}

	return &config, nil
}

type Config struct {
	Port    uint   `json:"port"`
	Address string `json:"address"`

	ContentPath  string `json:"contentPath"`
	UserDir      string `json:"userDir"`
	DataFileName string `json:"dataFileName"`
}
