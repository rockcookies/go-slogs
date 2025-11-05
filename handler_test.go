package slogs

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	assert.NotNil(t, h)
}

func TestHandler_WithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	h2 := h.WithAttrs([]slog.Attr{slog.String("key", "value")})

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
	h2.Handle(context.Background(), r)

	assert.Contains(t, buf.String(), "key")
	assert.Contains(t, buf.String(), "value")
}

func TestHandler_WithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	h2 := h.WithGroup("grp")

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
	r.AddAttrs(slog.String("k", "v"))
	h2.Handle(context.Background(), r)

	assert.Contains(t, buf.String(), "grp")
}

func TestHandler_Enabled(t *testing.T) {
	buf := &bytes.Buffer{}
	base := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	h := NewHandler(base)

	assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestHandler_WithLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	h2 := h.WithLevel(slog.LevelError)

	// When handler level is Error, only Error level should be enabled
	// Warn is below Error, so it should be disabled
	assert.True(t, h2.Enabled(context.Background(), slog.LevelError))
	assert.False(t, h2.Enabled(context.Background(), slog.LevelWarn))
}

func TestNewMiddleware(t *testing.T) {
	buf := &bytes.Buffer{}
	middleware := NewMiddleware(nil)
	handler := middleware(slog.NewJSONHandler(buf, nil))

	assert.NotNil(t, handler)
}

func TestNewHandler_NilPanic(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(nil)
	})
}

func TestHandler_Named(t *testing.T) {
	buf := &bytes.Buffer{}
	base := slog.NewJSONHandler(buf, nil)
	h := NewHandler(base)
	h2 := h.Named("myapp")

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
	h2.Handle(context.Background(), r)

	assert.Contains(t, buf.String(), "[myapp]")
}

func TestHandler_Named_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	base := slog.NewJSONHandler(buf, nil)
	h := NewHandler(base)
	h2 := h.Named("")

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
	h2.Handle(context.Background(), r)

	assert.NotContains(t, buf.String(), "[]")
}

func TestHandler_Name(t *testing.T) {
	buf := &bytes.Buffer{}
	base := slog.NewJSONHandler(buf, nil)
	h := NewHandler(base)
	h2 := h.Named("testname")

	assert.Equal(t, "testname", h2.Name())
}

func TestHandler_Name_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	base := slog.NewJSONHandler(buf, nil)
	h := NewHandler(base)

	assert.Equal(t, "", h.Name())
}

func TestHandler_WithLevel_Leveler(t *testing.T) {
	buf := &bytes.Buffer{}
	// Use LevelDebug for base handler so it doesn't filter anything
	base := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	h := NewHandler(base)
	h2 := h.WithLevel(slog.LevelWarn)

	// When handler level is Warn, levels below Warn should be disabled
	// and levels at or above Warn should be enabled
	assert.False(t, h2.Enabled(context.Background(), slog.LevelDebug))
	assert.False(t, h2.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, h2.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, h2.Enabled(context.Background(), slog.LevelError))
}
