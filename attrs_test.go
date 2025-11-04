package slogs_test

import (
	"log/slog"
	"testing"

	"github.com/rockcookies/go-slogs"
	"github.com/stretchr/testify/assert"
)

func TestGroupOrAttrs_WithGroup(t *testing.T) {
	var g *slogs.GroupOrAttrs
	g2 := g.WithGroup("group1")

	assert.NotNil(t, g2)
}

func TestGroupOrAttrs_WithGroup_Empty(t *testing.T) {
	var g *slogs.GroupOrAttrs
	g2 := g.WithGroup("")

	assert.Nil(t, g2)
}

func TestGroupOrAttrs_WithAttrs(t *testing.T) {
	var g *slogs.GroupOrAttrs
	attrs := []slog.Attr{slog.String("key", "value")}
	g2 := g.WithAttrs(attrs)

	assert.NotNil(t, g2)
}

func TestGroupOrAttrs_WithAttrs_Empty(t *testing.T) {
	var g *slogs.GroupOrAttrs
	g2 := g.WithAttrs(nil)

	assert.Nil(t, g2)
}

func TestGroupOrAttrs_Chain(t *testing.T) {
	var g *slogs.GroupOrAttrs
	g = g.WithAttrs([]slog.Attr{slog.String("k1", "v1")})
	g = g.WithGroup("group1")
	g = g.WithAttrs([]slog.Attr{slog.String("k2", "v2")})

	assert.NotNil(t, g)
}
