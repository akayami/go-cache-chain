package cache

import (
	"context"
	"fmt"
	"github.com/akayami/go-cache-chain/cache/schema"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	pbf := &schema.Payload{Payload: payload.Payload, Stale: timestamppb.New(payload.Stale)}
	data, err := proto.Marshal(pbf)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (l *Layer) unmarchal(value []byte) (payload, error) {
	ppbPayload := &schema.Payload{}
	err := proto.Unmarshal(value, ppbPayload)
	if err != nil {
		return payload{}, err
	}
	pl := payload{Payload: ppbPayload.Payload, Stale: ppbPayload.Stale.AsTime()}
	return pl, nil
}

func (l *Layer) Get(ctx context.Context, key string) CacheBackendResult {
	// Get Value from own backed
	r := l.backend.Get(ctx, key)
	//log.Printf("Getting key %s using backend %s", key, l.backend.GetName())
	if r.getError() != nil {
		// Will need a better handling here
		return CacheBackendResult{Value: "", Nil: false, Err: r.getError()}
	} else if r.isNil() {
		// Fetch Value from child
		if l.child != nil {
			// Get Value from the child - Should we use a lock here ? There is a risk of slamming backend for the first key
			b, err := l.lock.Acquire(ctx, key, l.lockTTL)
			if !b {
				return CacheBackendResult{Value: "", Nil: false, Err: CacheError{Message: "Unable to acquire refresh lock"}}
			}
			p := l.child.Get(ctx, key)

			l.lock.Release(ctx, key)

			if p.getError() != nil {
				return CacheBackendResult{Value: "", Nil: false, Err: p.getError()}
			}
			if p.isNil() {
				return CacheBackendResult{Value: "", Nil: true, Err: nil}
			}
			// Set/Update local Value
			go func(ctx context.Context, key string, v string, ttl time.Duration) {
				// Set Value in this layer
				err := l.Set(ctx, key, v)
				//log.Printf("Set cache Value in the background for key %s with Value %v", key, v)
				if err != nil {
					fmt.Println(err)
				}
			}(ctx, key, p.getValue(), l.ttl)
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
				go l.refresh(ctx, key)
			}
			return CacheBackendResult{Value: v.Payload, Nil: false, Err: nil}
		} else {
			// Return Value straight up
			return CacheBackendResult{Value: r.getValue(), Nil: false, Err: nil}
		}
	}
}

func (l *Layer) Set(ctx context.Context, key string, value string) error {
	// create Payload
	payload := payload{value, time.Now().Add(l.stale)}
	// marshal the Payload with Stale argument.
	marshaled_payload, err := l.marshal(payload)
	if err != nil {
		panic(err)
	}
	// Store marshaled data in the backend.
	return l.backend.Set(ctx, key, string(marshaled_payload), l.ttl)
}

func (l *Layer) refresh(ctx context.Context, key string) {
	// Refresh only possible when there is a child layer
	if l.child != nil {
		b, err := l.lock.Acquire(ctx, key, l.lockTTL)
		if !b {
			return
		}
		defer l.lock.Release(ctx, key)
		res := l.child.Get(ctx, key)
		if err != nil {
			// Need to handle this better ?
			log.Printf("Error occured while trying to refresh key %s: %s", key, err.Error())
		} else {
			if res.isNil() {
				log.Printf("Got noval for key %s. It will be cached.", key)
			}
			//@todo decide what should be done for noval cases here. Maybe the layer should have a setting to cache novals?
			// Attempt to store the Value in local cache
			err := l.Set(ctx, key, res.getValue())
			if err != nil {
				panic(err)
			}
		}
	}
}
