package cache

import (
	"context"
	"time"
)

type Getter func(ctx context.Context, key string) (string, bool, error)

type Setter func(ctx context.Context, key string, val string) (string, error)

type Creator func(ctx context.Context, keyPrefix string, value string) (string, error)

type Deleter func(ctx context.Context, key string) error

type CacheBackend interface {
	Get(ctx context.Context, key string) *CacheBackendResult
	Set(ctx context.Context, key string, value string, ttl time.Duration) (string, error)
	Del(ctx context.Context, key string) error
	GetName() string
	IsMarshaled() bool
}
