package aniskip

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"metachan/utils/logger"
	"metachan/utils/ratelimit"
	"net/http"
	"sync"
	"time"
)

const (
	aniskipBaseURL = "https://api.aniskip.com/v2"
)

// AniSkipClient provides methods for interacting with the AniSkip API
type AniSkipClient struct {
	client      *http.Client
	rateLimiter *ratelimit.RateLimiter
	maxRetries  int
	cache       map[string][]AnimeSkipTimes
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration
	cacheTime   map[string]time.Time
}

// EpisodeSkipTimesResult contains skip times for a specific episode
type EpisodeSkipTimesResult struct {
	EpisodeNumber int
	SkipTimes     []AnimeSkipTimes
}

// NewAniSkipClient creates a new client for the AniSkip API
func NewAniSkipClient() *AniSkipClient {
	return &AniSkipClient{
		client: &http.Client{
			Timeout: 5 * time.Second, // Reduced timeout for faster failure detection
		},
		rateLimiter: ratelimit.NewRateLimiter(10, 10*time.Second), // Conservative rate limit
		maxRetries:  2,
		cache:       make(map[string][]AnimeSkipTimes),
		cacheTime:   make(map[string]time.Time),
		cacheTTL:    24 * time.Hour, // Cache skip times for 24 hours
	}
}

// getCacheKey generates a cache key for skip times
func (c *AniSkipClient) getCacheKey(malID, episode int) string {
	return fmt.Sprintf("%d-%d", malID, episode)
}

// getFromCache tries to get skip times from cache
func (c *AniSkipClient) getFromCache(malID, episode int) ([]AnimeSkipTimes, bool) {
	key := c.getCacheKey(malID, episode)

	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	// Check if we have a valid cache entry
	if cacheTime, exists := c.cacheTime[key]; exists {
		// Check if cache is still valid
		if time.Since(cacheTime) < c.cacheTTL {
			return c.cache[key], true
		}
	}

	return nil, false
}

// saveToCache saves skip times to cache
func (c *AniSkipClient) saveToCache(malID, episode int, skipTimes []AnimeSkipTimes) {
	key := c.getCacheKey(malID, episode)

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cache[key] = skipTimes
	c.cacheTime[key] = time.Now()
}

// GetSkipTimesForEpisode fetches skip times for a specific anime episode
func (c *AniSkipClient) GetSkipTimesForEpisode(malID, episodeNumber int) ([]AnimeSkipTimes, error) {
	// Check cache first
	if skipTimes, found := c.getFromCache(malID, episodeNumber); found {
		return skipTimes, nil
	}

	// Wait for rate limiter before making request
	c.rateLimiter.Wait()

	// Using v2 API which is more efficient
	apiURL := fmt.Sprintf("%s/skip-times/%d/%d?types=op&types=ed", aniskipBaseURL, malID, episodeNumber)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request with retries
	var resp *http.Response
	var lastErr error
	success := false

	for i := 0; i <= c.maxRetries && !success; i++ {
		resp, err = c.client.Do(req)
		if err != nil {
			lastErr = err
			// Backoff with exponential delay
			time.Sleep(time.Duration((i+1)*300) * time.Millisecond)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			// No skip times found, not an error
			c.saveToCache(malID, episodeNumber, []AnimeSkipTimes{})
			return []AnimeSkipTimes{}, nil
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			lastErr = fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))

			// Longer backoff for rate limits
			if resp.StatusCode == http.StatusTooManyRequests {
				time.Sleep(time.Duration((i+1)*1000) * time.Millisecond)
			} else {
				time.Sleep(time.Duration((i+1)*300) * time.Millisecond)
			}
			continue
		}

		success = true
	}

	if !success {
		return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
	}

	// Parse response
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// The response format from AniSkip API v1 as shown in the prompt example
	type aniskipResponse struct {
		Found   bool `json:"found"`
		Results []struct {
			Interval struct {
				StartTime float64 `json:"start_time"`
				EndTime   float64 `json:"end_time"`
			} `json:"interval"`
			SkipType      string  `json:"skip_type"`
			SkipID        string  `json:"skip_id"`
			EpisodeLength float64 `json:"episode_length"`
		} `json:"results"`
	}

	var skipResp aniskipResponse
	if err := json.Unmarshal(bodyBytes, &skipResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// If no results found
	if !skipResp.Found || len(skipResp.Results) == 0 {
		c.saveToCache(malID, episodeNumber, []AnimeSkipTimes{})
		return []AnimeSkipTimes{}, nil
	}

	// Convert to our skip times format
	var skipTimes []AnimeSkipTimes
	for _, result := range skipResp.Results {
		skipTime := AnimeSkipTimes{
			SkipType:      result.SkipType,
			StartTime:     result.Interval.StartTime,
			EndTime:       result.Interval.EndTime,
			EpisodeLength: result.EpisodeLength,
		}
		skipTimes = append(skipTimes, skipTime)
	}

	// Save to cache
	c.saveToCache(malID, episodeNumber, skipTimes)

	return skipTimes, nil
}

// GetSkipTimesForEpisodesBatch fetches skip times for episodes in batches
func (c *AniSkipClient) GetSkipTimesForEpisodesBatch(malID int, episodes []int) (map[int][]AnimeSkipTimes, error) {
	// If we have fewer than 3 episodes, use individual requests instead
	if len(episodes) < 3 {
		results := make(map[int][]AnimeSkipTimes)
		for _, ep := range episodes {
			skipTimes, err := c.GetSkipTimesForEpisode(malID, ep)
			if err != nil {
				return nil, err
			}
			results[ep] = skipTimes
		}
		return results, nil
	}

	// Check if all episodes are cached and return them
	allCached := true
	cachedResults := make(map[int][]AnimeSkipTimes)

	for _, ep := range episodes {
		if skipTimes, found := c.getFromCache(malID, ep); found {
			cachedResults[ep] = skipTimes
		} else {
			allCached = false
			break
		}
	}

	if allCached {
		return cachedResults, nil
	}

	// Wait for rate limiter
	c.rateLimiter.Wait()

	// Construct episode IDs parameter
	episodeParams := ""
	for i, ep := range episodes {
		if i > 0 {
			episodeParams += ","
		}
		episodeParams += fmt.Sprintf("%d", ep)
	}

	// Batch endpoint URL
	apiURL := fmt.Sprintf("%s/skip-times-batch?malId=%d&episodeIds=%s&types=op&types=ed",
		aniskipBaseURL, malID, episodeParams)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}

	// Execute request with retries
	var resp *http.Response
	var lastErr error
	success := false

	for i := 0; i <= c.maxRetries && !success; i++ {
		resp, err = c.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration((i+1)*300) * time.Millisecond)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			lastErr = fmt.Errorf("batch request failed with status %d: %s", resp.StatusCode, string(bodyBytes))

			if resp.StatusCode == http.StatusTooManyRequests {
				time.Sleep(time.Duration((i+1)*1000) * time.Millisecond)
			} else {
				time.Sleep(time.Duration((i+1)*300) * time.Millisecond)
			}
			continue
		}

		success = true
	}

	if !success {
		return nil, fmt.Errorf("batch request failed after %d retries: %w", c.maxRetries, lastErr)
	}

	// Parse response
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read batch response: %w", err)
	}

	// Batch response format
	type batchSkipTime struct {
		Interval struct {
			StartTime float64 `json:"start_time"`
			EndTime   float64 `json:"end_time"`
		} `json:"interval"`
		SkipType      string  `json:"skip_type"`
		SkipID        string  `json:"skip_id"`
		EpisodeLength float64 `json:"episode_length"`
	}

	type episodeSkipTimes struct {
		Found   bool            `json:"found"`
		Results []batchSkipTime `json:"results"`
	}

	// Map of episode number to skip times
	type batchResponse map[string]episodeSkipTimes

	var skipResp batchResponse
	if err := json.Unmarshal(bodyBytes, &skipResp); err != nil {
		return nil, fmt.Errorf("failed to decode batch response: %w", err)
	}

	results := make(map[int][]AnimeSkipTimes)

	// Process results
	for epStr, epData := range skipResp {
		var epNum int
		if _, err := fmt.Sscanf(epStr, "%d", &epNum); err != nil {
			continue // Skip if we can't parse the episode number
		}

		var skipTimes []AnimeSkipTimes

		if epData.Found {
			for _, result := range epData.Results {
				skipTimes = append(skipTimes, AnimeSkipTimes{
					SkipType:      result.SkipType,
					StartTime:     result.Interval.StartTime,
					EndTime:       result.Interval.EndTime,
					EpisodeLength: result.EpisodeLength,
				})
			}
		}

		// Save to cache
		c.saveToCache(malID, epNum, skipTimes)
		results[epNum] = skipTimes
	}

	return results, nil
}

// GetSkipTimesForEpisodes fetches skip times for multiple episodes efficiently
func (c *AniSkipClient) GetSkipTimesForEpisodes(malID int, episodeCount int, maxConcurrent int) []EpisodeSkipResult {
	startTime := time.Now()

	// If episode count is small, just use single endpoint
	if episodeCount <= 5 {
		results := []EpisodeSkipResult{}
		for i := 1; i <= episodeCount; i++ {
			skipTimes, err := c.GetSkipTimesForEpisode(malID, i)
			if err == nil && len(skipTimes) > 0 {
				results = append(results, EpisodeSkipResult{
					EpisodeNumber: i,
					SkipTimes:     skipTimes,
				})
			}
		}
		return results
	}

	// Create episode numbers slice
	allEpisodes := make([]int, episodeCount)
	for i := 0; i < episodeCount; i++ {
		allEpisodes[i] = i + 1 // 1-indexed episodes
	}

	// Batch size - we'll process episodes in batches
	const batchSize = 25
	var results []EpisodeSkipResult

	// Process in batches
	for i := 0; i < episodeCount; i += batchSize {
		end := i + batchSize
		if end > episodeCount {
			end = episodeCount
		}

		batchEpisodes := allEpisodes[i:end]
		batchResults, err := c.GetSkipTimesForEpisodesBatch(malID, batchEpisodes)
		if err != nil {
			logger.Log(fmt.Sprintf("Error fetching skip times batch %d-%d: %v", i+1, end, err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AniSkip",
			})
			continue
		}

		// Add results to the final list
		for epNum, skipTimes := range batchResults {
			if len(skipTimes) > 0 {
				results = append(results, EpisodeSkipResult{
					EpisodeNumber: epNum,
					SkipTimes:     skipTimes,
				})
			}
		}
	}

	logger.Log(fmt.Sprintf("AniSkip: Fetched skip times for %d episodes of %d in %s",
		len(results), episodeCount, time.Since(startTime)), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AniSkip",
	})

	return results
}
