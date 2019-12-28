package main

import (
	"encoding/json"
	"os"
)

type JSONConfig struct {
	ConfigPath   string
	ManifestPath string   `json:"manifest"`
	Project      string   `json:"project"`
	Port         int      `json:"port"`
	Labels       LabelMap `json:"labels"`
}

func ParseJSONConfigFromPath(path string) (JSONConfig, error) {
	out := JSONConfig{ConfigPath: path}

	f, err := os.Open(path)
	if err != nil {
		return out, err
	}

	err = json.NewDecoder(f).Decode(&out)

	return out, err
}
