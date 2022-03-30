package main

import (
	"fmt"
	"github.com/akayami/go-cache-chain/cache"
	"strconv"
	"time"
)

/**
Basic test setup. Memory -> Backed
*/

func main() {
	counter := 0

	// Bootstrapping MemoryBackend
	MemBackend := cache.NewMemoryBackend(2)
	// Setting up Top layer
	TopLayer := cache.NewLayer(50*time.Second, 5*time.Second, MemBackend, cache.NewMemoryLock())

	// Creating an API Backend
	ApiBackend := cache.NewAPIBackend(func(key string) (string, bool, error) {
		fmt.Println("Backend Got Called")
		// This is a stub returning some value. Under normal circumstances, this should wrap some more complex logic fetching data from API, DB or some other store
		counter++
		return strconv.Itoa(counter), false, nil
	})

	// Setting up the layer
	BottomLayer := cache.NewLayer(2*time.Hour, 1*time.Hour, ApiBackend, cache.NewNoLock())

	// Append bottom layer as child, and wait 10 seconds before refresh
	TopLayer.AppendLayer(BottomLayer, 10*time.Second)

	// Spamming "Key"
	c := time.Tick(1 * time.Second)
	for now := range c {
		val, _, _ := TopLayer.Get("key")
		fmt.Println(now, val)
	}
}
