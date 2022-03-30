package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisLock struct {
	ctx    context.Context
	client *redis.Client
	prefix string
}

func NewRedisLock(c context.Context, client *redis.Client) *RedisLock {
	return &RedisLock{ctx: c, client: client, prefix: "lock_"}
}

func NewRedisLockWithPrefix(c context.Context, client *redis.Client, prefix string) *RedisLock {
	return &RedisLock{ctx: c, client: client, prefix: prefix}
}

func (c *RedisLock) getKey(key string) string {
	return c.prefix + key
}

func (c *RedisLock) Acquire(key string, ttl time.Duration) (bool, error) {
	res, err := c.client.SetNX(c.ctx, c.getKey(key), true, ttl).Result()
	if err != nil {
		return false, err
	}
	return res, err
}

func (c *RedisLock) Release(key string) error {
	_, err := c.client.Del(c.ctx, c.getKey(key)).Result()
	if err != nil {
		return err
	}
	return nil
}
