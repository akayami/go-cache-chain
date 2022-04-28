package cache

import (
	"context"
	"time"
)

type Getter func(context.Context, string) (string, bool, error)

type Setter func(context.Context, string, string) (string, error)

type Creator func(ctx context.Context, keyPrefix string, value string) (string, error)

type Deleter func(ctx context.Context, key string) error

type CacheBackend interface {
	Get(ctx context.Context, key string) *CacheBackendResult
	Set(ctx context.Context, key string, value string, ttl time.Duration) (string, error)
	Del(ctx context.Context, key string) error
	GetName() string
	IsMarshaled() bool
}
