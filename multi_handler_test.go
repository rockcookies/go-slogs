package slogs

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockHandler is a mock handler used for testing
type mockHandler struct {
	mu       sync.Mutex
	enabled  bool
	records  []slog.Record
	err      error
	callFunc func()
	mutate   func(*slog.Record)
}

// newMockHandler creates a test mock handler with the specified enabled state.
func newMockHandler(enabled bool) *mockHandler {
	return &mockHandler{
		enabled: enabled,
		records: make([]slog.Record, 0),
	}
}

func (h *mockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.enabled
}

func (h *mockHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.callFunc != nil {
		h.callFunc()
	}

	if h.mutate != nil {
		h.mutate(&r)
	}

	h.records = append(h.records, r)
	return h.err
}

func (h *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.records = make([]slog.Record, 0)
	return &newHandler
}

func (h *mockHandler) WithGroup(name string) slog.Handler {
	newHandler := *h
	newHandler.records = make([]slog.Record, 0)
	return &newHandler
}

func (h *mockHandler) getRecords() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]slog.Record{}, h.records...)
}

// recordHasAttr checks if a record contains the specified attribute.
func recordHasAttr(r slog.Record, key, want string) bool {
	found := false
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == key && a.Value.String() == want {
			found = true
		}
		return true
	})
	return found
}

func TestMultiHandler(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	h1 := slog.NewJSONHandler(buf1, nil)
	h2 := slog.NewJSONHandler(buf2, nil)

	multi := MultiHandler(h1, h2)

	logger := slog.New(multi)
	logger.Info("test message", "key", "value")

	// both handlers should receive the log
	if !strings.Contains(buf1.String(), "test message") {
		t.Fatalf("buf1 should contain message; got: %s", buf1.String())
	}
	if !strings.Contains(buf2.String(), "test message") {
		t.Fatalf("buf2 should contain message; got: %s", buf2.String())
	}
}

func TestMultiHandler_Enabled(t *testing.T) {
	cases := []struct {
		name     string
		handlers []slog.Handler
		want     bool
	}{
		{
			name:     "any enabled returns true",
			handlers: []slog.Handler{newMockHandler(true), newMockHandler(false)},
			want:     true,
		},
		{
			name:     "none enabled returns false",
			handlers: []slog.Handler{newMockHandler(false), newMockHandler(false)},
			want:     false,
		},
		{
			name:     "empty list returns false",
			handlers: nil,
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			multi := MultiHandler(tc.handlers...)
			got := multi.Enabled(context.Background(), slog.LevelInfo)
			if got != tc.want {
				t.Fatalf("Enabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMultiHandler_ErrorAggregation(t *testing.T) {
	t.Run("collects errors from multiple handlers", func(t *testing.T) {
		h1 := newMockHandler(true)
		h1.err = errors.New("handler 1 error")

		h2 := newMockHandler(true)
		h2.err = errors.New("handler 2 error")

		multi := MultiHandler(h1, h2)

		record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
		err := multi.Handle(context.Background(), record)

		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "handler 1 error") || !strings.Contains(err.Error(), "handler 2 error") {
			t.Fatalf("expected aggregated errors, got: %v", err)
		}
	})

	t.Run("returns nil when no errors occur", func(t *testing.T) {
		h1 := newMockHandler(true)
		h2 := newMockHandler(true)

		multi := MultiHandler(h1, h2)

		record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
		err := multi.Handle(context.Background(), record)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMultiHandler_WithAttrs(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	h1 := slog.NewJSONHandler(buf1, nil)
	h2 := slog.NewJSONHandler(buf2, nil)

	multi := MultiHandler(h1, h2)
	attrs := []slog.Attr{slog.String("key", "value")}

	newMulti := multi.WithAttrs(attrs)

	// verify that the returned type is multiHandler
	if _, ok := newMulti.(multiHandler); !ok {
		t.Fatalf("WithAttrs should return multiHandler type")
	}

	// verify that attributes are applied to all downstream handlers
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := newMulti.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if !strings.Contains(buf1.String(), "\"key\":\"value\"") {
		t.Fatalf("buf1 should contain attr; got: %s", buf1.String())
	}
	if !strings.Contains(buf2.String(), "\"key\":\"value\"") {
		t.Fatalf("buf2 should contain attr; got: %s", buf2.String())
	}
}

func TestMultiHandler_WithGroup(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	h1 := slog.NewJSONHandler(buf1, nil)
	h2 := slog.NewJSONHandler(buf2, nil)

	multi := MultiHandler(h1, h2)

	newMulti := multi.WithGroup("test").WithAttrs([]slog.Attr{slog.String("key", "value")})

	// verify that the returned type is multiHandler
	if _, ok := newMulti.(multiHandler); !ok {
		t.Fatalf("WithGroup should return multiHandler type")
	}

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := newMulti.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	// group name should wrap the attributes
	if !strings.Contains(buf1.String(), "\"test\":{\"key\":\"value\"}") {
		t.Fatalf("buf1 should contain grouped attr; got: %s", buf1.String())
	}
	if !strings.Contains(buf2.String(), "\"test\":{\"key\":\"value\"}") {
		t.Fatalf("buf2 should contain grouped attr; got: %s", buf2.String())
	}
}

func TestMultiHandler_RecordCloning(t *testing.T) {
	h1 := newMockHandler(true)
	// h1 modifies the record, simulating improper modification
	h1.mutate = func(r *slog.Record) {
		r.AddAttrs(slog.String("mutated", "yes"))
	}

	h2 := newMockHandler(true)

	multi := MultiHandler(h1, h2)

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	record.AddAttrs(slog.String("key", "value"))

	if err := multi.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	// verify that both handlers received the record
	records1 := h1.getRecords()
	records2 := h2.getRecords()
	if len(records1) != 1 || len(records2) != 1 {
		t.Fatalf("expected each handler to receive 1 record, got %d and %d", len(records1), len(records2))
	}

	// verify that the second handler was not affected by the first handler's modification
	if !recordHasAttr(records1[0], "mutated", "yes") {
		t.Fatalf("handler1 record should have mutated attr")
	}
	if recordHasAttr(records2[0], "mutated", "yes") {
		t.Fatalf("handler2 record should NOT have mutated attr")
	}
}

func TestMultiHandler_EdgeCases(t *testing.T) {
	t.Run("handles nil handlers", func(t *testing.T) {
		buf := &bytes.Buffer{}
		h := slog.NewJSONHandler(buf, nil)

		// including nil handlers should not cause panic
		multi := MultiHandler(nil, h, nil)

		logger := slog.New(multi)
		logger.Info("test message")

		if !strings.Contains(buf.String(), "test message") {
			t.Fatalf("buf should contain message; got: %s", buf.String())
		}
	})

	t.Run("handles empty handler list", func(t *testing.T) {
		multi := MultiHandler()

		record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
		err := multi.Handle(context.Background(), record)
		// empty handler list should return nil error
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMultiHandler_ConcurrentAccess(t *testing.T) {
	h1 := newMockHandler(true)
	h2 := newMockHandler(true)

	multi := MultiHandler(h1, h2)

	var wg sync.WaitGroup
	const numGoroutines = 10
	const numLogs = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numLogs; j++ {
				record := slog.NewRecord(time.Now(), slog.LevelInfo,
					"test", 0)
				multi.Handle(context.Background(), record)
			}
		}(i)
	}

	wg.Wait()

	// verify that all logs were processed
	records1 := h1.getRecords()
	records2 := h2.getRecords()

	if len(records1) != numGoroutines*numLogs {
		t.Fatalf("handler1 got %d records, want %d", len(records1), numGoroutines*numLogs)
	}
	if len(records2) != numGoroutines*numLogs {
		t.Fatalf("handler2 got %d records, want %d", len(records2), numGoroutines*numLogs)
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