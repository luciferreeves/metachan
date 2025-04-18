package anime

import (
	"metachan/types"
	"metachan/utils/logger"
	"sync"
	"time"
)

// JikanRateLimiter manages request rate limiting for the Jikan API
// Jikan allows maximum 3 requests per second, 60 requests per minute
type JikanRateLimiter struct {
	mu             sync.Mutex
	lastRequests   []time.Time
	perSecRequests int // Max requests per second
	perSecWindow   time.Duration
	perMinRequests int // Max requests per minute
	perMinWindow   time.Duration
}

var (
	// Global Jikan rate limiter instance with conservative settings
	jikanLimiter = &JikanRateLimiter{
		lastRequests:   make([]time.Time, 0, 60),
		perSecRequests: 3, // More conservative than the stated 3/sec
		perSecWindow:   time.Second,
		perMinRequests: 60, // More conservative than the stated 60/min
		perMinWindow:   time.Minute,
	}
)

// Wait blocks until a request can be made according to rate limiting rules
func (r *JikanRateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Clean up old requests
	r.cleanupOldRequests(now)

	// Count requests in the current windows
	secWindowRequests := 0
	minWindowRequests := 0

	for _, t := range r.lastRequests {
		if now.Sub(t) < r.perSecWindow {
			secWindowRequests++
		}
		if now.Sub(t) < r.perMinWindow {
			minWindowRequests++
		}
	}

	logger.Log(
		"Rate limit check - Second: "+
			time.Duration(secWindowRequests).String()+"/"+time.Duration(r.perSecRequests).String()+
			" - Minutes: "+time.Duration(minWindowRequests).String()+"/"+time.Duration(r.perMinRequests).String(),
		types.LogOptions{
			Level:  types.Debug,
			Prefix: "JikanAPI",
		},
	)

	// Calculate necessary delay
	var delay time.Duration

	// Check per-second limit
	if secWindowRequests >= r.perSecRequests && len(r.lastRequests) > 0 {
		// Find the oldest request within the second window
		var oldestInSecWindow time.Time
		foundInSecWindow := false

		for _, t := range r.lastRequests {
			if now.Sub(t) < r.perSecWindow {
				if !foundInSecWindow || t.Before(oldestInSecWindow) {
					oldestInSecWindow = t
					foundInSecWindow = true
				}
			}
		}

		if foundInSecWindow {
			// Calculate when we can make the next request
			secDelay := r.perSecWindow - now.Sub(oldestInSecWindow) + 200*time.Millisecond // Add buffer
			if secDelay > 0 {
				delay = secDelay
			}
		}
	}

	// Check per-minute limit
	if minWindowRequests >= r.perMinRequests && len(r.lastRequests) > 0 {
		// Find the oldest request within the minute window
		var oldestInMinWindow time.Time
		foundInMinWindow := false

		for _, t := range r.lastRequests {
			if now.Sub(t) < r.perMinWindow {
				if !foundInMinWindow || t.Before(oldestInMinWindow) {
					oldestInMinWindow = t
					foundInMinWindow = true
				}
			}
		}

		if foundInMinWindow {
			// Calculate when we can make the next request
			minDelay := r.perMinWindow - now.Sub(oldestInMinWindow) + 200*time.Millisecond // Add buffer
			if minDelay > delay {
				delay = minDelay
			}
		}
	}

	// If we need to wait, do so
	if delay > 0 {
		// Log and sleep
		r.mu.Unlock() // Unlock while sleeping
		logger.Log("Rate limiting Jikan API request - waiting "+delay.String(), types.LogOptions{
			Level:  types.Info,
			Prefix: "JikanAPI",
		})
		time.Sleep(delay)
		r.mu.Lock()      // Lock again before modifying state
		now = time.Now() // Update current time
	}

	// Record this request with current time
	r.lastRequests = append(r.lastRequests, now)
}

// cleanupOldRequests removes requests older than the longest window
func (r *JikanRateLimiter) cleanupOldRequests(now time.Time) {
	validRequests := make([]time.Time, 0, len(r.lastRequests))
	for _, t := range r.lastRequests {
		// Keep requests that are within our longest time window (per minute)
		if now.Sub(t) < r.perMinWindow {
			validRequests = append(validRequests, t)
		}
	}
	r.lastRequests = validRequests
}

// WaitForJikanRequest is a convenience function to access the global rate limiter
func WaitForJikanRequest() {
	jikanLimiter.Wait()
}
