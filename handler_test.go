package slogs_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/rockcookies/go-slogs"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	assert.NotNil(t, h)
}

func TestHandler_WithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	h2 := h.WithAttrs([]slog.Attr{slog.String("key", "value")})

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
	h2.Handle(context.Background(), r)

	assert.Contains(t, buf.String(), "key")
	assert.Contains(t, buf.String(), "value")
}

func TestHandler_WithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	h2 := h.WithGroup("grp")

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
	r.AddAttrs(slog.String("k", "v"))
	h2.Handle(context.Background(), r)

	assert.Contains(t, buf.String(), "grp")
}

func TestHandler_Enabled(t *testing.T) {
	buf := &bytes.Buffer{}
	base := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	h := slogs.NewHandler(base)

	assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestHandler_WithLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	h2 := h.WithLevel(slog.LevelError)

	assert.True(t, h2.Enabled(context.Background(), slog.LevelError))
	assert.True(t, h2.Enabled(context.Background(), slog.LevelWarn))
}

func TestNewMiddleware(t *testing.T) {
	buf := &bytes.Buffer{}
	middleware := slogs.NewMiddleware(nil)
	handler := middleware(slog.NewJSONHandler(buf, nil))

	assert.NotNil(t, handler)
}

func TestNewHandler_NilPanic(t *testing.T) {
	assert.Panics(t, func() {
		slogs.NewHandler(nil)
	})
}
