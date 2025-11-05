// Package bufferpool houses zap's shared internal buffer pool. Third-party
// packages can recreate the same functionality with buffers.NewPool.
package bufferpool

import "github.com/rockcookies/go-slogs/buffer"

var (
	_pool = buffer.NewPool()
	// Get retrieves a buffer from the pool, creating one if necessary.
	Get = _pool.Get
)
