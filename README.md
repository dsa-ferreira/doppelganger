# Doppelganger

Simple CLI mock server. Uses a JSON spec to setup a web server with stubbed results.

Available in two implementations:
- **Go** - Uses the Golang Gin framework
- **Python** - Uses Flask

## Current Features

- Basic endpoint mapping
- Responses based on request values (PATH, QUERY or BODY)
- Response codes

## Features under construction

- Response templating

## Installing

### Go Version

Clone the repo -> cd into the folder -> `sudo make install`

### Python Version

Requires Python 3.10+

**Recommended: Using pipx** (handles virtual environment automatically)

```bash
# Install pipx if not already installed
sudo apt install pipx

# Install doppelganger-py
pipx install ./python
```

**Or using make**

```bash
make install-py-pipx
```

On older systems or other distributions:

```bash
pip install --user ./python
```

**Or using make**

```bash
make install-py
```

## How to use

### Go Version

`doppelganger <json_file>`

### Python Version

`doppelganger-py <json_file>`

### Options

Can use `-verbose` or `--verbose` to log request payloads

### Json file schema (OUT OF DATE, will update soon)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["servers"],
  "properties": {
    "servers": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["endpoint"],
        "properties": {
          "port": {
            "type": integer,
            "description": "Port for which the server will listen to"
          },
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
                  "enum": ["GET", "POST", "PUT", "DELETE"]
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
                          "required": ["type"],
                          "properties": {
                            "key": { 
                              "type": "string",
                              "description": "The parameter/attribute's key"
                            },
                            "type": {
                              "type": "string",
                              "description": "The type of the parameter. Can be path, query or body params.",
                              "enum": [
                                "AND", 
                                "OR", 
                                "NOT", 
                                "QUERY_ARRAY", 
                                "STRING", 
                                "EQUALS", 
                                "REGEX", 
                                "CONTAINS", 
                                "BODY", 
                                "PATH", 
                                "QUERY"
                              ]
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
                        "required": ["data"],
                        "properties": {
                          "type": {
                            "type": "string",
                            "enum": ["JSON", "FILE"],
                            "default": "JSON"
                          },
                          "data": {
                            "type": "object",
                            "description": "Either an open json object that will be used as the response or a file path",
                            "properties": {
                              "path": {
                                "type": "string",
                                "description": "Path to the file, relative to where you botted the doppleganger"
                              }
                            },
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
        }
      }
    }
  }
}
```

#### Example

```json
{
  "servers": 
    [
      {
        "port": 8081,
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
                  "type": "JSON",
                  "data": {
                    "message": "Empty one!"
                  }
                }
              }
            ]
          },
          {
            "path": "/api/status",
            "mappings": [
              {
                "content": {
                  "type": "FILE",
                  "data": {
                    "path": "/path/to/file"
                  }
                }
              }
            ]
          }
        ]
      }
    }
  ]
}
```
