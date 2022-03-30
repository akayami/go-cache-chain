package cache

import (
	"testing"
	"time"
)

func TestNoLock(t *testing.T) {
	duration := time.Nanosecond
	t.Run("Acquire Lock", func(t *testing.T) {
		lock := NewNoLock()
		locked, err := lock.Acquire("test", 1*duration)
		if err != nil {
			t.Error(err)
		}
		if !locked {
			t.Errorf("Should lock")
		}
		t.Run("Lock Is Removed", func(t *testing.T) {
			err := lock.Release("test")
			if err != nil {
				t.Error(err)
			}
			locked, err := lock.Acquire("test", 1*duration)
			if err != nil {
				t.Error(err)
			}
			if !locked {
				t.Errorf("Should lock")
			}
		})
	})
}

func CommonLockTests(client Lock, t *testing.T) {

	duration := time.Millisecond

	t.Run("Acquire Lock", func(t *testing.T) {
		lock := client
		locked, err := lock.Acquire("test", 1*duration)
		if err != nil {
			t.Error(err)
		}
		if !locked {
			t.Errorf("Should lock")
		}
		t.Run("Lock Is Exclusive", func(t *testing.T) {
			locked, err := lock.Acquire("test", 1*duration)
			if err != nil {
				t.Error(err)
			}
			if locked {
				t.Errorf("Lock should have failed")
			}
		})
		t.Run("Lock Is Removed", func(t *testing.T) {
			err := lock.Release("test")
			if err != nil {
				t.Error(err)
			}
			locked, err := lock.Acquire("test", 1*duration)
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
		locked, err := lock.Acquire("test2", 1*duration)
		if err != nil {
			t.Error(err)
		}
		if !locked {
			t.Errorf("Should lock")
		}
		// Should fail to acquire due to timeout not passing
		locked2, err2 := lock.Acquire("test2", 1*duration)
		if err2 != nil {
			t.Error(err2)
		}
		if locked2 {
			t.Errorf("Should not lock")
		}
		time.Sleep(2 * duration)
		// Should acquire after 2ms wait-time
		locked3, err3 := lock.Acquire("test2", 1*duration)
		if err2 != nil {
			t.Error(err3)
		}
		if !locked3 {
			t.Errorf("Should lock")
		}
	})

}
