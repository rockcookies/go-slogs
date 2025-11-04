// Package attr provides utility functions for converting variadic arguments to slog attributes.
//
// This package contains internal implementation details copied from the Go standard library's
// slog package to maintain consistency with slog's argument processing behavior.
package attr

import "log/slog"

// badKey is the key used when an argument cannot be properly converted to an attribute.
// This matches the behavior of the Go standard library's slog package.
const badKey = "!BADKEY"

// ArgsToAttrSlice converts a slice of arguments into a slice of slog.Attr.
//
// This function handles mixed argument types:
//   - slog.Attr values are used directly
//   - String-value pairs are converted to key-value attributes
//   - Other values are treated as having a missing key and use "!BADKEY"
//
// This implementation is copied from the Go standard library's slog package
// to ensure consistent behavior with slog.Logger.Log.
func ArgsToAttrSlice(args []any) []slog.Attr {
	var (
		attr  slog.Attr
		attrs []slog.Attr
	)
	for len(args) > 0 {
		attr, args = ArgsToAttr(args)
		attrs = append(attrs, attr)
	}
	return attrs
}

// ArgsToAttr extracts a single attribute from the beginning of the args slice.
//
// It returns the extracted attribute and the remaining unconsumed portion of the slice.
// The conversion rules are:
//   - If args[0] is an slog.Attr, it is returned directly
//   - If args[0] is a string and there's a second element, they form a key-value pair
//   - If args[0] is a string with no second element, it uses "!BADKEY" as the key
//   - Otherwise, the value uses "!BADKEY" as the key
//
// This implementation is copied from the Go standard library's slog package.
func ArgsToAttr(args []any) (slog.Attr, []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]

	case slog.Attr:
		return x, args[1:]

	default:
		return slog.Any(badKey, x), args[1:]
	}
}
