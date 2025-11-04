package slogs

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// SugarLogger provides a more ergonomic API for logging with formatting support.
//
// It wraps a Logger and offers both Sprint-style (concatenation) and Sprintf-style (formatting)
// methods for each log level. This is similar to the Sugar logger in uber-go/zap.
//
// Use Desugar to convert back to a regular Logger when needed.
//
// Example:
//
//	sugar := logger.Sugar()
//	sugar.Info("User logged in")                           // Sprint style
//	sugar.Infof("User %s logged in from %s", user, ip)    // Sprintf style
//	sugar.Info("User", user, "logged in from", ip)        // Sprint with multiple args
type SugarLogger struct {
	base *Logger
}

// Handler returns the underlying Handler.
func (l *SugarLogger) Handler() *Handler {
	return l.base.Handler()
}

// Enabled reports whether the logger emits log records at the given level.
func (l *SugarLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.base.Enabled(ctx, level)
}

// With returns a new SugarLogger with the given attributes added.
//
// The arguments are converted to attributes using the same rules as Logger.Log.
func (l *SugarLogger) With(args ...any) *SugarLogger {
	return &SugarLogger{base: l.base.With(args...)}
}

// WithGroup returns a new SugarLogger that starts a group.
//
// All attributes added through the returned logger will be nested under the given group name.
func (l *SugarLogger) WithGroup(name string) *SugarLogger {
	return &SugarLogger{base: l.base.WithGroup(name)}
}

// WithOptions returns a new SugarLogger with the given options applied.
func (l *SugarLogger) WithOptions(opts ...Option) *SugarLogger {
	return &SugarLogger{base: l.base.WithOptions(opts...)}
}

// Desugar returns the underlying Logger.
//
// Use this when you need to access Logger-specific functionality or pass the logger
// to code that expects a regular Logger.
func (l *SugarLogger) Desugar() *Logger {
	return l.base
}

// Log logs at the given level. Uses Sprint to format the message.
func (l *SugarLogger) Log(level slog.Level, args ...any) {
	l.log(context.Background(), level, "", args)
}

// LogContext logs at the given level with the given context. Uses Sprint to format the message.
func (l *SugarLogger) LogContext(ctx context.Context, level slog.Level, args ...any) {
	l.log(ctx, level, "", args)
}

// Debug logs at LevelDebug. Uses Sprint to format the message.
func (l *SugarLogger) Debug(args ...any) {
	l.log(context.Background(), slog.LevelDebug, "", args)
}

// DebugContext logs at LevelDebug with the given context. Uses Sprint to format the message.
func (l *SugarLogger) DebugContext(ctx context.Context, args ...any) {
	l.log(ctx, slog.LevelDebug, "", args)
}

// Info logs at LevelInfo. Uses Sprint to format the message.
func (l *SugarLogger) Info(args ...any) {
	l.log(context.Background(), slog.LevelInfo, "", args)
}

// InfoContext logs at LevelInfo with the given context. Uses Sprint to format the message.
func (l *SugarLogger) InfoContext(ctx context.Context, args ...any) {
	l.log(ctx, slog.LevelInfo, "", args)
}

// Warn logs at LevelWarn. Uses Sprint to format the message.
func (l *SugarLogger) Warn(args ...any) {
	l.log(context.Background(), slog.LevelWarn, "", args)
}

// WarnContext logs at LevelWarn with the given context. Uses Sprint to format the message.
func (l *SugarLogger) WarnContext(ctx context.Context, args ...any) {
	l.log(ctx, slog.LevelWarn, "", args)
}

// Error logs at LevelError. Uses Sprint to format the message.
func (l *SugarLogger) Error(args ...any) {
	l.log(context.Background(), slog.LevelError, "", args)
}

// ErrorContext logs at LevelError with the given context. Uses Sprint to format the message.
func (l *SugarLogger) ErrorContext(ctx context.Context, args ...any) {
	l.log(ctx, slog.LevelError, "", args)
}

// Logf logs at the given level. Uses Sprintf to format the message.
func (l *SugarLogger) Logf(level slog.Level, template string, args ...any) {
	l.log(context.Background(), level, template, args)
}

// LogfContext logs at the given level with the given context. Uses Sprintf to format the message.
func (l *SugarLogger) LogfContext(ctx context.Context, level slog.Level, template string, args ...any) {
	l.log(ctx, level, template, args)
}

// Debugf logs at LevelDebug. Uses Sprintf to format the message.
func (l *SugarLogger) Debugf(template string, args ...any) {
	l.log(context.Background(), slog.LevelDebug, template, args)
}

// DebugfContext logs at LevelDebug with the given context. Uses Sprintf to format the message.
func (l *SugarLogger) DebugfContext(ctx context.Context, template string, args ...any) {
	l.log(ctx, slog.LevelDebug, template, args)
}

// Infof logs at LevelInfo. Uses Sprintf to format the message.
func (l *SugarLogger) Infof(template string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, template, args)
}

// InfofContext logs at LevelInfo with the given context. Uses Sprintf to format the message.
func (l *SugarLogger) InfofContext(ctx context.Context, template string, args ...any) {
	l.log(ctx, slog.LevelInfo, template, args)
}

// Warnf logs at LevelWarn. Uses Sprintf to format the message.
func (l *SugarLogger) Warnf(template string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, template, args)
}

// WarnfContext logs at LevelWarn with the given context. Uses Sprintf to format the message.
func (l *SugarLogger) WarnfContext(ctx context.Context, template string, args ...any) {
	l.log(ctx, slog.LevelWarn, template, args)
}

// Errorf logs at LevelError. Uses Sprintf to format the message.
func (l *SugarLogger) Errorf(template string, args ...any) {
	l.log(context.Background(), slog.LevelError, template, args)
}

// ErrorfContext logs at LevelError with the given context. Uses Sprintf to format the message.
func (l *SugarLogger) ErrorfContext(ctx context.Context, template string, args ...any) {
	l.log(ctx, slog.LevelError, template, args)
}

func (l *SugarLogger) log(ctx context.Context, level slog.Level, template string, fmtArgs []any) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !l.Enabled(ctx, level) {
		return
	}

	msg := getMessage(template, fmtArgs)
	pc := l.base.capturePC()
	r := slog.NewRecord(time.Now(), level, msg, pc)

	_ = l.base.handler.Handle(ctx, r)
}

// getMessage formats the message using Sprint, Sprintf, or returns as-is.
//
// If template is non-empty, uses Sprintf with fmtArgs.
// If fmtArgs has a single string element, returns that string.
// Otherwise, uses Sprintln and trims the trailing newline for consistent spacing.
func getMessage(template string, fmtArgs []any) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	// Use Sprintln and trim trailing newline for consistent spacing
	msg := fmt.Sprintln(fmtArgs...)
	return msg[:len(msg)-1]
}
