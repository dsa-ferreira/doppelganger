"""Configuration parsing for doppelganger mock server."""

import json
from dataclasses import dataclass
from enum import Enum
from typing import Any

from .expressions import Expression, build_expression


class ContentType(Enum):
    """Type of response content."""
    JSON = "JSON"
    FILE = "FILE"


@dataclass
class DataFile:
    """File path for FILE content type."""
    path: str


@dataclass
class Content:
    """Response content configuration."""
    type: ContentType
    data: Any  # Either dict/list for JSON or DataFile for FILE


@dataclass
class Mapping:
    """Endpoint mapping configuration."""
    params: list[Expression]
    resp_code: int
    content: Content | None


@dataclass
class Endpoint:
    """Endpoint configuration."""
    path: str
    verb: str
    mappings: list[Mapping]


@dataclass
class Configuration:
    """Server configuration."""
    port: int
    endpoints: list[Endpoint]


@dataclass
class Servers:
    """Root configuration containing all servers."""
    configurations: list[Configuration]


def parse_content(data: dict | None) -> Content | None:
    """Parse content configuration from JSON."""
    if data is None:
        return None
    
    content_type_str = data.get("type", "JSON")
    content_type = ContentType(content_type_str)
    
    if content_type == ContentType.FILE:
        file_data = data.get("data", {})
        return Content(type=content_type, data=DataFile(path=file_data.get("path", "")))
    else:  # JSON
        return Content(type=content_type, data=data.get("data"))


def parse_mapping(data: dict) -> Mapping:
    """Parse mapping configuration from JSON."""
    params_data = data.get("params", [])
    params = [build_expression(p) for p in params_data]
    
    content_data = data.get("content")
    content = parse_content(content_data)
    
    # Determine response code
    resp_code = data.get("code")
    if resp_code is None:
        if content is None:
            resp_code = 204
        else:
            resp_code = 200
    
    return Mapping(params=params, resp_code=resp_code, content=content)


def parse_endpoint(data: dict) -> Endpoint:
    """Parse endpoint configuration from JSON."""
    path = data.get("path", "/")
    verb = data.get("verb", "GET")
    mappings_data = data.get("mappings", [])
    mappings = [parse_mapping(m) for m in mappings_data]
    
    return Endpoint(path=path, verb=verb, mappings=mappings)


def parse_configuration(data: dict) -> Configuration:
    """Parse server configuration from JSON."""
    port = data.get("port", 8000)
    endpoints_data = data.get("endpoint", [])
    endpoints = [parse_endpoint(e) for e in endpoints_data]
    
    return Configuration(port=port, endpoints=endpoints)


def parse_servers(data: dict) -> Servers:
    """Parse servers configuration from JSON."""
    if "servers" in data:
        servers_data = data["servers"]
        if not servers_data:
            raise ValueError("No server found")
        configurations = [parse_configuration(s) for s in servers_data]
        return Servers(configurations=configurations)
    else:
        # Fallback: treat the whole config as a single server
        configuration = parse_configuration(data)
        return Servers(configurations=[configuration])


def parse_configuration_file(file_path: str) -> Servers:
    """Parse configuration from a JSON file."""
    with open(file_path, "r") as f:
        data = json.load(f)
    
    return parse_servers(data)
