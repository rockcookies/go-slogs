package slogs

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Basic(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	logger.Info("test message", "key", "value")
	assert.Contains(t, buf.String(), "test message")
	assert.Contains(t, buf.String(), "key")
}

func TestLogger_WithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h).With("app", "test")

	logger.Info("message")
	assert.Contains(t, buf.String(), "app")
	assert.Contains(t, buf.String(), "test")
}

func TestLogger_WithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h).WithGroup("group1")

	logger.Info("message", "key", "value")
	assert.Contains(t, buf.String(), "group1")
}

func TestLogger_Enabled(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	logger := New(h)

	assert.True(t, logger.Enabled(context.Background(), slog.LevelWarn))
	assert.False(t, logger.Enabled(context.Background(), slog.LevelDebug))
}

func TestLogger_Levels(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := New(h)

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	output := buf.String()
	assert.Contains(t, output, "debug msg")
	assert.Contains(t, output, "info msg")
	assert.Contains(t, output, "warn msg")
	assert.Contains(t, output, "error msg")
}

func TestLogger_WithOptions(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h, WithCaller(false))

	logger2 := logger.WithOptions(WithLevel(slog.LevelWarn))
	logger2.Warn("message")

	assert.Contains(t, buf.String(), "message")
}

func TestLogger_Handler(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	assert.NotNil(t, logger.Handler())
	assert.Equal(t, h, logger.Handler())
}

func TestLogger_Log(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	logger.Log(context.Background(), slog.LevelInfo, "log message", "key", "val")
	assert.Contains(t, buf.String(), "log message")
	assert.Contains(t, buf.String(), "key")
}

func TestLogger_LogAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	logger.LogAttrs(context.Background(), slog.LevelInfo, "attrs msg", slog.String("k", "v"))
	assert.Contains(t, buf.String(), "attrs msg")
	assert.Contains(t, buf.String(), "k")
}

func TestLogger_DebugContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := New(h)

	ctx := context.Background()
	logger.DebugContext(ctx, "debug ctx", "key", "value")
	assert.Contains(t, buf.String(), "debug ctx")
}

func TestLogger_WarnContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	ctx := context.Background()
	logger.WarnContext(ctx, "warn ctx", "key", "value")
	assert.Contains(t, buf.String(), "warn ctx")
}

func TestLogger_ErrorContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	ctx := context.Background()
	logger.ErrorContext(ctx, "error ctx", "key", "value")
	assert.Contains(t, buf.String(), "error ctx")
}

func TestLogger_With_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	logger2 := logger.With()
	assert.Equal(t, logger, logger2)
}

func TestLogger_WithGroup_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	logger2 := logger.WithGroup("")
	assert.Equal(t, logger, logger2)
}

func TestLogger_Enabled_NilContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	assert.True(t, logger.Enabled(nil, slog.LevelInfo))
}

func TestNew_NilHandler(t *testing.T) {
	assert.Panics(t, func() {
		New(nil)
	})
}

func TestLogger_Log_Disabled(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelError}))
	logger := New(h)

	logger.Log(context.Background(), slog.LevelInfo, "should not log")
	assert.Empty(t, buf.String())
}

func TestLogger_LogAttrs_Disabled(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelError}))
	logger := New(h)

	logger.LogAttrs(context.Background(), slog.LevelInfo, "should not log", slog.String("k", "v"))
	assert.Empty(t, buf.String())
}

func TestLogger_Named(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h).Named("myapp")

	logger.Info("message")
	assert.Contains(t, buf.String(), "[myapp]")
}

func TestLogger_Named_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h).Named("")

	logger.Info("message")
	assert.NotContains(t, buf.String(), "[]")
}

func TestLogger_Name(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h).Named("testname")

	assert.Equal(t, "testname", logger.Name())
}

func TestLogger_Name_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	assert.Equal(t, "", logger.Name())
}
