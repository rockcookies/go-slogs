package slogs

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithCaller(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h, WithCaller(true))

	logger.Info("test")
	assert.NotEmpty(t, buf.String())
}

func TestWithCallerSkip(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h, WithCallerSkip(1))

	logger.Info("test")
	assert.NotEmpty(t, buf.String())
}

func TestWithLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := New(h, WithLevel(slog.LevelWarn))

	logger.Info("should not appear")
	buf.Reset()

	logger.Warn("should appear")
	assert.Contains(t, buf.String(), "should appear")
}

func TestWithCallerAt(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))

	// Custom function that enables caller at Error level
	callerFunc := func(ctx context.Context, level slog.Level) bool {
		return level >= slog.LevelError
	}

	logger := New(h, WithCallerAt(callerFunc))
	logger.Info("info message")
	logger.Error("error message")

	assert.NotEmpty(t, buf.String())
}

func TestWithCallerAtLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h, WithCallerAtLevel(slog.LevelWarn))

	logger.Info("info message")
	logger.Warn("warn message")

	assert.NotEmpty(t, buf.String())
}

func TestWithCallerAt_Nil(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h, WithCallerAt(nil))

	logger.Info("test")
	assert.NotEmpty(t, buf.String())
}
