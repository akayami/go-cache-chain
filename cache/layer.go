package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type CacheError struct {
	Message string
}

func (e CacheError) Error() string {
	return fmt.Sprintf("Cache Chain Error occured with following message: " + e.Message)
}

type Layer struct {
	child   *Layer
	backend CacheBackend
	stale   time.Duration
	ttl     time.Duration
	lock    Lock
	lockTTL time.Duration
}

type payload struct {
	Payload string
	Stale   time.Time
}

func NewLayer(ttl time.Duration, stale time.Duration, backend CacheBackend, lock Lock) *Layer {
	return &Layer{ttl: ttl, stale: stale, backend: backend, lock: lock}
}

/*
Append a child layer to this layer
@layer - the layer to be appended
@lockTTL - the amount of time a the parent layer will protect the child layer against duplicate refresh requests.
*/
func (l *Layer) AppendLayer(layer *Layer, lockTTL time.Duration) {
	l.child = layer
	l.lockTTL = lockTTL
}

func (l *Layer) marshal(payload payload) ([]byte, error) {
	return json.Marshal(payload)
}

func (l *Layer) unmarchal(value []byte) (payload, error) {
	pl := payload{}
	err := json.Unmarshal(value, &pl)
	return pl, err
}

func (l *Layer) Get(key string) CacheBackendResult {
	// Get Value from own backed
	r := l.backend.Get(key)
	//log.Printf("Getting key %s using backend %s", key, l.backend.GetName())
	if r.getError() != nil {
		// Will need a better handling here
		return CacheBackendResult{Value: "", Nil: false, Err: r.getError()}
	} else if r.isNil() {
		// Fetch Value from child
		if l.child != nil {
			// Get Value from the child - Should we use a lock here ? There is a risk of slamming backend for the first key
			b, err := l.lock.Acquire(key, l.lockTTL)
			if !b {
				return CacheBackendResult{Value: "", Nil: false, Err: CacheError{Message: "Unable to acquire refresh lock"}}
			}
			p := l.child.Get(key)

			if p.getError() != nil {
				return CacheBackendResult{Value: "", Nil: false, Err: p.getError()}
			}
			if p.isNil() {
				return CacheBackendResult{Value: "", Nil: true, Err: nil}
			}
			// Set/Update local Value
			go func(key string, v string, ttl time.Duration) {
				// Set Value in this layer
				err := l.Set(key, v)
				//log.Printf("Set cache Value in the background for key %s with Value %v", key, v)
				if err != nil {
					fmt.Println(err)
				}
			}(key, p.getValue(), l.ttl)
			// return retrieve Value
			return CacheBackendResult{Value: p.Value, Nil: false, Err: err}
		} else {
			// No child, signaling no Value
			return CacheBackendResult{Value: "", Nil: true, Err: nil}
		}
	} else {
		// If there is no child, the data should not be marshaled.
		// @todo review this logic
		if l.child != nil {
			v, _ := l.unmarchal([]byte(r.getValue()))
			now := time.Now()
			if v.Stale.Before(now) {
				//log.Printf("Detected stale %s vs %s", v.Stale, now)
				go l.refresh(key)
			}
			return CacheBackendResult{Value: v.Payload, Nil: false, Err: nil}
		} else {
			// Return Value straight up
			return CacheBackendResult{Value: r.getValue(), Nil: false, Err: nil}
		}
	}
}

func (l *Layer) Set(key string, value string) error {
	// create Payload
	payload := payload{value, time.Now().Add(l.stale)}
	// marshal the Payload with Stale argument.
	marshaled_payload, err := l.marshal(payload)
	if err != nil {
		panic(err)
	}
	// Store marshaled data in the backend.
	return l.backend.Set(key, string(marshaled_payload), l.ttl)
}

func (l *Layer) refresh(key string) {
	// Refresh only possible when there is a child layer
	if l.child != nil {
		b, err := l.lock.Acquire(key, l.lockTTL)
		if !b {
			return
		}
		defer l.lock.Release(key)
		res := l.child.Get(key)
		if err != nil {
			// Need to handle this better ?
			log.Printf("Error occured while trying to refresh key %s: %s", key, err.Error())
		} else {
			if res.isNil() {
				log.Printf("Got noval for key %s. It will be cached.", key)
			}
			//@todo decide what should be done for noval cases here. Maybe the layer should have a setting to cache novals?
			// Attempt to store the Value in local cache
			err := l.Set(key, res.getValue())
			if err != nil {
				panic(err)
			}
		}
	}
}
