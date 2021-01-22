package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
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

	return &config, nil
}

type Config struct {
	Port    uint   `json:"port"`
	Address string `json:"address"`

	ContentPath  string `json:"contentPath"`
	UserDir      string `json:"userDir"`
	DataFileName string `json:"dataFileName"`
}
