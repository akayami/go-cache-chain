package helpers

import (
	"context"
	"strconv"
)

type ticker func()

type TestBackend struct {
	id      int
	store   map[string]string
	getFunc ticker
}

func NewTestBackend(arg ticker) *TestBackend {
	return &TestBackend{id: 0, store: map[string]string{}, getFunc: arg}
	//return &TestBackend{id: 0}
}

func (b *TestBackend) Create(ctx context.Context, keyPrefix string, value string) (string, error) {
	id := keyPrefix + strconv.Itoa(b.id)
	b.id++
	b.store[id] = value
	b.id++
	return id, nil
}

func (b *TestBackend) Get(ctx context.Context, keyPrefix string) (string, bool, error) {
	if val, ok := b.store[keyPrefix]; ok {
		if b.getFunc != nil {
			b.getFunc()
		}
		return val, false, nil
	} else {
		return "", true, nil
	}
}

func (b *TestBackend) Delete(ctx context.Context, key string) error {
	delete(b.store, key)
	return nil
}
