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
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.store[key]; !ok {
		c.store[key] = time.AfterFunc(ttl, func() {
			delete(c.store, key)
		})
		return true, nil
	} else {
		return false, nil
	}
}

func (c *MemoryLock) Release(ctx context.Context, key string) error {
	//c.mu.Lock()
	//defer c.mu.Unlock()
	delete(c.store, key)
	return nil
}
