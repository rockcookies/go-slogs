package slogs

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"github.com/rockcookies/go-slogs/internal/attr"
)

// A Logger records structured information about each call to its
// Log, Debug, Info, Warn, and Error methods.
// For each call, it creates a [Record] and passes it to a [Handler].
//
// To create a new Logger, call [New] or a Logger method
// that begins "With".
type Logger struct {
	handler    *Handler
	addCaller  bool
	callerSkip int
}

// New creates a new Logger with the given non-nil Handler.
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

// Handler returns l's Handler.
func (l *Logger) Handler() *Handler {
	return l.handler
}

// With returns a Logger that includes the given attributes
// in each output operation. Arguments are converted to
// attributes as if by [Logger.Log].
func (l *Logger) With(args ...any) *Logger {
	if len(args) == 0 {
		return l
	}

	l2 := l.clone()
	l2.handler = l2.handler.withAttrs(attr.ArgsToAttrSlice(args))
	return l2
}

// WithGroup returns a Logger that starts a group, if name is non-empty.
// The keys of all attributes added to the Logger will be qualified by the given
// name. (How that qualification happens depends on the [Handler.WithGroup]
// method of the Logger's Handler.)
//
// If name is empty, WithGroup returns the receiver.
func (l *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return l
	}

	l2 := l.clone()
	l2.handler = l2.handler.withGroup(name)
	return l2
}

// Enabled reports whether l emits log records at the given context and level.
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	return l.handler.Enabled(ctx, level)
}

// WithOptions returns a copy of l with the provided options applied.
func (l *Logger) WithOptions(opts ...Option) *Logger {
	l2 := l.clone()
	for _, opt := range opts {
		opt.apply(l2)
	}
	return l2
}

// Sugar returns a sugared logger that wraps l.
func (l *Logger) Sugar() *SugarLogger {
	return &SugarLogger{base: l}
}

// Log emits a log record with the current time and the given level and message.
// The Record's Attrs consist of the Logger's attributes followed by
// the Attrs specified by args.
//
// The attribute arguments are processed as follows:
//   - If an argument is an Attr, it is used as is.
//   - If an argument is a string and this is not the last argument,
//     the following argument is treated as the value and the two are combined
//     into an Attr.
//   - Otherwise, the argument is treated as a value with key "!BADKEY".
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

// LogAttrs is a more efficient version of [Logger.Log] that accepts only Attrs.
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

// Debug logs at [LevelDebug].
func (l *Logger) Debug(msg string, args ...any) {
	l.log(context.Background(), slog.LevelDebug, msg, args...)
}

// DebugContext logs at [LevelDebug] with the given context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args...)
}

// Info logs at [LevelInfo].
func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, msg, args...)
}

// InfoContext logs at [LevelInfo] with the given context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args...)
}

// Warn logs at [LevelWarn].
func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, msg, args...)
}

// WarnContext logs at [LevelWarn] with the given context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args...)
}

// Error logs at [LevelError].
func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), slog.LevelError, msg, args...)
}

// ErrorContext logs at [LevelError] with the given context.
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
