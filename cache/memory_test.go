package cache

import (
	"strconv"
	"testing"
)

func TestMemoryBacked(t *testing.T) {
	backend := NewMemoryBackend(10)
	t.Run("Empty Tests", func(t *testing.T) {
		res := backend.Get("key")
		if res.isNil() != true {
			t.Errorf("Invalid result")
		}

		if res.getValue() != "" {
			t.Errorf("Invalid result. String must be empty")
		}

		if res.getError() != nil {
			t.Errorf("Error has to be Nil")
		}
	})
	t.Run("Test with a Value", func(t *testing.T) {
		backend.Set("test", "testvalue", 0)
		res := backend.Get("test")
		if res.isNil() != false {
			t.Errorf("Invalid result")
		}

		if res.getValue() != "testvalue" {
			t.Errorf("Invalid result. String must be empty")
		}

		if res.getError() != nil {
			t.Errorf("Error has to be Nil")
		}
	})
	t.Run("Overwrite a Value", func(t *testing.T) {
		backend.Set("test", "testvalue2", 0)
		res := backend.Get("test")
		if res.isNil() != false {
			t.Errorf("Invalid result")
		}

		if res.getValue() != "testvalue2" {
			t.Errorf("Invalid result. String must be empty")
		}

		if res.getError() != nil {
			t.Errorf("Error has to be Nil")
		}
		if backend.GetSize() != 1 {
			t.Errorf("Wrong size %d", backend.GetSize())
		}
	})

	t.Run("Delete Value", func(t *testing.T) {
		backend.Del("test")
		if l := len(backend.list); l > 0 {
			t.Errorf("List should be empty - Actual length %d", l)
		}
		res := backend.Get("test")
		if res.isNil() != true {
			t.Errorf("Invalid result")
		}

		if res.getValue() != "" {
			t.Errorf("Invalid result. String must be empty")
		}

		if res.getError() != nil {
			t.Errorf("Error has to be Nil")
		}

	})

	t.Run("Exceed size limit", func(t *testing.T) {
		for i := 0; i < 15; i++ {
			backend.Set(strconv.Itoa(i), "Value"+strconv.Itoa(i), 0)
		}
		if len(backend.list) != 10 {
			t.Errorf("Wrong size")
		}
		if len(backend.store) != 10 {
			t.Errorf("Wrong size")
		}
		t.Run("Test element sorting", func(t *testing.T) {
			if backend.list[0] != "5" {
				t.Errorf("Wrong size")
			}
			if backend.list[9] != "14" {
				t.Errorf("Wrong size")
			}
			backend.Get("5")
			if backend.list[9] != "5" {
				t.Errorf("Wrong size")
			}
			if backend.list[8] != "14" {
				t.Errorf("Wrong size")
			}
			if backend.list[0] != "6" {
				t.Errorf("Wrong size")
			}
			backend.Get("6")
			if backend.list[9] != "6" {
				t.Errorf("Wrong size")
			}
			if backend.list[8] != "5" {
				t.Errorf("Wrong size")
			}
			if backend.list[0] != "7" {
				t.Errorf("Wrong size")
			}
		})
	})
}
