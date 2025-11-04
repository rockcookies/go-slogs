package slogs_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/rockcookies/go-slogs"
	"github.com/stretchr/testify/assert"
)

func TestWithCaller(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	logger := slogs.New(h, slogs.WithCaller(true))

	logger.Info("test")
	assert.NotEmpty(t, buf.String())
}

func TestWithCallerSkip(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	logger := slogs.New(h, slogs.WithCallerSkip(1))

	logger.Info("test")
	assert.NotEmpty(t, buf.String())
}

func TestWithLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := slogs.New(h, slogs.WithLevel(slog.LevelWarn))

	logger.Info("should not appear")
	buf.Reset()

	logger.Warn("should appear")
	assert.Contains(t, buf.String(), "should appear")
}
