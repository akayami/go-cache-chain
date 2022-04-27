package cache

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestLayerFallback(t *testing.T) {
	ctx := context.Background()
	t.Run("Basic Logic", func(t *testing.T) {
		topvalue := "TopValue"
		getter := func(ctx context.Context, k string) (string, bool, error) {
			if k == "error" {
				return "", false, errors.New("Error")
			}
			if k == "noval" {
				return "", true, nil
			}
			return topvalue, false, nil
		}
		backend := NewAPIBackend(getter, nil, nil, nil)
		layer := NewLayer(10*time.Millisecond, 5*time.Millisecond, backend, NewNoLock())

		t.Run("with childLayer backend", func(t *testing.T) {
			mem := NewMemoryBackend(10)
			layer2 := NewLayer(20*time.Millisecond, 10*time.Millisecond, mem, NewNoLock())
			layer2.AppendLayer(layer, 0)
			//t.Run("Noval", func(t *testing.T) {

			t.Run("Noval", func(t *testing.T) {

				res := layer2.Get(ctx, "noval", getter)
				if (res.Value == "" && res.Nil == true && res.Err == nil) == false {
					t.Errorf("Invalid response")
				}
			})
			t.Run("Error", func(t *testing.T) {
				res := layer2.Get(ctx, "error", getter)
				if (res.Value == "" && res.Nil == false && res.Err != nil) == false {
					t.Errorf("Invalid response")
				}
			})

			t.Run("Get Val", func(t *testing.T) {
				res := layer2.Get(ctx, "key", getter)
				assert.Nil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, topvalue, res.Value)
			})

			//})
			t.Run("Error", func(t *testing.T) {
				res := layer2.Get(ctx, "error", getter)
				if (res.Value == "" && res.Nil == false && res.Err != nil) == false {
					t.Errorf("Invalid response")
				}
			})

			t.Run("Get Val", func(t *testing.T) {

				res := layer2.Get(ctx, "key", getter)
				assert.Nil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, topvalue, res.Value)

				time.Sleep(1 * time.Millisecond)
				t.Run("Get cached Value", func(t *testing.T) {
					res := layer2.Get(ctx, "key", getter)
					assert.Nil(t, res.Err)
					assert.False(t, res.Nil)
					assert.Equal(t, topvalue, res.Value)
				})
			})
		})
	})

	t.Run("Test Refreshing Stale", func(t *testing.T) {

		timeUnit := time.Millisecond
		mem := NewMemoryBackend(10)
		toplayer := NewLayer(100*timeUnit, 10*timeUnit, mem, NewNoLock())

		t.Run("Simple noval test on top layer", func(t *testing.T) {
			res := toplayer.Get(ctx, "key", nil)
			if res.Err != nil {
				t.Error(res.Err)
			}
			if res.Value != "" {
				t.Errorf("Value should be empty string")
			}
			if !res.Nil {
				t.Errorf("Should be a noval")
			}
		})
		counter := 0
		//topvalue := "TopValue"
		getter := func(ctx context.Context, k string) (string, bool, error) {
			if k == "error" {
				return "", false, errors.New("Error")
			}
			if k == "noval" {
				return "", true, nil
			}
			if k == "inc" {
				//log.Printf("Calling inc %d", counter)
				counter++
				return strconv.Itoa(counter), false, nil
			}
			panic("Unknown Key requested")
			//return topvalue, false, nil
		}
		//backend := NewAPIBackend(getter)
		//bottomLayer := NewLayer(200*timeUnit, 150*timeUnit, backend, NewNoLock())
		//toplayer.AppendLayer(bottomLayer, 0)
		t.Run("Triggering lookup in lower level", func(t *testing.T) {
			assert.Equal(t, counter, 0)
			res := toplayer.Get(ctx, "inc", getter)
			assert.Equal(t, counter, 1)
			assert.Nil(t, res.Err)
			assert.False(t, res.Nil)
			assert.Equal(t, res.Value, "1")
			time.Sleep(1 * timeUnit) // Wait to let the lookup update cache
			assert.NotNil(t, mem.store["inc"])
			t.Run("This should not go to lower level, should get it from local store", func(t *testing.T) {
				assert.Equal(t, counter, 1)
				res := toplayer.Get(ctx, "inc", getter)
				assert.Equal(t, counter, 1)
				assert.Nil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, "1", res.Value)
			})
			time.Sleep(11 * timeUnit) // Wait 35 to exceed the stale Value
			assert.NotNil(t, mem.store["inc"])
			t.Run("Should get the stale Value and trigger refresh", func(t *testing.T) {
				assert.Equal(t, 1, counter)
				res := toplayer.Get(ctx, "inc", getter)
				assert.Nil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, "1", res.Value)
				time.Sleep(1 * timeUnit) // Wait one ms to let the lookup update cache
				assert.Equal(t, 2, counter)
				t.Run("Should grab the new Value from cache and not trigger a refresh", func(t *testing.T) {
					assert.Equal(t, 2, counter)
					res := toplayer.Get(ctx, "inc", getter)
					assert.Nil(t, res.Err)
					assert.False(t, res.Nil)
					assert.Equal(t, "2", res.Value)
					time.Sleep(1 * timeUnit)
					assert.Equal(t, 2, counter)
				})
			})
		})
	})

	t.Run("Test first fetch race condition prevention (Locking)", func(t *testing.T) {
		timeUnit := time.Second
		mem := NewMemoryBackend(10)
		memLock := NewMemoryLock()
		toplayer := NewLayer(100*timeUnit, 50*timeUnit, mem, memLock)

		var getter = func(ctx context.Context, key string) (string, bool, error) {
			time.Sleep(1000 * time.Millisecond)
			return "val", false, nil
		}

		backend := NewAPIBackend(getter, nil, nil, nil)
		bottomLayer := NewLayer(200*timeUnit, 150*timeUnit, backend, NewNoLock())
		toplayer.AppendLayer(bottomLayer, 2*time.Second)

		// Under normal circumstances we should be able to get the key
		t.Run("Test key locking mechanism and returning cache error", func(t *testing.T) {
			done := make(chan bool)
			go func(done chan bool) {
				res := toplayer.Get(ctx, "key1", getter)
				lock := memLock.store["key1"]
				assert.NotNil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, "", res.Value)
				assert.IsType(t, lock, &time.Timer{})
				assert.IsType(t, res.Err, CacheError{})
				done <- true
			}(done)
			t.Run("This should get the key normally", func(t *testing.T) {
				res := toplayer.Get(ctx, "key1", getter)
				lock := memLock.store["key1"]
				assert.Nil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, "val", res.Value)
				assert.IsType(t, lock, &time.Timer{})
			})
			<-done
			close(done)
		})
	})
}
