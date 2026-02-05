package aniskip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"metachan/types"
	"metachan/utils/logger"
	"metachan/utils/ratelimit"
	"net/http"
	"strconv"
	"time"
)

const (
	aniskipBaseURL    = "https://api.aniskip.com/v2"
	rateLimitPerSec   = 10
	rateLimitPer10Sec = 100
	contextTimeout    = 10 * time.Second
)

var (
	rateLimiter = ratelimit.NewMultiLimiter(
		ratelimit.NewRateLimiter(rateLimitPerSec, time.Second),
		ratelimit.NewRateLimiter(rateLimitPer10Sec, 10*time.Second),
	)
	clientInstance = &client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
		backoff:    1 * time.Second,
	}
)

func (c *client) getBackOffDuration(attempt int) time.Duration {
	return time.Duration(float64(c.backoff) * math.Pow(2, float64(attempt-1)))
}

func (c *client) getRetryAfterDuration(resp *http.Response) time.Duration {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return c.backoff
}

func (c *client) handleRetry(retries *int, url string, reason string, retryAfter time.Duration) bool {
	*retries++
	if *retries >= c.maxRetries {
		return false
	}

	backoffDuration := c.getBackOffDuration(*retries)
	if retryAfter > backoffDuration {
		backoffDuration = retryAfter
	}

	logger.Warnf("AniskipClient", "%s for %s (attempt %d/%d)", reason, url, *retries, c.maxRetries)
	time.Sleep(backoffDuration)
	return true
}

func (c *client) makeRequest(ctx context.Context, url string) ([]byte, error) {
	var response *http.Response
	var retries int

	for retries < c.maxRetries {
		rateLimiter.Wait()

		request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			logger.Errorf("AniskipClient", "Failed to create request: %v", err)
			return nil, errors.New("failed to create request to Aniskip API")
		}

		response, err = c.httpClient.Do(request)
		if err != nil {
			if !c.handleRetry(&retries, url, fmt.Sprintf("Request failed: %v", err), 0) {
				logger.Errorf("AniskipClient", "All retries exhausted for request to %s: %v", url, err)
				return nil, errors.New("failed to make request to Aniskip API after max retries")
			}
			continue
		}

		defer response.Body.Close()

		switch response.StatusCode {
		case http.StatusNotFound:
			// No skip times found, return empty slice
			return []byte("{\"found\":false,\"results\":[]}"), nil
		case http.StatusTooManyRequests:
			retryAfter := c.getRetryAfterDuration(response)
			if !c.handleRetry(&retries, url, "Rate limited", retryAfter) {
				logger.Errorf("AniskipClient", "All retries exhausted for request to %s", url)
				return nil, errors.New("failed to make request to Aniskip API after max retries")
			}
		case http.StatusOK:
			bytes, err := io.ReadAll(response.Body)

			if err != nil {
				logger.Errorf("AniskipClient", "Failed to read response body from %s: %v", url, err)
				return nil, errors.New("failed to read response from Aniskip API")
			}

			return bytes, nil
		default:
			retries++
			backoffDuration := c.getBackOffDuration(retries)

			logger.Warnf("AniskipClient", "Request to %s returned status %d (attempt %d/%d)", url, response.StatusCode, retries, c.maxRetries)

			time.Sleep(backoffDuration)
		}
	}

	logger.Errorf("AniskipClient", "All retries exhausted for request to %s", url)
	return nil, errors.New("failed to make request to Aniskip API after max retries")
}

func GetSkipTimesForEpisode(malID, episodeNumber int) ([]types.AniskipResult, error) {
	url := fmt.Sprintf("%s/skip-times/%d/%d?types=op&types=ed", aniskipBaseURL, malID, episodeNumber)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("AniskipClient", "GetSkipTimesForEpisode failed for MAL ID %d, episode %d: %v", malID, episodeNumber, err)
		return nil, errors.New("failed to fetch skip times from Aniskip API")
	}

	var response types.AniskipResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("AniskipClient", "Failed to unmarshal response for MAL ID %d, episode %d: %v", malID, episodeNumber, err)
		return nil, errors.New("failed to parse skip times from Aniskip API")
	}

	return response.Results, nil
}
