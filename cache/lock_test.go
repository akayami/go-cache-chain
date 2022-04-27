package cache

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNoLock(t *testing.T) {
	ctx := context.Background()
	duration := time.Nanosecond
	t.Run("Acquire Lock", func(t *testing.T) {
		lock := NewNoLock()
		locked, err := lock.Acquire(ctx, "test", 1*duration)
		if err != nil {
			t.Error(err)
		}
		if !locked {
			t.Errorf("Should lock")
		}
		t.Run("Lock Is Removed", func(t *testing.T) {
			err := lock.Release(ctx, "test")
			if err != nil {
				t.Error(err)
			}
			locked, err := lock.Acquire(ctx, "test", 1*duration)
			if err != nil {
				t.Error(err)
			}
			if !locked {
				t.Errorf("Should lock")
			}
		})
	})
}

func CommonLockTests(ctx context.Context, client Lock, t *testing.T) {
	duration := time.Millisecond

	t.Run("Acquire Lock", func(t *testing.T) {
		lock := client
		locked, err := lock.Acquire(ctx, "test", 1*duration)
		if err != nil {
			t.Error(err)
		}
		if !locked {
			t.Errorf("Should lock")
		}
		t.Run("Lock Is Exclusive", func(t *testing.T) {
			locked, err := lock.Acquire(ctx, "test", 1*duration)
			if err != nil {
				t.Error(err)
			}
			if locked {
				t.Errorf("Lock should have failed")
			}
		})
		t.Run("Lock Is Removed", func(t *testing.T) {
			err := lock.Release(ctx, "test")
			if err != nil {
				t.Error(err)
			}
			locked, err := lock.Acquire(ctx, "test", 1*duration)
			if err != nil {
				t.Error(err)
			}
			if !locked {
				t.Errorf("Should lock")
			}
		})
	})

	t.Run("Lock Timeout", func(t *testing.T) {
		lock := client
		// Should acquire a lock
		locked, err := lock.Acquire(ctx, "test2", 1*duration)
		assert.Nil(t, err)
		assert.True(t, locked)
		// Should fail to acquire due to timeout not passing
		locked2, err2 := lock.Acquire(ctx, "test2", 1*duration)
		assert.Nil(t, err2)
		assert.False(t, locked2)
		time.Sleep(3 * duration)
		// Should acquire after 2ms wait-time
		locked3, err3 := lock.Acquire(ctx, "test2", 1*duration)
		fmt.Println(err3)
		assert.Nil(t, err3)
		assert.True(t, locked3)
	})

}
