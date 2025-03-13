package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Test struct {
	Thing []Config `json:"mappings"`
}

type Config struct {
	Mapping string `json:"mapping"`
	Verb    string `json:"verb"`
	Content any    `json:"content"`
}

func ParseConfiguration(filePath string) []Config {
	var value Test
	err := json.Unmarshal(readFile(filePath), &value)
	if err != nil {
		fmt.Printf("Error parsing configuration: %s\n", err)
		os.Exit(2)
	}
	return value.Thing
}

func readFile(file string) []byte {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	return fileBytes
}
