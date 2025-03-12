package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dsa-ferreira/doppelganger/internal/config"
	"github.com/gin-gonic/gin"
)

type mappers func(*gin.Engine, config.Config)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Can't do this without a config file")
		os.Exit(1)
	}
	configFile := os.Args[1]
	mappings := config.ParseConfiguration(configFile)

	r := gin.Default()

	for _, mapping := range mappings {
		mapper, err := selectMap(mapping.Verb)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		mapper(r, mapping)
	}

	r.Run()
}

func selectMap(verb string) (mappers, error) {
	switch verb {
	case "GET":
		return getMap, nil
	case "POST":
		return postMap, nil
	}
	return nil, errors.New("No verb match found for verb " + verb)
}

func getMap(router *gin.Engine, config config.Config) {
	router.GET(config.Mapping, func(c *gin.Context) {
		c.JSON(200, config.Content)
	})
}

func postMap(router *gin.Engine, config config.Config) {
	router.POST(config.Mapping, func(c *gin.Context) {
		c.JSON(200, config.Content)
	})
}
