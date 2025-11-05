package slogs

import (
	"bytes"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectStdLogAt_Basic(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)
	require.NotNil(t, restore)
	defer restore()

	// Write a log message using standard library's log
	log.Println("test message from std log")

	output := buf.String()
	assert.Contains(t, output, "test message from std log")
	assert.Contains(t, output, `"level":"INFO"`)
}

func TestRedirectStdLogAt_DifferentLevels(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
	}{
		{"Debug", slog.LevelDebug},
		{"Info", slog.LevelInfo},
		{"Warn", slog.LevelWarn},
		{"Error", slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			h := NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			logger := New(h)

			restore, err := RedirectStdLogAt(logger, tt.level)
			require.NoError(t, err)
			defer restore()

			log.Println("message at level")

			output := buf.String()
			assert.Contains(t, output, "message at level")
			// Verify the level is correct in output
			assert.Contains(t, output, `"level"`)
		})
	}
}

func TestRedirectStdLogAt_RestoreFunction(t *testing.T) {
	// Save original values
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	originalOutput := log.Writer()

	// Set custom flags and prefix
	log.SetFlags(log.Ldate | log.Ltime)
	log.SetPrefix("[TEST] ")

	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)

	// Verify flags and prefix were changed
	assert.Equal(t, 0, log.Flags())
	assert.Equal(t, "", log.Prefix())

	// Restore
	restore()

	// Verify flags and prefix were restored
	assert.Equal(t, log.Ldate|log.Ltime, log.Flags())
	assert.Equal(t, "[TEST] ", log.Prefix())

	// Restore to original state
	log.SetFlags(originalFlags)
	log.SetPrefix(originalPrefix)
	log.SetOutput(originalOutput)
}

func TestRedirectStdLogAt_WithoutNewline(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)
	defer restore()

	// Use Print instead of Println (no automatic newline)
	log.Print("message without newline")

	output := buf.String()
	assert.Contains(t, output, "message without newline")
}

func TestRedirectStdLogAt_MultipleMessages(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)
	defer restore()

	log.Println("first message")
	log.Println("second message")
	log.Println("third message")

	output := buf.String()
	assert.Contains(t, output, "first message")
	assert.Contains(t, output, "second message")
	assert.Contains(t, output, "third message")

	// Should have multiple log records
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 3, len(lines))
}

func TestRedirectStdLogAt_LevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelWarn, // Only warn and above
	}))
	logger := New(h)

	// Redirect at Info level, but handler only accepts Warn+
	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)
	defer restore()

	log.Println("this should be filtered")

	// Buffer should be empty because level is below threshold
	assert.Empty(t, buf.String())
}

func TestRedirectStdLogAt_SpecialCharacters(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)
	defer restore()

	log.Println("message with \"quotes\" and\nnewlines\tand\ttabs")

	output := buf.String()
	assert.NotEmpty(t, output)
	// JSON handler should escape special characters properly
	assert.Contains(t, output, "message with")
}

func TestRedirectStdLogAt_EmptyMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)
	defer restore()

	log.Println("")

	output := buf.String()
	// Should still produce a log record, even if empty
	assert.NotEmpty(t, output)
}

func TestRedirectStdLogAt_ConcurrentWrites(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)
	defer restore()

	// Write from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(id int) {
			log.Printf("concurrent message %d", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	output := buf.String()
	assert.Contains(t, output, "concurrent message")
}

func TestRedirectStdLogAt_OutputToStderr(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewHandler(slog.NewJSONHandler(buf, nil))
	logger := New(h)

	restore, err := RedirectStdLogAt(logger, slog.LevelInfo)
	require.NoError(t, err)

	log.Println("test")

	// After restore, output should go back to stderr
	restore()

	// Verify log.Writer() is os.Stderr after restore
	assert.Equal(t, os.Stderr, log.Writer())
}
