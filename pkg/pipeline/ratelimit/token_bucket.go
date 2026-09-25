package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket limits event throughput.
type TokenBucket struct {
	mu       sync.Mutex
	capacity int
	tokens   int
	rate     time.Duration
	last     time.Time
}

// New creates a token bucket.
func New(capacity int, refillEvery time.Duration) *TokenBucket {
	if capacity < 1 {
		capacity = 1
	}
	return &TokenBucket{capacity: capacity, tokens: capacity, rate: refillEvery, last: time.Now()}
}

// Allow returns true if a token is available.
func (t *TokenBucket) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	if t.rate > 0 {
		elapsed := now.Sub(t.last)
		refills := int(elapsed / t.rate)
		if refills > 0 {
			t.tokens += refills
			if t.tokens > t.capacity {
				t.tokens = t.capacity
			}
			t.last = now
		}
	}
	if t.tokens <= 0 {
		return false
	}
	t.tokens--
	return true
}
