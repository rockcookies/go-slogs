package slogs_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/rockcookies/go-slogs"
	"github.com/stretchr/testify/assert"
)

func TestPrepend(t *testing.T) {
	ctx := slogs.Prepend(context.Background(), "key1", "val1")
	attrs := slogs.ExtractPrepended(ctx)

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key1", attrs[0].Key)
	assert.Equal(t, "val1", attrs[0].Value.String())
}

func TestAppend(t *testing.T) {
	ctx := slogs.Append(context.Background(), "key2", "val2")
	attrs := slogs.ExtractAppended(ctx)

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key2", attrs[0].Key)
	assert.Equal(t, "val2", attrs[0].Value.String())
}

func TestPrepend_Multiple(t *testing.T) {
	ctx := slogs.Prepend(context.Background(), "k1", "v1")
	ctx = slogs.Prepend(ctx, "k2", "v2")
	attrs := slogs.ExtractPrepended(ctx)

	assert.Len(t, attrs, 2)
}

func TestContextAttrs_Integration(t *testing.T) {
	buf := &bytes.Buffer{}
	h := slogs.NewHandler(slog.NewJSONHandler(buf, nil))
	logger := slogs.New(h)

	ctx := slogs.Prepend(context.Background(), "pre", "first")
	ctx = slogs.Append(ctx, "app", "last")

	logger.InfoContext(ctx, "message", "mid", "value")
	output := buf.String()

	assert.Contains(t, output, "pre")
	assert.Contains(t, output, "app")
	assert.Contains(t, output, "mid")
}

func TestExtract_Nil(t *testing.T) {
	attrs := slogs.ExtractPrepended(context.Background())
	assert.Nil(t, attrs)

	attrs = slogs.ExtractAppended(context.Background())
	assert.Nil(t, attrs)
}

func TestPrepend_NilContext(t *testing.T) {
	ctx := slogs.Prepend(nil, "key", "value")
	attrs := slogs.ExtractPrepended(ctx)

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key", attrs[0].Key)
}

func TestAppend_NilContext(t *testing.T) {
	ctx := slogs.Append(nil, "key", "value")
	attrs := slogs.ExtractAppended(ctx)

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key", attrs[0].Key)
}

func TestAppend_Multiple(t *testing.T) {
	ctx := slogs.Append(context.Background(), "k1", "v1")
	ctx = slogs.Append(ctx, "k2", "v2")
	attrs := slogs.ExtractAppended(ctx)

	assert.Len(t, attrs, 2)
}
