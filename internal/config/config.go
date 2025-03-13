package config

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Mappings []Mapping `json:"mappings"`
}

type Mapping struct {
	Mapping  string `json:"mapping"`
	Verb     string `json:"verb"`
	RespCode int    `json:"code"`
	Content  any    `json:"content"`
}

func ParseConfiguration(filePath string) (*Configuration, error) {
	file, err := readFile(filePath)
	if err != nil {
		return nil, err
	}

	var value Configuration
	err = json.Unmarshal(file, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func readFile(file string) ([]byte, error) {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}
