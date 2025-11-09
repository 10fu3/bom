package alias

import "sync"

// AtomicCounter issues deterministic integers, optionally in a concurrent-safe manner.
type AtomicCounter struct {
	mu  sync.Mutex
	val int
}

// Next returns the next integer (starting from 1).
func (c *AtomicCounter) Next() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.val++
	return c.val
}

// Reset zeros the counter.
func (c *AtomicCounter) Reset() {
	c.mu.Lock()
	c.val = 0
	c.mu.Unlock()
}
