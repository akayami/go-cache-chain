package cache

import (
	"context"
	lru "github.com/hashicorp/golang-lru"
	"testing"
	"time"
)

func TestGolangLRUCacheBackend(t *testing.T) {
	ctx := context.TODO()
	handle, e := lru.New(10)
	if e != nil {
		t.Error(e)
	}
	key := "test"
	cache := NewGolangLRUBackend(handle)

	t.Run("Gets Nil on no key request", func(t *testing.T) {
		if !cache.Get(ctx, key).Nil {
			t.Error("Expected Nil to be false")
		}
	})

	t.Run("Set real key", func(t *testing.T) {
		set, err := cache.Set(ctx, key, "testval", 1*time.Millisecond)
		if err != nil {
			t.Error("Error should not have occurred")
		}
		if set != "testval" {
			t.Error("Invalid return value")
		}
		t.Run("Get key that was set", func(t *testing.T) {
			res := cache.Get(ctx, key)
			if res.Err != nil {
				t.Error(res.Err)
			}
			if res.Nil {
				t.Error("Should not be nil")
			}
			if res.Value != "testval" {
				t.Error("Value has wrong content")
			}
		})
		t.Run("Delete key", func(t *testing.T) {
			e := cache.Del(ctx, key)
			if e != nil {
				t.Error(e)
			}
			t.Run("Verify key deleted", func(t *testing.T) {
				res := cache.Get(ctx, key)
				if res.Err != nil {
					t.Error(res.Err)
				}
				if !res.Nil {
					t.Error("Should be nil")
				}
			})
		})
	})
}
