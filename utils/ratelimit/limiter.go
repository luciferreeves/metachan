package ratelimit

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu           sync.Mutex
	lastRequest  time.Time
	lastRequests []time.Time
	maxRequests  int
	window       time.Duration
	minDelay     time.Duration
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	minDelay := window / time.Duration(maxRequests)
	return &RateLimiter{
		lastRequests: make([]time.Time, 0, maxRequests),
		maxRequests:  maxRequests,
		window:       window,
		minDelay:     minDelay,
	}
}

func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	if !r.lastRequest.IsZero() {
		elapsed := now.Sub(r.lastRequest)
		if elapsed < r.minDelay {
			waitTime := r.minDelay - elapsed
			r.mu.Unlock()
			time.Sleep(waitTime)
			r.mu.Lock()
			now = time.Now()
		}
	}

	cutoff := now.Add(-r.window)
	i := 0
	for i < len(r.lastRequests) && r.lastRequests[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		r.lastRequests = r.lastRequests[i:]
	}

	if len(r.lastRequests) >= r.maxRequests {
		oldestInWindow := r.lastRequests[0]
		waitDuration := r.window - now.Sub(oldestInWindow)

		r.mu.Unlock()
		time.Sleep(waitDuration + time.Millisecond)
		r.mu.Lock()

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

	r.lastRequest = now
	r.lastRequests = append(r.lastRequests, now)
}

func (r *RateLimiter) RemainingRequests() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	i := 0
	for i < len(r.lastRequests) && r.lastRequests[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		r.lastRequests = r.lastRequests[i:]
	}

	return r.maxRequests - len(r.lastRequests)
}

type MultiLimiter struct {
	limiters []*RateLimiter
}

func NewMultiLimiter(limiters ...*RateLimiter) *MultiLimiter {
	return &MultiLimiter{
		limiters: limiters,
	}
}

func (m *MultiLimiter) Wait() {
	for _, limiter := range m.limiters {
		limiter.Wait()
	}
}
