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

# Run tests with race detection (recommended for development)
go test -race -v ./...

# Run specific test file
go test -v ./logger_test.go

# Run tests for specific package
go test -v ./internal/attr

# Run benchmarks
go test -bench=. -benchmem ./...

# Run specific test function
go test -v -run TestLoggerInfo ./...

# Run tests with verbose output and specific timeout
go test -v -timeout 30s ./...
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

# Verify module integrity
go mod verify
```

### CI/CD Commands
```bash
# Run tests with race detection (as used in CI)
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Run golangci-lint locally
golangci-lint run

# Upload coverage to Codecov (requires token)
go tool cover -html=coverage.out -o coverage.html
# Or use codecov CLI if installed
codecov -f coverage.out
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

3. **SugaredLogger (`sugar.go`)** - Zap-like Sugar API for convenient logging
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

7. **Stack Trace System (`attr.go`, `internal/stacktrace/`)** - Stack capture and formatting
   - `Stack()` and `StackSkip()` functions for creating stack trace attributes
   - Efficient stack trace capture with pooling
   - Formatted stack trace output similar to zap

8. **Buffer Pool System (`buffer/`, `internal/pool/`)** - High-performance buffer management
   - Generic type-safe pool implementation
   - Reduces memory allocations for string operations
   - Used internally by stack trace formatting

9. **Clock Abstraction (`clock.go`)** - Time source abstraction for testing
   - `Clock` interface with Now() and NewTicker() methods
   - Default `systemClock` implementation using system time
   - Enables time mocking in tests

#### Key Design Patterns

- **Middleware Pattern**: Handler wraps another handler for processing pipelines
- **Linked List**: Attribute groups use linked list for efficient composition
- **Interface Segregation**: Separate Logger and SugaredLogger interfaces
- **Context Propagation**: Uses Go context for automatic attribute inclusion
- **Object Pooling**: Stack traces and buffers use pooling to minimize allocations
- **Immutable Pattern**: Logger and Handler derivation uses cloning to prevent race conditions
- **Clock Abstraction**: Time source abstraction enables testable code

### Handler Flow

1. Log record created by Logger with caller information and timestamp using Clock
2. Record attributes extracted and converted to `slog.Attr` slice
3. `HandleFunc` processes attributes in this order:
   - Appends context attributes from `Append()` to the end
   - Processes attribute group chain (newest to oldest), applying nesting
   - Prepends context attributes from `Prepend()` to the start
   - Prefixes message with logger names (e.g., "[service.database] message")
4. New record created with processed message and attributes
5. Next handler in chain receives processed record

### Middleware Pattern

The Handler implements a middleware pattern that allows for:

- **Handler Composition**: Multiple handlers can be chained together
- **Attribute Transformation**: Custom `HandleFunc` can modify/filter attributes
- **Context Injection**: Automatic inclusion of request-scoped attributes
- **Name Propagation**: Logger names flow through the handler chain

**Example Pipeline:**
```
Logger → slogs.Handler → slog-multi.Pipe → slog.NewJSONHandler → Output
```

### slog-multi Integration

The library provides `NewMiddleware()` for creating slog-multi compatible handlers:

```go
slog.SetDefault(slog.New(
    slogmulti.Pipe(
        slogs.NewMiddleware(&slogs.HandlerOptions{}),  // Add context/names
        customMiddleware,                               // Custom processing
    ).Handler(
        slog.NewJSONHandler(os.Stdout, nil),           // Final output
    ),
))
```

## Performance Considerations

### General Guidelines

- **Use `LogAttrs` over `Log`** when you already have `slog.Attr` values to avoid argument conversion overhead
- **Check `Enabled()` before expensive operations** to avoid unnecessary work
- **Disable caller info** in performance-critical paths using `slogs.WithCaller(false)`
- **Use context attributes** for automatic inclusion without manual attribute passing

### Caller Information Overhead

The Logger captures caller information by default (`runtime.Callers(4+callerSkip, pcs[:])`). This adds:
- Stack trace traversal cost
- PC resolution to function/file/line information
- Additional memory allocation for PC storage

**Mitigation:** Use `WithCaller(false)` for high-frequency logging paths.

### Context Attribute Performance

- **`Prepend()`** attributes are added at the root level and incur minimal overhead
- **`Append()`** attributes respect group structure but require additional processing
- Context attribute extraction is O(1) using type assertions
- Multiple calls to `Prepend()`/`Append()` create new context values (immutable)

### Handler Processing Cost

The `DefaultHandleFunc` processes attributes in this order:
1. Appended context attributes (linear append)
2. Attribute group chain traversal (linked list, O(n) where n = groups)
3. Prepended context attributes (linear prepend)
4. Name prefixing (string concatenation)

**Optimization:** Minimize the number of attribute groups and context attributes in hot paths.

### Memory Allocation Patterns

- **Log records**: Allocate new `slog.Record` for each log entry
- **Attribute slices**: Create new slices during processing (immutable pattern)
- **Handler cloning**: Deep copy mutable state when deriving handlers
- **Sugar formatting**: May allocate for `Sprintf` operations
- **Buffer pooling**: Reduces allocations for string operations via buffer reuse
- **Stack trace capture**: Uses pooling to minimize allocation overhead

### Best Practices

1. **Reuse loggers** rather than creating new ones for each log entry
2. **Pool expensive attribute creation** for frequently logged data
3. **Consider async logging** for high-throughput scenarios
4. **Profile with `pprof`** to identify actual bottlenecks in production

## Testing Approach

The project uses comprehensive test coverage:

### Test Structure
- **Unit tests for each component** (`*_test.go` files)
- **Tests cover both positive and negative cases**
- **Mock handlers for testing handler behavior**
- **Context attribute propagation testing**

### Mock Handler Pattern

Tests use a custom mock handler to verify behavior:

```go
type mockHandler struct {
    records []slog.Record
    enabled bool
}

func (m *mockHandler) Handle(ctx context.Context, r slog.Record) error {
    m.records = append(m.records, r)
    return nil
}
```

This pattern allows testing of:
- Attribute processing order
- Context attribute extraction
- Name prefixing behavior
- Group nesting structure

### Race Condition Testing

All CI tests run with `-race` flag to detect:
- Data races in concurrent logging
- Handler cloning race conditions
- Context attribute mutation issues
- Logger sharing between goroutines

### Coverage Requirements

- CI generates coverage with `go test -coverprofile=coverage.out -covermode=atomic`
- Coverage uploaded to Codecov for tracking
- Atomic coverage mode ensures accurate measurements in concurrent tests

### Test Patterns

**Context Testing:**
```go
func TestContextPrependAppend(t *testing.T) {
    ctx := Prepend(context.Background(), "request_id", "123")
    ctx = Append(ctx, "duration", "100ms")
    // Verify order: request_id first, duration last
}
```

**Handler Chain Testing:**
```go
func TestHandlerChaining(t *testing.T) {
    base := &mockHandler{}
    handler := NewHandler(base)
    // Test handler processes attributes correctly
}
```

**Named Logger Testing:**
```go
func TestNamedLogger(t *testing.T) {
    logger := New(handler, WithName("service"))
    // Verify name appears in output: "[service] message"
}
```

## Architectural Principles

### Design Philosophy

The library follows several key architectural principles:

1. **Standard Library Compatibility**: Drop-in replacement for `log/slog` - existing code works unchanged
2. **Middleware-First**: Handler designed for composable processing pipelines
3. **Performance Conscious**: Minimal allocations beyond standard slog, with pooling optimizations
4. **Testability**: Clock abstraction and immutable patterns enable comprehensive testing
5. **Context-First**: Leverages Go contexts for automatic attribute propagation
6. **Type Safety**: Generic pool implementation and strong typing throughout

### Memory Management

- **Immutable Patterns**: Logger and Handler cloning prevents race conditions
- **Object Pooling**: Stack traces and buffers reuse objects to minimize GC pressure
- **Zero-Allocation Extraction**: Context attributes extracted with O(1) cost via type assertions
- **Linked List Efficiency**: Attribute groups use linked lists for efficient composition

### Concurrency Safety

- **Race Condition Free**: All components designed for safe concurrent usage
- **Immutable Context**: Context values are never mutated, only replaced
- **Handler Cloning**: Deep copy mutable state when deriving handlers
- **Atomic Operations**: Pool operations use atomic patterns for safety

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


## Development Workflow

### Pre-commit Validation

Before committing changes, run the following local validation:

```bash
# Run full test suite with race detection
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Run linter
golangci-lint run

# Verify module integrity
go mod verify
go mod tidy
```

### Branch Strategy

- **main**: Stable releases (tags like v1.0.0)
- **dev**: Development branch for features
- **feature/***: Individual feature branches
- **fix/***: Bug fix branches

### Release Process

Releases are automated via GitHub Actions:

1. **Release Drafter**: Creates draft releases from PRs
2. **Semantic Versioning**: Follows SemVer for versioning
3. **Git Tags**: Tags created automatically on release
4. **Go Module**: Version reflected in go.mod

### Version Requirements

- **Go 1.21+**: Minimum required version
- **Standard Library Only**: No external runtime dependencies
- **Testing Dependencies**: `github.com/stretchr/testify` for tests only

### Local Development Setup

```bash
# Clone repository
git clone https://github.com/rockcookies/go-slogs.git
cd go-slogs

# Install dependencies
go mod download

# Run tests to verify setup
go test -v ./...

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Code Quality Standards

- **Race Condition Free**: All code must pass `-race` testing
- **No Linting Issues**: Must pass `golangci-lint run`
- **Test Coverage**: Maintain high coverage for critical paths
- **Documentation**: Public APIs must have godoc comments
- **Examples**: Include usage examples in documentation

## Dependencies

- Go 1.21+ (required)
- `log/slog` (standard library)
- `github.com/stretchr/testify` (testing only)

## File Structure

```
.
├── attr.go            # Stack trace attribute functions
├── attrs.go           # Attribute grouping linked list
├── clock.go           # Clock abstraction for testing
├── context.go         # Context attribute management
├── handler.go         # Middleware handler implementation
├── logger.go          # Main Logger implementation
├── option.go          # Configuration options
├── redirect.go        # Standard library log redirection
├── sugar.go           # Sugar API implementation
├── buffer/            # Buffer pool implementation
│   ├── buffer.go
│   └── pool.go
├── internal/
│   ├── attr/
│   │   └── attr.go    # Argument conversion utilities
│   ├── bufferpool/
│   │   └── bufferpool.go # Internal buffer pool
│   ├── pool/
│   │   └── pool.go    # Generic pool implementation
│   └── stacktrace/
│       └── stack.go   # Stack trace capture and formatting
└── *_test.go          # Comprehensive test suite
```

## Additional Features

### Standard Library Log Redirection

The library provides `RedirectStdLogAt` function to redirect the standard library's global logger to slogs with specific level handling:

```go
// Redirect standard lib log to slogs at Info level
restore, err := slogs.RedirectStdLogAt(logger, slog.LevelInfo)
if err != nil {
    log.Fatal(err)
}
defer restore() // Restore original logger

// Standard library log calls now go through slogs
log.Print("This will be handled by slogs")
```

This feature automatically handles caller information and disables the standard library's annotations to avoid duplication.

### Stack Trace Attributes

The library provides stack trace attribute creation for debugging:

```go
// Add stack trace to log entries
logger.Error("Database connection failed", slogs.Stack("stack"))

// Skip frames when capturing stack trace
logger.Error("Handler error", slogs.StackSkip("stack", 2)) // skip 2 frames
```

Stack trace capture is optimized with pooling to minimize performance impact.

### Clock Abstraction and Testing

The library provides a Clock abstraction for testable time-dependent code:

```go
// DefaultClock uses system time
logger := slogs.New(handler)

// For testing, you can create a mock clock
type mockClock struct {
    now time.Time
}

func (m *mockClock) Now() time.Time { return m.now }
func (m *mockClock) NewTicker(d time.Duration) *time.Ticker {
    return time.NewTicker(d) // or implement mock ticker
}

// Use in tests for deterministic timestamps
mock := &mockClock{now: time.Unix(1234567890, 0)}
// Inject mock clock into logger for consistent test results
```