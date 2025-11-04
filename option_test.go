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

func TestWithName(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	logger := slogs.New(h, slogs.WithName("myapp"))

	logger.Info("message")
	assert.Contains(t, buf.String(), "[myapp]")
}

func TestWithNameOverride(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	logger := slogs.New(h, slogs.WithName("first"), slogs.WithNameOverride("second"))

	logger.Info("message")
	assert.Contains(t, buf.String(), "[second]")
	assert.NotContains(t, buf.String(), "[first]")
}

func TestWithName_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	logger := slogs.New(h, slogs.WithName(""))

	logger.Info("message")
	assert.NotContains(t, buf.String(), "[]")
}

func TestWithNameOverride_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	logger := slogs.New(h, slogs.WithName("first"), slogs.WithNameOverride(""))

	logger.Info("message")
	assert.Contains(t, buf.String(), "[first]")
}
