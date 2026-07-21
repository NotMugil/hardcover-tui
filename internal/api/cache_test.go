package api

import (
	"testing"
	"time"
)

func TestQueryCacheSetGet(t *testing.T) {
	cache := NewQueryCache()

	cache.Set("user_123", "John Doe", 100*time.Millisecond)

	val, found := cache.Get("user_123")
	if !found {
		t.Fatalf("expected key user_123 to be found")
	}
	if val.(string) != "John Doe" {
		t.Fatalf("expected 'John Doe', got %v", val)
	}

	time.Sleep(150 * time.Millisecond)

	_, foundAfterExp := cache.Get("user_123")
	if foundAfterExp {
		t.Fatalf("expected key user_123 to expire after TTL")
	}
}

func TestQueryCacheInvalidate(t *testing.T) {
	cache := NewQueryCache()

	cache.Set("books_1", "Book 1", 1*time.Minute)
	cache.Set("books_2", "Book 2", 1*time.Minute)
	cache.Set("user_1", "User 1", 1*time.Minute)

	cache.Invalidate("books_")

	if _, found := cache.Get("books_1"); found {
		t.Fatalf("expected books_1 to be invalidated")
	}
	if _, found := cache.Get("books_2"); found {
		t.Fatalf("expected books_2 to be invalidated")
	}
	if _, found := cache.Get("user_1"); !found {
		t.Fatalf("expected user_1 to remain in cache")
	}
}

func TestQueryCacheClear(t *testing.T) {
	cache := NewQueryCache()

	cache.Set("k1", "v1", 1*time.Minute)
	cache.Set("k2", "v2", 1*time.Minute)

	cache.Clear()

	if _, found := cache.Get("k1"); found {
		t.Fatalf("expected cache to be cleared")
	}
}
