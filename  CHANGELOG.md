# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
---
## [v0.4.0] - 2026-01-24

### 🎉 Major Release : Introducing Intelligent Response Caching System

### Added
- **Intelligent Response Caching System**
  - LRU cache with automatic eviction
  - TTL-based expiration (short/long modes)
  - Per-tool cache configuration
  - Deterministic SHA-256 cache key generation
  - Background cleanup of expired entries
  - Cache statistics API (hits, misses, evictions, hit rate)
  
- **Performance Improvements**
  - 53x speedup in real-world Weather API calls (478ms → 9ms)
  - 286x speedup in benchmarks
  - 100% cache hit rate in production workloads
  - Memory efficient: ~1KB per cached response
  
- **Developer Experience**
  - `framework.WithCache("short", 60)` - Enable caching in 1 line
  - `framework.WithToolCacheTTL(tool, ttl)` - Per-tool TTL overrides
  - `tool.WithCache(true, 5*time.Minute)` - Mark tools as cacheable
  - `cache.Stats()` - Get cache statistics
  
- **New Package: `cache/`**
  - `cache.Cache` interface for pluggable cache backends
  - `cache.MemoryCache` - In-memory LRU implementation
  - `cache.NoOpCache` - Disabled cache (safe default)
  - `cache.Config` - Cache configuration
  - `cache.KeyGenerator` - Deterministic key generation
  - `cache.New()` - Factory function

- **Backend Enhancements**
  - `ToolCacheConfig` - Cache metadata for tools
  - `IsCacheable()` - Check if tool supports caching
  - `GetCacheTTL()` - Get tool-specific TTL
  - `HasCacheTags()` - Tag-based cache grouping

- **Framework Integration**
  - Cache initialized before backend
  - Protocol handler cache-aware
  - Background cleanup goroutine
  - Graceful cache shutdown
  
- **Tests & Benchmarks**
  - 120+ new tests
  - 97%+ code coverage
  - Integration tests prove end-to-end functionality
  - Performance benchmarks show 286x speedup

### Changed
- Updated README badges (added Cache, Performance)
- Updated framework comparison table
- Enhanced Weather Server example with caching
- Improved documentation with performance benchmarks

### Performance
- Real-World (Weather API):

- Without cache: 478ms
- With cache:    9ms
- Speedup:       53x

Integration Tests:

- Without cache: 10.67ms
- With cache:    0.053ms
- Speedup:       202x

Benchmarks:

- Without cache: 5,204 ns/op
- With cache:    18 ns/op
- Speedup:       286x

### Security
- Cache disabled by default (must opt-in)
- Tools non-cacheable by default (must opt-in)
- Thread-safe operations (sync.RWMutex)
- No sensitive data cached (user must ensure)

### Breaking Changes
- None! Fully backward compatible with v0.3.0
---

## [0.3.0] - 2026-01-22

### 🎉 Major Release: Enterprise Authentication & Beautiful Terminal Output

This release adds enterprise-grade authentication, beautiful terminal UI components, and enhanced observability features. v0.3.0 represents a significant step toward production-ready enterprise deployments.

### Added

#### 🔐 Enterprise Authentication System
- **OAuth2 Integration** - Full OAuth2 2.0 support with PKCE
  - GitHub provider with automatic token refresh
  - Google provider support
  - Microsoft provider support
  - Slack provider support
  - Facebook provider support
  - Generic OAuth2 provider for custom implementations
- **API Key Authentication** - Simple token-based authentication with resource scoping
- **Database Authentication** - Direct database connection authentication
- **Token Management** - Automatic token refresh before expiry with configurable thresholds
- **Encrypted Storage** - AES-256-GCM encryption for sensitive tokens
- **Multi-Provider Support** - Use multiple authentication providers simultaneously
- **Provider Factory** - Easy OAuth2 provider creation with standard configurations
- **Resource Scoping** - Per-resource authentication configuration and validation

#### 🎨 Beautiful Terminal Output System
- **ANSI Color Support** - Full 256-color palette with automatic terminal detection
- **Rich Components**:
  - Banners with box drawing characters
  - ASCII tables with borders and alignment
  - Progress bars with percentage and custom messages
  - Spinners for long-running operations
  - Colored boxes for important messages
- **Colored Logging** - slog integration with colored output by level (INFO, WARN, ERROR, SUCCESS)
- **Smart Terminal Detection** - Auto-disable colors in CI/CD environments and when NO_COLOR is set
- **Semantic Colors** - Helper functions for success, error, warning, info messages
- **Reusable Package** - Use color system in your own applications

#### 📊 Enhanced Observability
- **Auth-Specific Metrics**:
  - `mcp_auth_validations_total` - Track validation attempts and status
  - `mcp_auth_token_refresh_total` - Monitor token refresh operations
  - `mcp_auth_token_expiry_seconds` - Time until token expiry
  - `mcp_auth_resource_duration_seconds` - Resource acquisition performance
  - `mcp_oauth2_flows_total` - OAuth2 flow tracking (initiated, completed, failed)
  - `mcp_token_store_operations_total` - Token storage operations
- **Auth Health Checks**:
  - Provider validation health checks
  - OAuth2 token status monitoring
  - Database connection health checks
  - Token store health verification
  - Multi-provider aggregate health
- **Instrumented Providers** - Transparent metrics wrapper for all auth providers
- **Colored Log Output** - Beautiful terminal logs with semantic colors

#### 📚 Complete Examples
- **GitHub Server** (Production-Ready!)
  - 13 GitHub API tools (repos, issues, stars, search, user)
  - Real-time streaming support (list_repos, list_issues, search_repos)
  - 100% test coverage with automated test suite
  - Beautiful colored terminal output
  - Complete curl test examples
  - Claude Desktop integration guide
  - OAuth2 token support
  - Rate limit monitoring

#### 🏗️ Architecture Improvements
- **Modular Auth Package** - Clean separation of authentication concerns
- **Provider Interface** - Standardized auth provider interface
- **Auth Manager** - Centralized multi-provider management
- **Metrics Recorder Interface** - Allows auth package independence from observability
- **Better Error Handling** - Context-rich errors with proper error wrapping
- **Graceful Shutdown** - Proper cleanup for auth providers and database connections

### Changed

#### Framework Improvements
- **Enhanced Server Options** - New auth and color configuration options
- **Improved Logging** - Colored output integrated with structured logging
- **Better Configuration** - Support for auth provider configuration
- **Updated Examples** - All examples now use colored output
- **Documentation** - Comprehensive auth and color system documentation

#### Backend Enhancements
- **Streaming Tools** - Better separation of regular and streaming tools
- **Tool Registration** - Improved tool builder with parameter validation
- **Error Responses** - Enhanced error messages with context

#### Observability Updates
- **Metrics Expansion** - 6 new auth-specific metric types
- **Health Check System** - New auth-focused health check functions
- **Logging Enhancement** - Colored output for better readability

### Fixed

- **Tool Registration Bug** - Fixed closure issue in streaming tool registration that caused all tools to reference the last tool
- **Parameter Building** - Fixed parameter addition after Build() call (now builds before)
- **Streaming Emitter** - Added missing Context() method to emitter adapter
- **CORS Handling** - Improved CORS headers for cross-origin requests
- **Graceful Shutdown** - Better cleanup of streaming connections

### Documentation

- **Comprehensive README** - Updated with all v0.3.0 features
- **Auth Guide** - Complete authentication setup and usage guide
- **Color System Guide** - Terminal output customization documentation
- **GitHub Server Example** - Full production-ready example with tests
- **API Documentation** - Updated endpoint documentation
- **Migration Notes** - Guidance for upgrading from v0.2.0
- **Architecture Diagrams** - Updated to show auth and color systems

### Performance

- **Auth Operations** - ~15,000 validations/second
- **Color Rendering** - ~150,000 renders/second (minimal overhead)
- **Token Encryption** - AES-256-GCM with optimized performance
- **Memory Usage** - +3KB per auth provider (negligible overhead)

### Dependencies

No new external dependencies. All new features use existing Go standard library and already-included packages.

### Breaking Changes

**None** - v0.3.0 is fully backward compatible with v0.2.0. All existing code will continue to work without modifications.

### Security

- **Encrypted Token Storage** - All OAuth2 tokens encrypted at rest using AES-256-GCM
- **Secure Defaults** - HTTPS-only for OAuth2 redirects in production
- **Token Expiry** - Automatic monitoring and refresh before expiry
- **Resource Scoping** - Fine-grained access control per resource

### Migration from v0.2.0

No migration needed! v0.3.0 is fully compatible. New features are opt-in:

```go
// v0.2.0 code works as-is
server := framework.NewServer(
    framework.WithBackendType("mybackend"),
    framework.WithTransport("http"),
)

// Add v0.3.0 features when ready
server := framework.NewServer(
    framework.WithBackendType("mybackend"),
    framework.WithTransport("http"),
    framework.WithAutoColors(),  // New: colored output
    framework.WithGitHub(...),   // New: OAuth2 auth
)
```

---

## [0.2.0] - 2026-01-17

### Added
- **Real-Time Streaming** - Server-Sent Events (SSE) support
- **Streaming Tools** - RegisterStreamingTool() for real-time results
- **Progress Updates** - EmitProgress() for live status tracking
- **SSE Endpoint** - `/stream?tool=<name>` for streaming requests
- **Concurrent Control** - Semaphore-based execution limits
- **Event System** - Start, Data, Progress, End, Error events
- **Streaming Metrics** - Track active streams and event counts
- **Examples** - Grep server with streaming search

### Changed
- Enhanced tool builder with Streaming() option
- Improved observability with streaming metrics
- Better error handling for streaming operations

### Fixed
- Memory leaks in long-running streams
- Context cancellation in streaming operations

---

## [0.1.0] - 2026-01-15

### Added
- **Initial Release** - Production-ready MCP framework
- **Multiple Transports** - stdio and HTTP support
- **Backend Registry** - Plugin system for backends
- **Tool Builder** - Fluent API for tool definition
- **Observability** - Prometheus metrics and structured logging
- **Security** - Path validation and sandboxing
- **Configuration** - YAML, environment, and code configuration
- **Examples** - Filesystem and weather servers

---

## Legend

- 🔐 Security features
- 🎨 UI/UX improvements
- 📊 Observability enhancements
- 🏗️ Architecture changes
- 📚 Documentation updates
- ⚡ Performance improvements
- 🐛 Bug fixes

---

## Upgrade Guide

### From v0.2.0 to v0.3.0

**No breaking changes!** Simply update your dependency:

```bash
go get -u github.com/SaherElMasry/go-mcp-framework@v0.3.0
```

**Optional enhancements:**

1. **Add colored output:**
```go
server := framework.NewServer(
    // ... existing options ...
    framework.WithAutoColors(),
)
```

2. **Add OAuth2 authentication:**
```go
server := framework.NewServer(
    // ... existing options ...
    framework.WithGitHub(clientID, clientSecret, redirectURL, scopes),
)
```

3. **Monitor auth metrics:**
```
# New metrics available at :9091/metrics
mcp_auth_validations_total
mcp_auth_token_refresh_total
mcp_oauth2_flows_total
```

### From v0.1.0 to v0.3.0

Update dependency and optionally add streaming support:

```bash
go get -u github.com/SaherElMasry/go-mcp-framework@v0.3.0
```

**To add streaming:**
```go
b.RegisterStreamingTool(
    backend.NewTool("my_tool").
        Streaming(true).
        Build(),
    myStreamingHandler,
)
```

---

## Support

- **Issues:** https://github.com/SaherElMasry/go-mcp-framework/issues
- **Discussions:** https://github.com/SaherElMasry/go-mcp-framework/discussions
- **Documentation:** https://pkg.go.dev/github.com/SaherElMasry/go-mcp-framework

---

## Contributors

Thank you to everyone who contributed to v0.3.0!

- [@SaherElMasry](https://github.com/SaherElMasry) - Main development
- Community feedback and testing
- Early adopters who tested GitHub server

---

**Full Changelog:** https://github.com/SaherElMasry/go-mcp-framework/compare/v0.2.0...v0.3.0
