package slogs

import (
	"bytes"
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
