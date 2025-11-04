# go-slogs

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A powerful Go structured logging library built on top of the standard library `log/slog`, providing enhanced logging capabilities and flexible middleware support.

## Features

- 🚀 **Fully Compatible with `log/slog`**: Built on Go's standard library for seamless integration
- 🎯 **Middleware Architecture**: Support for Handler middleware, easy to extend
- 📝 **Sugar Logger**: Provides Zap-like Sugar API with formatting support
- 🏷️ **Named Loggers**: Set logger names for easy identification of log sources
- 📦 **Context Attributes**: Extract and add log attributes from Context
- ⚙️ **Flexible Configuration**: Rich configuration options including log levels, caller info, etc.
- 🎨 **Attribute Grouping**: Group attributes for better log structure organization

## Installation

```bash
go get github.com/rockcookies/go-slogs
```

# go-slogs

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A powerful Go structured logging library built on top of the standard library `log/slog`, providing enhanced logging capabilities and flexible middleware support.

## Features

- 🚀 **Fully Compatible with `log/slog`**: Built on Go's standard library for seamless integration
- 🎯 **Middleware Architecture**: Support for Handler middleware, easy to extend
- 📝 **Sugar Logger**: Provides Zap-like Sugar API with formatting support
- 🏷️ **Named Loggers**: Set logger names for easy identification of log sources
- 📦 **Context Attributes**: Extract and add log attributes from Context
- ⚙️ **Flexible Configuration**: Rich configuration options including log levels, caller info, etc.
- 🎨 **Attribute Grouping**: Group attributes for better log structure organization

## Installation

```bash
go get github.com/rockcookies/go-slogs
```

## Quick Start

### Basic Usage

```go
package main

import (
    "log/slog"
    "os"

    "github.com/rockcookies/go-slogs"
)

func main() {
    // Create a base JSON handler
    baseHandler := slog.NewJSONHandler(os.Stdout, nil)

    // Wrap with slogs Handler
    handler := slogs.NewHandler(baseHandler)

    // Create Logger
    logger := slogs.New(handler)

    // Log messages
    logger.Info("Hello, World!", "user", "alice", "action", "login")
}
```

### Using Sugar Logger

Sugar Logger provides a more concise API with formatting support:

```go
// Get Sugar Logger
sugar := logger.Sugar()

// Sprint style
sugar.Info("User logged in")
sugar.Info("User", "alice", "logged in")

// Sprintf style
sugar.Infof("User %s logged in from %s", "alice", "192.168.1.1")
sugar.Debugf("Processing request %d", 12345)
```

### Named Loggers

Create named loggers for different modules or components:

```go
// Create named loggers
dbLogger := slogs.New(handler, slogs.WithName("database"))
apiLogger := slogs.New(handler, slogs.WithName("api"))

dbLogger.Info("Connected to database")  // Output: [database] Connected to database
apiLogger.Info("Server started")        // Output: [api] Server started

// Nested names
childLogger := dbLogger.WithOptions(slogs.WithName("pool"))
childLogger.Info("Pool initialized")    // Output: [database.pool] Pool initialized
```

### Context Attributes

Pass and extract log attributes using Context:

```go
import "context"

// Add attributes to Context
ctx := context.Background()
ctx = slogs.Prepend(ctx, "request_id", "abc-123")  // Add to beginning
ctx = slogs.Append(ctx, "duration", "100ms")       // Add to end

// Log with Context attributes
logger.InfoContext(ctx, "Request completed", "status", 200)
// Output: {"request_id":"abc-123","status":200,"duration":"100ms","msg":"Request completed"}
```

### Attributes and Grouping

```go
// Add persistent attributes
logger = logger.With("app", "myapp", "env", "production")

// Create attribute groups
logger = logger.WithGroup("http")
logger.Info("Request received", "method", "GET", "path", "/api/users")
// Output: {"app":"myapp","env":"production","http":{"method":"GET","path":"/api/users"},"msg":"Request received"}
```

## Configuration Options

### Logger Options

```go
logger := slogs.New(handler,
    // Enable/disable caller information
    slogs.WithCaller(true),

    // Skip call stack frames (useful when wrapping the logger)
    slogs.WithCallerSkip(1),

    // Set log level
    slogs.WithLevel(slog.LevelInfo),

    // Set logger name
    slogs.WithName("myapp"),

    // Override logger name (clears previous names)
    slogs.WithNameOverride("newapp"),
)
```

### Handler Options

```go
// Custom handle function
options := &slogs.HandlerOptions{
    HandleFunc: func(ctx context.Context, hc *slogs.HandlerContext,
                     rt time.Time, rl slog.Level, rm string,
                     attrs []slog.Attr) (string, []slog.Attr) {
        // Custom log processing logic
        return rm, attrs
    },
}

handler := slogs.NewHandlerWithOptions(baseHandler, options)
```

## Middleware Integration

go-slogs can be used as a slog Handler middleware, for example with `slog-multi`:

```go
import (
    "log/slog"
    "os"

    slogmulti "github.com/samber/slog-multi"
    "github.com/rockcookies/go-slogs"
)

func main() {
    logger := slog.New(
        slogmulti.
            Pipe(slogs.NewMiddleware(&slogs.HandlerOptions{})).
            Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})),
    )

    slog.SetDefault(logger)
}
```

## API Reference

### Logger Methods

- `New(h *Handler, options ...Option) *Logger` - Create a new Logger
- `With(args ...any) *Logger` - Add attributes
- `WithGroup(name string) *Logger` - Create attribute group
- `WithOptions(opts ...Option) *Logger` - Apply options
- `Sugar() *SugarLogger` - Convert to Sugar Logger

**Logging Methods:**
- `Debug(msg string, args ...any)`
- `Info(msg string, args ...any)`
- `Warn(msg string, args ...any)`
- `Error(msg string, args ...any)`
- `DebugContext(ctx context.Context, msg string, args ...any)`
- `InfoContext(ctx context.Context, msg string, args ...any)`
- `WarnContext(ctx context.Context, msg string, args ...any)`
- `ErrorContext(ctx context.Context, msg string, args ...any)`
- `Log(ctx context.Context, level slog.Level, msg string, args ...any)`
- `LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)`

### SugarLogger Methods

- `With(args ...any) *SugarLogger` - Add attributes
- `WithGroup(name string) *SugarLogger` - Create attribute group
- `Desugar() *Logger` - Convert back to standard Logger

**Logging Methods:**
- `Debug(args ...any)` / `Debugf(template string, args ...any)`
- `Info(args ...any)` / `Infof(template string, args ...any)`
- `Warn(args ...any)` / `Warnf(template string, args ...any)`
- `Error(args ...any)` / `Errorf(template string, args ...any)`
- All methods have corresponding `*Context` versions

### Context Functions

- `Prepend(parent context.Context, args ...any) context.Context` - Add attributes to beginning
- `Append(parent context.Context, args ...any) context.Context` - Add attributes to end
- `ExtractPrepended(ctx context.Context) []slog.Attr` - Extract prepended attributes
- `ExtractAppended(ctx context.Context) []slog.Attr` - Extract appended attributes

## Advanced Usage

### Custom Handler Function

```go
options := &slogs.HandlerOptions{
    HandleFunc: func(ctx context.Context, hc *slogs.HandlerContext,
                     rt time.Time, rl slog.Level, rm string,
                     attrs []slog.Attr) (string, []slog.Attr) {
        // Add custom logic, e.g., filter sensitive information
        for i, attr := range attrs {
            if attr.Key == "password" {
                attrs[i].Value = slog.StringValue("***")
            }
        }
        return rm, attrs
    },
}

handler := slogs.NewHandlerWithOptions(baseHandler, options)
```

### Conditional Logging

```go
if logger.Enabled(ctx, slog.LevelDebug) {
    // Only execute expensive operations when Debug level is enabled
    expensiveData := computeExpensiveDebugInfo()
    logger.Debug("Debug info", "data", expensiveData)
}
```

## Comparison with Standard Library

go-slogs provides the following enhancements while maintaining full compatibility with `log/slog`:

| Feature | log/slog | go-slogs |
|---------|----------|----------|
| Basic Logging | ✅ | ✅ |
| Structured Logging | ✅ | ✅ |
| Handler Middleware | ❌ | ✅ |
| Sugar API | ❌ | ✅ |
| Named Loggers | ❌ | ✅ |
| Context Attribute Extraction | ❌ | ✅ |
| Attribute Prepend/Append | ❌ | ✅ |

## Best Practices

1. **Use Named Loggers**: Create named loggers for different modules to facilitate log filtering and analysis
2. **Leverage Context**: Pass request-level attributes (like request_id, trace_id) through Context
3. **Choose the Right API**: Use Logger for simple scenarios, SugarLogger when formatting is needed
4. **Control Log Levels**: Use Info or higher levels in production to avoid excessive Debug logs
5. **Use LogAttrs**: In performance-sensitive scenarios, use `LogAttrs` method to avoid parameter conversion overhead

## Testing

Run tests:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -v -cover ./...
```

## Dependencies

- Go 1.21 or higher
- `log/slog` (Go standard library)
- `github.com/stretchr/testify` (testing only)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Issues and Pull Requests are welcome!

## Acknowledgments

This project draws inspiration from:

- Go standard library `log/slog`
- [uber-go/zap](https://github.com/uber-go/zap) Sugar API
- [jba/slog](https://github.com/jba/slog) Handler implementation ideas

## Author

[@rockcookies](https://github.com/rockcookies)
