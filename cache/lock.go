package cache

import (
	"context"
	"time"
)

type Lock interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, key string) error
}

type NoLock struct {
}

func NewNoLock() *NoLock {
	return &NoLock{}
}

func (n *NoLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return true, nil
}

func (n *NoLock) Release(ctx context.Context, key string) error {
	return nil
}
