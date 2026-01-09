package cache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	ttl     time.Duration
}
