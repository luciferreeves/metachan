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

type MultiLimiter struct {
	limiters []*RateLimiter
}
