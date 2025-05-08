package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter manages request rate limiting for APIs
type RateLimiter struct {
	mu           sync.Mutex
	lastRequests []time.Time
	maxRequests  int           // Maximum requests per time window
	window       time.Duration // Time window duration
}

// NewRateLimiter creates a new rate limiter
// maxRequests is the maximum number of requests allowed in the specified time window
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		lastRequests: make([]time.Time, 0, maxRequests),
		maxRequests:  maxRequests,
		window:       window,
	}
}

// Wait blocks until a request can be made according to rate limiting rules
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Clean up old requests
	cutoff := now.Add(-r.window)
	i := 0
	for i < len(r.lastRequests) && r.lastRequests[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		r.lastRequests = r.lastRequests[i:]
	}

	// If we've reached max requests in the window, wait until we can make another
	if len(r.lastRequests) >= r.maxRequests {
		// Calculate wait time based on the oldest request in the window
		oldestInWindow := r.lastRequests[0]
		waitDuration := r.window - now.Sub(oldestInWindow)

		// Release lock while waiting
		r.mu.Unlock()
		time.Sleep(waitDuration + time.Millisecond) // Add 1ms to be safe
		r.mu.Lock()                                 // Re-acquire lock

		// Refresh current time and clean up again after waiting
		now = time.Now()
		cutoff = now.Add(-r.window)
		i = 0
		for i < len(r.lastRequests) && r.lastRequests[i].Before(cutoff) {
			i++
		}
		if i > 0 {
			r.lastRequests = r.lastRequests[i:]
		}
	}

	// Add current request timestamp
	r.lastRequests = append(r.lastRequests, now)
}

// RemainingRequests returns the number of requests that can still be made in the current window
func (r *RateLimiter) RemainingRequests() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Clean up old requests
	i := 0
	for i < len(r.lastRequests) && r.lastRequests[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		r.lastRequests = r.lastRequests[i:]
	}

	return r.maxRequests - len(r.lastRequests)
}

// MultiLimiter combines multiple rate limiters
type MultiLimiter struct {
	limiters []*RateLimiter
}

// NewMultiLimiter creates a new multi-limiter from the given limiters
func NewMultiLimiter(limiters ...*RateLimiter) *MultiLimiter {
	return &MultiLimiter{
		limiters: limiters,
	}
}

// Wait waits for all underlying limiters
func (m *MultiLimiter) Wait() {
	for _, limiter := range m.limiters {
		limiter.Wait()
	}
}
