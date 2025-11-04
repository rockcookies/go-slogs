package slogs

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"
)

type HandleFunc func(ctx context.Context, hc *HandlerContext, rt time.Time, rl slog.Level, rm string, attrs []slog.Attr) (string, []slog.Attr)

// HandlerOptions are options for a Handler
type HandlerOptions struct {
	handleFunc HandleFunc
}

type Handler struct {
	next    slog.Handler
	handle  HandleFunc
	level   *slog.Level
	context *HandlerContext
}

type HandlerContext struct {
	Names []string
	Attrs *GroupOrAttrs
}

var _ slog.Handler = (*Handler)(nil)

// NewMiddleware creates a slogs.Handler slog.Handler middleware
// that conforms to [github.com/samber/slog-multi.Middleware] interface.
// It can be used with slogmulti methods such as Pipe to easily setup a pipeline of slog handlers:
//
//	slog.SetDefault(slog.New(slogmulti.
//		Pipe(slogs.NewMiddleware(&slogs.HandlerOptions{})).
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

func NewHandler(next slog.Handler) *Handler {
	return NewHandlerWithOptions(next, nil)
}

// NewHandlerWithOptions creates a Handler slog.Handler middleware that will Prepend and
// Append attributes to log lines. The attributes are extracted out of the log
// record's context by the provided AttrExtractor methods.
// It passes the final record and attributes off to the next handler when finished.
// If opts is nil, the default options are used.
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
// The handler ignores records whose level is lower.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	if h.level != nil {
		if *h.level < level {
			return false
		}
	}

	return h.next.Enabled(ctx, level)
}

// Handle de-duplicates all attributes and groups, then passes the new set of attributes to the next handler.
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

func (h *Handler) Clone() *Handler {
	h2 := *h
	hc2 := *h.context
	// Deep copy Names slice to avoid race conditions
	hc2.Names = slices.Clone(h.context.Names)
	h2.context = &hc2
	return &h2
}

func (h *Handler) withGroup(name string) *Handler {
	h2 := h.Clone()
	h2.context.Attrs = h.context.Attrs.WithGroup(name)
	return h2
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return h.withGroup(name)
}

func (h *Handler) withAttrs(attrs []slog.Attr) *Handler {
	h2 := h.Clone()
	h2.context.Attrs = h.context.Attrs.WithAttrs(attrs)
	return h2
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.withAttrs(attrs)
}

func (h *Handler) WithLevel(level slog.Level) *Handler {
	h2 := h.Clone()
	h2.level = &level
	return h2
}

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

	if len(hc.Names) > 0 {
		rm = "[" + strings.Join(hc.Names, ".") + "] " + rm
	}

	return rm, attrs
}
