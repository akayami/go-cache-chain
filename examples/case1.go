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
	counter := 0

	// Bootstrapping MemoryBackend
	MemBackend := cache.NewMemoryBackend(10)
	// Setting up Top layer
	TopLayer := cache.NewLayer(50*time.Second, 5*time.Second, MemBackend)

	var ctx = context.TODO()
	var redisOptions = redis.Options{}
	// Redis Clients
	client := redis.NewClient(&redisOptions)
	// Redis Backend
	RedisBackend := cache.NewRedisCacheBackend(ctx, client)
	// Redis Layer
	MidLayer := cache.NewLayer(60*time.Second, 30*time.Second, RedisBackend)

	// Creating an API Backend
	ApiBackend := cache.NewAPIBackend(func(key string) (string, bool, error) {
		// This is a stub returning some value. Under normal circumstances, this should wrap some more complex logic fetching data from API, DB or some other store
		counter++
		return strconv.Itoa(counter), false, nil
	})

	// Setting up the layer
	BottomLayer := cache.NewLayer(2*time.Hour, 1*time.Hour, ApiBackend)

	// Connecting all Layers in a chain
	// TopLayer => MidLayer => BottomLayer
	MidLayer.AppendLayer(BottomLayer)
	TopLayer.AppendLayer(MidLayer)

	// Spamming "Key"
	c := time.Tick(1 * time.Second)
	for now := range c {
		val, _, _ := TopLayer.Get("key")
		fmt.Println(now, val)
	}
}
