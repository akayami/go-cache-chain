package cache

import "time"

type Lock interface {
	Acquire(key string, ttl time.Duration) (bool, error)
	Release(key string) error
}

type NoLock struct {
}

func NewNoLock() *NoLock {
	return &NoLock{}
}

func (n *NoLock) Acquire(key string, ttl time.Duration) (bool, error) {
	return true, nil
}

func (n *NoLock) Release(key string) error {
	return nil
}
