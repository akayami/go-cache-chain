package cache

import (
	"context"
	"time"
)

type Getter func(context.Context, string) (string, bool, error)

type Setter func(context.Context, string, string) (string, error)

type Creator func(ctx context.Context, keyPrefix string, value string) (string, error)

type Deleter func(ctx context.Context, key string) error

type APIBackend struct {
	Backend
	get_handler Getter
	set_handler Setter
	del_handler Deleter
	add_handler Creator
}

func (A *APIBackend) Get(ctx context.Context, key string) *CacheBackendResult {
	val, noval, err := A.get_handler(ctx, key)
	return &CacheBackendResult{Value: val, Nil: noval, Err: err, needsMarshalling: false}
}

func (A *APIBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) (string, error) {
	v, err := A.set_handler(ctx, key, value)
	return v, err
}

func (A *APIBackend) Del(ctx context.Context, key string) error {
	err := A.del_handler(ctx, key)
	return err
}

func NewAPIBackend(getter Getter, setter Setter, creator Creator, deleter Deleter) *APIBackend {
	return &APIBackend{Backend{name: "API", marshal: false}, getter, setter, deleter, creator}
}
