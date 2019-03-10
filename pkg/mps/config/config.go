package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ResourceManifestURL string `json:"resourceManifestURL"`
	TokenP1             string `json:"tokenP1"`
}

func FromJSONFile(file string) (cnf Config, err error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	if err = json.Unmarshal(raw, &cnf); err != nil {
		return
	}
	return
}

func (config Config) IsValid() bool {
	return config.ResourceManifestURL != "" && config.TokenP1 != ""
}
