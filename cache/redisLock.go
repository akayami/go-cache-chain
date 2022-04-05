package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisLock struct {
	client *redis.Client
	prefix string
}

func NewRedisLock(client *redis.Client) *RedisLock {
	return &RedisLock{client: client, prefix: "lock_"}
}

func NewRedisLockWithPrefix(client *redis.Client, prefix string) *RedisLock {
	return &RedisLock{client: client, prefix: prefix}
}

func (c *RedisLock) getKey(key string) string {
	return c.prefix + key
}

func (c *RedisLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	res, err := c.client.SetNX(ctx, c.getKey(key), true, ttl).Result()
	if err != nil {
		return false, err
	}
	return res, err
}

func (c *RedisLock) Release(ctx context.Context, key string) error {
	_, err := c.client.Del(ctx, c.getKey(key)).Result()
	if err != nil {
		return err
	}
	return nil
}
