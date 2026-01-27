"""CLI entry point for doppelganger mock server."""

import argparse
import signal
import sys
import threading
from typing import NoReturn

from .config import parse_configuration_file
from .server import start_server


def main() -> NoReturn:
    """Main entry point for the CLI."""
    parser = argparse.ArgumentParser(
        prog="doppelganger-py",
        description="Simple CLI mock server - Python implementation",
    )
    parser.add_argument(
        "config_file",
        help="Path to the JSON configuration file",
    )
    parser.add_argument(
        "-verbose",
        "--verbose",
        action="store_true",
        help="Increase verbosity (log request payloads)",
    )
    
    args = parser.parse_args()
    
    # Parse configuration
    try:
        servers = parse_configuration_file(args.config_file)
    except Exception as e:
        print(f"Error parsing configuration: {e}", file=sys.stderr)
        sys.exit(2)
    
    # Track server threads
    server_threads: list[threading.Thread] = []
    shutdown_event = threading.Event()
    
    def signal_handler():
        """Handle shutdown signals."""
        print("\nShutting down...")
        shutdown_event.set()
        sys.exit(0)
    
    # Register signal handlers
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    # Start servers in threads
    for config in servers.configurations:
        thread = threading.Thread(
            target=start_server,
            args=(config, args.verbose),
            daemon=True,
        )
        thread.start()
        server_threads.append(thread)
    
    # Wait for shutdown signal
    try:
        shutdown_event.wait()
    except KeyboardInterrupt:
        print("\nShutting down...")
    
    sys.exit(0)


if __name__ == "__main__":
    main()
