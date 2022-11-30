package cache

import (
	"context"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"time"
)

type GolangLRUCacheBackend struct {
	Backend
	provider *lru.Cache
}

func NewGolangLRUBackend(cache *lru.Cache) *GolangLRUCacheBackend {
	return &GolangLRUCacheBackend{provider: cache}
}

func (c *GolangLRUCacheBackend) Get(ctx context.Context, key string) *CacheBackendResult {
	res := NewCacheBackendResult()
	value, ok := c.provider.Get(key)
	if ok {
		res.Nil = false
		res.Value = fmt.Sprint(value)
	} else {
		res.Nil = true
	}
	return res
}

func (c *GolangLRUCacheBackend) Set(ctx context.Context, key string, val string, ttl time.Duration) (string, error) {
	c.provider.Add(key, val)
	return val, nil
}

func (c *GolangLRUCacheBackend) Del(ctx context.Context, key string) error {
	c.provider.Remove(key)
	return nil
}
