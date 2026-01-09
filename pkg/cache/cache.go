package cache

import "time"

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, exists := c.entries[key]
	if !exists || time.Since(entry.Timestamp) > c.ttl {
		return nil, false
	}
	return entry.Data, true
}

func (c *Cache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = CacheEntry{Data: data, Timestamp: time.Now()}
}
