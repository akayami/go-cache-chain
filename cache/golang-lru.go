package cache

import (
	"context"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"time"
)

type GolangLRUCacheBackend[K comparable, V any] struct {
	Backend
	provider *lru.Cache[K, V]
}

func NewGolangLRUBackend[K comparable, V any](cache *lru.Cache[K, V]) *GolangLRUCacheBackend[K, V] {
	return &GolangLRUCacheBackend[K, V]{provider: cache}
}

func (c *GolangLRUCacheBackend[K, V]) Get(ctx context.Context, key K) *CacheBackendResult {
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

func (c *GolangLRUCacheBackend[K, V]) Set(ctx context.Context, key K, val V, ttl time.Duration) (V, error) {
	c.provider.Add(key, val)
	return val, nil
}

func (c *GolangLRUCacheBackend[K, V]) Del(ctx context.Context, key K) error {
	c.provider.Remove(key)
	return nil
}
