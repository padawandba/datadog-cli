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
- Migrated to latest Datadog API client v2 features and best practices
- Improved HTTP response handling with detailed error messages
- Enhanced parameter initialization for API calls
- Optimized tag filtering logic for better performance
- Implemented consistent formatting for all resource types (hosts, monitors, tags)

### Fixed
- Fixed nil pointer issues in formatter package
- Improved error messages for configuration errors
- Enhanced HTTP request/response logging
- Fixed potential memory leaks in API client usage
- Fixed Datadog API URL handling to ensure proper 'api.' prefix
- Fixed host list display to properly format pointer values instead of showing memory addresses
- Fixed monitor and tag display formatting for better readability

## [0.1.0] - 2023-06-01

### Added
- Initial release with basic functionality
- Support for hosts, tags, and monitors commands
- Table, JSON, and YAML output formats 