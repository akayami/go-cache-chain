package cache

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
