package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisCacheBackend struct {
	Backend
	client *redis.Client
}

func NewRedisCacheBackend(client *redis.Client) *RedisCacheBackend {
	backend := Backend{"Redis", true}
	return &RedisCacheBackend{
		backend,
		client,
	}
}

func (t *RedisCacheBackend) Get(ctx context.Context, key string) *CacheBackendResult {
	res := NewCacheBackendResult()
	val, err := t.client.Get(ctx, key).Result()
	if err == redis.Nil {
		res.Nil = true
	} else if err != nil {
		res.Err = err
	} else {
		res.Value = val
	}
	return res
}

func (t *RedisCacheBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) (string, error) {
	return value, t.client.Set(ctx, key, value, ttl).Err()
}

func (t *RedisCacheBackend) Del(ctx context.Context, key string) error {
	return t.client.Del(ctx, key).Err()
}
