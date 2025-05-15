# Doppelganger

Simple CLI mock server. Uses a JSON spec and the Golang Gin framework to setup a web server with stubbed result.

### Current Features

- Basic endpoint mapping
- Responses based on request values (PATH, QUERY or BODY)
- Response codes


### Features under construction

- Response templating

### Installing

Clone the repo -> cd into the folder -> `sudo make install`

### How to use

`doppelganger <json_file>`

```
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["endpoint"],
  "properties": {
    "endpoint": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["path", "mappings"],
        "properties": {
          "path": { 
            "type": "string",
            "description": "Path for the enpoint's mapping"
          },
          "verb": {
            "type": "string",
            "description": "HTTP verb being mapped",
            "enum": ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"]
          },
          "mappings": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["content"],
              "properties": {
                "params": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "required": ["key", "type", "value"],
                    "properties": {
                      "key": { 
                        "type": "string",
                        "description": "The parameter/attribute's key"
                      },
                      "type": {
                        "type": "string",
                        "description": "The type of the parameter. Can be path, query or body params.",
                        "enum": ["BODY", "PATH", "QUERY"]
                      },
                      "value": { 
                        "type": "string",
                        "description": "The value that will need to match the received parameter."
                      }
                    }
                  }
                },
                "code": { 
                  "type": "integer",
                  "description": "Http status code for the response"
                },
                "content": {
                  "type": "object",
                  "description": "Open json object that will be used as the response. No validation or parsing made on this field.",
                  "additionalProperties": true
                }
              }
            }
          }
        }
      }
    }
  }
}

```

###### Example

```
{
  "endpoint": [
    {
      "path": "/api/user/:id",
      "verb": "POST",
      "mappings": [
        {
          "params": [
            {
              "key": "id",
              "type": "BODY",
              "value": "123"
            },
            {
              "key": "id",
              "type": "PATH",
              "value": "123"
            },
            {
              "key": "name",
              "type": "QUERY",
              "value": "John"
            }
          ],
          "code": 201,
          "content": {
            "message": "User created"
          }
        },
        {
          "params": [
          ],
          "content": {
            "message": "Empty one!"
          }
        }
      ]
    },
    {
      "path": "/api/status",
      "mappings": [
        {
          "content": {
            "status": "ok"
          }
        }
      ]
    }
  ]
}
```

