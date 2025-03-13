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

func (mapping *Mapping) UnmarshalJSON(data []byte) error {
	type Alias Mapping
	type Aux struct {
		RespCode *int    `json:"append"` // We override the type of Append
		Verb     *string `json:"verb"`
		*Alias
	}
	aux := &Aux{Alias: (*Alias)(mapping)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.RespCode == nil {
		// Field "append" is not set: we want the default value to be true.
		mapping.RespCode = 204
	} else {
		// Field "append" is set: dereference and assign the value.
		mapping.RespCode = *aux.RespCode
	}

	if aux.Verb == nil {
		mapping.Verb = "GET"
	} else {
		mapping.Verb = *aux.Verb
	}

	return nil
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
