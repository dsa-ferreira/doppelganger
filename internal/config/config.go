package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/dsa-ferreira/doppelganger/internal/expressions"
)

type Servers struct {
	Configurations []Configuration `json:"servers"`
}

func (servers *Servers) UnmarshalJSON(data []byte) error {
	type Alias Servers
	type Aux struct {
		*Alias
	}

	aux := &Aux{Alias: (*Alias)(servers)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(servers.Configurations) == 0 {
		return errors.New("No server found")
	}

	return nil
}

type Configuration struct {
	Endpoints []Endpoint `json:"endpoint"`
	Port      int        `json:"port"`
}

func (configuration *Configuration) UnmarshalJSON(data []byte) error {
	type Alias Configuration
	type Aux struct {
		Port *int `json:"port"`
		*Alias
	}

	aux := &Aux{Alias: (*Alias)(configuration)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Port == nil {
		configuration.Port = 8000
	} else {
		configuration.Port = *aux.Port
	}

	return nil
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
	Params   []expressions.Expression `json:"params"`
	RespCode int                      `json:"code"`
	Content  any                      `json:"content"`
}

func (mapping *Mapping) UnmarshalJSON(data []byte) error {
	type Alias Mapping
	type Aux struct {
		Params   []json.RawMessage `json:"params"`
		RespCode *int              `json:"code"`
		*Alias
	}
	aux := &Aux{Alias: (*Alias)(mapping)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	mapping.Params = make([]expressions.Expression, len(aux.Params))
	for i, v := range aux.Params {
		result, err := expressions.BuildExpression([]byte(v))
		if err != nil {
			panic("error building param n: " + strconv.Itoa(i))
		}

		mapping.Params[i] = result
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

func ParseConfiguration(filePath string) (*Servers, error) {
	file, err := readFile(filePath)
	if err != nil {
		return nil, err
	}

	var value Servers
	err = json.Unmarshal(file, &value)
	if err != nil {
		var fallback Configuration
		err = json.Unmarshal(file, &fallback)
		if err != nil {
			return nil, err
		}

		return &Servers{Configurations: []Configuration{fallback}}, nil
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
