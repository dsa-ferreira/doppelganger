package server

import (
	"errors"
	"fmt"
	"os"

	"github.com/dsa-ferreira/doppelganger/internal/config"
	"github.com/gin-gonic/gin"
)

type mappers func(*gin.Engine, config.Mapping)

func StartServer(configuration *config.Configuration) {
	r := gin.Default()

	for _, mapping := range configuration.Mappings {
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
	case "PUT":
		return putMap, nil
	case "DELETE":
		return deleteMap, nil
	}
	return nil, errors.New("No verb match found for verb " + verb)
}

func getMap(router *gin.Engine, config config.Mapping) {
	router.GET(config.Mapping, func(c *gin.Context) {
		c.JSON(config.RespCode, config.Content)
	})
}

func postMap(router *gin.Engine, config config.Mapping) {
	router.POST(config.Mapping, func(c *gin.Context) {
		c.JSON(config.RespCode, config.Content)
	})
}

func putMap(router *gin.Engine, config config.Mapping) {
	router.PUT(config.Mapping, func(c *gin.Context) {
		c.JSON(config.RespCode, config.Content)
	})
}

func deleteMap(router *gin.Engine, config config.Mapping) {
	router.DELETE(config.Mapping, func(c *gin.Context) {
		c.JSON(config.RespCode, config.Content)
	})
}
