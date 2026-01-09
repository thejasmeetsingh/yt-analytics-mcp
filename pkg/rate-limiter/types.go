package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu             sync.Mutex
	tokens         int
	maxTokens      int
	refillRate     time.Duration
	lastRefillTime time.Time
}
