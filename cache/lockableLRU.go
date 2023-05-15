package cache

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Loader[K Key, V any] func(ctx context.Context, key K) (V, error)
type Saver[K Key, V any] func(ctx context.Context, value V) error

type lockableItem[V any] struct {
	value V
	lock  sync.Mutex
}

type Lockable[K Key, V any] struct {
	*LRU[K, *lockableItem[V]]
	loader Loader[K, V]
	saver  Saver[K, V]
}

func NewLockable[K Key, V any](size int, loader Loader[K, V], saver Saver[K, V]) *Lockable[K, V] {
	l := NewLRU[K, *lockableItem[V]](size)
	c := &Lockable[K, V]{
		LRU:    l,
		loader: loader,
		saver:  saver,
	}
	return c
}

func (c *Lockable[K, V]) GetAndLock(ctx context.Context, key K) (value V, err error) {
	c.lock.Lock()
	it, ok := c.findItem(key)
	if ok {
		value = it.value.value
	} else {
		value, err = c.loader(ctx, key)
		if err != nil {
			return
		}
		it = c.setItem(key, &lockableItem[V]{value: value, lock: sync.Mutex{}})
	}
	c.lock.Unlock()
	it.value.lock.Lock()
	return
}

func (c *Lockable[K, V]) PutAndLock(key K, value V) {
	c.lock.Lock()
	it, ok := c.findItem(key)
	if ok {
		it.value.value = value
	} else {
		it = c.setItem(key, &lockableItem[V]{value: value, lock: sync.Mutex{}})
	}
	c.lock.Unlock()
	it.value.lock.Lock()
}

func (c *Lockable[K, V]) Unlock(key K) (value V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if it, ok := c.index[key]; ok {
		value = it.value.value
		it.value.lock.Unlock()
	}
	return
}

func (c *Lockable[K, V]) SaveAndUnlock(ctx context.Context, key K, value V) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if it, ok := c.findItem(key); ok {
		it.value.value = value
		c.saver(ctx, value)
		it.value.lock.Unlock()
	} else {
		return ErrNotFound
	}
	return
}
