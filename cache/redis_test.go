package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

var ctx = context.TODO()
var redisOptions = redis.Options{}

func TestRedisCacheBackend(t *testing.T) {
	client := redis.NewClient(&redisOptions)
	rcb := NewRedisCacheBackend(ctx, client)
	key := "test"

	t.Run("Gets nil on no key request", func(t *testing.T) {
		if !rcb.Get(key).isNil() {
			t.Errorf("Expected isNil to be false")
		}
	})

	t.Run("Can return real key", func(t *testing.T) {
		// Create a key value pair to test against
		res := client.Set(ctx, key, "value", 30000000)
		if res.Err() != nil {
			t.Error(res.Err())
		}
		get := rcb.Get("test")

		if get.isNil() {
			t.Error("Should not be nil")
		}

		if get.getError() != nil {
			t.Error("Should not have errors")
		}

		if get.getValue() != "value" {
			t.Error("Should have a value")
		}

	})

	t.Run("Need to set a key", func(t *testing.T) {
		rcb.Set(key, "value2", 3*time.Second)
		if rcb.Get(key).getValue() != "value2" {
			t.Error("Should have a value equal to value2")
		}
	})

	t.Run("Need to delete a key", func(t *testing.T) {
		rcb.Set(key, "value2", 3*time.Second)
		if rcb.Get(key).getValue() != "value2" {
			t.Error("Should have a value equal to value2")
		}
		if rcb.Del(key) != nil {
			t.Error("Should have deleted without errors")
		}
		if !rcb.Get(key).isNil() {
			t.Error("Should be nil")
		}
	})
}
