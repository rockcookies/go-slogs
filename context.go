package slogs

import (
	"context"
	"log/slog"
	"slices"

	"github.com/rockcookies/go-slogs/internal/attr"
)

// prependKey is the context key for storing prepended attributes.
type prependKey struct{}

// appendKey is the context key for storing appended attributes.
type appendKey struct{}

// Prepend adds attributes to a context that will be prepended to the start of log records.
//
// The attributes are added at the root level of the log record, not in any groups.
// This is useful for adding common context like request IDs, trace IDs, or user information
// that should appear at the beginning of every log entry.
//
// If parent is nil, a new background context is created. The args are converted to attributes
// using the same rules as slog.Logger.Log: string pairs become key-value attributes, and
// slog.Attr values are used directly.
//
// Example:
//
//	ctx := slogs.Prepend(context.Background(), "request_id", "abc-123", "user", "alice")
//	logger.InfoContext(ctx, "Processing request")
//	// Output: {"request_id":"abc-123","user":"alice","msg":"Processing request"}
func Prepend(parent context.Context, args ...any) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(prependKey{}).([]slog.Attr); ok {
		// Use slices.Clip to ensure we create a new slice with a separate backing array.
		// This prevents accidental modifications to the original slice and avoids
		// potential race conditions when the context is used across goroutines.
		return context.WithValue(parent, prependKey{}, append(slices.Clip(v), attr.ArgsToAttrSlice(args)...))
	}
	return context.WithValue(parent, prependKey{}, attr.ArgsToAttrSlice(args))
}

// ExtractPrepended retrieves the prepended attributes stored in the context.
//
// Returns nil if no prepended attributes are found in the context.
// The returned slice should not be modified as it may cause race conditions.
//
// This function is typically used by custom handlers to extract context attributes.
// Most users should use Prepend to add attributes rather than calling this directly.
func ExtractPrepended(ctx context.Context) []slog.Attr {
	if v, ok := ctx.Value(prependKey{}).([]slog.Attr); ok {
		return v
	}
	return nil
}

// Append adds attributes to a context that will be appended to the end of log records.
//
// Unlike Prepend, appended attributes respect the current group structure established by
// Logger.WithGroup. This means they will be placed inside any active groups.
//
// If parent is nil, a new background context is created. The args are converted to attributes
// using the same rules as slog.Logger.Log.
//
// Example:
//
//	logger := logger.WithGroup("http")
//	ctx := slogs.Append(context.Background(), "duration", "100ms")
//	logger.InfoContext(ctx, "Request completed", "status", 200)
//	// Output: {"http":{"status":200,"duration":"100ms"},"msg":"Request completed"}
func Append(parent context.Context, args ...any) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(appendKey{}).([]slog.Attr); ok {
		// Use slices.Clip to ensure we create a new slice with a separate backing array.
		// This prevents accidental modifications to the original slice and avoids
		// potential race conditions when the context is used across goroutines.
		return context.WithValue(parent, appendKey{}, append(slices.Clip(v), attr.ArgsToAttrSlice(args)...))
	}
	return context.WithValue(parent, appendKey{}, attr.ArgsToAttrSlice(args))
}

// ExtractAppended retrieves the appended attributes stored in the context.
//
// Returns nil if no appended attributes are found in the context.
// The returned slice should not be modified as it may cause race conditions.
//
// This function is typically used by custom handlers to extract context attributes.
// Most users should use Append to add attributes rather than calling this directly.
func ExtractAppended(ctx context.Context) []slog.Attr {
	if v, ok := ctx.Value(appendKey{}).([]slog.Attr); ok {
		return v
	}
	return nil
}
