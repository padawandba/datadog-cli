# Datadog Logging Integration

This document explains how the Datadog CLI tool integrates with Datadog's logging system to provide enhanced observability.

## Overview

The Datadog CLI tool includes built-in functionality to send logs directly to your Datadog account. This integration provides several benefits:

1. **Centralized Logging**: All CLI operations are logged to your Datadog account, providing a centralized view of CLI usage across your organization.
2. **Advanced Analytics**: Leverage Datadog's powerful log analytics capabilities to understand CLI usage patterns.
3. **Correlation**: Correlate CLI operations with other metrics and events in your Datadog dashboards.
4. **Troubleshooting**: Easily troubleshoot issues by examining detailed logs of CLI operations.

## How It Works

The CLI uses a custom `slog.Handler` implementation that sends logs to the Datadog Logs API. Key features include:

- **Asynchronous Processing**: Logs are sent asynchronously to avoid blocking CLI operations.
- **Batching**: Logs are batched to reduce API calls and improve performance.
- **Automatic Context**: Each log includes metadata about the command being executed, the environment, and the user.
- **Error Handling**: Failed log submissions are retried with exponential backoff.

## Configuration

### Environment Variables

```bash
# Required for logging to Datadog
export DD_API_KEY="your_api_key"
export DD_APP_KEY="your_application_key"

# Optional logging configuration
export DD_ENV="prod"            # Environment tag (default: dev)
export DD_SERVICE="custom-name" # Service name (default: datadog-cli)
export DD_DEBUG=true            # Enable debug-level logging
```

### Command-Line Flags

```bash
# Enable debug logging
./dd --debug <command>

# Set environment tag
./dd --env production <command>
```

## Log Structure

Each log sent to Datadog includes the following information:

- **Message**: The log message describing the operation
- **Status**: The log level (INFO, ERROR, DEBUG, etc.)
- **Service**: Always set to "datadog-cli" (or custom value from DD_SERVICE)
- **Hostname**: The hostname of the machine running the CLI
- **Timestamp**: The time the log was generated
- **Attributes**:
  - `command`: The command being executed
  - `duration`: The execution time (for completed operations)
  - `error`: Error details (for failed operations)
  - `user`: The username of the user running the CLI
  - Custom attributes from the log context

## Example Logs

### Command Execution

```json
{
  "message": "Executing command: hosts list",
  "status": "INFO",
  "service": "datadog-cli",
  "hostname": "user-laptop",
  "timestamp": "2023-06-15T14:30:45Z",
  "attributes": {
    "command": "hosts list",
    "user": "admin"
  }
}
```

### API Request

```json
{
  "message": "API request to /api/v1/hosts",
  "status": "DEBUG",
  "service": "datadog-cli",
  "hostname": "user-laptop",
  "timestamp": "2023-06-15T14:30:46Z",
  "attributes": {
    "command": "hosts list",
    "method": "GET",
    "url": "https://api.datadoghq.com/api/v1/hosts",
    "duration_ms": 235
  }
}
```

### Error

```json
{
  "message": "Failed to execute command: hosts mute",
  "status": "ERROR",
  "service": "datadog-cli",
  "hostname": "user-laptop",
  "timestamp": "2023-06-15T14:35:12Z",
  "attributes": {
    "command": "hosts mute web-server-01",
    "error": "Host not found: web-server-01",
    "user": "admin"
  }
}
```

## Viewing Logs in Datadog

To view logs from the CLI in your Datadog account:

1. Navigate to the Logs Explorer in your Datadog account
2. Search for `service:datadog-cli`
3. Filter by environment using `env:your_environment`

You can create saved views and alerts based on these logs to monitor CLI usage and detect issues.

## Implementation Details

The logging implementation is contained in the `internal/platform/datadog/logger.go` file. It uses the Go `slog` package introduced in Go 1.21 to provide structured logging capabilities.

Key components:

- `DatadogLogEntry`: Represents a log entry formatted for Datadog
- `DatadogHandler`: Implements the `slog.Handler` interface to send logs to Datadog
- Background worker: Processes logs asynchronously and handles batching

## Disabling Logging

If you need to disable logging to Datadog for any reason, you can do so by setting an empty API key:

```bash
export DD_API_KEY=""
```

This will cause the CLI to fall back to console-only logging. 