package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Layer struct {
	child   *Layer
	backend CacheBackend
	stale   time.Duration
	ttl     time.Duration
}

type payload struct {
	Payload string
	Stale   time.Time
}

func NewLayer(ttl time.Duration, stale time.Duration, backend CacheBackend) *Layer {
	return &Layer{ttl: ttl, stale: stale, backend: backend}
}

func (l *Layer) AppendLayer(layer *Layer) {
	l.child = layer
}

func (l *Layer) marshal(payload payload) ([]byte, error) {
	return json.Marshal(payload)
}

func (l *Layer) unmarchal(value []byte) (payload, error) {
	pl := payload{}
	err := json.Unmarshal(value, &pl)
	return pl, err
}

func (l *Layer) Get(key string) (string, bool, error) {
	// Get value from own backed
	r := l.backend.Get(key)
	//log.Printf("Getting key %s using backend %s", key, l.backend.GetName())
	if r.getError() != nil {
		// Will need a better handling here
		return "", false, r.getError()
	} else if r.isNil() {
		// Fetch Value from child
		if l.child != nil {
			v, noval, err := l.child.Get(key)
			if err != nil {
				return "", false, err
			}
			if noval {
				return "", true, nil
			}
			go func(key string, v string, ttl time.Duration) {
				// Set value in this layer
				err := l.Set(key, v)
				//log.Printf("Set cache value in the background for key %s with value %v", key, v)
				if err != nil {
					fmt.Println(err)
				}
			}(key, v, l.ttl)
			return v, false, nil
		} else {
			// No child, signaling no value
			return "", true, nil
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
			return v.Payload, false, nil
		} else {
			// Return value straight up
			return r.getValue(), false, nil
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
		v, noval, err := l.child.Get(key)
		if err != nil {
			// Need to handle this better ?
			log.Printf("Error occured while trying to refresh key %s: %s", key, err.Error())
		}
		if noval {
			//@todo decide what should be done for noval cases here. Maybe the layer should have a setting to cache novals?
		} else {
			// Attempt to store the value in local cache
			err := l.Set(key, v)
			if err != nil {
				panic(err)
			}
		}
	}
}