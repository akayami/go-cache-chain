package cache

import (
	"context"
	"time"
)

type CacheBackend interface {
	Get(ctx context.Context, key string) *CacheBackendResult
	Set(ctx context.Context, key string, value string, ttl time.Duration) (string, error)
	Del(ctx context.Context, key string) error
	GetName() string
	IsMarshaled() bool
}
