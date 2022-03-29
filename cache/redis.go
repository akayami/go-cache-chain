package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisCacheBackend struct {
	Backend
	ctx    context.Context
	client *redis.Client
}

func NewRedisCacheBackend(c context.Context, client *redis.Client) *RedisCacheBackend {
	backend := Backend{"Redis"}
	return &RedisCacheBackend{
		backend,
		c,
		client,
	}
}

func (t *RedisCacheBackend) Get(key string) *CacheBackendResult {
	res := NewCacheBackendResult()
	val, err := t.client.Get(t.ctx, key).Result()
	if err == redis.Nil {
		res.setNil(true)
	} else if err != nil {
		res.setError(err)
	} else {
		res.setValue(val)
	}
	return res
}

func (t *RedisCacheBackend) Set(key string, value string, ttl time.Duration) error {
	return t.client.Set(t.ctx, key, value, ttl).Err()
}

func (t *RedisCacheBackend) Del(key string) error {
	return t.client.Del(t.ctx, key).Err()
}
