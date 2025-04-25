package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/dsa-ferreira/doppelganger/internal/config"
	"github.com/gin-gonic/gin"
)

type mappers func(*gin.Engine, config.Endpoint)

func StartServer(configuration *config.Configuration) {
	r := gin.Default()

	for _, endpoint := range configuration.Endpoints {
		mapper, err := selectMap(endpoint.Verb)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		mapper(r, endpoint)
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

func getMap(router *gin.Engine, config config.Endpoint) {
	router.GET(config.Path, func(c *gin.Context) {
		mapReturns(c, nil, config.Mappings)
	})
}

func postMap(router *gin.Engine, config config.Endpoint) {
	router.POST(config.Path, func(c *gin.Context) {
		mapReturnsWithBody(c, config.Mappings)
	})
}

func putMap(router *gin.Engine, config config.Endpoint) {
	router.PUT(config.Path, func(c *gin.Context) {
		mapReturnsWithBody(c, config.Mappings)
	})
}

func deleteMap(router *gin.Engine, config config.Endpoint) {
	router.DELETE(config.Path, func(c *gin.Context) {
		mapReturnsWithBody(c, config.Mappings)
	})
}

func mapReturnsWithBody(c *gin.Context, mappings []config.Mapping) {
	var body map[string]interface{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mapReturns(c, body, mappings)
}

func mapReturns(c *gin.Context, body map[string]interface{}, mappings []config.Mapping) {
	for _, mapping := range mappings {
		if allMatch(c, body, mapping.Params) {
			c.JSON(mapping.RespCode, mapping.Content)
			return
		}
	}
}

func allMatch(c *gin.Context, body map[string]interface{}, params []config.Param) bool {
	for _, param := range params {
		var value string

		switch param.Type {
		case "BODY":
			aux := body[param.Key]
			if aux == nil {
				value = ""
			} else {
				value = fmt.Sprintf("%v", body[param.Key])
			}
		case "PATH":
			value = c.Param(param.Key)
		case "QUERY":
			value = c.Query(param.Key)
		default:
			return false
		}

		if value == "" {
			fmt.Println(fmt.Sprintf("WARNING: No parameter of type %s found for key %s", param.Type, param.Key))
			return false
		}
		if value != param.Value {
			return false
		}
	}

	return true
}
