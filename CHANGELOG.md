# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-09-06

### Added
- Repository-based filtering for PR and issue list handlers
- Support for `repository` parameter in API calls
- CWD-to-repository resolution functionality
- Repository metadata in API responses
- Comprehensive repository validation and access control

### Changed
- Updated PR and issue handlers to use repository-specific queries
- Enhanced response format with repository context
- Improved error handling for repository-related operations

### Technical
- Migrated from user-based to repository-based filtering
- Updated SDK client methods to support repository parameters
- Added repository validation logic
- Enhanced test coverage for repository-based functionality

## [1.0.0] - 2025-08-26

### Added
- Initial release of Forgejo MCP server
- Basic PR and issue listing functionality
- MCP protocol implementation
- Gitea SDK integration
- Cobra CLI framework
- Configuration management
- Logging system