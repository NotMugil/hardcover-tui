package api

import (
	"sync"
	"time"
)

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// QueryCache stores in-memory query results with TTL expiration.
type QueryCache struct {
	items map[string]cacheItem
	mu    sync.RWMutex
}

// NewQueryCache creates a thread-safe QueryCache instance.
func NewQueryCache() *QueryCache {
	return &QueryCache{
		items: make(map[string]cacheItem),
	}
}

// Get fetches a value from cache if it exists and hasn't expired.
func (c *QueryCache) Get(key string) (interface{}, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}
	if time.Now().After(item.expiresAt) {
		return nil, false
	}
	return item.value, true
}

// Set stores a value in cache with a TTL duration.
func (c *QueryCache) Set(key string, value interface{}, ttl time.Duration) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Invalidate removes keys starting with keyPrefix.
func (c *QueryCache) Invalidate(keyPrefix string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	for k := range c.items {
		if keyPrefix == "" || (len(k) >= len(keyPrefix) && k[:len(keyPrefix)] == keyPrefix) {
			delete(c.items, k)
		}
	}
}

// Clear flushes all cached entries.
func (c *QueryCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cacheItem)
}
