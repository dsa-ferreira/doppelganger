package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dsa-ferreira/doppelganger/internal/config"
	"github.com/dsa-ferreira/doppelganger/internal/server"
)

func main() {

	verbose := flag.Bool("verbose", false, "increase verbosity")

	flag.Parse()

	configFile := flag.Args()[0]
	mappings, err := config.ParseConfiguration(configFile)
	if err != nil {
		fmt.Printf("Error parsing configuration: %s\n", err)
		os.Exit(2)
	}

	server.StartServer(mappings, *verbose)
}
