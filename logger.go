// Package slogs provides enhanced logging capabilities built on top of Go's standard library log/slog.
//
// It offers middleware architecture for Handler composition, named loggers for better log organization,
// context-based attribute management, and a sugar API with formatting support for more convenient logging.
//
// # Basic Usage
//
//	handler := slogs.NewHandler(slog.NewJSONHandler(os.Stdout, nil))
//	logger := slogs.New(handler)
//	logger.Info("Hello, World!", "user", "alice")
//
// # Named Loggers
//
// Create named loggers to identify log sources:
//
//	dbLogger := slogs.New(handler, slogs.WithName("database"))
//	dbLogger.Info("Connected to database")  // Output: [database] Connected to database
//
// # Context Attributes
//
// Add attributes to context for automatic inclusion in logs:
//
//	ctx := slogs.Prepend(context.Background(), "request_id", "abc-123")
//	logger.InfoContext(ctx, "Request processed")
//
// # Sugar Logger
//
// Use Sugar API for formatted logging:
//
//	sugar := logger.Sugar()
//	sugar.Infof("User %s logged in from %s", "alice", "192.168.1.1")
//
// For more details, see https://github.com/rockcookies/go-slogs
package slogs

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"github.com/rockcookies/go-slogs/internal/attr"
)

// A Logger records structured information about each call to its logging methods.
//
// Logger wraps a Handler and provides convenient methods for structured logging at different levels.
// It supports the standard log levels (Debug, Info, Warn, Error) with both simple and context-aware variants.
//
// Unlike slog.Logger, this Logger automatically captures caller information by default
// and supports additional features like named loggers and custom options.
//
// Create a Logger with New, or derive a new Logger from an existing one using With, WithGroup, or WithOptions.
type Logger struct {
	handler    *Handler
	addCaller  bool
	callerSkip int
}

// New creates a new Logger with the given Handler and options.
//
// The handler must not be nil, or this function will panic.
// Options can be used to configure caller tracking, log levels, and logger names.
//
// Example:
//
//	handler := slogs.NewHandler(slog.NewJSONHandler(os.Stdout, nil))
//	logger := slogs.New(handler,
//		slogs.WithName("myapp"),
//		slogs.WithLevel(slog.LevelInfo),
//	)
func New(h *Handler, options ...Option) *Logger {
	if h == nil {
		panic("nil Handler")
	}

	l := &Logger{
		handler:    h,
		addCaller:  true,
		callerSkip: 0,
	}

	for _, opt := range options {
		opt.apply(l)
	}

	return l
}

// clone creates a shallow copy of l.
func (l *Logger) clone() *Logger {
	l2 := *l
	return &l2
}

// Handler returns the Logger's Handler.
//
// This is useful when you need to access the underlying handler,
// for example to use it with other logging libraries or middleware.
func (l *Logger) Handler() *Handler {
	return l.handler
}

// With returns a new Logger that includes the given attributes in each output operation.
//
// The arguments are converted to attributes using the same rules as Logger.Log:
//   - If an argument is an slog.Attr, it is used as is
//   - If an argument is a string followed by any value, they form a key-value pair
//   - Otherwise, the argument is treated as a value with key "!BADKEY"
//
// The returned Logger shares the same handler but with additional attributes.
// If no arguments are provided, the original logger is returned.
//
// Example:
//
//	logger := logger.With("app", "myapp", "env", "prod")
//	logger.Info("Server started")  // Will include app=myapp and env=prod
func (l *Logger) With(args ...any) *Logger {
	if len(args) == 0 {
		return l
	}

	l2 := l.clone()
	l2.handler = l2.handler.withAttrs(attr.ArgsToAttrSlice(args))
	return l2
}

// WithGroup returns a new Logger that starts a group.
//
// If name is non-empty, all attributes added to the returned Logger will be
// nested under a group with the given name. How this qualification appears
// depends on the Handler implementation.
//
// If name is empty, WithGroup returns the receiver unchanged.
//
// Example:
//
//	logger := logger.WithGroup("http")
//	logger.Info("Request", "method", "GET", "path", "/api")
//	// Output: {"http":{"method":"GET","path":"/api"},"msg":"Request"}
func (l *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return l
	}

	l2 := l.clone()
	l2.handler = l2.handler.withGroup(name)
	return l2
}

// Enabled reports whether the Logger emits log records at the given level.
//
// This can be used to avoid expensive operations when a log statement
// would not produce output.
//
// If ctx is nil, context.Background() is used.
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	return l.handler.Enabled(ctx, level)
}

// WithOptions returns a new Logger with the given options applied.
//
// This allows you to create a logger variant with modified behavior,
// such as different caller skip levels, names, or log levels.
//
// Example:
//
//	logger2 := logger.WithOptions(
//		slogs.WithName("worker"),
//		slogs.WithLevel(slog.LevelDebug),
//	)
func (l *Logger) WithOptions(opts ...Option) *Logger {
	l2 := l.clone()
	for _, opt := range opts {
		opt.apply(l2)
	}
	return l2
}

// Sugar returns a SugarLogger that wraps this Logger.
//
// SugarLogger provides a more ergonomic API with Sprint-style and Sprintf-style methods.
// Use Sugar when you need printf-style formatting or when you want a more concise API.
//
// Example:
//
//	sugar := logger.Sugar()
//	sugar.Infof("User %s logged in from %s", username, ipAddr)
func (l *Logger) Sugar() *SugarLogger {
	return &SugarLogger{base: l}
}

// Log emits a log record with the current time and the given level and message.
//
// The Record's attributes consist of the Logger's attributes followed by
// the attributes specified by args.
//
// The args are processed as follows:
//   - If an argument is an slog.Attr, it is used as is
//   - If an argument is a string and not the last argument, the following argument
//     is treated as the value and the two are combined into an Attr
//   - Otherwise, the argument is treated as a value with key "!BADKEY"
//
// If ctx is nil, context.Background() is used.
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

// LogAttrs is a more efficient version of Logger.Log that accepts only Attrs.
//
// Use this method when you already have slog.Attr values and want to avoid
// the overhead of argument parsing.
//
// If ctx is nil, context.Background() is used.
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

// Debug logs at LevelDebug with the given message and attributes.
func (l *Logger) Debug(msg string, args ...any) {
	l.log(context.Background(), slog.LevelDebug, msg, args...)
}

// DebugContext logs at LevelDebug with the given context, message, and attributes.
//
// The context can contain additional attributes added via Prepend or Append.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args...)
}

// Info logs at LevelInfo with the given message and attributes.
func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, msg, args...)
}

// InfoContext logs at LevelInfo with the given context, message, and attributes.
//
// The context can contain additional attributes added via Prepend or Append.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args...)
}

// Warn logs at LevelWarn with the given message and attributes.
func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, msg, args...)
}

// WarnContext logs at LevelWarn with the given context, message, and attributes.
//
// The context can contain additional attributes added via Prepend or Append.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args...)
}

// Error logs at LevelError with the given message and attributes.
func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), slog.LevelError, msg, args...)
}

// ErrorContext logs at LevelError with the given context, message, and attributes.
//
// The context can contain additional attributes added via Prepend or Append.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args...)
}

func (l *Logger) capturePC() uintptr {
	var pc uintptr
	if l.addCaller {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, log function, this function's caller]
		runtime.Callers(4+l.callerSkip, pcs[:])
		pc = pcs[0]
	}
	return pc
}

// log is the internal logging method.
func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !l.Enabled(ctx, level) {
		return
	}

	pc := l.capturePC()
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attr.ArgsToAttrSlice(args)...)

	_ = l.handler.Handle(ctx, r)
}

// logAttrs is the internal logging method that accepts only Attrs.
func (l *Logger) logAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !l.Enabled(ctx, level) {
		return
	}

	pc := l.capturePC()
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)

	_ = l.Handler().Handle(ctx, r)
}
