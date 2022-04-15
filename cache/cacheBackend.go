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
	Value string
	Err   error
	Nil   bool
}

func NewCacheBackendResult() *CacheBackendResult {
	o := CacheBackendResult{Nil: false}
	return &o
}

func (c *CacheBackendResult) setValue(v string) *CacheBackendResult {
	c.Value = v
	return c
}

func (c *CacheBackendResult) setError(e error) *CacheBackendResult {
	c.Err = e
	return c
}

func (c *CacheBackendResult) setNil(nil bool) *CacheBackendResult {
	c.Nil = nil
	return c
}

func (c *CacheBackendResult) isNil() bool {
	return c.Nil
}

func (c *CacheBackendResult) getError() error {
	return c.Err
}

func (c *CacheBackendResult) getValue() string {
	return c.Value

}

func (c CacheBackendResult) Expand() (string, bool, error) {
	return c.Value, c.Nil, c.Err
}
