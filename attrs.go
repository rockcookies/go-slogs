package slogs

import "log/slog"

// GroupOrAttrs represents a node in a linked list that holds either a group name or attributes.
//
// This structure forms a chain where each node can contain either:
//   - A group name (for attribute grouping)
//   - A list of slog.Attr (for actual attributes)
//
// The linked list is built from newest to oldest, with each node pointing to its parent.
// This design enables efficient attribute and group management in the handler chain.
//
// Implementation courtesy of https://github.com/jba/slog/blob/b5eef75b08965b871bd5214891313b73d5a30432/withsupport/withsupport.go
type GroupOrAttrs struct {
	group string        // group name if non-empty
	attrs []slog.Attr   // attrs if non-empty
	next  *GroupOrAttrs // parent node in the linked list
}

// WithGroup creates a new GroupOrAttrs node with the given group name and links it to the current node.
//
// If name is empty, the group is inlined (no grouping occurs) and the current node is returned unchanged.
// This method is safe to call on a nil receiver.
//
// Example:
//
//	var g *GroupOrAttrs
//	g = g.WithGroup("http")          // Creates a group named "http"
//	g = g.WithAttrs([]slog.Attr{...}) // Adds attributes under the "http" group
func (g *GroupOrAttrs) WithGroup(name string) *GroupOrAttrs {
	// Empty-name groups are inlined as if they didn't exist
	if name == "" {
		return g
	}
	return &GroupOrAttrs{
		group: name,
		next:  g,
	}
}

// WithAttrs creates a new GroupOrAttrs node with the given attributes and links it to the current node.
//
// If attrs is empty, no new node is created and the current node is returned unchanged.
// This method is safe to call on a nil receiver.
//
// Example:
//
//	var g *GroupOrAttrs
//	g = g.WithAttrs([]slog.Attr{slog.String("key", "value")})
//	g = g.WithAttrs([]slog.Attr{slog.Int("count", 42)})
func (g *GroupOrAttrs) WithAttrs(attrs []slog.Attr) *GroupOrAttrs {
	if len(attrs) == 0 {
		return g
	}
	return &GroupOrAttrs{
		attrs: attrs,
		next:  g,
	}
}
