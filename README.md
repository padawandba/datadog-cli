# Datadog CLI

A command-line interface tool for managing Datadog resources.

## Overview

The Datadog CLI provides a convenient way to interact with the Datadog API from the command line. It allows you to manage various Datadog resources such as hosts, tags, and monitors without having to use the web interface.

## Features

- **Hosts Management**: List, mute, and unmute hosts
- **Tags Management**: List, add, and remove tags from hosts
- **Monitors Management**: List, mute, and unmute monitors
- **Flexible Output Formats**: Display results in table, JSON, or YAML format

## Quick Start

```bash
# Build the CLI
go build -o dd ./cmd/dd

# Set your Datadog API and application keys
export DD_API_KEY="your_api_key"
export DD_APP_KEY="your_application_key"

# Run a command
./dd hosts list
```

## Documentation

For detailed documentation, including installation instructions, configuration options, and command references, see the [CLI Documentation](cmd/dd/README.md).

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.
