package cache

import (
	"errors"
	"testing"
)

func TestNewCacheBackendResult(t *testing.T) {
	obj := NewCacheBackendResult()
	t.Run("Needs to store and retrieve a value", func(t *testing.T) {
		obj.setValue("value")
		if obj.getValue() != "value" {
			t.Errorf("Invalid value")
		}
	})
	t.Run("Needs to store and retrieve nil state", func(t *testing.T) {
		obj.setNil(true)
		if !obj.isNil() {
			t.Errorf("Invalid value. isNil should be true")
		}
	})

	t.Run("Needs to store and retrieve nil state", func(t *testing.T) {
		obj.setError(errors.New("Error"))
		if obj.getError() == nil {
			t.Errorf("Invalid value. isNil should be true")
		}
	})
}
