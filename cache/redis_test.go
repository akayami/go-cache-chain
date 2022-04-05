package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

func TestRedisCacheBackend(t *testing.T) {
	var ctx = context.TODO()
	var redisOptions = redis.Options{}
	client := redis.NewClient(&redisOptions)
	rcb := NewRedisCacheBackend(client)
	key := "test"

	t.Run("Gets Nil on no key request", func(t *testing.T) {
		if !rcb.Get(ctx, key).isNil() {
			t.Errorf("Expected isNil to be false")
		}
	})

	t.Run("Can return real key", func(t *testing.T) {
		// Create a key Value pair to test against
		res := client.Set(ctx, key, "Value", 30000000)
		if res.Err() != nil {
			t.Error(res.Err())
		}
		get := rcb.Get(ctx, "test")

		if get.isNil() {
			t.Error("Should not be Nil")
		}

		if get.getError() != nil {
			t.Error("Should not have errors")
		}

		if get.getValue() != "Value" {
			t.Error("Should have a Value")
		}

	})

	t.Run("Need to set a key", func(t *testing.T) {
		rcb.Set(ctx, key, "value2", 3*time.Second)
		if rcb.Get(ctx, key).getValue() != "value2" {
			t.Error("Should have a Value equal to value2")
		}
	})

	t.Run("Need to delete a key", func(t *testing.T) {
		rcb.Set(ctx, key, "value2", 3*time.Second)
		if rcb.Get(ctx, key).getValue() != "value2" {
			t.Error("Should have a Value equal to value2")
		}
		if rcb.Del(ctx, key) != nil {
			t.Error("Should have deleted without errors")
		}
		if !rcb.Get(ctx, key).isNil() {
			t.Error("Should be Nil")
		}
	})
}
