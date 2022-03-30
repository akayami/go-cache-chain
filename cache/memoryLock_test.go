package cache

import (
	"testing"
)

func TestMemoryLock(t *testing.T) {

	CommonLockTests(NewMemoryLock(), t)
}
