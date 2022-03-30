package cache

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestLayer(t *testing.T) {

	t.Run("Basic Logic", func(t *testing.T) {

		topvalue := "TopValue"
		getter := func(k string) (string, bool, error) {
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
			//	t.Error("Wrong value")
			//}
			t.Run("Payload Unmarshalling", func(t *testing.T) {
				_, e := layer.unmarchal(val)
				if e != nil {
					t.Error(e)
				}
			})
		})

		t.Run("without child backend", func(t *testing.T) {
			t.Run("Noval", func(t *testing.T) {
				val, noval, err := layer.Get("noval")
				if (val == "" && noval == true && err == nil) == false {
					t.Errorf("Invalid response")
				}
			})
			t.Run("Error", func(t *testing.T) {
				val, noval, err := layer.Get("error")
				if (val == "" && noval == false && err != nil) == false {
					t.Errorf("Invalid response")
				}
			})

			t.Run("Get Val", func(t *testing.T) {
				val, noval, err := layer.Get("key")
				if (val == topvalue && noval == false && err == nil) == false {
					t.Errorf("Invalid response")
				}
			})
		})

		t.Run("with child backed", func(t *testing.T) {
			mem := NewMemoryBackend(10)
			layer2 := NewLayer(2*time.Millisecond, 1*time.Millisecond, mem, NewNoLock())
			layer2.AppendLayer(layer, 0)
			t.Run("Noval", func(t *testing.T) {
				val, noval, err := layer2.Get("noval")
				if (val == "" && noval == true && err == nil) == false {
					t.Errorf("Invalid response")
				}
			})
			t.Run("Error", func(t *testing.T) {
				val, noval, err := layer2.Get("error")
				if (val == "" && noval == false && err != nil) == false {
					t.Errorf("Invalid response")
				}
			})

			t.Run("Get Val", func(t *testing.T) {
				val, noval, err := layer2.Get("key")
				if (val == topvalue && noval == false && err == nil) == false {
					t.Errorf("Invalid response")
				}
				time.Sleep(1 * time.Millisecond)
				t.Run("Get cached value", func(t *testing.T) {
					val, noval, err := layer2.Get("key")
					if (val == topvalue && noval == false && err == nil) == false {
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
			val, noval, err := toplayer.Get("key")
			if err != nil {
				t.Error(err)
			}
			if val != "" {
				t.Errorf("Value should be empty string")
			}
			if !noval {
				t.Errorf("Should be a noval")
			}
		})
		counter := 0
		topvalue := "TopValue"
		getter := func(k string) (string, bool, error) {
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
			val, noval, err := toplayer.Get("inc")
			if err != nil {
				t.Error(err)
			}
			if noval != false {
				t.Errorf("Should be no val")
			}
			if val != "1" {
				t.Errorf("Should be top val")
			}
			time.Sleep(10 * timeUnit) // Wait to let the lookup update cache
			t.Run("This should not go to lower level", func(t *testing.T) {
				val, noval, err := toplayer.Get("inc")
				if err != nil {
					t.Error(err)
				}
				if noval != false {
					t.Errorf("Should be no val")
				}
				if val != "1" {
					t.Errorf("Should be top val")
				}
			})
			time.Sleep(55 * timeUnit) // Wait 35 to exceed the stale value
			t.Run("Should get the stale value and trigger refresh", func(t *testing.T) {
				val, noval, err := toplayer.Get("inc")
				if err != nil {
					t.Error(err)
				}
				if noval != false {
					t.Errorf("Should be no val")
				}
				if val != "1" {
					t.Errorf("Should be 1, got %s", val)
				}
				time.Sleep(5 * timeUnit) // Wait one ms to let the lookup update cache
				t.Run("Should grab the new value from cache and not trigger a refresh", func(t *testing.T) {
					val, noval, err := toplayer.Get("inc")
					if err != nil {
						t.Error(err)
					}
					if noval != false {
						t.Errorf("Should be no val")
					}
					if val != "2" {
						t.Errorf("Should be 2. Backend must have refreshed the cache.")
					}
					time.Sleep(5 * timeUnit)
				})
			})
		})
	})

	t.Run("Test first fetch race condition prevention", func(t *testing.T) {
		timeUnit := time.Second
		mem := NewMemoryBackend(10)
		toplayer := NewLayer(100*timeUnit, 50*timeUnit, mem, NewMemoryLock())

		var getter = func(key string) (string, bool, error) {
			time.Sleep(500 * time.Millisecond)
			return "val", false, nil
		}

		backend := NewAPIBackend(getter)
		bottomLayer := NewLayer(200*timeUnit, 150*timeUnit, backend, NewNoLock())
		toplayer.AppendLayer(bottomLayer, time.Second)

		t.Run("Should get key", func(t *testing.T) {
			val, noval, err := toplayer.Get("key1")
			if err != nil {
				t.Error(err)
			}
			if val != "val" {
				t.Errorf("Value should equal value")
			}
			if noval {
				t.Errorf("Should be a noval")
			}
			t.Run("Should Fail to get key as backend is slow", func(t *testing.T) {
				_, _, err := toplayer.Get("key1")
				if _, ok := err.(CacheError); !ok {
					t.Errorf("Expected Cache Error %s", err.(CacheError))
				}
			})
		})
	})
}
