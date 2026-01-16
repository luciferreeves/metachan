package jikan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"metachan/utils/ratelimit"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

var (
	// Global Jikan rate limiters
	jikanPerSecLimiter = ratelimit.NewRateLimiter(3, time.Second)
	jikanPerMinLimiter = ratelimit.NewRateLimiter(60, time.Minute)
	jikanLimiter       = ratelimit.NewMultiLimiter(jikanPerSecLimiter, jikanPerMinLimiter)
)

// JikanClient provides methods to interact with the Jikan API
type JikanClient struct {
	client      *http.Client
	maxRetries  int
	baseBackoff time.Duration
}

// NewJikanClient creates a new Jikan API client
func NewJikanClient() *JikanClient {
	return &JikanClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		maxRetries:  3,
		baseBackoff: 1 * time.Second,
	}
}

// WaitForRateLimit waits until a request can be made according to rate limiting rules
func (c *JikanClient) WaitForRateLimit() {
	jikanLimiter.Wait()
}

// makeRequest makes an HTTP request with retries and proper error handling
func (c *JikanClient) makeRequest(ctx context.Context, url string) ([]byte, error) {
	var bodyBytes []byte
	var statusCode int

	retries := 0
	for retries <= c.maxRetries {
		// Wait for rate limiter before attempting request
		c.WaitForRateLimit()

		// Create the request with timeout context
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Execute the request
		resp, err := c.client.Do(req)
		if err != nil {
			if retries < c.maxRetries {
				retries++
				backoffTime := time.Duration(float64(c.baseBackoff) * math.Pow(2, float64(retries-1)))
				time.Sleep(backoffTime)
				continue
			}
			return nil, fmt.Errorf("failed to execute request after %d retries: %w", c.maxRetries, err)
		}
		defer resp.Body.Close()

		statusCode = resp.StatusCode

		// Handle rate limiting with exponential backoff
		if statusCode == http.StatusTooManyRequests {
			if retries < c.maxRetries {
				retries++
				backoffTime := time.Duration(float64(c.baseBackoff) * math.Pow(1.5, float64(retries-1)))

				// Respect Retry-After header if available
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, err := strconv.Atoi(retryAfter); err == nil {
						backoffTime = time.Duration(seconds) * time.Second
					}
				}

				time.Sleep(backoffTime)
				continue
			}
			return nil, fmt.Errorf("rate limited after %d retries", c.maxRetries)
		} else if statusCode != http.StatusOK {
			if retries < c.maxRetries {
				retries++
				backoffTime := time.Duration(float64(c.baseBackoff) * math.Pow(2, float64(retries-1)))
				time.Sleep(backoffTime)
				continue
			}
			return nil, fmt.Errorf("request failed with status: %d", statusCode)
		}

		// Limit response body size to prevent memory issues
		bodyBytes, err = io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
		if err != nil {
			if retries < c.maxRetries {
				retries++
				backoffTime := time.Duration(float64(c.baseBackoff) * math.Pow(2, float64(retries-1)))
				time.Sleep(backoffTime)
				continue
			}
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		// Success, break the retry loop
		return bodyBytes, nil
	}

	return nil, fmt.Errorf("exhausted all retries with status code: %d", statusCode)
}

// GetAnime fetches basic anime information by MAL ID
func (c *JikanClient) GetAnime(malID int) (*JikanAnimeResponse, error) {
	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d", malID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bodyBytes, err := c.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime data: %w", err)
	}

	var animeResponse JikanAnimeResponse
	if err := json.Unmarshal(bodyBytes, &animeResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if animeResponse.Data.MALID == 0 {
		return nil, fmt.Errorf("no data found for MAL ID %d", malID)
	}

	return &animeResponse, nil
}

// GetFullAnime fetches detailed anime information by MAL ID
func (c *JikanClient) GetFullAnime(malID int) (*JikanAnimeResponse, error) {
	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/full", malID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bodyBytes, err := c.makeRequest(ctx, apiURL)
	if err != nil {
		// Fallback to curl if HTTP client fails
		var curlErr error
		bodyBytes, curlErr = c.makeRequestWithCurl(apiURL)
		if curlErr != nil {
			return nil, fmt.Errorf("failed to get anime full data via HTTP (%w) and curl (%v)", err, curlErr)
		}
	}

	var animeResponse JikanAnimeResponse
	if err := json.Unmarshal(bodyBytes, &animeResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if animeResponse.Data.MALID == 0 {
		return nil, fmt.Errorf("no data found for MAL ID %d", malID)
	}

	return &animeResponse, nil
}

// makeRequestWithCurl uses curl as a fallback when Go HTTP client fails
func (c *JikanClient) makeRequestWithCurl(url string) ([]byte, error) {
	c.WaitForRateLimit()

	cmd := exec.Command("curl", "-s", "-H", "Accept: application/json", url)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("curl command failed: %w", err)
	}

	return output, nil
}

// GetAnimeEpisodes fetches all episodes for an anime by MAL ID
func (c *JikanClient) GetAnimeEpisodes(malID int) (*JikanAnimeEpisodeResponse, error) {
	result := JikanAnimeEpisodeResponse{
		Data: []JikanAnimeEpisode{},
	}

	maxPages := 25 // Safety limit to avoid excessive requests
	page := 1
	maxAttempts := 15 // Maximum number of attempts across all pages
	totalAttempts := 0

	for page <= maxPages && totalAttempts < maxAttempts {
		apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/episodes?page=%d", malID, page)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

		totalAttempts++

		bodyBytes, err := c.makeRequest(ctx, apiURL)
		cancel()

		if err != nil {
			// If we have some episodes already, return them rather than failing
			if len(result.Data) > 0 {
				result.Pagination.HasNextPage = false
				break
			}
			return nil, fmt.Errorf("failed to get anime episodes page %d: %w", page, err)
		}

		var pageResponse JikanAnimeEpisodeResponse
		if err := json.Unmarshal(bodyBytes, &pageResponse); err != nil {
			// Return what we have if we got some pages successfully
			if len(result.Data) > 0 {
				result.Pagination.HasNextPage = false
				break
			}
			return nil, fmt.Errorf("failed to decode episodes response: %w", err)
		}

		// Append episodes from this page
		result.Data = append(result.Data, pageResponse.Data...)
		result.Pagination = pageResponse.Pagination

		// Check if we need to fetch more pages
		if !pageResponse.Pagination.HasNextPage {
			break
		}

		page++
	}

	return &result, nil
}

// GetAnimeCharacters fetches all characters for an anime by MAL ID
func (c *JikanClient) GetAnimeCharacters(malID int) (*JikanAnimeCharacterResponse, error) {
	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/characters", malID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bodyBytes, err := c.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime characters: %w", err)
	}

	var characterResponse JikanAnimeCharacterResponse
	if err := json.Unmarshal(bodyBytes, &characterResponse); err != nil {
		return nil, fmt.Errorf("failed to decode characters response: %w", err)
	}

	return &characterResponse, nil
}
