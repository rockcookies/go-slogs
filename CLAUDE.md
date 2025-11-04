# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-slogs is an enhanced structured logging library for Go built on top of the standard library `log/slog`. It provides middleware architecture, named loggers, context attribute management, and a Sugar API similar to uber-go/zap.

## Development Commands

### Testing
```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test file
go test -v ./logger_test.go

# Run tests for specific package
go test -v ./internal/attr
```

### Building
```bash
# Build the module
go build ./...

# Verify module
go mod verify
```

### Module Management
```bash
# Tidy dependencies
go mod tidy

# Download dependencies
go mod download
```

## Architecture Overview

### Core Components

1. **Logger (`logger.go`)** - Main logging interface with enhanced features
   - Wraps slog.Handler with additional functionality
   - Supports named loggers for better identification
   - Provides both regular and context-aware logging methods

2. **Handler (`handler.go`)** - Middleware handler that processes log records
   - Implements `slog.Handler` interface for middleware compatibility
   - Manages attribute groups and context attributes
   - Supports custom processing via `HandleFunc`
   - Designed to work with handler pipelines (e.g., slog-multi)

3. **SugarLogger (`sugar.go`)** - Zap-like Sugar API for convenient logging
   - Provides both Sprint-style and Sprintf-style methods
   - Wraps the main Logger for backward compatibility
   - Offers `Desugar()` method to convert back to regular Logger

4. **Context Management (`context.go`)** - Context-based attribute propagation
   - `Prepend()` - Adds attributes at the root level
   - `Append()` - Adds attributes respecting current group structure
   - Extract functions for custom handler implementations

5. **Attribute System (`attrs.go`, `internal/attr/`)** - Linked list for attribute grouping
   - `GroupOrAttrs` structure for efficient attribute/group chain management
   - Copied from slog source for consistency
   - Supports nested attribute groups

6. **Options (`option.go`)** - Configuration system
   - Caller information control
   - Log level management
   - Named logger configuration

### Key Design Patterns

- **Middleware Pattern**: Handler wraps another handler for processing pipelines
- **Linked List**: Attribute groups use linked list for efficient composition
- **Interface Segregation**: Separate Logger and SugarLogger interfaces
- **Context Propagation**: Uses Go context for automatic attribute inclusion

### Handler Flow

1. Log record created by Logger
2. Record attributes extracted
3. `HandleFunc` processes attributes (adds context, applies grouping, adds names)
4. New record created with processed attributes
5. Next handler in chain receives processed record

## Testing Approach

The project uses comprehensive test coverage:
- Unit tests for each component (`*_test.go` files)
- Tests cover both positive and negative cases
- Mock handlers for testing handler behavior
- Context attribute propagation testing

## Integration Notes

### slog-multi Compatibility
The library provides `NewMiddleware()` function for creating slog-multi compatible middleware:

```go
slog.SetDefault(slog.New(
    slogmulti.Pipe(slogs.NewMiddleware(&slogs.HandlerOptions{})).
    Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})),
))
```

### Context Usage
Use context for request-scoped attributes:
```go
ctx := slogs.Prepend(context.Background(), "request_id", "abc-123")
logger.InfoContext(ctx, "Processing request")
```

### Named Loggers
Create named loggers for better organization:
```go
dbLogger := slogs.New(handler, slogs.WithName("database"))
```

## Performance Considerations

- Use `LogAttrs` method when you already have `slog.Attr` values
- Use `Enabled()` check before expensive operations
- Consider disabling caller info for performance-critical paths
- Use context attributes for automatic inclusion without manual passing

## Dependencies

- Go 1.21+ (required)
- `log/slog` (standard library)
- `github.com/stretchr/testify` (testing only)

## File Structure

```
.
├── attrs.go           # Attribute grouping linked list
├── context.go         # Context attribute management
├── handler.go         # Middleware handler implementation
├── logger.go          # Main Logger implementation
├── option.go          # Configuration options
├── sugar.go           # Sugar API implementation
├── internal/
│   └── attr/
│       └── attr.go    # Argument conversion utilities
└── *_test.go          # Comprehensive test suite
```