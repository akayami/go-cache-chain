package cache

import (
	"context"
	"sync"
	"time"
)

type MemoryBackend struct {
	Backend
	size    int
	store   map[string]string
	list    []string
	storeMU sync.Mutex
	listMU  sync.Mutex
}

func NewMemoryBackend(size int) *MemoryBackend {
	return &MemoryBackend{Backend{name: "Memory", marshal: true}, size, map[string]string{}, []string{}, sync.Mutex{}, sync.Mutex{}}
}

func (c *MemoryBackend) Get(ctx context.Context, key string) *CacheBackendResult {
	res := NewCacheBackendResult()
	if val, ok := c.store[key]; ok {
		res.Value = val
		c.listMU.Lock()
		first := c.list[0]
		c.list = c.list[1:]
		c.list = append(c.list, first)
		c.listMU.Unlock()
	} else {
		res.Nil = true
	}
	return res
}

func (c *MemoryBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) (string, error) {
	if _, ok := c.store[key]; ok == false {
		c.storeMU.Lock()
		c.store[key] = value
		c.storeMU.Unlock()
		if len(c.list) >= c.size {
			// Remove first element
			c.storeMU.Lock()
			delete(c.store, c.list[0])
			c.storeMU.Unlock()
			c.listMU.Lock()
			c.list = c.list[1:]
			c.listMU.Unlock()
		}
		// Add New element
		c.listMU.Lock()
		c.list = append(c.list, key)
		c.listMU.Unlock()
	}
	c.storeMU.Lock()
	c.store[key] = value
	c.storeMU.Unlock()
	return value, nil
}

func (c *MemoryBackend) Del(ctx context.Context, key string) error {
	// delete key
	c.storeMU.Lock()
	delete(c.store, key)
	c.storeMU.Unlock()
	// cleanup the list
	for i, v := range c.list {
		if v == key {
			// Remove index from list
			c.storeMU.Lock()
			c.list = append(c.list[:i], c.list[i+1:]...)
			c.storeMU.Unlock()
		}
	}
	return nil
}

func (c *MemoryBackend) GetSize() int {
	return len(c.list)
}
