package buffer

import "github.com/rockcookies/go-slogs/internal/pool"

// A Pool is a type-safe wrapper around a sync.Pool.
type Pool struct {
	p *pool.Pool[*Buffer]
}

// NewPool constructs a new Pool.
func NewPool() Pool {
	return Pool{
		p: pool.New(func() *Buffer {
			return &Buffer{
				bs: make([]byte, 0, _size),
			}
		}),
	}
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (p Pool) Get() *Buffer {
	buf := p.p.Get()
	buf.Reset()
	buf.pool = p
	return buf
}

func (p Pool) put(buf *Buffer) {
	p.p.Put(buf)
}
