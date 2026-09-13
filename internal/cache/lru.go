package cache

import (
	"container/list"
	"errors"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"sync"
)

var ErrInvalidCapacity = errors.New("cache capacity must be greater than zero")

type entry struct {
	key  string
	link domain.Link
}

type LRUCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

func NewLRU(capacity int) (*LRUCache, error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}, nil
}

func (c *LRUCache) Get(key string) (domain.Link, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.items[key]
	if !ok {
		return domain.Link{}, false
	}

	c.order.MoveToFront(element)

	item := element.Value.(entry)

	return item.link, true
}

func (c *LRUCache) Set(key string, link domain.Link) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {
		element.Value = entry{
			key:  key,
			link: link,
		}
		c.order.MoveToFront(element)
		return
	}

	element := c.order.PushFront(entry{
		key:  key,
		link: link,
	})

	c.items[key] = element

	if c.order.Len() <= c.capacity {
		return
	}

	oldest := c.order.Back()
	if oldest == nil {
		return
	}

	item := oldest.Value.(entry)

	delete(c.items, item.key)
	c.order.Remove(oldest)
}
