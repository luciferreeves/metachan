package ratelimit

import (
	"time"
)

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	interval := window / time.Duration(maxRequests)
	rl := &RateLimiter{
		tokens: make(chan struct{}, maxRequests),
		done:   make(chan struct{}),
	}
	rl.tokens <- struct{}{}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case rl.tokens <- struct{}{}:
				default:
				}
			case <-rl.done:
				return
			}
		}
	}()
	return rl
}

func (r *RateLimiter) Wait() {
	<-r.tokens
}

func (r *RateLimiter) Stop() {
	close(r.done)
}

func (r *RateLimiter) RemainingRequests() int {
	return len(r.tokens)
}

func NewMultiLimiter(limiters ...*RateLimiter) *MultiLimiter {
	return &MultiLimiter{
		limiters: limiters,
	}
}

func (m *MultiLimiter) Wait() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, limiter := range m.limiters {
		limiter.Wait()
	}
}
