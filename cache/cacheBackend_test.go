package cache

import (
	"errors"
	"testing"
)

func TestNewCacheBackendResult(t *testing.T) {
	obj := NewCacheBackendResult()
	t.Run("Needs to store and retrieve a Value", func(t *testing.T) {
		obj.Value = "Value"
		if obj.Value != "Value" {
			t.Errorf("Invalid Value")
		}
	})
	t.Run("Needs to store and retrieve Nil state", func(t *testing.T) {
		obj.Nil = true
		if !obj.Nil {
			t.Errorf("Invalid Value. isNil should be true")
		}
	})

	t.Run("Needs to store and retrieve Nil state", func(t *testing.T) {
		obj.Err = errors.New("Error")
		if obj.Err == nil {
			t.Errorf("Invalid Value. isNil should be true")
		}
	})
}
