package ratelimit

import (
	"sync"
)

type RateLimiter struct {
	tokens chan struct{}
	done   chan struct{}
}

type MultiLimiter struct {
	mu       sync.Mutex
	limiters []*RateLimiter
}
