package cache

import (
	"sync"
	"testing"
)

func TestLRUCache_SetAndGet(t *testing.T) {
	cache, err := NewLRU[string, string](2)
	if err != nil {
		t.Fatalf("NewLRU() error = %v", err)
	}

	cache.Set("first", "value")

	got, ok := cache.Get("first")
	if !ok {
		t.Fatal("Get() returned false")
	}

	if got != "value" {
		t.Fatalf("Get() = %q, want %q", got, "value")
	}
}

func TestLRUCache_EvictsLeastRecentlyUsed(t *testing.T) {
	cache, err := NewLRU[string, string](2)
	if err != nil {
		t.Fatalf("NewLRU() error = %v", err)
	}

	cache.Set("first", "first value")
	cache.Set("second", "second value")

	if _, ok := cache.Get("first"); !ok {
		t.Fatal("first value was not found")
	}

	cache.Set("third", "third value")

	if _, ok := cache.Get("second"); ok {
		t.Fatal("second value should have been evicted")
	}

	if _, ok := cache.Get("first"); !ok {
		t.Fatal("first value should still exist")
	}

	if _, ok := cache.Get("third"); !ok {
		t.Fatal("third value should exist")
	}
}

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	cache, err := NewLRU[string, string](100)
	if err != nil {
		t.Fatalf("NewLRU() error = %v", err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			key := string(rune('a' + i%26))

			cache.Set(key, "value")
			cache.Get(key)
		}(i)
	}

	wg.Wait()
}

func TestLRUCache_InvalidCapacity(t *testing.T) {
	_, err := NewLRU[string, string](0)
	if err == nil {
		t.Fatal("NewLRU(0) expected error")
	}
}
