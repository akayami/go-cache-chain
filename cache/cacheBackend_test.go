package cache

import (
	"errors"
	"testing"
)

func TestNewCacheBackendResult(t *testing.T) {
	obj := NewCacheBackendResult()
	t.Run("Needs to store and retrieve a Value", func(t *testing.T) {
		obj.setValue("Value")
		if obj.getValue() != "Value" {
			t.Errorf("Invalid Value")
		}
	})
	t.Run("Needs to store and retrieve Nil state", func(t *testing.T) {
		obj.setNil(true)
		if !obj.isNil() {
			t.Errorf("Invalid Value. isNil should be true")
		}
	})

	t.Run("Needs to store and retrieve Nil state", func(t *testing.T) {
		obj.setError(errors.New("Error"))
		if obj.getError() == nil {
			t.Errorf("Invalid Value. isNil should be true")
		}
	})
}
