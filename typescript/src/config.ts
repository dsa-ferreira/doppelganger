/**
 * Configuration parsing for doppelganger mock server.
 */

import { readFileSync } from 'fs';
import { Expression, ExpressionData, buildExpression } from './expressions.js';

// Content type enum
export enum ContentType {
  JSON = 'JSON',
  FILE = 'FILE',
}

// File data for FILE content type
export interface DataFile {
  path: string;
}

// Response content configuration
export interface Content {
  type: ContentType;
  data: unknown; // Either JSON data or DataFile for FILE
}

// Endpoint mapping configuration
export interface Mapping {
  params: Expression[];
  respCode: number;
  content: Content | null;
}

// Endpoint configuration
export interface Endpoint {
  path: string;
  verb: string;
  mappings: Mapping[];
}

// Server configuration
export interface Configuration {
  port: number;
  endpoints: Endpoint[];
}

// Root configuration containing all servers
export interface Servers {
  configurations: Configuration[];
}

// Raw JSON types for parsing
interface RawContent {
  type?: string;
  data?: unknown;
}

interface RawMapping {
  params?: ExpressionData[];
  code?: number;
  content?: RawContent;
}

interface RawEndpoint {
  path?: string;
  verb?: string;
  mappings?: RawMapping[];
}

interface RawConfiguration {
  port?: number;
  endpoint?: RawEndpoint[];
}

interface RawServers {
  servers?: RawConfiguration[];
}

/**
 * Parse content configuration from JSON
 */
function parseContent(data: RawContent | undefined): Content | null {
  if (!data) {
    return null;
  }

  const contentTypeStr = data.type ?? 'JSON';
  const contentType = contentTypeStr === 'FILE' ? ContentType.FILE : ContentType.JSON;

  if (contentType === ContentType.FILE) {
    const fileData = data.data as { path?: string } | undefined;
    return {
      type: contentType,
      data: { path: fileData?.path ?? '' } as DataFile,
    };
  } else {
    return {
      type: contentType,
      data: data.data,
    };
  }
}

/**
 * Parse mapping configuration from JSON
 */
function parseMapping(data: RawMapping): Mapping {
  const paramsData = data.params ?? [];
  const params = paramsData.map((p) => buildExpression(p));

  const content = parseContent(data.content);

  // Determine response code
  let respCode = data.code;
  if (respCode === undefined) {
    if (content === null) {
      respCode = 204;
    } else {
      respCode = 200;
    }
  }

  return { params, respCode, content };
}

/**
 * Parse endpoint configuration from JSON
 */
function parseEndpoint(data: RawEndpoint): Endpoint {
  const path = data.path ?? '/';
  const verb = data.verb ?? 'GET';
  const mappingsData = data.mappings ?? [];
  const mappings = mappingsData.map((m) => parseMapping(m));

  return { path, verb, mappings };
}

/**
 * Parse server configuration from JSON
 */
function parseConfiguration(data: RawConfiguration): Configuration {
  const port = data.port ?? 8000;
  const endpointsData = data.endpoint ?? [];
  const endpoints = endpointsData.map((e) => parseEndpoint(e));

  return { port, endpoints };
}

/**
 * Parse servers configuration from JSON
 */
function parseServers(data: RawServers | RawConfiguration): Servers {
  if ('servers' in data && data.servers) {
    const serversData = data.servers;
    if (serversData.length === 0) {
      throw new Error('No server found');
    }
    const configurations = serversData.map((s) => parseConfiguration(s));
    return { configurations };
  } else {
    // Fallback: treat the whole config as a single server
    const configuration = parseConfiguration(data as RawConfiguration);
    return { configurations: [configuration] };
  }
}

/**
 * Parse configuration from a JSON file
 */
export function parseConfigurationFile(filePath: string): Servers {
  const fileContent = readFileSync(filePath, 'utf-8');
  const data = JSON.parse(fileContent) as RawServers | RawConfiguration;
  return parseServers(data);
}
