package cache

import (
	"context"
	"time"
)

type Getter func(context.Context, string) (string, bool, error)

type Setter func(context.Context, string, string) (string, error)

type Creator func(ctx context.Context, keyPrefix string, value string) (string, error)

type APIBackend struct {
	Backend
	get_handler Getter
}

func (A *APIBackend) Get(ctx context.Context, key string) *CacheBackendResult {
	val, noval, err := A.get_handler(ctx, key)
	return &CacheBackendResult{Value: val, Nil: noval, Err: err, needsMarshalling: false}
}

func (A *APIBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (A *APIBackend) Del(ctx context.Context, key string) error {
	return nil
}

func NewAPIBackend(fn Getter) *APIBackend {
	return &APIBackend{Backend{name: "API", marshal: false}, fn}
}
