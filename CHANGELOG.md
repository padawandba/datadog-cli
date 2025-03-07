# Changelog

## [Unreleased]

### Added
- Integrated Datadog logging: CLI now sends logs directly to Datadog for better observability
- Added `--debug` flag to enable debug-level logging
- Added `--env` flag to set environment tag for logs
- Created comprehensive documentation for the Datadog logging feature
- Added example script to demonstrate logging functionality

### Changed
- Enhanced error handling across the codebase
- Improved configuration validation
- Updated main application to use structured logging
- Added graceful shutdown handling

### Fixed
- Fixed nil pointer issues in formatter package
- Improved error messages for configuration errors
- Enhanced HTTP request/response logging

## [0.1.0] - 2023-06-01

### Added
- Initial release with basic functionality
- Support for hosts, tags, and monitors commands
- Table, JSON, and YAML output formats 