# Datadog CLI Commands

This document provides detailed information about the available commands in the Datadog CLI tool.

## Global Flags

The following flags can be used with any command:

```bash
--dd-api-key string      Datadog API key (can also use DD_API_KEY env var)
--dd-app-key string      Datadog Application key (can also use DD_APP_KEY env var)
--dd-site string         Datadog site to use (default "datadoghq.com")
--debug                  Enable debug logging
--env string             Environment tag for logs (default "dev")
--output string          Output format: table, json, yaml (default "table")
--help, -h               Show help for any command
```

## Hosts Commands

Commands for managing Datadog hosts.

### List Hosts

```bash
./dd hosts list [flags]
```

**Flags:**
```bash
--filter string      Filter hosts by name (supports wildcards)
--limit int          Maximum number of hosts to return (default 100)
--status string      Filter by host status (up, down, all) (default "up")
```

**Examples:**
```bash
# List all up hosts
./dd hosts list

# List hosts with "web" in the name
./dd hosts list --filter "*web*"

# List all hosts (including down) in JSON format
./dd hosts list --status all --output json
```

### Mute Host

```bash
./dd hosts mute <hostname> [flags]
```

**Flags:**
```bash
--end int            End time for muting in seconds from now
--message string     Message explaining the reason for muting
```

**Examples:**
```bash
# Mute a host indefinitely
./dd hosts mute web-server-01 --message "Maintenance in progress"

# Mute a host for 1 hour
./dd hosts mute web-server-01 --end 3600 --message "Scheduled maintenance"
```

### Unmute Host

```bash
./dd hosts unmute <hostname>
```

**Examples:**
```bash
./dd hosts unmute web-server-01
```

## Tags Commands

Commands for managing Datadog tags.

### List Tags

```bash
./dd tags list [flags]
```

**Flags:**
```bash
--source string      Filter tags by source
```

**Examples:**
```bash
# List all tags
./dd tags list

# List tags from a specific source
./dd tags list --source user
```

### Add Tags

```bash
./dd tags add <hostname> <tags> [flags]
```

**Flags:**
```bash
--source string      Tag source (default "user")
```

**Examples:**
```bash
# Add tags to a host
./dd tags add web-server-01 env:prod,role:web

# Add tags with a specific source
./dd tags add web-server-01 team:platform --source chef
```

### Remove Tags

```bash
./dd tags remove <hostname> <tags> [flags]
```

**Flags:**
```bash
--source string      Tag source (default "user")
```

**Examples:**
```bash
# Remove specific tags from a host
./dd tags remove web-server-01 env:prod,role:web

# Remove all tags from a host
./dd tags remove web-server-01 "*"
```

## Monitors Commands

Commands for managing Datadog monitors.

### List Monitors

```bash
./dd monitors list [flags]
```

**Flags:**
```bash
--filter string      Filter monitors by name (supports wildcards)
--limit int          Maximum number of monitors to return (default 100)
--status string      Filter by monitor status (alert, warn, no data, ok, all) (default "all")
--tags string        Filter monitors by tags (comma-separated)
```

**Examples:**
```bash
# List all monitors
./dd monitors list

# List monitors with "api" in the name
./dd monitors list --filter "*api*"

# List monitors with specific tags
./dd monitors list --tags "service:api,env:prod"
```

### Mute Monitor

```bash
./dd monitors mute <monitor_id> [flags]
```

**Flags:**
```bash
--end int            End time for muting in seconds from now
--message string     Message explaining the reason for muting
--scope string       Scope to mute (e.g., "host:web-server-01")
```

**Examples:**
```bash
# Mute a monitor indefinitely
./dd monitors mute 12345 --message "Investigating issues"

# Mute a monitor for 1 hour with a specific scope
./dd monitors mute 12345 --end 3600 --scope "host:web-server-01" --message "Maintenance"
```

### Unmute Monitor

```bash
./dd monitors unmute <monitor_id> [flags]
```

**Flags:**
```bash
--scope string       Scope to unmute (e.g., "host:web-server-01")
```

**Examples:**
```bash
# Unmute a monitor
./dd monitors unmute 12345

# Unmute a monitor for a specific scope
./dd monitors unmute 12345 --scope "host:web-server-01"
```

## Logging

The CLI automatically sends logs to your Datadog account. You can control the logging behavior with the following options:

```bash
# Enable debug logging
./dd --debug <command>

# Set the environment tag for logs
./dd --env production <command>
```

Logs are sent to Datadog with the service name `datadog-cli` and can be viewed in your Datadog logs explorer. 