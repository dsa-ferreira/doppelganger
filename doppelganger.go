package main

import (
	"fmt"
	"os"

	"github.com/dsa-ferreira/doppelganger/internal/config"
	"github.com/dsa-ferreira/doppelganger/internal/server"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Can't do this without a config file")
		os.Exit(1)
	}
	configFile := os.Args[1]
	mappings, err := config.ParseConfiguration(configFile)
	if err != nil {
		fmt.Printf("Error parsing configuration: %s\n", err)
		os.Exit(2)
	}

	server.StartServer(mappings)
}
