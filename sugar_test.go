package slogs_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/rockcookies/go-slogs"
	"github.com/stretchr/testify/assert"
)

func TestSugarLogger_Basic(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	sugar.Info("test message")
	assert.Contains(t, buf.String(), "test message")
}

func TestSugarLogger_Formatted(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	sugar.Infof("formatted %s", "message")
	assert.Contains(t, buf.String(), "formatted message")
}

func TestSugarLogger_Levels(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sugar := slogs.New(h).Sugar()

	sugar.Debug("debug")
	sugar.Info("info")
	sugar.Warn("warn")
	sugar.Error("error")

	output := buf.String()
	assert.Contains(t, output, "debug")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "warn")
	assert.Contains(t, output, "error")
}

func TestSugarLogger_WithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.InfoContext(ctx, "context message")
	assert.Contains(t, buf.String(), "context message")
}

func TestSugarLogger_With(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar().With("key", "value")

	sugar.Info("test")
	assert.Contains(t, buf.String(), "key")
	assert.Contains(t, buf.String(), "value")
}

func TestSugarLogger_WithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar().WithGroup("group")

	sugar.Info("test", "k", "v")
	output := buf.String()
	// Sugar logger concatenates args as message, not attributes
	assert.Contains(t, output, "test")
}

func TestSugarLogger_Desugar(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	logger := sugar.Desugar()
	assert.NotNil(t, logger)

	logger.Info("from desugar")
	assert.Contains(t, buf.String(), "from desugar")
}

func TestSugarLogger_LogLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	sugar.Log(slog.LevelInfo, "custom level")
	assert.Contains(t, buf.String(), "custom level")
}

func TestSugarLogger_Handler(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	assert.NotNil(t, sugar.Handler())
}

func TestSugarLogger_WithOptions(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar().WithOptions(slogs.WithLevel(slog.LevelWarn))

	sugar.Warn("msg")
	assert.Contains(t, buf.String(), "msg")
}

func TestSugarLogger_LogContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.LogContext(ctx, slog.LevelInfo, "ctx msg")
	assert.Contains(t, buf.String(), "ctx msg")
}

func TestSugarLogger_DebugContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.DebugContext(ctx, "debug ctx")
	assert.Contains(t, buf.String(), "debug ctx")
}

func TestSugarLogger_WarnContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.WarnContext(ctx, "warn ctx")
	assert.Contains(t, buf.String(), "warn ctx")
}

func TestSugarLogger_ErrorContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.ErrorContext(ctx, "error ctx")
	assert.Contains(t, buf.String(), "error ctx")
}

func TestSugarLogger_Logf(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	sugar.Logf(slog.LevelInfo, "formatted %s", "message")
	assert.Contains(t, buf.String(), "formatted message")
}

func TestSugarLogger_LogfContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.LogfContext(ctx, slog.LevelInfo, "ctx %s", "formatted")
	assert.Contains(t, buf.String(), "ctx formatted")
}

func TestSugarLogger_Debugf(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sugar := slogs.New(h).Sugar()

	sugar.Debugf("debug %d", 123)
	assert.Contains(t, buf.String(), "debug 123")
}

func TestSugarLogger_DebugfContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.DebugfContext(ctx, "debug ctx %s", "msg")
	assert.Contains(t, buf.String(), "debug ctx msg")
}

func TestSugarLogger_InfofContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.InfofContext(ctx, "info ctx %s", "msg")
	assert.Contains(t, buf.String(), "info ctx msg")
}

func TestSugarLogger_Warnf(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	sugar.Warnf("warn %s", "message")
	assert.Contains(t, buf.String(), "warn message")
}

func TestSugarLogger_WarnfContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.WarnfContext(ctx, "warn ctx %s", "msg")
	assert.Contains(t, buf.String(), "warn ctx msg")
}

func TestSugarLogger_Errorf(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	sugar.Errorf("error %d", 500)
	assert.Contains(t, buf.String(), "error 500")
}

func TestSugarLogger_ErrorfContext(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	ctx := context.Background()
	sugar.ErrorfContext(ctx, "error ctx %s", "msg")
	assert.Contains(t, buf.String(), "error ctx msg")
}

func TestSugarLogger_GetMessage_NoArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	sugar.Info()
	assert.NotEmpty(t, buf.String())
}

func TestSugarLogger_GetMessage_Multiple(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	sugar.Info("msg1", "msg2", 123)
	assert.Contains(t, buf.String(), "msg1")
}

func TestSugarLogger_Log_Disabled(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelError}))
	sugar := slogs.New(h).Sugar()

	sugar.Log(slog.LevelInfo, "should not log")
	assert.Empty(t, buf.String())
}

func TestSugarLogger_Named(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar().Named("myapp")

	sugar.Info("message")
	assert.Contains(t, buf.String(), "[myapp]")
}

func TestSugarLogger_Named_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar().Named("")

	sugar.Info("message")
	assert.NotContains(t, buf.String(), "[]")
}

func TestSugarLogger_Name(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar().Named("testname")

	assert.Equal(t, "testname", sugar.Name())
}

func TestSugarLogger_Name_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	sugar := slogs.New(h).Sugar()

	assert.Equal(t, "", sugar.Name())
}
