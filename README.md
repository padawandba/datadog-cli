# Datadog CLI

A command-line interface tool for managing Datadog resources.

## Overview

The Datadog CLI provides a convenient way to interact with the Datadog API from the command line. It allows you to manage various Datadog resources such as hosts, tags, and monitors without having to use the web interface.

## Features

- **Hosts Management**: List, mute, and unmute hosts
- **Tags Management**: List, add, and remove tags from hosts
- **Monitors Management**: List, mute, and unmute monitors
- **Flexible Output Formats**: Display results in table, JSON, or YAML format
- **Integrated Logging**: Automatically sends logs to your Datadog account for better observability

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

## Configuration

### Environment Variables

```bash
# Required for API access
export DD_API_KEY="your_api_key"
export DD_APP_KEY="your_application_key"

# Optional settings
export DD_SITE="datadoghq.com"  # Datadog site (default: datadoghq.com)
export DD_ENV="prod"            # Environment tag for logs (default: dev)
export DD_DEBUG=true            # Enable debug logging
```

### Command-Line Flags

```bash
# API credentials
./dd --dd-api-key="your_api_key" --dd-app-key="your_application_key" [command]

# Output format
./dd --output=json hosts list

# Logging options
./dd --debug --env=staging hosts list
```

## Logging to Datadog

The CLI automatically sends logs to your Datadog account when you provide API credentials. This gives you better visibility into CLI usage and helps with troubleshooting.

Logs include:
- Command execution details
- API request/response information
- Error messages and stack traces
- Performance metrics

You can view these logs in your Datadog account by searching for:
- `service:datadog-cli`
- Environment tag matching your `DD_ENV` setting

## Documentation

For detailed documentation, including installation instructions, configuration options, and command references, see the [CLI Documentation](cmd/dd/README.md).

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.
