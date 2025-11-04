package slogs

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type SugarLogger struct {
	base *Logger
}

// Handler returns l's Handler.
func (l *SugarLogger) Handler() *Handler {
	return l.base.Handler()
}

func (l *SugarLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.base.Enabled(ctx, level)
}

func (l *SugarLogger) With(args ...any) *SugarLogger {
	return &SugarLogger{base: l.base.With(args...)}
}

func (l *SugarLogger) WithGroup(name string) *SugarLogger {
	return &SugarLogger{base: l.base.WithGroup(name)}
}

func (l *SugarLogger) WithOptions(opts ...Option) *SugarLogger {
	return &SugarLogger{base: l.base.WithOptions(opts...)}
}

func (l *SugarLogger) Desugar() *Logger {
	return l.base
}

func (l *SugarLogger) Log(level slog.Level, args ...any) {
	l.log(context.Background(), level, "", args)
}

func (l *SugarLogger) LogContext(ctx context.Context, level slog.Level, args ...any) {
	l.log(ctx, level, "", args)
}

func (l *SugarLogger) Debug(args ...any) {
	l.log(context.Background(), slog.LevelDebug, "", args)
}

func (l *SugarLogger) DebugContext(ctx context.Context, args ...any) {
	l.log(ctx, slog.LevelDebug, "", args)
}

func (l *SugarLogger) Info(args ...any) {
	l.log(context.Background(), slog.LevelInfo, "", args)
}

func (l *SugarLogger) InfoContext(ctx context.Context, args ...any) {
	l.log(ctx, slog.LevelInfo, "", args)
}

func (l *SugarLogger) Warn(args ...any) {
	l.log(context.Background(), slog.LevelWarn, "", args)
}

func (l *SugarLogger) WarnContext(ctx context.Context, args ...any) {
	l.log(ctx, slog.LevelWarn, "", args)
}

func (l *SugarLogger) Error(args ...any) {
	l.log(context.Background(), slog.LevelError, "", args)
}

func (l *SugarLogger) ErrorContext(ctx context.Context, args ...any) {
	l.log(ctx, slog.LevelError, "", args)
}

func (l *SugarLogger) Logf(level slog.Level, template string, args ...any) {
	l.log(context.Background(), level, template, args)
}

func (l *SugarLogger) LogfContext(ctx context.Context, level slog.Level, template string, args ...any) {
	l.log(ctx, level, template, args)
}

func (l *SugarLogger) Debugf(template string, args ...any) {
	l.log(context.Background(), slog.LevelDebug, template, args)
}

func (l *SugarLogger) DebugfContext(ctx context.Context, template string, args ...any) {
	l.log(ctx, slog.LevelDebug, template, args)
}

func (l *SugarLogger) Infof(template string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, template, args)
}

func (l *SugarLogger) InfofContext(ctx context.Context, template string, args ...any) {
	l.log(ctx, slog.LevelInfo, template, args)
}

func (l *SugarLogger) Warnf(template string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, template, args)
}

func (l *SugarLogger) WarnfContext(ctx context.Context, template string, args ...any) {
	l.log(ctx, slog.LevelWarn, template, args)
}

func (l *SugarLogger) Errorf(template string, args ...any) {
	l.log(context.Background(), slog.LevelError, template, args)
}

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

// getMessage format with Sprint, Sprintf, or neither.
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
