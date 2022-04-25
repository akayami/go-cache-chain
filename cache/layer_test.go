package cache

import (
	"context"
	"errors"
	"github.com/akayami/go-cache-chain/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strconv"
	"testing"
	"time"
)

//func NewTestBackend() *TestBackend {
//	return &TestBackend{Backend{name: "Test", marshal: true}, nil}
//}
//
//type TestBackend struct {
//	Backend
//	get_handler Getter
//}
//
//func (t TestBackend) Get(ctx context.Context, key string) *CacheBackendResult {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (t TestBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (t TestBackend) Del(ctx context.Context, key string) error {
//	//TODO implement me
//	panic("implement me")
//}

type InvalidValue struct {
	mock.Mock
}

func (m *InvalidValue) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (m *InvalidValue) Del(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}

func (m *InvalidValue) GetName() string {
	//TODO implement me
	panic("implement me")
}

func (m *InvalidValue) IsMarshaled() bool {
	//TODO implement me
	panic("implement me")
}

func (m *InvalidValue) Get(ctx context.Context, key string) *CacheBackendResult {
	//args := m.Called(ctx, key)
	return &CacheBackendResult{Value: "invalid", Err: nil, Nil: false, needsMarshalling: true}

}

func TestLayer(t *testing.T) {
	ctx := context.Background()
	t.Run("Test CacheError Object", func(t *testing.T) {
		e := CacheError{Message: "Test"}
		assert.Equal(t, "Cache Chain Error occured with following message: Test", e.Error())
	})

	t.Run("Test Error Handling for invalid payload while unmarshalling", func(t *testing.T) {
		backend := new(InvalidValue)
		l := NewLayer(2*time.Millisecond, 1*time.Millisecond, backend, nil)
		result := l.getFromBackend(ctx, "key")
		assert.NotNil(t, result.Err)
	})

	t.Run("Test Error Invalid Payload", func(t *testing.T) {
		backend := new(InvalidValue)
		l := NewLayer(2*time.Millisecond, 1*time.Millisecond, backend, nil)
		result := l.getFromBackend(ctx, "key")
		assert.NotNil(t, result.Err)
	})

	t.Run("Test error handling from fallback", func(t *testing.T) {
		l := NewLayer(2*time.Millisecond, 1*time.Millisecond, NewMemoryBackend(10), NewNoLock())
		result := l.Get(ctx, "key", func(ctx context.Context, s string) (string, bool, error) {
			return "", false, errors.New("Fake error")
		})
		assert.NotNil(t, result.Err)
	})

	t.Run("Basic Get Logic", func(t *testing.T) {
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
		backend := NewAPIBackend(getter)
		layer := NewLayer(10*time.Millisecond, 5*time.Millisecond, backend, NewNoLock())

		t.Run("Payload Marshaling", func(t *testing.T) {
			// Payload{Payload: "test", Stale: 5 * int64(time.Second)}
			p := payload{Payload: "Test", Stale: time.Now().Add(5)}
			val, err := layer.marshal(p)
			if err != nil {
				t.Error(err)
			}
			//sval := string(val)
			//if sval != "{\"Payload\":\"Test\",\"Stale\":10}" {
			//	t.Error("Wrong Value")
			//}
			t.Run("Payload Unmarshalling", func(t *testing.T) {
				_, e := layer.unmarshal(val)
				if e != nil {
					t.Error(e)
				}
			})
		})

		t.Run("Single Layer", func(t *testing.T) {
			t.Run("Noval", func(t *testing.T) {

				res := layer.Get(ctx, "noval", nil)
				assert.Nil(t, res.Err)
				assert.True(t, res.Nil)
				assert.Equal(t, "", res.Value)
			})
			t.Run("Error", func(t *testing.T) {
				res := layer.Get(ctx, "error", nil)
				assert.NotNil(t, res.Err)
			})

			t.Run("Get Val", func(t *testing.T) {
				res := layer.Get(ctx, "key", nil)
				assert.Nil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, topvalue, res.Value)
			})
		})

		t.Run("with childLayer backed", func(t *testing.T) {
			mem := NewMemoryBackend(10)
			layer2 := NewLayer(2*time.Millisecond, 1*time.Millisecond, mem, NewNoLock())
			layer2.AppendLayer(layer, 0)
			t.Run("Noval", func(t *testing.T) {

				t.Run("Noval", func(t *testing.T) {

					res := layer2.Get(ctx, "noval", nil)
					if (res.Value == "" && res.Nil == true && res.Err == nil) == false {
						t.Errorf("Invalid response")
					}
				})
				t.Run("Error", func(t *testing.T) {
					res := layer2.Get(ctx, "error", nil)
					if (res.Value == "" && res.Nil == false && res.Err != nil) == false {
						t.Errorf("Invalid response")
					}
				})

				t.Run("Get Val", func(t *testing.T) {
					res := layer2.Get(ctx, "key", nil)
					if (res.Value == topvalue && res.Nil == false && res.Err == nil) == false {
						t.Errorf("Invalid response")
					}
				})

			})
			t.Run("Error", func(t *testing.T) {
				res := layer2.Get(ctx, "error", nil)
				if (res.Value == "" && res.Nil == false && res.Err != nil) == false {
					t.Errorf("Invalid response")
				}
			})

			t.Run("Get Val", func(t *testing.T) {

				res := layer2.Get(ctx, "key", nil)
				if (res.Value == topvalue && res.Nil == false && res.Err == nil) == false {
					t.Errorf("Invalid response")
				}

				time.Sleep(1 * time.Millisecond)
				t.Run("Get cached Value", func(t *testing.T) {
					res := layer2.Get(ctx, "key", nil)
					if (res.Value == topvalue && res.Nil == false && res.Err == nil) == false {
						t.Errorf("Invalid response")
					}
				})
			})
		})
	})

	t.Run("Test Refreshing Stale", func(t *testing.T) {

		timeUnit := time.Millisecond
		mem := NewMemoryBackend(10)
		toplayer := NewLayer(100*timeUnit, 50*timeUnit, mem, NewNoLock())

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
		topvalue := "TopValue"
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
			return topvalue, false, nil
		}
		backend := NewAPIBackend(getter)
		bottomLayer := NewLayer(200*timeUnit, 150*timeUnit, backend, NewNoLock())
		toplayer.AppendLayer(bottomLayer, 0)
		t.Run("Triggering lookup in lower level", func(t *testing.T) {
			res := toplayer.Get(ctx, "inc", nil)
			if res.Err != nil {
				t.Error(res.Err)
			}
			if res.Nil != false {
				t.Errorf("Should be no noval")
			}
			if res.Value != "1" {
				t.Errorf("Should be top val")
			}
			time.Sleep(10 * timeUnit) // Wait to let the lookup update cache
			t.Run("This should not go to lower level", func(t *testing.T) {
				res := toplayer.Get(ctx, "inc", nil)
				if res.Err != nil {
					t.Error(res.Err)
				}
				if res.Nil != false {
					t.Errorf("Should be no noval")
				}
				if res.Value != "1" {
					t.Errorf("Should be top val")
				}
			})
			time.Sleep(55 * timeUnit) // Wait 35 to exceed the stale Value
			t.Run("Should get the stale Value and trigger refresh", func(t *testing.T) {
				res := toplayer.Get(ctx, "inc", nil)
				if res.Err != nil {
					t.Error(res.Err)
				}
				if res.Nil != false {
					t.Errorf("Should be no noval")
				}
				if res.Value != "1" {
					t.Errorf("Should be top val")
				}
				time.Sleep(5 * timeUnit) // Wait one ms to let the lookup update cache
				t.Run("Should grab the new Value from cache and not trigger a refresh", func(t *testing.T) {
					res := toplayer.Get(ctx, "inc", nil)
					if res.Err != nil {
						t.Error(res.Err)
					}
					if res.Nil != false {
						t.Errorf("Should be no noval")
					}
					if res.Value != "2" {
						t.Errorf("Should be 2. Backend must have refreshed the cache.")
					}
					time.Sleep(5 * timeUnit)
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

		backend := NewAPIBackend(getter)
		bottomLayer := NewLayer(200*timeUnit, 150*timeUnit, backend, NewNoLock())
		toplayer.AppendLayer(bottomLayer, 2*time.Second)

		// Under normal circumstances we should be able to get the key
		t.Run("Test key locking mechanism and returning cache error", func(t *testing.T) {
			done := make(chan bool)
			go func(done chan bool) {
				res := toplayer.Get(ctx, "key1", nil)
				lock := memLock.store["key1"]
				assert.NotNil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, "", res.Value)
				assert.IsType(t, lock, &time.Timer{})
				assert.IsType(t, res.Err, CacheError{})
				done <- true
			}(done)
			t.Run("This should get the key normally", func(t *testing.T) {
				res := toplayer.Get(ctx, "key1", nil)
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

	t.Run("Basic create logic", func(t *testing.T) {
		val := "value"
		counter := 1
		fakeBackend := helpers.NewTestBackend(func() {
			counter++
		})
		mem := NewMemoryBackend(10)
		memLock := NewMemoryLock()
		layer := NewLayer(100*time.Millisecond, 20*time.Millisecond, mem, memLock)
		id, val, err := layer.Insert(ctx, "somekey:", val, fakeBackend.Create)
		lock := memLock.store["key1"]
		assert.IsType(t, lock, &time.Timer{})
		assert.Equal(t, "somekey:0", id)
		assert.Equal(t, "value", val)
		assert.Nil(t, err)
		t.Run("Get created value by key", func(t *testing.T) {
			assert.Equal(t, 1, counter)
			res := layer.Get(ctx, "somekey:0", fakeBackend.Get)
			assert.Equal(t, 2, counter)
			assert.Nil(t, res.Err)
			assert.False(t, res.Nil)
			assert.Equal(t, val, res.Value)
			// Sleep for 3 ms for key to expire
			time.Sleep(30 * time.Millisecond)
			t.Run("Get created value by key", func(t *testing.T) {
				res := layer.Get(ctx, "somekey:0", fakeBackend.Get)
				assert.Nil(t, res.Err)
				assert.False(t, res.Nil)
				assert.Equal(t, val, res.Value)
				// Sleep for 1ms to let the backend be called and counter inc
				time.Sleep(time.Millisecond)
				assert.Equal(t, 3, counter)
			})
		})
	})

	t.Run("Test Set Logic", func(t *testing.T) {
		t.Run("Basic Set", func(t *testing.T) {
			mem := NewMemoryBackend(10)
			memLock := NewMemoryLock()
			layer := NewLayer(100*time.Millisecond, 20*time.Millisecond, mem, memLock)
			setter := Setter(func(ctx context.Context, key string, val string) (string, error) {
				assert.Equal(t, "test", key)
				assert.Equal(t, "value", val)
				return val, nil
			})
			val, e := layer.Set(ctx, "test", "value", setter)
			assert.Nil(t, e)
			assert.Equal(t, "value", val)
			time.Sleep(5 * time.Millisecond)
			res := layer.Get(ctx, "test", nil)
			assert.Equal(t, "value", res.Value)
		})

		/**
		This test tests the scenario when client set a key, (which goes all the way to the backend) which is then returned
		1. Set a k/v pair test/value
		2. Immediately get the test value
		3. Check if deep set lock is acquired.
		4. Sleep 5ms to make sure the deep refresh is over and the lock is removed
		5. Fetch the value again and see that a newval value is stored instead of orignal value
		*/
		t.Run("Backend changes value, needs to overwrite local cache", func(t *testing.T) {
			mem := NewMemoryBackend(10)
			memLock := NewMemoryLock()
			layer := NewLayer(100*time.Millisecond, 20*time.Millisecond, mem, memLock)
			setter := Setter(func(ctx context.Context, key string, val string) (string, error) {
				assert.Equal(t, "test", key)
				assert.Equal(t, "value", val)
				return "newval", nil
			})
			val, e := layer.Set(ctx, "test", "value", setter)
			assert.Equal(t, "value", val)
			res1 := layer.Get(ctx, "test", nil)
			lockState, ok := memLock.store["set_test"]
			assert.True(t, ok)
			assert.NotNil(t, lockState)
			// at first the value is value as the provided value is set as real value
			assert.Equal(t, "value", res1.Value)
			assert.Nil(t, e)
			time.Sleep(5 * time.Millisecond)
			// Checking if lock was removed
			lockState, ok = memLock.store["set_test"]
			assert.False(t, ok)
			assert.Nil(t, lockState)
			// Fetching the value egain
			res := layer.Get(ctx, "test", nil)
			// After backend is called and new value is retrieved local cache is changed.
			assert.Equal(t, "newval", res.Value)
		})
	})
}
