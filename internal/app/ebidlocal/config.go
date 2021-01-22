package ebidlocal

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

//Config for ebidlocal app
type Config struct {
	//ContentPath all config paths should be relative to the content path.
	ContentPath   string `json:"contentPath"`
	TemplateDir   string `json:"templateDir"`
	DataFileName  string `json:"dataFileName"`
	WatchlistDir  string `json:"watchlistDir"`
	ScanInterval  int64  `json:"scanIntervalSeconds"`
	AsyncRequests int64  `json:"asyncRequests"`
}
