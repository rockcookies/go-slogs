package slogs

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepend(t *testing.T) {
	ctx := Prepend(context.Background(), "key1", "val1")
	attrs := ExtractPrepended(ctx)

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key1", attrs[0].Key)
	assert.Equal(t, "val1", attrs[0].Value.String())
}

func TestAppend(t *testing.T) {
	ctx := Append(context.Background(), "key2", "val2")
	attrs := ExtractAppended(ctx)

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key2", attrs[0].Key)
	assert.Equal(t, "val2", attrs[0].Value.String())
}

func TestPrepend_Multiple(t *testing.T) {
	ctx := Prepend(context.Background(), "k1", "v1")
	ctx = Prepend(ctx, "k2", "v2")
	attrs := ExtractPrepended(ctx)

	assert.Len(t, attrs, 2)
}

func TestContextAttrs_Integration(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	ctx := Prepend(context.Background(), "pre", "first")
	ctx = Append(ctx, "app", "last")

	logger.InfoContext(ctx, "message", "mid", "value")
	output := buf.String()

	assert.Contains(t, output, "pre")
	assert.Contains(t, output, "app")
	assert.Contains(t, output, "mid")
}

func TestExtract_Nil(t *testing.T) {
	attrs := ExtractPrepended(context.Background())
	assert.Nil(t, attrs)

	attrs = ExtractAppended(context.Background())
	assert.Nil(t, attrs)
}

func TestPrepend_NilContext(t *testing.T) {
	ctx := Prepend(nil, "key", "value")
	attrs := ExtractPrepended(ctx)

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key", attrs[0].Key)
}

func TestAppend_NilContext(t *testing.T) {
	ctx := Append(nil, "key", "value")
	attrs := ExtractAppended(ctx)

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key", attrs[0].Key)
}

func TestAppend_Multiple(t *testing.T) {
	ctx := Append(context.Background(), "k1", "v1")
	ctx = Append(ctx, "k2", "v2")
	attrs := ExtractAppended(ctx)

	assert.Len(t, attrs, 2)
}
