"""Flask-based HTTP server for doppelganger mock server."""

import re
from typing import Any

from flask import Flask, request, jsonify, send_file, Response

from .config import Configuration, Endpoint, Mapping, Content, ContentType, DataFile
from .expressions import EvaluationFetchers, Expression


def convert_path_params(path: str) -> str:
    """Convert Gin-style path params (:param) to Flask-style (<param>).
    
    Also ensures the path starts with a slash as required by Flask.
    """
    if not path.startswith("/"):
        path = "/" + path
    return re.sub(r":(\w+)", r"<\1>", path)


def create_fetchers(body: dict[str, Any] | None, path_params: dict[str, str]) -> EvaluationFetchers:
    """Create evaluation fetchers from the current request context."""
    
    def query_fetcher(key: str) -> str:
        return request.args.get(key, "")
    
    def query_array_fetcher(key: str) -> list[str]:
        return request.args.getlist(key)
    
    def param_fetcher(key: str) -> str:
        return path_params.get(key, "")
    
    return EvaluationFetchers(
        body_fetcher=body or {},
        query_fetcher=query_fetcher,
        query_array_fetcher=query_array_fetcher,
        param_fetcher=param_fetcher,
    )


def all_match(fetchers: EvaluationFetchers, params: list[Expression]) -> bool:
    """Check if all parameter expressions match."""
    for param in params:
        if not param.evaluate(fetchers):
            return False
    return True


def build_response(mapping: Mapping) -> tuple[Response | Any, int]:
    """Build the response for a matching mapping."""
    if mapping.content is None:
        return "", mapping.resp_code
    
    content = mapping.content
    
    if content.type == ContentType.JSON:
        return jsonify(content.data), mapping.resp_code
    elif content.type == ContentType.FILE:
        file_data: DataFile = content.data
        return send_file(file_data.path), mapping.resp_code
    
    return "", mapping.resp_code


def read_body() -> dict[str, Any] | None:
    """Read the request body based on content type."""
    content_type = request.content_type or ""
    
    if "application/json" in content_type:
        try:
            return request.get_json(force=True, silent=True) or {}
        except Exception:
            return {}
    elif "application/x-www-form-urlencoded" in content_type or "multipart/form-data" in content_type:
        form_data: dict[str, Any] = {}
        for key in request.form:
            values = request.form.getlist(key)
            if len(values) > 1:
                form_data[key] = values
            else:
                form_data[key] = values[0]
        return form_data
    
    return None


def create_endpoint_handler(endpoint: Endpoint, verbose: bool):
    """Create a request handler for an endpoint."""
    
    def handler(**path_params):
        if verbose:
            body_bytes = request.get_data()
            if body_bytes:
                print(f"Request body: {body_bytes.decode('utf-8', errors='replace')}")
        
        body = None
        if request.method in ("POST", "PUT", "DELETE"):
            body = read_body()
        
        fetchers = create_fetchers(body, path_params)
        
        for mapping in endpoint.mappings:
            if all_match(fetchers, mapping.params):
                return build_response(mapping)
        
        return jsonify({"error": "No matching mapping found"}), 404
    
    return handler


def create_app(configuration: Configuration, verbose: bool = False) -> Flask:
    """Create a Flask app from configuration."""
    app = Flask(__name__)
    
    # Register endpoints
    for i, endpoint in enumerate(configuration.endpoints):
        flask_path = convert_path_params(endpoint.path)
        handler = create_endpoint_handler(endpoint, verbose)
        
        # Create unique endpoint name
        endpoint_name = f"endpoint_{i}_{endpoint.verb}_{endpoint.path.replace('/', '_').replace(':', '_')}"
        
        app.add_url_rule(
            flask_path,
            endpoint_name,
            handler,
            methods=[endpoint.verb],
        )
    
    return app


def start_server(configuration: Configuration, verbose: bool = False):
    """Start a Flask server for the given configuration."""
    app = create_app(configuration, verbose)
    
    # Use werkzeug's serving directly to allow threading
    from werkzeug.serving import make_server
    
    server = make_server("0.0.0.0", configuration.port, app, threaded=True)
    print(f" * Running on http://0.0.0.0:{configuration.port}")
    server.serve_forever()
