package slogs

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	attr := Stack("test_stack")

	assert.Equal(t, "test_stack", attr.Key)
	assert.Equal(t, slog.KindString, attr.Value.Kind())

	stackStr := attr.Value.String()
	assert.NotEmpty(t, stackStr)
	assert.Contains(t, stackStr, "TestStack")
}

func TestStackSkip(t *testing.T) {
	attr := StackSkip("test_stack_skip", 1)

	assert.Equal(t, "test_stack_skip", attr.Key)
	assert.Equal(t, slog.KindString, attr.Value.Kind())

	stackStr := attr.Value.String()
	assert.NotEmpty(t, stackStr)
	assert.NotContains(t, stackStr, "TestStackSkip")
}

func TestStackSkipDifferentValues(t *testing.T) {
	tests := []struct {
		name     string
		skip     int
		contains string
		excludes string
	}{
		{"skip 0", 0, "TestStackSkipDifferentValues", ""},
		{"skip 1", 1, "", "TestStackSkipDifferentValues"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := StackSkip("stack", tt.skip)
			stackStr := attr.Value.String()
			assert.NotEmpty(t, stackStr)

			if tt.contains != "" {
				assert.Contains(t, stackStr, tt.contains)
			}
			if tt.excludes != "" {
				assert.NotContains(t, stackStr, tt.excludes)
			}
		})
	}
}

func TestStackReturnType(t *testing.T) {
	attr := Stack("test")

	assert.IsType(t, slog.Attr{}, attr)
	assert.Equal(t, slog.String("test", attr.Value.String()), attr)
}

func TestStackSkipNegative(t *testing.T) {
	attr := StackSkip("test", -1)

	assert.Equal(t, "test", attr.Key)
	assert.Equal(t, slog.KindString, attr.Value.Kind())

	stackStr := attr.Value.String()
	assert.NotEmpty(t, stackStr)
}

func TestStackLargeSkip(t *testing.T) {
	attr := StackSkip("test", 100)

	assert.Equal(t, "test", attr.Key)
	assert.Equal(t, slog.KindString, attr.Value.Kind())

	stackStr := attr.Value.String()

	lines := strings.Count(stackStr, "\n")
	assert.Less(t, lines, 10)
}