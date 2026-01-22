package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/dsa-ferreira/doppelganger/internal/config"
	"github.com/dsa-ferreira/doppelganger/internal/expressions"
	"github.com/gin-gonic/gin"
)

type mappers func(*gin.Engine, config.Endpoint)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		buf, _ := io.ReadAll(c.Request.Body)
		rdr1 := io.NopCloser(bytes.NewBuffer(buf))
		rdr2 := io.NopCloser(bytes.NewBuffer(buf))

		body := readBody(rdr1)
		if body != "" {
			fmt.Println("Request body: " + body)
		}

		c.Request.Body = rdr2
		c.Next()
	}
}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	if buf != nil {
		return buf.String()
	}

	return ""
}

func StartServer(configuration *config.Configuration, verbose bool) {
	r := gin.Default()

	if verbose {
		r.Use(RequestLogger())
	}

	for _, endpoint := range configuration.Endpoints {
		mapper, err := selectMap(endpoint.Verb)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		mapper(r, endpoint)
	}

	r.Run(fmt.Sprintf(":%d", configuration.Port))
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
	contentType := c.GetHeader("Content-Type")

	// Check if body is empty when content type is specified
	if contentType != "" {
		if c.Request.ContentLength == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Request body is empty but Content-Type header is present",
			})
			return
		}
	}

	var body map[string]any
	var err error
	switch contentType {
	case "application/json", "application/json; charset=utf-8":
		body, err = readFromJson(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON payload: " + err.Error(),
			})
			return
		}
	case "application/x-www-form-urlencoded", "multipart/form-data":
		body, err = readFromForm(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid form data: " + err.Error(),
			})
			return
		}
	case "":
		// No content type specified, try to read as JSON if body exists
		if c.Request.ContentLength > 0 {
			body, err = readFromJson(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Unable to parse request body. Please specify Content-Type header or ensure valid JSON format",
				})
				return
			}
		} else {
			body = make(map[string]any)
		}
	default:
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error": "Unsupported Content-Type: " + contentType + ". Supported types are: application/json, application/x-www-form-urlencoded, multipart/form-data",
		})
		return
	}

	mapReturns(c, body, mappings)
}

func mapReturns(c *gin.Context, body map[string]any, mappings []config.Mapping) {
	for _, mapping := range mappings {
		if allMatch(c, body, mapping.Params) {
			buildResponse(c, mapping.RespCode, mapping.Content)
			return
		}
	}
	// No mapping matched - return 404 with error message
	c.JSON(http.StatusNotFound, gin.H{
		"error": "No matching endpoint configuration found for this request",
	})
}

func allMatch(c *gin.Context, body map[string]interface{}, params []expressions.Expression) bool {
	for _, param := range params {
		if !param.Evaluate(expressions.EvaluationFetchers{BodyFetcher: body, QueryFetcher: c.Query, QueryArrayFetcher: c.QueryArray, ParamFetcher: c.Param}).(bool) {
			return false
		}
	}

	return true
}

func buildResponse(c *gin.Context, code int, content config.Content) {
	switch content.Type {
	case config.ContentTypeJson:
		c.JSON(code, content.Data)
	case config.ContentTypeFile:
		c.Status(code)
		c.File(content.Data.(config.DataFile).Path)
	}
}

func readFromJson(c *gin.Context) (map[string]any, error) {
	// Check if body is actually empty
	if c.Request.ContentLength == 0 {
		return nil, errors.New("request body is empty")
	}

	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return body, nil
}

func readFromForm(c *gin.Context) (map[string]any, error) {
	// Check if body is actually empty
	if c.Request.ContentLength == 0 {
		return nil, errors.New("request body is empty")
	}

	formData := c.Request.PostForm
	if formData == nil {
		if err := c.Request.ParseForm(); err != nil {
			return nil, fmt.Errorf("failed to parse form data: %w", err)
		}
		return squashFormData(c.Request.PostForm), nil
	}
	return squashFormData(formData), nil
}

func squashFormData(formData url.Values) map[string]any {
	result := make(map[string]any)

	for key, values := range formData {
		if len(values) > 1 {
			result[key] = values // keep as []string
		} else {
			result[key] = values[0] // collapse single value
		}
	}
	return result
}
