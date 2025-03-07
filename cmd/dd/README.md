# Datadog CLI (`dd`)

A command-line interface tool for managing Datadog resources.

## Overview

The Datadog CLI (`dd`) provides a convenient way to interact with the Datadog API from the command line. It allows you to manage various Datadog resources such as hosts, tags, and monitors without having to use the web interface.

## Installation

### Prerequisites

- Go 1.18 or higher
- Datadog API and Application keys

### Building from Source

Clone the repository and build the application:

```bash
git clone https://github.com/padawandba/datadog-cli.git
cd datadog-cli
go build -o dd ./cmd/dd
```

Move the binary to a location in your PATH to make it globally accessible:

```bash
# Linux/macOS
sudo mv dd /usr/local/bin/

# Or add to your user bin directory
mv dd ~/bin/
```

## Configuration

The CLI requires Datadog API and Application keys to authenticate with the Datadog API. You can provide these in several ways:

### Environment Variables

```bash
export DD_API_KEY="your_api_key"
export DD_APP_KEY="your_application_key"
export DD_SITE="datadoghq.com"  # Optional, defaults to datadoghq.com
```

### Command-Line Flags

```bash
dd --dd-api-key="your_api_key" --dd-app-key="your_application_key" [command]
```

### Configuration File

The CLI will look for a configuration file at `~/.config/datadog/config.yaml` with the following format:

```yaml
api_key: your_api_key
app_key: your_application_key
site: datadoghq.com  # Optional
output: table  # Optional, can be table, json, or yaml
```

## Usage

### General Syntax

```bash
dd [global options] command [command options] [arguments...]
```

### Global Options

- `--dd-api-key value`: Datadog API key
- `--dd-app-key value`: Datadog application key
- `--dd-site value`: Datadog site (e.g., datadoghq.com, datadoghq.eu) (default: "datadoghq.com")
- `--output value, -o value`: Output format (table, json, yaml) (default: "table")
- `--help, -h`: Show help

### Commands

#### Hosts

Manage Datadog hosts:

```bash
# List hosts
dd hosts list [--filter value]

# Mute a host
dd hosts mute HOSTNAME [--end value] [--message value]

# Unmute a host
dd hosts unmute HOSTNAME
```

#### Tags

Manage Datadog tags:

```bash
# List tags for a host
dd tags list HOSTNAME [--source value]

# Add tags to a host
dd tags add HOSTNAME TAG [TAG...]

# Remove tags from a host
dd tags remove HOSTNAME [TAG...] [--all]
```

#### Monitors

Manage Datadog monitors:

```bash
# List monitors
dd monitors list [--query value] [--tags value]

# Mute a monitor
dd monitors mute MONITOR_ID [--scope value] [--duration value]

# Unmute a monitor
dd monitors unmute MONITOR_ID [--scope value]
```

## Examples

### List all hosts

```bash
dd hosts list
```

### Add tags to a host

```bash
dd tags add web-server-01 env:production role:web
```

### List monitors with specific tags

```bash
dd monitors list --tags "service:api,env:production"
```

### Mute a monitor for 2 hours

```bash
dd monitors mute 12345678 --duration 2h
```

### Output results in JSON format

```bash
dd hosts list -o json > hosts.json
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details. 