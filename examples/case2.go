package main

import (
	"context"
	"fmt"
	"github.com/akayami/go-cache-chain/cache"
	"github.com/akayami/go-cache-chain/examples/jsons"
	"github.com/go-redis/redis/v8"
	"github.com/mailru/easyjson"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"
)

func main() {
	//counter := 0

	// Bootstrapping MemoryBackend
	MemBackend := cache.NewMemoryBackend(2)
	// Setting up Top layer
	TopLayer := cache.NewLayer(120*time.Second, 30*time.Second, MemBackend, cache.NewMemoryLock())

	var ctx = context.TODO()
	var redisOptions = redis.Options{}
	// Redis Clients
	client := redis.NewClient(&redisOptions)
	// Redis Backend
	RedisBackend := cache.NewRedisCacheBackend(client)
	// Redis Layer
	MidLayer := cache.NewLayer(300*time.Second, 60*time.Second, RedisBackend, cache.NewRedisLock(client))

	// Creating an API Backend
	//ApiBackend := cache.NewAPIBackend(func(key string) (string, bool, error) {
	//	fmt.Println("Backend Got Called")
	//	// This is a stub returning some value. Under normal circumstances, this should wrap some more complex logic fetching data from API, DB or some other store
	//	counter++
	//	return strconv.Itoa(counter), false, nil
	//})

	ApiBackend := cache.NewAPIBackend(func(ctx context.Context, key string) (string, bool, error) {
		fmt.Println("Backend Got Called")
		client := http.Client{
			Timeout: 11 * time.Second,
		}
		resp, err := client.Get("http://localhost:3000/user")
		if err != nil {
			return "", false, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", false, err
		}
		fmt.Println(string(body))
		return string(body), false, nil
	})

	// Setting up the layer
	BottomLayer := cache.NewLayer(2*time.Hour, 1*time.Hour, ApiBackend, cache.NewNoLock())

	// Connecting all Layers in a chain
	// TopLayer => MidLayer => BottomLayer
	MidLayer.AppendLayer(BottomLayer, 3*time.Second)
	TopLayer.AppendLayer(MidLayer, 50*time.Millisecond)

	//TopLayer.AppendLayer(BottomLayer)

	// Spamming "Key"

	c := time.Tick(100 * time.Millisecond)
	for now := range c {
		fmt.Println("Go routines", runtime.NumGoroutine())
		go Cycle(ctx, TopLayer, now)
	}
}

func Cycle(ctx context.Context, l *cache.Layer, now time.Time) {
	getResult := make(chan cache.Result)
	go cache.PersistentGet(ctx, l, "key", 1000*time.Millisecond, 100*time.Millisecond, getResult)
	result := <-getResult

	val, noval, err := result.Value, result.Noval, result.Error
	close(getResult)

	if noval {
		fmt.Println("No value found")
	} else if err != nil {
		fmt.Println(err)
	} else {
		userSchema := &jsons.User{}
		easyjson.Unmarshal([]byte(val), userSchema)
		fmt.Println(now, userSchema)
	}
}
