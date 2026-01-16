
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- WebSocket transport
- gRPC transport  
- Streaming responses
- Multi-backend routing

## [0.1.0] - 2026-01-16

### Added
- Initial framework release
- stdio and HTTP transports
- Prometheus metrics integration
- Structured logging with slog
- Plugin-based backend system
- Fluent tool definition API
- Configuration via YAML/env/flags
- Graceful shutdown
- Health check endpoints
- Filesystem server example with full security
- Path traversal prevention
- Workspace sandboxing
- Request/response size tracking
- System metrics (memory, goroutines, uptime)

### Architecture
- Backend registry for plugin system
- BaseBackend for automatic request routing
- Transport abstraction layer
- Protocol layer with JSON-RPC 2.0
- Observability stack with metrics server

### Security
- Path validation and sandboxing
- File size limits
- Extension filtering (whitelist/blacklist)
- Read-only mode support

[Unreleased]: https://github.com/SaherElMasry/go-mcp-framework/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/SaherElMasry/go-mcp-framework/releases/tag/v0.1.0
