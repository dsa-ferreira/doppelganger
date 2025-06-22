package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dsa-ferreira/doppelganger/internal/config"
	"github.com/dsa-ferreira/doppelganger/internal/server"
)

func main() {
	verbose := flag.Bool("verbose", false, "increase verbosity")

	flag.Parse()

	configFile := flag.Args()[0]
	servers, err := config.ParseConfiguration(configFile)
	if err != nil {
		fmt.Printf("Error parsing configuration: %s\n", err)
		os.Exit(2)
	}

	for i := 0; i < len(servers.Configurations); i++ {
		go server.StartServer(&servers.Configurations[i], *verbose)
	}

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	<-gracefulShutdown

	fmt.Printf("Shuting down")
}
