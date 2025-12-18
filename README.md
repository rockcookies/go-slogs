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
- **Stack Trace Support**: Built-in stack trace attribute creation
- **MultiHandler**: Broadcast logs to multiple handlers simultaneously
- **Standard Log Redirect**: Redirect `log` package output to structured logging

## Installation

```bash
go get github.com/rockcookies/go-slogs
```

## Quick Start

```go
package main

import (
    "context"
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

    // MultiHandler - broadcast to multiple handlers
    h1 := slog.NewJSONHandler(os.Stdout, nil)
    h2 := slog.NewTextHandler(os.Stderr, nil)
    multi := slogs.MultiHandler(h1, h2)
    logger := slog.New(multi)
    logger.Info("This log will be output to both stdout and stderr")

    // Stack trace for debugging
    logger.Error("Something went wrong", slogs.Stack("stack"))
}
```

## Key Concepts

### Named Loggers

Create hierarchical loggers for better organization:

```go
dbLogger := slogs.New(handler).Named("database")
poolLogger := dbLogger.Named("pool") // [database.pool]
```

### Context Attributes

Propagate attributes through contexts:

```go
ctx := slogs.Prepend(ctx, "request_id", "123")
logger.InfoContext(ctx, "Request completed")
```

### Sugar API

Convenient formatted logging:

```go
sugar := logger.Sugar()
sugar.Infof("User %s logged in", "alice")
```

### MultiHandler

Broadcast log records to multiple handlers simultaneously:

```go
// Create multiple handlers
h1 := slog.NewJSONHandler(os.Stdout, nil)      // JSON to stdout
h2 := slog.NewTextHandler(os.Stderr, nil)      // Text to stderr
fileHandler := slog.NewJSONHandler(file, nil)   // JSON to file

// Combine with MultiHandler
multi := slogs.MultiHandler(h1, h2, fileHandler)
logger := slog.New(multi)

// This log will be written to all three handlers
logger.Info("Broadcast to multiple outputs")

// MultiHandler properly handles:
// - Nil handler filtering (nil handlers are automatically removed)
// - Record isolation (each handler gets a cloned copy)
// - Attribute independence (WithAttrs/WithGroup applied per handler)
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

## Configuration

```go
logger := slogs.New(handler).
    Named("myapp").
    WithOptions(
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

## API Overview

### Core Methods
- `New(h slog.Handler, opts ...Option) *Logger`
- `With(args ...any) *Logger`
- `WithGroup(name string) *Logger`
- `Sugar() *SugaredLogger`
- `Named(name string) *Logger`
- `MultiHandler(handlers ...slog.Handler) slog.Handler`

### Context Functions
- `Prepend(ctx, args...) context.Context`
- `Append(ctx, args...) context.Context`
- `RedirectStdLogAt(logger, level) (func(), error)`

### Stack Trace Functions
- `Stack(key string) slog.Attr`
- `StackSkip(key string, skip int) slog.Attr`

## Performance

- **Zero allocation** for attribute extraction from contexts
- **Minimal overhead** compared to standard slog
- **Caller info optional** - disable in performance-critical paths
- **Use LogAttrs** when you already have `slog.Attr` values

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