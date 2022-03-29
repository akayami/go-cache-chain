package cache

import "time"

//type CacheBackend interface {
//	Set(key string, value string, ttl int) bool
//	Get(key string) string
//	Del(key string) bool
//}

type CacheBackend interface {
	Get(key string) *CacheBackendResult
	Set(key string, value string, ttl time.Duration) error
	Del(key string) error
	GetName() string
}

type CacheBackendResult struct {
	value string
	err   error
	nil   bool
}

func NewCacheBackendResult() *CacheBackendResult {
	o := CacheBackendResult{nil: false}
	return &o
}

func (c *CacheBackendResult) setValue(v string) {
	c.value = v
}

func (c *CacheBackendResult) setError(e error) {
	c.err = e
}

func (c *CacheBackendResult) setNil(nil bool) {
	c.nil = nil
}

func (c *CacheBackendResult) isNil() bool {
	return c.nil
}

func (c *CacheBackendResult) getError() error {
	return c.err
}

func (c *CacheBackendResult) getValue() string {
	return c.value

}
