// Package slogs provides structured logging utilities and handlers
// that extend the standard log/slog package functionality.
package slogs

import (
	"context"
	"errors"
	"log/slog"
)

// multiHandler is an implementation that broadcasts logs to multiple handlers.
// It implements the slog.Handler interface, ensuring full compatibility with the standard library.
// multiHandler broadcasts each log record to all downstream handlers,
// ensuring each handler receives a cloned copy of the record to prevent interference.
type multiHandler []slog.Handler

// MultiHandler creates a new handler that broadcasts logs to all provided handlers.
//
// If nil handlers are passed in, they will be filtered out and will not affect broadcasting.
// Passing an empty list will return a handler that does not process any logs.
//
// Example:
//
//	h1 := slog.NewJSONHandler(os.Stdout, nil)
//	h2 := slog.NewTextHandler(os.Stderr, nil)
//	multi := slogs.MultiHandler(h1, h2)
//	logger := slog.New(multi)
//	logger.Info("this log will be output to both stdout and stderr")
func MultiHandler(handlers ...slog.Handler) slog.Handler {
	// Filter out nil handlers
	var valid []slog.Handler
	for _, h := range handlers {
		if h != nil {
			valid = append(valid, h)
		}
	}
	return multiHandler(valid)
}

// Enabled reports whether any downstream handler will process logs at the specified level.
//
// It returns true as long as at least one handler is enabled.
// If the handler list is empty or all handlers are disabled, it returns false.
func (h multiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h {
		if h[i].Enabled(ctx, l) {
			return true // enable if any handler needs it
		}
	}
	return false
}

// Handle broadcasts the log record to all enabled downstream handlers.
//
// For each enabled handler, it receives a cloned copy of the record
// to prevent one handler from modifying the record and affecting other handlers.
//
// Errors from all handlers will be collected and merged using errors.Join.
// If all handlers process successfully, it returns nil.
func (h multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error

	for i := range h {
		// Check Enabled again inside Handle to ensure logs are only sent to needed handlers
		if h[i].Enabled(ctx, r.Level) {
			// Clone Record to prevent handler modification from affecting subsequent handlers
			if err := h[i].Handle(ctx, r.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...) // merge all handler errors
}

// WithAttrs returns a new multiHandler where each downstream handler has the same attributes added.
//
// Each handler creates its own WithAttrs copy, ensuring attribute isolation.
func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h))
	for i := range h {
		handlers = append(handlers, h[i].WithAttrs(attrs))
	}
	return multiHandler(handlers)
}

// WithGroup returns a new multiHandler where each downstream handler has the same group name added.
//
// Each handler creates its own WithGroup copy, ensuring group isolation.
func (h multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h))
	for i := range h {
		handlers = append(handlers, h[i].WithGroup(name))
	}
	return multiHandler(handlers)
}
