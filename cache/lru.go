package cache

import "sync"

type Key interface {
	int | string
}

type lruItem[K Key, V any] struct {
	key   K
	value V
	next  *lruItem[K, V]
	prev  *lruItem[K, V]
}

type LRU[K Key, V any] struct {
	index map[K]*lruItem[K, V]
	head  *lruItem[K, V]
	tail  *lruItem[K, V]
	lock  sync.Mutex
	size  int
}

func NewLRU[K Key, V any](size int) *LRU[K, V] {
	c := &LRU[K, V]{
		index: make(map[K]*lruItem[K, V], size),
		lock:  sync.Mutex{},
		size:  size,
	}
	return c
}

func (c *LRU[K, V]) Set(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.setItem(key, value)
}

func (c *LRU[K, V]) Get(key K) (value V) {
	value, _ = c.Find(key)
	return
}

func (c *LRU[K, V]) Find(key K) (value V, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if it, ok := c.findItem(key); ok {
		return it.value, true
	}
	return
}

func (c *LRU[K, V]) setItem(key K, value V) *lruItem[K, V] {
	if it, ok := c.index[key]; ok {
		it.value = value
		c.moveToHead(it)
		return it
	}
	var it *lruItem[K, V]
	if len(c.index) == c.size {
		it = c.removeTail()
		delete(c.index, it.key)
	} else {
		it = &lruItem[K, V]{}
	}
	it.value = value
	it.key = key
	c.index[key] = it
	c.addHead(it)
	return it
}

func (c *LRU[K, V]) findItem(key K) (value *lruItem[K, V], ok bool) {
	if it, ok := c.index[key]; ok {
		c.moveToHead(it)
		value = it
		ok = true
	}
	return
}

func (c *LRU[K, V]) Remove(key K) (value V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if it, ok := c.index[key]; ok {
		c.remove(it)
		delete(c.index, key)
		value = it.value
	}
	return
}
func (c *LRU[K, V]) addHead(item *lruItem[K, V]) {
	item.next = c.head
	c.head = item
	item.prev = nil
	if c.tail == nil {
		c.tail = item
	} else {
		item.next.prev = item
	}
}

func (c *LRU[K, V]) removeTail() (item *lruItem[K, V]) {
	if c.tail != nil {
		item = c.tail
		c.tail = c.tail.prev
		if c.tail != nil {
			c.tail.next = nil
		} else {
			c.head = nil
		}
	}
	return
}

func (c *LRU[K, V]) moveToHead(item *lruItem[K, V]) {
	if c.head != item {
		item.prev.next = item.next
		if item.next != nil {
			item.next.prev = item.prev
		} else {
			c.tail = item.prev
		}
		item.next = c.head
		item.prev = nil
		c.head.prev = item
		c.head = item
	}
}

func (c *LRU[K, V]) remove(item *lruItem[K, V]) {
	if item.prev != nil {
		item.prev.next = item.next
	} else {
		c.head = item.next
	}
	if item.next != nil {
		item.next.prev = item.prev
	} else {
		c.tail = item.prev
	}
}
