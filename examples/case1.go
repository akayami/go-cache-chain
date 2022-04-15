package main

import (
	"context"
	"fmt"
	"github.com/akayami/go-cache-chain/cache"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

func main() {
	ctx := context.Background()
	counter := 0

	// Bootstrapping MemoryBackend
	MemBackend := cache.NewMemoryBackend(10)
	// Setting up Top layer
	TopLayer := cache.NewLayer(50*time.Second, 5*time.Second, MemBackend, cache.NewMemoryLock())

	var redisOptions = redis.Options{}
	// Redis Clients
	client := redis.NewClient(&redisOptions)
	// Redis Backend
	RedisBackend := cache.NewRedisCacheBackend(client)
	// Redis Layer
	MidLayer := cache.NewLayer(60*time.Second, 30*time.Second, RedisBackend, cache.NewRedisLock(client))

	// Creating an API Backend
	ApiBackend := cache.NewAPIBackend(func(ctx context.Context, key string) (string, bool, error) {
		// This is a stub returning some value. Under normal circumstances, this should wrap some more complex logic fetching data from API, DB or some other store
		time.Sleep(time.Second)
		counter++
		return strconv.Itoa(counter), false, nil
	})

	// Setting up the layer
	BottomLayer := cache.NewLayer(2*time.Hour, 1*time.Hour, ApiBackend, cache.NewNoLock())

	// Connecting all Layers in a chain
	// TopLayer => MidLayer => BottomLayer
	// Append bottom layer as child, and wait 10 seconds before refresh
	MidLayer.AppendLayer(BottomLayer, 10*time.Second)
	// Append mid layer as child, and wait 1 seconds before refresh
	TopLayer.AppendLayer(MidLayer, 1*time.Second)

	// Spamming "Key"
	c := time.Tick(1 * time.Second)
	for now := range c {
		res := TopLayer.Get(ctx, "key", nil)
		if res.Err != nil {
			fmt.Errorf(res.Err.Error())
		} else {
			fmt.Println(now, res.Value)
		}
	}
}
