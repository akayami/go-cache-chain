package cache

import (
	"context"
	"sync"
	"time"
)

type MemoryLock struct {
	store map[string]*time.Timer
	mu    sync.Mutex
}

func NewMemoryLock() *MemoryLock {
	return &MemoryLock{store: map[string]*time.Timer{}, mu: sync.Mutex{}}
}

func (c *MemoryLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if _, ok := c.store[key]; !ok {
		c.mu.Lock()
		c.store[key] = time.AfterFunc(ttl, func() {
			c.mu.Lock()
			delete(c.store, key)
			c.mu.Unlock()
		})
		c.mu.Unlock()
		return true, nil
	} else {
		return false, nil
	}
}

func (c *MemoryLock) Release(ctx context.Context, key string) error {
	c.mu.Lock()
	delete(c.store, key)
	c.mu.Unlock()
	return nil
}
