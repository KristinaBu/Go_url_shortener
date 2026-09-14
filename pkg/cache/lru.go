package cache

import (
	"container/list"
	"errors"
	"sync"
)

var ErrInvalidCapacity = errors.New("cache capacity must be greater than zero")

type entry[K comparable, V any] struct {
	key   K
	value V
}

type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	items    map[K]*list.Element
	order    *list.List
}

func NewLRU[K comparable, V any](capacity int) (*LRUCache[K, V], error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element, capacity),
		order:    list.New(),
	}, nil
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	c.order.MoveToFront(element)

	item := element.Value.(entry[K, V])
	return item.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {
		element.Value = entry[K, V]{
			key:   key,
			value: value,
		}
		c.order.MoveToFront(element)
		return
	}

	element := c.order.PushFront(entry[K, V]{
		key:   key,
		value: value,
	})

	c.items[key] = element

	if c.order.Len() <= c.capacity {
		return
	}

	oldest := c.order.Back()
	if oldest == nil {
		return
	}

	item := oldest.Value.(entry[K, V])
	delete(c.items, item.key)
	c.order.Remove(oldest)
}
