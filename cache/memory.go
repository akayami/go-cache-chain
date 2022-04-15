package cache

import (
	"context"
	"time"
)

type MemoryBackend struct {
	Backend
	size  int
	store map[string]string
	list  []string
}

func NewMemoryBackend(size int) *MemoryBackend {
	return &MemoryBackend{Backend{name: "Memory", marshal: true}, size, map[string]string{}, []string{}}
}

func (c *MemoryBackend) Get(ctx context.Context, key string) *CacheBackendResult {
	res := NewCacheBackendResult()
	if val, ok := c.store[key]; ok {
		res.setValue(val)
		first := c.list[0]
		c.list = c.list[1:]
		c.list = append(c.list, first)
	} else {
		res.setNil(true)
	}
	return res
}

func (c *MemoryBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if _, ok := c.store[key]; ok == false {
		c.store[key] = value
		if len(c.list) >= c.size {
			// Remove first element
			delete(c.store, c.list[0])
			c.list = c.list[1:]
		}
		// Add New element
		c.list = append(c.list, key)
	}
	c.store[key] = value
	return nil
}

func (c *MemoryBackend) Del(ctx context.Context, key string) error {
	// delete key
	delete(c.store, key)
	// cleanup the list
	for i, v := range c.list {
		if v == key {
			// Remove index from list
			c.list = append(c.list[:i], c.list[i+1:]...)
		}
	}
	return nil
}

func (c *MemoryBackend) GetSize() int {
	return len(c.list)
}
