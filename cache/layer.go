package cache

import (
	"context"
	"fmt"
	"github.com/akayami/go-cache-chain/cache/schema"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"time"
)

type LayerInterface interface {
	AppendLayer(layer *Layer, childLockTTL time.Duration)
	Get(ctx context.Context, key string, fallback Getter) CacheBackendResult
	Set(ctx context.Context, key string, value string, setter Setter) (string, error)
	Insert(ctx context.Context, keycontext string, value string, creator Creator) (string, string, error)
	Delete(ctx context.Context, key string, deleter Deleter) error
}

type CacheError struct {
	Message string
}

func (e CacheError) Error() string {
	return fmt.Sprintf("Cache Chain Error occured with following message: " + e.Message)
}

type Layer struct {
	childLayer   *Layer
	childLockTTL time.Duration
	backend      CacheBackend
	stale        time.Duration
	ttl          time.Duration
	lock         Lock
}

type payload struct {
	Payload string
	Stale   time.Time
}

func NewLayer(ttl time.Duration, stale time.Duration, backend CacheBackend, lock Lock) *Layer {
	return &Layer{ttl: ttl, stale: stale, backend: backend, lock: lock}
}

/*
Append a childLayer layer to this layer
@layer - the layer to be appended
@childLockTTL - the amount of time a the parent layer will protect the childLayer layer against duplicate refresh requests.
*/
func (l *Layer) AppendLayer(layer *Layer, childLockTTL time.Duration) {
	l.childLayer = layer
	l.childLockTTL = childLockTTL
}

func (l *Layer) getFromBackend(ctx context.Context, key string) *UnmarshaledBackendResult {
	res := l.backend.Get(ctx, key)
	r := &UnmarshaledBackendResult{}
	if res.Err != nil {
		r.Err = res.Err
		return r
	}
	if res.Nil {
		r.Nil = res.Nil
		return r
	}
	if l.backend.IsMarshaled() {
		payload, e := l.unmarshal([]byte(res.Value))
		if e != nil {
			r.Err = e
			return r
		}
		r.Value = payload
	} else {
		r.Value = payload{Payload: res.Value, Stale: time.Now().Add(24 * time.Hour)}
	}
	return r
}

func (l *Layer) putIntoBackend(ctx context.Context, key string, payload payload) error {
	if l.backend.IsMarshaled() {
		marshaled, err := l.marshal(payload)
		if err != nil {
			return err
		}
		_, e := l.backend.Set(ctx, key, string(marshaled), l.ttl)
		return e
	} else {
		_, e := l.backend.Set(ctx, key, payload.Payload, l.ttl)
		return e
	}
}

func (l *Layer) marshal(payload payload) ([]byte, error) {
	pbf := &schema.Payload{Payload: payload.Payload, Stale: timestamppb.New(payload.Stale)}
	data, err := proto.Marshal(pbf)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (l *Layer) unmarshal(value []byte) (payload, error) {
	ppbPayload := &schema.Payload{}
	err := proto.Unmarshal(value, ppbPayload)
	if err != nil {
		return payload{}, err
	}
	pl := payload{Payload: ppbPayload.Payload, Stale: ppbPayload.Stale.AsTime()}
	return pl, nil
}

func (l *Layer) Get(ctx context.Context, key string, fallback Getter) CacheBackendResult {
	// Get Value from own backed
	r := l.getFromBackend(ctx, key)
	//r := l.backend.Get(ctx, key)
	//log.Printf("Getting key %s using backend %s", key, l.backend.GetName())
	if r.Err != nil {
		// Will need a better handling here
		return CacheBackendResult{Value: "", Nil: false, Err: r.Err}
	} else if r.Nil {
		// Fetch Value from childLayer
		if l.childLayer != nil || fallback != nil {
			// Get Value from the childLayer - Should we use a lock here ? There is a risk of slamming backend for the first key
			b, err := l.lock.Acquire(ctx, key, l.childLockTTL)
			if !b {
				return CacheBackendResult{Value: "", Nil: false, Err: CacheError{Message: "Unable to acquire refresh lock"}}
			}
			var p CacheBackendResult
			if l.childLayer != nil {
				p = l.childLayer.Get(ctx, key, fallback)
			} else {
				val, noval, err := fallback(ctx, key)
				if err != nil {
					p = CacheBackendResult{Value: "", Nil: false, Err: err}
				} else {
					p = CacheBackendResult{Value: val, Nil: noval, Err: err}
				}
			}

			l.lock.Release(ctx, key)

			if p.Err != nil {
				return CacheBackendResult{Value: "", Nil: false, Err: p.Err}
			}
			if p.Nil {
				return CacheBackendResult{Value: "", Nil: true, Err: nil}
			}
			// Set/Update local Value
			go func(ctx context.Context, key string, v string, ttl time.Duration) {
				// Set Value in this layer
				err := l.putIntoBackend(ctx, key, *l.getPayload(v))
				if err != nil {
					fmt.Println(err)
				}
			}(ctx, key, p.Value, l.ttl)
			// return retrieve Value
			return CacheBackendResult{Value: p.Value, Nil: false, Err: err}
		} else {
			return CacheBackendResult{Value: "", Nil: true, Err: nil}
		}
	} else {
		//return CacheBackendResult{Value: r.getValue(), Nil: false, Err: nil}
		// If there is no childLayer, the data should not be marshaled.
		// @todo review this logic

		now := time.Now()
		if r.Value.Stale.Before(now) {
			go l.refresh(ctx, key, fallback)
		}
		return CacheBackendResult{Value: r.Value.Payload, Nil: false, Err: nil}
	}
}

func (l *Layer) getPayload(value string) *payload {
	return &payload{value, time.Now().Add(l.stale)}
}

func (l *Layer) Set(ctx context.Context, key string, value string, setter Setter) (string, error) {
	// create Payload
	payload := l.getPayload(value)
	lockKey := "set_" + key

	err := l.putIntoBackend(ctx, key, *payload)
	if err != nil {
		return "", err
	}
	l.lock.Acquire(ctx, lockKey, 100*time.Millisecond)

	go func(ctx context.Context, key string, value string, setter Setter, lockKey string) {
		defer l.lock.Release(ctx, lockKey)
		var v string
		var err error
		if l.childLayer != nil {
			v, err = l.childLayer.Set(ctx, key, value, setter)
		} else if setter != nil {
			v, err = setter(ctx, key, value)
		} else {
			return
		}
		if err != nil {
			log.Println(err)
		}
		// Update local value in case backed returned a different value than frontend
		// Typically, some backend, like db may, for example, truncate the saved value,
		//so it is always good to check if the value after insertion was changed and overwrite it
		if v != value {
			err := l.putIntoBackend(ctx, key, *l.getPayload(v))
			if err != nil {
				log.Println(err)
			}
		}
	}(ctx, key, value, setter, lockKey)
	return value, nil
}

func (l *Layer) Insert(ctx context.Context, keycontext string, value string, creator Creator) (string, string, error) {
	id, err := creator(ctx, keycontext, value)
	if err != nil {
		return "", "", err
	}
	v, e := l.Set(ctx, keycontext+id, value, nil)
	if e != nil {
		return "", "", e
	}
	return id, v, nil
}

func (l *Layer) Delete(ctx context.Context, key string, deleter Deleter) error {
	if l.childLayer != nil {
		err := l.childLayer.Delete(ctx, key, deleter)
		if err != nil {
			return err
		}
	} else if deleter != nil {
		err := deleter(ctx, key)
		if err != nil {
			return err
		}
	}
	err := l.backend.Del(ctx, key)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (l *Layer) refresh(ctx context.Context, key string, fallback Getter) {
	// Refresh only possible when there is a childLayer layer
	if l.childLayer != nil || fallback != nil {
		b, err := l.lock.Acquire(ctx, key, l.childLockTTL)
		if !b {
			return
		}
		defer l.lock.Release(ctx, key)
		var res CacheBackendResult
		if l.childLayer != nil {
			res = l.childLayer.Get(ctx, key, fallback)
		} else {
			val, noval, err := fallback(ctx, key)
			res = CacheBackendResult{Value: val, Nil: noval, Err: err}
		}
		if err != nil {
			// Need to handle this better ?
			log.Printf("Error occured while trying to refresh key %s: %s", key, err.Error())
		} else {
			if res.Nil {
				log.Printf("Got noval for key %s. It will be cached.", key)
			}
			//@todo decide what should be done for noval cases here. Maybe the layer should have a setting to cache novals?
			// Attempt to store the Value in local cache
			err := l.putIntoBackend(ctx, key, *l.getPayload(res.Value))
			//err := l.backend.Set(ctx, key, res.getValue(), l.ttl)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}
