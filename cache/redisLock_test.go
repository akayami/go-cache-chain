package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"testing"
)

func TestRedisLock(t *testing.T) {

	var ctx = context.TODO()
	var redisOptions = redis.Options{}
	client := redis.NewClient(&redisOptions)
	h := NewRedisLock(client)

	CommonLockTests(ctx, h, t)
}
