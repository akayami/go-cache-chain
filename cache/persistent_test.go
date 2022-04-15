package cache

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func fakeCall() {
	time.Sleep(1000 * time.Millisecond)
}

func TestPresistent(t *testing.T) {
	var ctx = context.TODO()
	val := "{\"Id\":\"54\",\"FirstName\":\"John\",\"LastName\":\"Doe\",\"Stamp\":\"1\"}"
	t.Run("Check standard persistent request", func(t *testing.T) {
		MemBackend := NewMemoryBackend(2)
		TopLayer := NewLayer(2*time.Second, 1*time.Second, MemBackend, NewMemoryLock())

		getter := func(ctx context.Context, key string) (string, bool, error) {
			fmt.Println("Backend Called")

			go fakeCall()
			return val, false, nil
		}
		//fallback := Fallback{Getter: getter, TTL: time.Minute}
		ApiBackend := NewAPIBackend(getter)

		BottomLayer := NewLayer(2*time.Hour, 1*time.Hour, ApiBackend, NewNoLock())
		//MidLayer.AppendLayer(BottomLayer, 3*time.Second)
		TopLayer.AppendLayer(BottomLayer, 1000*time.Millisecond)

		ch := make(chan Result)
		go PersistentGet(ctx, TopLayer, "key", getter, 5*time.Second, 10*time.Millisecond, ch)
		result := <-ch
		close(ch)
		if result.Noval {
			t.Error("No val")
		}
		if result.Error != nil {
			t.Error(result.Error)
		}
		if result.Value != val {
			t.Error("Invalid value")
		}
	})

	t.Run("Check timeout on persistent request", func(t *testing.T) {

		MemBackend := NewMemoryBackend(2)
		TopLayer := NewLayer(2*time.Second, 1*time.Second, MemBackend, NewMemoryLock())

		getter := func(ctx context.Context, key string) (string, bool, error) {
			fmt.Println("Backend Called")

			go fakeCall()
			return val, false, nil
		}

		//fallback := Fallback{Getter: getter, TTL: time.Minute}
		ApiBackend := NewAPIBackend(getter)

		BottomLayer := NewLayer(2*time.Hour, 1*time.Hour, ApiBackend, NewNoLock())
		//MidLayer.AppendLayer(BottomLayer, 3*time.Second)
		TopLayer.AppendLayer(BottomLayer, 1000*time.Millisecond)

		ch := make(chan Result)
		go PersistentGet(ctx, TopLayer, "key", getter, 10*time.Millisecond, 1*time.Millisecond, ch)
		result := <-ch
		close(ch)
		if result.Noval {
			t.Error("No val")
		}
		if result.Error != nil {
			t.Error(result.Error)
		}
		if result.Value != val {
			t.Error("Invalid value")
		}
	})
}
