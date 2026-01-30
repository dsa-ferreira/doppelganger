#!/usr/bin/env node
/**
 * CLI entry point for doppelganger mock server.
 */

import { program } from 'commander';
import { Server } from 'http';
import { parseConfigurationFile } from './config.js';
import { startServer } from './server.js';

function main(): void {
  program
    .name('doppelganger-ts')
    .description('Simple CLI mock server - TypeScript implementation')
    .argument('<config_file>', 'Path to the JSON configuration file')
    .option('-v, --verbose', 'Increase verbosity (log request payloads)')
    .parse();

  const configFile = program.args[0];
  const options = program.opts();
  const verbose = options.verbose ?? false;

  // Parse configuration
  let servers;
  try {
    servers = parseConfigurationFile(configFile);
  } catch (error) {
    console.error(`Error parsing configuration: ${error}`);
    process.exit(2);
  }

  // Track servers
  const runningServers: Server[] = [];

  // Start servers
  for (const config of servers.configurations) {
    const server = startServer(config, verbose);
    runningServers.push(server);
  }

  // Handle shutdown signals
  const shutdown = (): void => {
    console.log('\nShutting down...');
    for (const server of runningServers) {
      server.close();
    }
    process.exit(0);
  };

  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
}

main();
