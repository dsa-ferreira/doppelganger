package config

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Endpoints []Endpoint `json:"endpoint"`
}

type Endpoint struct {
	Path     string    `json:"path"`
	Verb     string    `json:"verb"`
	Mappings []Mapping `json:"mappings"`
}

func (endpoint *Endpoint) UnmarshalJSON(data []byte) error {
	type Alias Endpoint
	type Aux struct {
		Verb *string `json:"verb"`
		*Alias
	}

	aux := &Aux{Alias: (*Alias)(endpoint)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Verb == nil {
		endpoint.Verb = "GET"
	} else {
		endpoint.Verb = *aux.Verb
	}

	return nil
}

type Mapping struct {
	Params   []Param `json:"params"`
	RespCode int     `json:"code"`
	Content  any     `json:"content"`
}

func (mapping *Mapping) UnmarshalJSON(data []byte) error {
	type Alias Mapping
	type Aux struct {
		RespCode *int `json:"code"`
		*Alias
	}
	aux := &Aux{Alias: (*Alias)(mapping)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.RespCode == nil {
		if aux.Content == nil {
			mapping.RespCode = 204
		} else {
			mapping.RespCode = 200
		}
	} else {
		mapping.RespCode = *aux.RespCode
	}

	return nil
}

type Param struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
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
