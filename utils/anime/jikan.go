package anime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"strconv"
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
func waitForJikanRequest() {
	jikanLimiter.Wait()
}

func getAnimeViaJikan(malID int) (*types.JikanAnimeResponse, error) {
	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/full", malID)
	maxRetries := 3
	baseBackoff := 1 * time.Second

	var animeResponse types.JikanAnimeResponse
	success := false
	retries := 0

	for !success && retries <= maxRetries {
		// Use rate limiter before making the request
		logger.Log(fmt.Sprintf("Waiting for rate limiter before requesting anime %d details", malID), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		waitForJikanRequest()

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		client := &http.Client{
			Timeout: 10 * time.Second, // Add timeout to prevent hanging requests
		}
		resp, err := client.Do(req)

		if err != nil {
			if retries < maxRetries {
				retries++
				backoffTime := time.Duration(float64(baseBackoff) * math.Pow(2, float64(retries-1)))
				logger.Log(fmt.Sprintf("Request error for anime details, retrying in %v (retry %d/%d): %v",
					backoffTime, retries, maxRetries, err), types.LogOptions{
					Level:  types.Warn,
					Prefix: "AnimeAPI",
				})
				time.Sleep(backoffTime)
				continue
			}
			return nil, fmt.Errorf("failed to execute request after %d retries: %w", maxRetries, err)
		}

		defer resp.Body.Close()

		// Handle rate limiting with exponential backoff
		if resp.StatusCode == http.StatusTooManyRequests {
			if retries < maxRetries {
				retries++
				backoffTime := time.Duration(float64(baseBackoff) * math.Pow(2, float64(retries-1)))

				// If we have a Retry-After header, respect it
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, err := strconv.Atoi(retryAfter); err == nil {
						backoffTime = time.Duration(seconds) * time.Second
					}
				}

				logger.Log(fmt.Sprintf("Rate limited on anime details, backing off for %v (retry %d/%d)",
					backoffTime, retries, maxRetries), types.LogOptions{
					Level:  types.Warn,
					Prefix: "AnimeAPI",
				})
				time.Sleep(backoffTime)
				continue
			}
			return nil, fmt.Errorf("failed to get anime data: rate limited after %d retries", maxRetries)
		} else if resp.StatusCode != http.StatusOK {
			if retries < maxRetries {
				retries++
				backoffTime := time.Duration(float64(baseBackoff) * math.Pow(2, float64(retries-1)))
				logger.Log(fmt.Sprintf("HTTP error %d for anime details, retrying in %v (retry %d/%d)",
					resp.StatusCode, backoffTime, retries, maxRetries), types.LogOptions{
					Level:  types.Warn,
					Prefix: "AnimeAPI",
				})
				time.Sleep(backoffTime)
				continue
			}
			return nil, fmt.Errorf("failed to get anime data: %s", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&animeResponse); err != nil {
			if retries < maxRetries {
				retries++
				backoffTime := time.Duration(float64(baseBackoff) * math.Pow(2, float64(retries-1)))
				logger.Log(fmt.Sprintf("JSON decode error for anime details, retrying in %v (retry %d/%d): %v",
					backoffTime, retries, maxRetries, err), types.LogOptions{
					Level:  types.Warn,
					Prefix: "AnimeAPI",
				})
				time.Sleep(backoffTime)
				continue
			}
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		success = true
	}

	if !success {
		return nil, fmt.Errorf("failed to fetch anime details after maximum retries")
	}

	if animeResponse.Data.MALID == 0 {
		return nil, fmt.Errorf("no data found for MAL ID %d", malID)
	}

	return &animeResponse, nil
}

func getAnimeEpisodesViaJikan(malId int) (*types.JikanAnimeEpisodeResponse, error) {
	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/episodes", malId)
	var allEpisodes []types.JikanAnimeEpisode
	page := 1
	var lastVisiblePage int

	maxRetries := 3
	baseBackoff := 1 * time.Second
	maxAttempts := 15 // Maximum number of attempts across all pages to prevent infinite loops

	logger.Log(fmt.Sprintf("Fetching episodes for anime %d", malId), types.LogOptions{
		Level:  types.Info,
		Prefix: "AnimeAPI",
	})

	totalAttempts := 0
	for {
		if totalAttempts >= maxAttempts {
			logger.Log(fmt.Sprintf("Reached maximum total attempts (%d) for anime %d. Returning collected episodes so far.",
				maxAttempts, malId), types.LogOptions{
				Level:  types.Warn,
				Prefix: "AnimeAPI",
			})
			break
		}

		var pageResponse types.JikanAnimeEpisodeResponse
		success := false
		retries := 0

		for !success && retries <= maxRetries {
			totalAttempts++

			// Use rate limiter before making the request
			logger.Log(fmt.Sprintf("Waiting for rate limiter before requesting page %d for anime %d", page, malId), types.LogOptions{
				Level:  types.Debug,
				Prefix: "AnimeAPI",
			})
			waitForJikanRequest()

			pageURL := fmt.Sprintf("%s?page=%d", apiURL, page)
			req, err := http.NewRequest("GET", pageURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			// Add a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			req = req.WithContext(ctx)

			client := &http.Client{
				Timeout: 15 * time.Second, // Add timeout to prevent hanging requests
			}

			resp, err := client.Do(req)
			cancel() // Cancel the context regardless of the outcome

			if err != nil {
				if retries < maxRetries {
					retries++
					backoffTime := time.Duration(float64(baseBackoff) * math.Pow(2, float64(retries-1)))
					logger.Log(fmt.Sprintf("Request error, retrying in %v (retry %d/%d): %v",
						backoffTime, retries, maxRetries, err), types.LogOptions{
						Level:  types.Warn,
						Prefix: "AnimeAPI",
					})
					time.Sleep(backoffTime)
					continue
				}
				return nil, fmt.Errorf("failed to execute request after %d retries: %w", maxRetries, err)
			}

			defer resp.Body.Close()

			// Handle rate limiting with exponential backoff
			if resp.StatusCode == http.StatusTooManyRequests {
				if retries < maxRetries {
					retries++

					// Start with a reasonable base backoff
					backoffTime := time.Duration(float64(baseBackoff) * math.Pow(1.5, float64(retries-1)))

					// Respect Retry-After header if available
					if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
						if seconds, err := strconv.Atoi(retryAfter); err == nil {
							backoffTime = time.Duration(seconds) * time.Second
						}
					}

					logger.Log(fmt.Sprintf("Rate limited, backing off for %v (retry %d/%d)",
						backoffTime, retries, maxRetries), types.LogOptions{
						Level:  types.Warn,
						Prefix: "AnimeAPI",
					})
					time.Sleep(backoffTime)
					continue
				}

				// If we've reached maximum retries and still getting rate limited,
				// return what we have so far rather than failing completely
				if len(allEpisodes) > 0 {
					logger.Log(fmt.Sprintf("Rate limited after maximum retries. Returning %d episodes collected so far.",
						len(allEpisodes)), types.LogOptions{
						Level:  types.Warn,
						Prefix: "AnimeAPI",
					})

					return &types.JikanAnimeEpisodeResponse{
						Pagination: types.JikanPagination{
							LastVisiblePage: lastVisiblePage,
							HasNextPage:     false,
						},
						Data: allEpisodes,
					}, nil
				}

				return nil, fmt.Errorf("failed to get anime episodes (page %d): rate limited after %d retries", page, maxRetries)
			} else if resp.StatusCode != http.StatusOK {
				if retries < maxRetries {
					retries++
					backoffTime := time.Duration(float64(baseBackoff) * math.Pow(2, float64(retries-1)))
					logger.Log(fmt.Sprintf("HTTP error %d, retrying in %v (retry %d/%d)",
						resp.StatusCode, backoffTime, retries, maxRetries), types.LogOptions{
						Level:  types.Warn,
						Prefix: "AnimeAPI",
					})
					time.Sleep(backoffTime)
					continue
				}
				return nil, fmt.Errorf("failed to get anime episodes (page %d): %s", page, resp.Status)
			}

			// Limit response body size to prevent memory issues
			bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
			if err != nil {
				if retries < maxRetries {
					retries++
					backoffTime := time.Duration(float64(baseBackoff) * math.Pow(2, float64(retries-1)))
					logger.Log(fmt.Sprintf("Error reading response body, retrying in %v (retry %d/%d): %v",
						backoffTime, retries, maxRetries, err), types.LogOptions{
						Level:  types.Warn,
						Prefix: "AnimeAPI",
					})
					time.Sleep(backoffTime)
					continue
				}
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			if err := json.Unmarshal(bodyBytes, &pageResponse); err != nil {
				if retries < maxRetries {
					retries++
					backoffTime := time.Duration(float64(baseBackoff) * math.Pow(2, float64(retries-1)))
					logger.Log(fmt.Sprintf("JSON decode error, retrying in %v (retry %d/%d): %v",
						backoffTime, retries, maxRetries, err), types.LogOptions{
						Level:  types.Warn,
						Prefix: "AnimeAPI",
					})
					time.Sleep(backoffTime)
					continue
				}
				return nil, fmt.Errorf("failed to decode response: %w", err)
			}

			success = true
		}

		if !success {
			// If we've collected some episodes, return them instead of completely failing
			if len(allEpisodes) > 0 {
				logger.Log(fmt.Sprintf("Failed to fetch page %d after maximum retries. Returning %d episodes collected so far.",
					page, len(allEpisodes)), types.LogOptions{
					Level:  types.Warn,
					Prefix: "AnimeAPI",
				})

				return &types.JikanAnimeEpisodeResponse{
					Pagination: types.JikanPagination{
						LastVisiblePage: page - 1,
						HasNextPage:     false,
					},
					Data: allEpisodes,
				}, nil
			}

			return nil, fmt.Errorf("failed to fetch page %d after maximum retries", page)
		}

		// Convert and append episodes from this page
		for _, episode := range pageResponse.Data {
			allEpisodes = append(allEpisodes, types.JikanAnimeEpisode{
				MALID:         episode.MALID,
				URL:           episode.URL,
				Title:         episode.Title,
				TitleJapanese: episode.TitleJapanese,
				TitleRomaji:   episode.TitleRomaji,
				Aired:         episode.Aired,
				Score:         episode.Score,
				Filler:        episode.Filler,
				Recap:         episode.Recap,
				ForumURL:      episode.ForumURL,
			})
		}

		// Update pagination info
		lastVisiblePage = pageResponse.Pagination.LastVisiblePage

		logger.Log(fmt.Sprintf("Fetched page %d/%d with %d episodes",
			page, lastVisiblePage, len(pageResponse.Data)), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})

		// Check if there are more pages
		if !pageResponse.Pagination.HasNextPage || page >= lastVisiblePage {
			break
		}

		// Safety check - don't fetch more than a reasonable number of pages
		if page >= 25 {
			logger.Log(fmt.Sprintf("Reached maximum page limit (25) for anime %d. Returning collected episodes so far.",
				malId), types.LogOptions{
				Level:  types.Warn,
				Prefix: "AnimeAPI",
			})
			break
		}

		// No need for explicit waiting between pages anymore
		// The rate limiter will handle the pacing automatically
		page++
	}

	logger.Log(fmt.Sprintf("Completed fetching all %d episodes for anime %d",
		len(allEpisodes), malId), types.LogOptions{
		Level:  types.Success,
		Prefix: "AnimeAPI",
	})

	// Return the complete response with all collected episodes
	return &types.JikanAnimeEpisodeResponse{
		Pagination: types.JikanPagination{
			LastVisiblePage: lastVisiblePage,
			HasNextPage:     false,
		},
		Data: allEpisodes,
	}, nil
}
