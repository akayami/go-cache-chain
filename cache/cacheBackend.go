package cache

import (
	"context"
	"time"
)

type CacheBackend interface {
	Get(ctx context.Context, key string) *CacheBackendResult
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	GetName() string
	IsMarshaled() bool
}

type CacheBackendResult struct {
	Value            string
	Err              error
	Nil              bool
	needsMarshalling bool
}

type UnmarshaledBackendResult struct {
	Value payload
	Err   error
	Nil   bool
}

func NewCacheBackendResult() *CacheBackendResult {
	o := CacheBackendResult{Nil: false, needsMarshalling: true}
	return &o
}
