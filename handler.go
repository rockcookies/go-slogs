package slogs

import (
	"context"
	"log/slog"
	"slices"
	"time"
)

// HandleFunc is a function that processes log records and their attributes.
//
// It receives the context, handler context (including names and attribute groups),
// record time, level, message, and attributes. It returns the potentially modified
// message and attributes.
//
// Custom HandleFunc implementations can be used to:
//   - Transform or filter attributes
//   - Modify log messages
//   - Add custom formatting
//   - Implement security features like sensitive data masking
type HandleFunc func(ctx context.Context, hc *HandlerContext, rt time.Time, rl slog.Level, rm string, attrs []slog.Attr) (string, []slog.Attr)

// HandlerOptions configures the behavior of a Handler.
type HandlerOptions struct {
	// handleFunc is the function that processes log records.
	// If nil, DefaultHandleFunc is used.
	handleFunc HandleFunc
}

// Handler is a middleware slog.Handler that manages attribute groups and context attributes.
//
// It wraps another slog.Handler and processes log records before passing them to the next handler.
// Handler supports:
//   - Prepending and appending context attributes
//   - Attribute grouping via WithGroup
//   - Named loggers for better log organization
//   - Custom handle functions for advanced processing
type Handler struct {
	next    slog.Handler
	handle  HandleFunc
	level   slog.Leveler
	context *HandlerContext
}

// HandlerContext holds the state for a handler instance.
//
// It maintains the chain of logger names and the linked list of attribute groups
// that will be applied to log records.
type HandlerContext struct {
	Name string

	// Attrs is the linked list of attribute groups.
	// Newest groups are at the head, forming a chain to the oldest.
	Attrs *GroupOrAttrs
}

var _ slog.Handler = (*Handler)(nil)

// NewMiddleware creates a Handler middleware constructor compatible with slogmulti.
//
// This is designed to work with github.com/samber/slog-multi for building handler pipelines.
// The returned function creates a Handler that wraps the next handler in the chain.
//
// Example:
//
//	slog.SetDefault(slog.New(
//		slogmulti.Pipe(slogs.NewMiddleware(&slogs.HandlerOptions{})).
//		Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})),
//	))
func NewMiddleware(options *HandlerOptions) func(slog.Handler) slog.Handler {
	return func(next slog.Handler) slog.Handler {
		return NewHandlerWithOptions(
			next,
			options,
		)
	}
}

// NewHandler creates a new Handler with default options.
//
// The Handler wraps the provided next handler and uses DefaultHandleFunc for processing.
// This is a convenience function equivalent to NewHandlerWithOptions(next, nil).
//
// Panics if next is nil.
func NewHandler(next slog.Handler) *Handler {
	return NewHandlerWithOptions(next, nil)
}

// NewHandlerWithOptions creates a Handler with custom options.
//
// The Handler wraps the next handler in the chain and applies the specified options.
// If opts is nil, default options are used. If opts.handleFunc is nil, DefaultHandleFunc is used.
//
// Panics if next is nil.
//
// Example:
//
//	opts := &slogs.HandlerOptions{
//		HandleFunc: func(ctx context.Context, hc *slogs.HandlerContext,
//			rt time.Time, rl slog.Level, rm string, attrs []slog.Attr) (string, []slog.Attr) {
//			// Custom processing
//			return rm, attrs
//		},
//	}
//	handler := slogs.NewHandlerWithOptions(baseHandler, opts)
func NewHandlerWithOptions(next slog.Handler, opts *HandlerOptions) *Handler {
	if next == nil {
		panic("slogs: next handler cannot be nil")
	}

	if opts == nil {
		opts = &HandlerOptions{}
	}

	handlerFunc := opts.handleFunc
	if handlerFunc == nil {
		handlerFunc = DefaultHandleFunc
	}

	return &Handler{
		next:    next,
		handle:  handlerFunc,
		context: &HandlerContext{},
	}
}

// Enabled reports whether the handler handles records at the given level.
//
// This respects both the handler's own level setting (if configured via WithLevel)
// and the next handler's level settings. The record is enabled only if both this
// handler and the next handler would handle it.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	if h.level != nil {
		// If the incoming level is less than the configured minimum level, disable it
		if level < h.level.Level() {
			return false
		}
	}

	return h.next.Enabled(ctx, level)
}

// Handle processes a log record and passes it to the next handler.
//
// It extracts all attributes from the record, processes them through the handle function
// (which may add context attributes, apply grouping, and add names), and creates a new
// record with the processed message and attributes.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	// Collect all attributes from the record (which is the most recent attribute set).
	// These attributes are ordered from oldest to newest, and our collection will be too.
	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	message := r.Message
	message, attrs = h.handle(ctx, h.context, r.Time, r.Level, message, attrs)

	// Add all attributes to new record (because old record has all the old attributes as private members)
	newR := &slog.Record{
		Time:    r.Time,
		Level:   r.Level,
		Message: message,
		PC:      r.PC,
	}

	// Add attributes back in
	newR.AddAttrs(attrs...)
	return h.next.Handle(ctx, *newR)
}

// Clone creates a shallow copy of the handler with a deep copy of mutable state.
//
// The Names slice in the handler context is cloned to prevent race conditions
// when multiple goroutines use derived handlers.
func (h *Handler) Clone() *Handler {
	h2 := *h
	hc2 := *h.context
	h2.context = &hc2
	return &h2
}

// withGroup returns a new Handler with the given group name added to the attribute chain.
func (h *Handler) withGroup(name string) *Handler {
	h2 := h.Clone()
	h2.context.Attrs = h.context.Attrs.WithGroup(name)
	return h2
}

// WithGroup returns a Handler that starts a group.
//
// The keys of all attributes added through this handler will be qualified by the given name.
// This implements the slog.Handler interface requirement.
func (h *Handler) WithGroup(name string) slog.Handler {
	return h.withGroup(name)
}

// withAttrs returns a new Handler with the given attributes added to the attribute chain.
func (h *Handler) withAttrs(attrs []slog.Attr) *Handler {
	h2 := h.Clone()
	h2.context.Attrs = h.context.Attrs.WithAttrs(attrs)
	return h2
}

// WithAttrs returns a Handler whose attributes include the given attributes.
//
// This implements the slog.Handler interface requirement.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.withAttrs(attrs)
}

// WithLevel returns a new Handler with the specified minimum log level.
//
// Records below this level will be discarded before reaching the next handler.
// This is useful for creating loggers with different verbosity levels.
func (h *Handler) WithLevel(level slog.Leveler) *Handler {
	h2 := h.Clone()
	h2.level = level
	return h2
}

// Named returns a new Handler with the given name set as the logger's name.
func (h *Handler) Named(name string) *Handler {
	h2 := h.Clone()
	h2.context.Name = name
	return h2
}

// Name returns the handler's name.
func (h *Handler) Name() string {
	return h.context.Name
}

// DefaultHandleFunc is the default handler function used when no custom HandleFunc is provided.
//
// It implements the standard slogs behavior:
//  1. Appends context attributes from Append() to the end
//  2. Processes the attribute group chain, applying groups and flattening attributes
//  3. Prepends context attributes from Prepend() to the start
//  4. Prefixes the message with logger names if any (e.g., "[service.database]")
//
// This function maintains attribute ordering and ensures proper group structure.
func DefaultHandleFunc(ctx context.Context, hc *HandlerContext, rt time.Time, rl slog.Level, rm string, attrs []slog.Attr) (string, []slog.Attr) {
	// Add our 'appended' context attributes to the end
	appended := ExtractAppended(ctx)
	attrs = append(attrs, appended...)

	// Iterate through the goa (group Or Attributes) linked list, which is ordered from newest to oldest
	for g := hc.Attrs; g != nil; g = g.next {
		if g.group != "" {
			// If a group, but all the previous attributes (the newest ones) in it
			attrs = []slog.Attr{{
				Key:   g.group,
				Value: slog.GroupValue(attrs...),
			}}
		} else {
			// Prepend to the front of finalAttrs, thereby making finalAttrs ordered from oldest to newest
			attrs = append(slices.Clip(g.attrs), attrs...)
		}
	}

	// Add our 'prepended' context attributes to the start.
	// Go in reverse order, since each is prepending to the front.
	prepended := ExtractPrepended(ctx)
	attrs = append(prepended, attrs...)

	if hc.Name != "" {
		rm = "[" + hc.Name + "] " + rm
	}

	return rm, attrs
}
