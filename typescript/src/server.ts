/**
 * Express-based HTTP server for doppelganger mock server.
 */

import express, {
  Application,
  Request,
  Response,
  NextFunction,
  RequestHandler,
} from 'express';
import { Server } from 'http';
import { resolve } from 'path';
import {
  Configuration,
  Endpoint,
  Mapping,
  Content,
  ContentType,
  DataFile,
} from './config.js';
import { EvaluationFetchers, Expression } from './expressions.js';

/**
 * Create evaluation fetchers from the current request context
 */
function createFetchers(
  body: Record<string, unknown> | null,
  pathParams: Record<string, string>,
  req: Request
): EvaluationFetchers {
  return {
    bodyFetcher: body ?? {},
    queryFetcher: (key: string): string => {
      const value = req.query[key];
      if (typeof value === 'string') {
        return value;
      }
      if (Array.isArray(value) && value.length > 0) {
        return String(value[0]);
      }
      return '';
    },
    queryArrayFetcher: (key: string): string[] => {
      const value = req.query[key];
      if (typeof value === 'string') {
        return [value];
      }
      if (Array.isArray(value)) {
        return value.map((v) => String(v));
      }
      return [];
    },
    paramFetcher: (key: string): string => {
      return pathParams[key] ?? '';
    },
  };
}

/**
 * Check if all parameter expressions match
 */
function allMatch(fetchers: EvaluationFetchers, params: Expression[]): boolean {
  for (const param of params) {
    if (!param.evaluate(fetchers)) {
      return false;
    }
  }
  return true;
}

/**
 * Build the response for a matching mapping
 */
function buildResponse(res: Response, mapping: Mapping): void {
  if (mapping.content === null) {
    res.status(mapping.respCode).send('');
    return;
  }

  const content: Content = mapping.content;

  if (content.type === ContentType.JSON) {
    res.status(mapping.respCode).json(content.data);
  } else if (content.type === ContentType.FILE) {
    const fileData = content.data as DataFile;
    const filePath = resolve(fileData.path);
    res.status(mapping.respCode).sendFile(filePath);
  } else {
    res.status(mapping.respCode).send('');
  }
}

/**
 * Read the request body based on content type
 */
function readBody(req: Request): Record<string, unknown> | null {
  const contentType = req.get('Content-Type') ?? '';

  if (contentType.includes('application/json')) {
    return (req.body as Record<string, unknown>) ?? {};
  } else if (
    contentType.includes('application/x-www-form-urlencoded') ||
    contentType.includes('multipart/form-data')
  ) {
    const formData: Record<string, unknown> = {};
    for (const key of Object.keys(req.body)) {
      const values = req.body[key];
      if (Array.isArray(values) && values.length > 1) {
        formData[key] = values;
      } else if (Array.isArray(values)) {
        formData[key] = values[0];
      } else {
        formData[key] = values;
      }
    }
    return formData;
  }

  return null;
}

/**
 * Convert Express params to a simple string record
 */
function normalizeParams(params: Record<string, string | string[]>): Record<string, string> {
  const result: Record<string, string> = {};
  for (const key of Object.keys(params)) {
    const value = params[key];
    if (Array.isArray(value)) {
      result[key] = value[0] ?? '';
    } else {
      result[key] = value;
    }
  }
  return result;
}

/**
 * Create a request handler for an endpoint
 */
function createEndpointHandler(endpoint: Endpoint, verbose: boolean): RequestHandler {
  return (req: Request, res: Response): void => {
    if (verbose) {
      const bodyStr = JSON.stringify(req.body);
      if (bodyStr && bodyStr !== '{}') {
        console.log(`Request body: ${bodyStr}`);
      }
    }

    let body: Record<string, unknown> | null = null;
    if (['POST', 'PUT', 'DELETE'].includes(req.method)) {
      body = readBody(req);
    }

    const pathParams = normalizeParams(req.params as Record<string, string | string[]>);
    const fetchers = createFetchers(body, pathParams, req);

    for (const mapping of endpoint.mappings) {
      if (allMatch(fetchers, mapping.params)) {
        buildResponse(res, mapping);
        return;
      }
    }

    res.status(404).json({ error: 'No matching mapping found' });
  };
}

/**
 * Request logger middleware for verbose mode
 */
function requestLogger(): RequestHandler {
  return (req: Request, _res: Response, next: NextFunction): void => {
    // Body will be logged in the endpoint handler after parsing
    next();
  };
}

/**
 * Create an Express app from configuration
 */
export function createApp(configuration: Configuration, verbose: boolean = false): Application {
  const app = express();

  // Body parsing middleware
  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));

  if (verbose) {
    app.use(requestLogger());
  }

  // Register endpoints
  for (const endpoint of configuration.endpoints) {
    const handler = createEndpointHandler(endpoint, verbose);
    const path = endpoint.path.startsWith('/') ? endpoint.path : `/${endpoint.path}`;

    switch (endpoint.verb.toUpperCase()) {
      case 'GET':
        app.get(path, handler);
        break;
      case 'POST':
        app.post(path, handler);
        break;
      case 'PUT':
        app.put(path, handler);
        break;
      case 'DELETE':
        app.delete(path, handler);
        break;
      default:
        console.error(`Unknown HTTP verb: ${endpoint.verb}`);
    }
  }

  return app;
}

/**
 * Start an Express server for the given configuration
 */
export function startServer(
  configuration: Configuration,
  verbose: boolean = false
): Server {
  const app = createApp(configuration, verbose);

  const server = app.listen(configuration.port, '0.0.0.0', () => {
    console.log(` * Running on http://0.0.0.0:${configuration.port}`);
  });

  return server;
}
