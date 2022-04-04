package cache

import (
	"errors"
	"time"
)

type Result struct {
	Value string
	Noval bool
	Error error
}

/*
Go routine to provide more persistent key pooling.
In this case, subsequent gets are waiting longer for the main get to pull the data.
*/
func PersistentGet(l *Layer, key string, maxTTL time.Duration, checkDelay time.Duration, result chan Result) {
	start := time.Now()
	loop := time.Tick(checkDelay)
	for now := range loop {
		res := l.Get(key)
		r := Result{Value: res.Value, Noval: res.Nil, Error: res.Err}
		if res.Err != nil || res.Nil == false {
			if _, ok := res.Err.(CacheError); !ok {
				result <- r
				break
			}
		}
		if start.Add(maxTTL).Before(now) {
			result <- Result{Value: "", Noval: false, Error: errors.New("Timeout")}
			break
		}
	}
}
