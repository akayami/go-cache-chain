package main

import (
	"context"
	"fmt"
	"github.com/akayami/go-cache-chain/cache"
	"github.com/akayami/go-cache-chain/helpers"
	"github.com/go-redis/redis/v8"
	"time"
)

func main() {

	someAPI := helpers.NewTestBackend(nil)

	// Bootstrapping MemoryBackend
	MemBackend := cache.NewMemoryBackend(2)
	// Setting up Top layer
	TopLayer := cache.NewLayer(120*time.Second, 30*time.Second, MemBackend, cache.NewMemoryLock())

	var ctx = context.TODO()
	var redisOptions = redis.Options{}
	// Redis Clients
	client := redis.NewClient(&redisOptions)
	// Redis TestBackend
	RedisBackend := cache.NewRedisCacheBackend(client)
	// Redis Layer
	MidLayer := cache.NewLayer(300*time.Second, 60*time.Second, RedisBackend, cache.NewRedisLock(client))
	TopLayer.AppendLayer(MidLayer, 50*time.Millisecond)

	fmt.Println(TopLayer.Insert(ctx, "keycontext", "Some value", someAPI.Create))
	fmt.Println(TopLayer.Get(ctx, "keycontext0", someAPI.Get))
	fmt.Println(TopLayer.Insert(ctx, "keycontext", "Some value", someAPI.Create))
	fmt.Println(TopLayer.Get(ctx, "keycontext1", someAPI.Get))

	fmt.Println(TopLayer.Delete(ctx, "keycontext0", someAPI.Delete))
	fmt.Println(TopLayer.Get(ctx, "keycontext0", someAPI.Get))

}
