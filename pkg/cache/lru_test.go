package cache

import (
	"sync"
	"testing"

	"github.com/KristinaBu/Go_url_shortener/internal/domain"
)

func TestLRUCache_SetAndGet(t *testing.T) {
	cache, err := NewLRU[string, domain.Link](2)
	if err != nil {
		t.Fatalf("NewLRU() error = %v", err)
	}

	link := domain.Link{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
	}

	cache.Set(link.ShortCode, link)

	got, ok := cache.Get(link.ShortCode)
	if !ok {
		t.Fatal("Get() returned false")
	}

	if got != link {
		t.Fatalf("Get() = %+v, want %+v", got, link)
	}
}

func TestLRUCache_EvictsLeastRecentlyUsed(t *testing.T) {
	cache, err := NewLRU[string, domain.Link](2)
	if err != nil {
		t.Fatalf("NewLRU() error = %v", err)
	}

	cache.Set("first", domain.Link{ShortCode: "first"})
	cache.Set("second", domain.Link{ShortCode: "second"})

	if _, ok := cache.Get("first"); !ok {
		t.Fatal("first link was not found")
	}

	cache.Set("third", domain.Link{ShortCode: "third"})

	if _, ok := cache.Get("second"); ok {
		t.Fatal("second link should have been evicted")
	}

	if _, ok := cache.Get("first"); !ok {
		t.Fatal("first link should still exist")
	}

	if _, ok := cache.Get("third"); !ok {
		t.Fatal("third link should exist")
	}
}

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	cache, err := NewLRU[string, domain.Link](100)
	if err != nil {
		t.Fatalf("NewLRU() error = %v", err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			key := string(rune('a' + i%26))

			cache.Set(key, domain.Link{
				ShortCode: key,
			})

			cache.Get(key)
		}(i)
	}

	wg.Wait()
}

func TestLRUCache_InvalidCapacity(t *testing.T) {
	_, err := NewLRU[string, domain.Link](0)
	if err == nil {
		t.Fatal("NewLRU(0) expected error")
	}
}
