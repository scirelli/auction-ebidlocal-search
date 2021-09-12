package extract

import (
	"encoding/json"
	"io/ioutil"
	"os"

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
	if config == nil {
		config = &Config{}
	}

	if config.LogLevel == 0 {
		config.LogLevel = log.DEFAULT_LOG_LEVEL
	}
	return config
}

//Config for scanner app
type Config struct {
	LogLevel log.LogLevel `json:"logLevel"`
}
