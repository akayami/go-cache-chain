package cache

import "time"

type Getter func(string) (string, bool, error)

type APIBackend struct {
	Backend
	get_handler Getter
}

func (A *APIBackend) Get(key string) *CacheBackendResult {
	val, noval, err := A.get_handler(key)
	return &CacheBackendResult{Value: val, Nil: noval, Err: err}
}

func (A *APIBackend) Set(key string, value string, ttl time.Duration) error {
	return nil
}

func (A *APIBackend) Del(key string) error {
	return nil
}

func NewAPIBackend(fn Getter) *APIBackend {
	return &APIBackend{Backend{name: "API"}, fn}
}
