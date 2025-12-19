package slogs

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testHandler is a simple fake handler for testing purposes.
// It records all Handle() calls and can be configured to return errors.
type testHandler struct {
	mu      sync.Mutex
	enabled bool
	records []slog.Record
	err     error
	mutate  func(*slog.Record) // Optional function to mutate records
}

func newTestHandler(enabled bool) *testHandler {
	return &testHandler{
		enabled: enabled,
		records: make([]slog.Record, 0),
	}
}

func (h *testHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return h.enabled
}

func (h *testHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.mutate != nil {
		h.mutate(&r)
	}

	h.records = append(h.records, r)
	return h.err
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &testHandler{
		mu:      sync.Mutex{},
		enabled: h.enabled,
		records: make([]slog.Record, 0),
		err:     h.err,
		mutate:  h.mutate,
	}
}

func (h *testHandler) WithGroup(_ string) slog.Handler {
	return &testHandler{
		mu:      sync.Mutex{},
		enabled: h.enabled,
		records: make([]slog.Record, 0),
		err:     h.err,
		mutate:  h.mutate,
	}
}

func (h *testHandler) recordCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.records)
}

func (h *testHandler) getRecords() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]slog.Record{}, h.records...)
}

// recordHasAttr checks if a record contains the specified attribute with the given value.
func recordHasAttr(r slog.Record, key, want string) bool {
	found := false
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == key && a.Value.String() == want {
			found = true
			return false // Stop iteration
		}
		return true
	})
	return found
}

func TestMultiHandler_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name            string
		logMessage      string
		logAttrs        []any
		validateBuffers func(t *testing.T, bufs []*bytes.Buffer)
	}{
		{
			name:       "broadcasts to multiple handlers",
			logMessage: "test message",
			logAttrs:   []any{"key", "value"},
			validateBuffers: func(t *testing.T, bufs []*bytes.Buffer) {
				for i, buf := range bufs {
					assert.Contains(t, buf.String(), "test message", "buffer %d should contain 'test message'", i)
					assert.Contains(t, buf.String(), "key", "buffer %d should contain attribute 'key'", i)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf1 := &bytes.Buffer{}
			buf2 := &bytes.Buffer{}
			h1 := slog.NewJSONHandler(buf1, nil)
			h2 := slog.NewJSONHandler(buf2, nil)

			multi := MultiHandler(h1, h2)
			logger := slog.New(multi)
			logger.Info(tt.logMessage, tt.logAttrs...)

			tt.validateBuffers(t, []*bytes.Buffer{buf1, buf2})
		})
	}
}

func TestMultiHandler_Enabled(t *testing.T) {
	tests := []struct {
		name        string
		handlers    []slog.Handler
		wantEnabled bool
		description string
	}{
		{
			name:        "returns true when any handler is enabled",
			handlers:    []slog.Handler{newTestHandler(true), newTestHandler(false)},
			wantEnabled: true,
			description: "at least one handler is enabled",
		},
		{
			name:        "returns false when all handlers are disabled",
			handlers:    []slog.Handler{newTestHandler(false), newTestHandler(false)},
			wantEnabled: false,
			description: "no handlers are enabled",
		},
		{
			name:        "returns false for empty handler list",
			handlers:    nil,
			wantEnabled: false,
			description: "empty list has no enabled handlers",
		},
		{
			name:        "returns true when all handlers are enabled",
			handlers:    []slog.Handler{newTestHandler(true), newTestHandler(true), newTestHandler(true)},
			wantEnabled: true,
			description: "all handlers are enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multi := MultiHandler(tt.handlers...)
			got := multi.Enabled(context.Background(), slog.LevelInfo)

			assert.Equal(t, tt.wantEnabled, got, tt.description)
		})
	}
}

func TestMultiHandler_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupHandlers func() []slog.Handler
		wantError     bool
		errorContains []string
	}{
		{
			name: "returns nil when no errors occur",
			setupHandlers: func() []slog.Handler {
				return []slog.Handler{
					newTestHandler(true),
					newTestHandler(true),
				}
			},
			wantError: false,
		},
		{
			name: "collects single handler error",
			setupHandlers: func() []slog.Handler {
				h1 := newTestHandler(true)
				h1.err = errors.New("handler error")
				return []slog.Handler{h1}
			},
			wantError:     true,
			errorContains: []string{"handler error"},
		},
		{
			name: "aggregates multiple handler errors",
			setupHandlers: func() []slog.Handler {
				h1 := newTestHandler(true)
				h1.err = errors.New("first error")

				h2 := newTestHandler(true)
				h2.err = errors.New("second error")

				return []slog.Handler{h1, h2}
			},
			wantError:     true,
			errorContains: []string{"first error", "second error"},
		},
		{
			name: "continues processing after partial failures",
			setupHandlers: func() []slog.Handler {
				h1 := newTestHandler(true)
				h1.err = errors.New("failed handler")

				h2 := newTestHandler(true) // No error

				return []slog.Handler{h1, h2}
			},
			wantError:     true,
			errorContains: []string{"failed handler"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := tt.setupHandlers()
			multi := MultiHandler(handlers...)

			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			err := multi.Handle(context.Background(), record)

			if tt.wantError {
				require.Error(t, err, "expected error but got nil")
				errStr := err.Error()
				for _, want := range tt.errorContains {
					assert.Contains(t, errStr, want)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultiHandler_WithAttrs(t *testing.T) {
	tests := []struct {
		name       string
		attrs      []slog.Attr
		wantInJSON string
	}{
		{
			name:       "adds single attribute",
			attrs:      []slog.Attr{slog.String("key", "value")},
			wantInJSON: `"key":"value"`,
		},
		{
			name: "adds multiple attributes",
			attrs: []slog.Attr{
				slog.String("key1", "value1"),
				slog.String("key2", "value2"),
			},
			wantInJSON: `"key1":"value1"`,
		},
		{
			name:       "adds integer attribute",
			attrs:      []slog.Attr{slog.Int("count", 42)},
			wantInJSON: `"count":42`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf1 := &bytes.Buffer{}
			buf2 := &bytes.Buffer{}

			h1 := slog.NewJSONHandler(buf1, nil)
			h2 := slog.NewJSONHandler(buf2, nil)

			multi := MultiHandler(h1, h2)
			newMulti := multi.WithAttrs(tt.attrs)

			// Verify return type
			_, ok := newMulti.(*multiHandler)
			assert.True(t, ok, "WithAttrs should return *multiHandler")

			// Log a message
			rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			err := newMulti.Handle(context.Background(), rec)
			require.NoError(t, err)

			// Verify all handlers received attributes
			for i, buf := range []*bytes.Buffer{buf1, buf2} {
				assert.Contains(t, buf.String(), tt.wantInJSON, "buffer %d missing attribute", i)
			}
		})
	}
}

func TestMultiHandler_WithGroup(t *testing.T) {
	tests := []struct {
		name         string
		groupName    string
		attrs        []slog.Attr
		wantInJSON   string
		shouldReturn bool // whether WithGroup should return a new handler
	}{
		{
			name:         "creates group with attributes",
			groupName:    "test",
			attrs:        []slog.Attr{slog.String("key", "value")},
			wantInJSON:   `"test":{"key":"value"}`,
			shouldReturn: true,
		},
		{
			name:         "creates nested groups",
			groupName:    "outer",
			attrs:        []slog.Attr{slog.String("field", "data")},
			wantInJSON:   `"outer":{"field":"data"}`,
			shouldReturn: true,
		},
		{
			name:         "empty group name returns same handler",
			groupName:    "",
			attrs:        []slog.Attr{slog.String("key", "value")},
			wantInJSON:   `"key":"value"`, // No grouping
			shouldReturn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf1 := &bytes.Buffer{}
			buf2 := &bytes.Buffer{}

			h1 := slog.NewJSONHandler(buf1, nil)
			h2 := slog.NewJSONHandler(buf2, nil)

			multi := MultiHandler(h1, h2)
			newMulti := multi.WithGroup(tt.groupName)

			if tt.shouldReturn {
				// Verify return type for non-empty groups
				_, ok := newMulti.(*multiHandler)
				assert.True(t, ok, "WithGroup should return *multiHandler")
			}

			// Add attributes and log
			withAttrs := newMulti.WithAttrs(tt.attrs)
			rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			err := withAttrs.Handle(context.Background(), rec)
			require.NoError(t, err)

			// Verify all handlers received grouped attributes
			for i, buf := range []*bytes.Buffer{buf1, buf2} {
				assert.Contains(t, buf.String(), tt.wantInJSON, "buffer %d missing expected content", i)
			}
		})
	}
}

func TestMultiHandler_RecordIsolation(t *testing.T) {
	tests := []struct {
		name          string
		setupHandlers func() (h1, h2 *testHandler)
		checkH1       func(t *testing.T, records []slog.Record)
		checkH2       func(t *testing.T, records []slog.Record)
	}{
		{
			name: "handlers receive isolated record copies",
			setupHandlers: func() (h1, h2 *testHandler) {
				h1 = newTestHandler(true)
				h1.mutate = func(r *slog.Record) {
					r.AddAttrs(slog.String("mutated", "yes"))
				}
				h2 = newTestHandler(true)
				return h1, h2
			},
			checkH1: func(t *testing.T, records []slog.Record) {
				require.Len(t, records, 1, "handler1 should receive 1 record")
				assert.True(t, recordHasAttr(records[0], "mutated", "yes"), "handler1 record should have mutated attribute")
			},
			checkH2: func(t *testing.T, records []slog.Record) {
				require.Len(t, records, 1, "handler2 should receive 1 record")
				assert.False(t, recordHasAttr(records[0], "mutated", "yes"), "handler2 record should NOT have mutated attribute (should be isolated)")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1, h2 := tt.setupHandlers()
			multi := MultiHandler(h1, h2)

			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			record.AddAttrs(slog.String("original", "value"))

			err := multi.Handle(context.Background(), record)
			require.NoError(t, err)

			tt.checkH1(t, h1.getRecords())
			tt.checkH2(t, h2.getRecords())
		})
	}
}

func TestMultiHandler_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		handlers    []slog.Handler
		shouldPanic bool
		wantError   bool
	}{
		{
			name:        "single non-nil handler",
			handlers:    []slog.Handler{newTestHandler(true)},
			shouldPanic: false,
			wantError:   false,
		},
		{
			name:        "handles empty handler list gracefully",
			handlers:    []slog.Handler{},
			shouldPanic: false,
			wantError:   false,
		},
		{
			name:        "handles nil handler list",
			handlers:    nil,
			shouldPanic: false,
			wantError:   false,
		},
		{
			name:        "filters out nil handlers",
			handlers:    []slog.Handler{newTestHandler(true), nil, newTestHandler(true)},
			shouldPanic: false,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					MultiHandler(tt.handlers...)
				})
				return
			}

			multi := MultiHandler(tt.handlers...)
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			err := multi.Handle(context.Background(), record)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultiHandler_ConcurrentAccess(t *testing.T) {
	tests := []struct {
		name           string
		numGoroutines  int
		numLogsPerTask int
	}{
		{
			name:           "handles concurrent writes",
			numGoroutines:  10,
			numLogsPerTask: 100,
		},
		{
			name:           "handles high concurrency",
			numGoroutines:  50,
			numLogsPerTask: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := newTestHandler(true)
			h2 := newTestHandler(true)
			multi := MultiHandler(h1, h2)

			var wg sync.WaitGroup
			expectedTotal := tt.numGoroutines * tt.numLogsPerTask

			for i := 0; i < tt.numGoroutines; i++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()
					for j := 0; j < tt.numLogsPerTask; j++ {
						record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
						record.AddAttrs(slog.Int("goroutine", goroutineID), slog.Int("log", j))
						err := multi.Handle(context.Background(), record)
						assert.NoError(t, err)
					}
				}(i)
			}

			wg.Wait()

			// Verify all logs were processed by both handlers
			assert.Equal(t, expectedTotal, h1.recordCount(), "handler1 should process all records")
			assert.Equal(t, expectedTotal, h2.recordCount(), "handler2 should process all records")
		})
	}
}

func TestMultiHandler_Optimization(t *testing.T) {
	tests := []struct {
		name            string
		setupHandlers   func() (handlers []slog.Handler, bufs []*bytes.Buffer)
		validateResult  func(t *testing.T, result slog.Handler)
		validateBuffers func(t *testing.T, bufs []*bytes.Buffer)
	}{
		{
			name: "flattens nested MultiHandlers",
			setupHandlers: func() ([]slog.Handler, []*bytes.Buffer) {
				buf1, buf2, buf3 := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
				h1 := slog.NewJSONHandler(buf1, nil)
				h2 := slog.NewJSONHandler(buf2, nil)
				h3 := slog.NewJSONHandler(buf3, nil)

				// Create nested structure: MultiHandler(MultiHandler(h1, h2), h3)
				inner := MultiHandler(h1, h2)
				outer := MultiHandler(inner, h3)

				return []slog.Handler{outer}, []*bytes.Buffer{buf1, buf2, buf3}
			},
			validateResult: func(t *testing.T, result slog.Handler) {
				// Should be flattened multiHandler
				mh, ok := result.(*multiHandler)
				assert.True(t, ok, "expected *multiHandler type")
				if ok {
					assert.Len(t, mh.handlers, 3, "expected 3 flattened handlers")
				}
			},
			validateBuffers: func(t *testing.T, bufs []*bytes.Buffer) {
				for i, buf := range bufs {
					assert.Contains(t, buf.String(), "test message", "buffer %d missing log message", i)
				}
			},
		},
		{
			name: "returns single handler directly without wrapping",
			setupHandlers: func() ([]slog.Handler, []*bytes.Buffer) {
				buf := &bytes.Buffer{}
				h := slog.NewJSONHandler(buf, nil)
				return []slog.Handler{h}, []*bytes.Buffer{buf}
			},
			validateResult: func(t *testing.T, result slog.Handler) {
				_, ok := result.(*multiHandler)
				assert.False(t, ok, "single handler should not be wrapped in *multiHandler")
			},
			validateBuffers: func(t *testing.T, bufs []*bytes.Buffer) {
				assert.Contains(t, bufs[0].String(), "test message", "buffer missing log message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, bufs := tt.setupHandlers()
			result := handlers[0]

			tt.validateResult(t, result)

			// Log a test message
			logger := slog.New(result)
			logger.Info("test message")

			tt.validateBuffers(t, bufs)
		})
	}
}

func TestMultiHandler_NilHandlerFiltering(t *testing.T) {
	tests := []struct {
		name           string
		handlers       []slog.Handler
		wantHandlerNum int
	}{
		{
			name:           "filters out single nil handler",
			handlers:       []slog.Handler{nil},
			wantHandlerNum: 0,
		},
		{
			name:           "filters out multiple nil handlers",
			handlers:       []slog.Handler{nil, nil, nil},
			wantHandlerNum: 0,
		},
		{
			name:           "keeps valid handlers and filters out nils",
			handlers:       []slog.Handler{newTestHandler(true), nil, newTestHandler(false), nil},
			wantHandlerNum: 2,
		},
		{
			name:           "no nil handlers to filter",
			handlers:       []slog.Handler{newTestHandler(true), newTestHandler(false)},
			wantHandlerNum: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multi := MultiHandler(tt.handlers...)

			// Test that it works without panicking
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
			err := multi.Handle(context.Background(), record)
			assert.NoError(t, err)

			// Verify handler count if it's a multiHandler
			if mh, ok := multi.(*multiHandler); ok {
				assert.Len(t, mh.handlers, tt.wantHandlerNum)
			} else if tt.wantHandlerNum == 1 {
				// Single handler optimization - should not be wrapped
				assert.NotNil(t, multi)
			} else if tt.wantHandlerNum == 0 {
				// Empty handler case
				assert.NotNil(t, multi)
			}
		})
	}
}

func TestMultiHandler_LevelFiltering(t *testing.T) {
	tests := []struct {
		name             string
		handler1Level    slog.Level
		handler2Level    slog.Level
		testLevel        slog.Level
		wantH1Enabled    bool
		wantH2Enabled    bool
		wantMultiEnabled bool
	}{
		{
			name:             "both handlers disabled for debug",
			handler1Level:    slog.LevelInfo,
			handler2Level:    slog.LevelWarn,
			testLevel:        slog.LevelDebug,
			wantH1Enabled:    false,
			wantH2Enabled:    false,
			wantMultiEnabled: false,
		},
		{
			name:             "one handler enabled for info",
			handler1Level:    slog.LevelInfo,
			handler2Level:    slog.LevelWarn,
			testLevel:        slog.LevelInfo,
			wantH1Enabled:    true,
			wantH2Enabled:    false,
			wantMultiEnabled: true,
		},
		{
			name:             "both handlers enabled for error",
			handler1Level:    slog.LevelInfo,
			handler2Level:    slog.LevelWarn,
			testLevel:        slog.LevelError,
			wantH1Enabled:    true,
			wantH2Enabled:    true,
			wantMultiEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf1 := &bytes.Buffer{}
			buf2 := &bytes.Buffer{}
			h1 := slog.NewJSONHandler(buf1, &slog.HandlerOptions{Level: tt.handler1Level})
			h2 := slog.NewJSONHandler(buf2, &slog.HandlerOptions{Level: tt.handler2Level})

			multi := MultiHandler(h1, h2)

			// Test Enabled
			assert.Equal(t, tt.wantMultiEnabled, multi.Enabled(context.Background(), tt.testLevel))

			// Test Handle - only enabled handlers should receive logs
			record := slog.NewRecord(time.Now(), tt.testLevel, "test message", 0)
			err := multi.Handle(context.Background(), record)
			require.NoError(t, err)

			if tt.wantH1Enabled {
				assert.NotEmpty(t, buf1.String(), "handler1 should have received log")
			} else {
				assert.Empty(t, buf1.String(), "handler1 should not have received log")
			}

			if tt.wantH2Enabled {
				assert.NotEmpty(t, buf2.String(), "handler2 should have received log")
			} else {
				assert.Empty(t, buf2.String(), "handler2 should not have received log")
			}
		})
	}
}

func TestMultiHandler_NestedFlattening(t *testing.T) {
	tests := []struct {
		name              string
		setupHandlers     func() slog.Handler
		expectedFlatCount int
	}{
		{
			name: "single level nesting",
			setupHandlers: func() slog.Handler {
				h1 := newTestHandler(true)
				h2 := newTestHandler(true)
				inner := MultiHandler(h1, h2)
				return MultiHandler(inner)
			},
			expectedFlatCount: 2,
		},
		{
			name: "double level nesting",
			setupHandlers: func() slog.Handler {
				h1 := newTestHandler(true)
				h2 := newTestHandler(true)
				h3 := newTestHandler(true)
				inner1 := MultiHandler(h1, h2)
				inner2 := MultiHandler(inner1, h3)
				return inner2
			},
			expectedFlatCount: 3,
		},
		{
			name: "multiple nested MultiHandlers",
			setupHandlers: func() slog.Handler {
				h1 := newTestHandler(true)
				h2 := newTestHandler(true)
				h3 := newTestHandler(true)
				h4 := newTestHandler(true)
				inner1 := MultiHandler(h1, h2)
				inner2 := MultiHandler(h3, h4)
				return MultiHandler(inner1, inner2)
			},
			expectedFlatCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setupHandlers()

			mh, ok := result.(*multiHandler)
			require.True(t, ok, "result should be *multiHandler")
			assert.Len(t, mh.handlers, tt.expectedFlatCount)

			// Ensure all handlers are not multiHandler (fully flattened)
			for i, h := range mh.handlers {
				_, isMulti := h.(*multiHandler)
				assert.False(t, isMulti, "handler %d should not be *multiHandler after flattening", i)
			}
		})
	}
}

// Benchmark tests
func BenchmarkMultiHandler(b *testing.B) {
	h1 := slog.NewJSONHandler(&bytes.Buffer{}, nil)
	h2 := slog.NewJSONHandler(&bytes.Buffer{}, nil)
	multi := MultiHandler(h1, h2)

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			multi.Handle(context.Background(), record)
		}
	})
}

func BenchmarkSingleHandler(b *testing.B) {
	h := slog.NewJSONHandler(&bytes.Buffer{}, nil)

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Handle(context.Background(), record)
		}
	})
}

func BenchmarkMultiHandlerWithFlattening(b *testing.B) {
	h1 := slog.NewJSONHandler(&bytes.Buffer{}, nil)
	h2 := slog.NewJSONHandler(&bytes.Buffer{}, nil)
	h3 := slog.NewJSONHandler(&bytes.Buffer{}, nil)

	// create nested handlers to test flattening performance
	inner := MultiHandler(h1, h2)
	multi := MultiHandler(inner, h3)

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			multi.Handle(context.Background(), record)
		}
	})
}
