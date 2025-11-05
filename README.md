# go-slogs

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Enhanced structured logging for Go built on `log/slog` with middleware support, Sugar API, and advanced context management.

## Features

- **Full slog Compatibility**: Drop-in replacement for `log/slog`
- **Handler Middleware**: Composable handlers for processing pipelines
- **Sugar API**: Zap-like formatted logging methods
- **Named Loggers**: Hierarchical logger naming for better organization
- **Context Attributes**: Automatic attribute propagation via Go contexts
- **Standard Log Redirect**: Redirect `log` package output to structured logging

## Installation

```bash
go get github.com/rockcookies/go-slogs
```

## Quick Start

```go
package main

import (
    "log/slog"
    "os"
    "github.com/rockcookies/go-slogs"
)

func main() {
    // Create handler and logger
    handler := slogs.NewHandler(slog.NewJSONHandler(os.Stdout, nil))
    logger := slogs.New(handler)

    // Basic logging
    logger.Info("Hello, World!", "user", "alice")

    // Named logger
    dbLogger := logger.Named("database")
    dbLogger.Info("Connected") // Output: [database] Connected

    // Sugar API
    sugar := logger.Sugar()
    sugar.Infof("User %s logged in", "alice")

    // Context attributes
    ctx := slogs.Prepend(context.Background(), "request_id", "123")
    logger.InfoContext(ctx, "Request processed")
}
```

## Key Concepts

### Named Loggers

Create hierarchical loggers for better organization:

```go
// Create named loggers
dbLogger := slogs.New(handler, slogs.WithName("database"))
apiLogger := slogs.New(handler, slogs.WithName("api"))

// Nested naming
poolLogger := dbLogger.Named("pool") // [database.pool]
```

### Context Attributes

Propagate attributes through contexts:

```go
ctx := slogs.Prepend(ctx, "request_id", "123")
ctx = slogs.Append(ctx, "duration", "100ms")
logger.InfoContext(ctx, "Request completed")
// Output: {"request_id":"123","duration":"100ms","msg":"Request completed"}
```

### Sugar API

Convenient formatted logging:

```go
sugar := logger.Sugar()
sugar.Info("Simple message")
sugar.Infof("Formatted %s", "message")
sugar.Infow("With fields", "key", "value")
```

### Standard Log Redirection

Redirect standard library log calls:

```go
restore, err := slogs.RedirectStdLogAt(logger, slog.LevelInfo)
if err != nil {
    log.Fatal(err)
}
defer restore()

log.Print("This goes through slogs")
```

## Middleware Integration

Use with slog-multi for handler pipelines:

```go
import slogmulti "github.com/samber/slog-multi"

logger := slog.New(
    slogmulti.Pipe(
        slogs.NewMiddleware(&slogs.HandlerOptions{}),
        // other middleware...
    ).Handler(slog.NewJSONHandler(os.Stdout, nil)),
)
```

## Configuration

```go
logger := slogs.New(handler,
    slogs.WithName("myapp"),
    slogs.WithCaller(true),
    slogs.WithLevel(slog.LevelInfo),
)

// Custom handler processing
options := &slogs.HandlerOptions{
    HandleFunc: func(ctx context.Context, hc *slogs.HandlerContext,
                     rt time.Time, rl slog.Level, rm string,
                     attrs []slog.Attr) (string, []slog.Attr) {
        // Custom logic (filtering, transformation, etc.)
        return rm, attrs
    },
}
handler := slogs.NewHandlerWithOptions(baseHandler, options)
```

## Migration from slog

Replace existing slog usage with minimal changes:

```go
// Before
logger := slog.New(handler)
logger.Info("message", "key", "value")

// After
logger := slogs.New(slogs.NewHandler(handler))
logger.Info("message", "key", "value")
```

## Performance

- **Zero allocation** for attribute extraction from contexts
- **Minimal overhead** compared to standard slog
- **Caller info optional** - disable in performance-critical paths
- **Use LogAttrs** when you already have `slog.Attr` values

## API Overview

### Core Methods
- `New(h slog.Handler, opts ...Option) *Logger`
- `With(args ...any) *Logger`
- `WithGroup(name string) *Logger`
- `Sugar() *SugaredLogger`
- `Named(name string) *Logger`

### Context Functions
- `Prepend(ctx, args...) context.Context`
- `Append(ctx, args...) context.Context`
- `RedirectStdLogAt(logger, level) (func(), error)`

For detailed API documentation, see [GoDoc](https://pkg.go.dev/github.com/rockcookies/go-slogs).

## Comparison with log/slog

| Feature | log/slog | go-slogs |
|---------|----------|----------|
| Basic Logging | ✅ | ✅ |
| Handler Middleware | ❌ | ✅ |
| Sugar API | ❌ | ✅ |
| Named Loggers | ❌ | ✅ |
| Context Attributes | ❌ | ✅ |
| Log Redirection | ❌ | ✅ |

## Best Practices

1. Use named loggers for module identification
2. Leverage context for request-scoped attributes
3. Choose Logger for simple cases, Sugar for formatting
4. Use `LogAttrs` in performance-critical code
5. Disable caller info in high-frequency logging paths

## Testing

```bash
go test -v ./...
go test -race -v ./...
go test -v -cover ./...
```

## Dependencies

- Go 1.21+
- `log/slog` (standard library)
- `github.com/stretchr/testify` (testing only)

## License

MIT License - see [LICENSE](LICENSE) file.

## Contributing

Issues and pull requests are welcome.