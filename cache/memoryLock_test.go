package cache

import (
	"context"
	"testing"
)

func TestMemoryLock(t *testing.T) {

	CommonLockTests(context.Background(), NewMemoryLock(), t)
}
