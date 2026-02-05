package malsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"strconv"
	"time"
)

const (
	malsyncAPIBaseURL = "https://api.malsync.moe/mal"
	contextTimeout    = 10 * time.Second
)

var (
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

	logger.Warnf("MalsyncClient", "%s for %s (attempt %d/%d)", reason, url, *retries, c.maxRetries)
	time.Sleep(backoffDuration)
	return true
}

func (c *client) makeRequest(ctx context.Context, url string) ([]byte, error) {
	var response *http.Response
	var retries int

	for retries < c.maxRetries {
		request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			logger.Errorf("MalsyncClient", "Failed to create request: %v", err)
			return nil, errors.New("failed to create request to Malsync API")
		}

		request.Header.Set("Accept", "application/json")

		response, err = c.httpClient.Do(request)
		if err != nil {
			if !c.handleRetry(&retries, url, fmt.Sprintf("Request failed: %v", err), 0) {
				logger.Errorf("MalsyncClient", "All retries exhausted for request to %s: %v", url, err)
				return nil, errors.New("failed to make request to Malsync API after max retries")
			}
			continue
		}

		defer response.Body.Close()

		switch response.StatusCode {
		case http.StatusNotFound:
			// Not found is not an error, return nil
			return nil, nil
		case http.StatusTooManyRequests:
			retryAfter := c.getRetryAfterDuration(response)
			if !c.handleRetry(&retries, url, "Rate limited", retryAfter) {
				logger.Errorf("MalsyncClient", "All retries exhausted for request to %s", url)
				return nil, errors.New("failed to make request to Malsync API after max retries")
			}
		case http.StatusOK:
			bytes, err := io.ReadAll(response.Body)

			if err != nil {
				logger.Errorf("MalsyncClient", "Failed to read response body from %s: %v", url, err)
				return nil, errors.New("failed to read response from Malsync API")
			}

			return bytes, nil
		default:
			retries++
			backoffDuration := c.getBackOffDuration(retries)

			logger.Warnf("MalsyncClient", "Request to %s returned status %d (attempt %d/%d)", url, response.StatusCode, retries, c.maxRetries)

			time.Sleep(backoffDuration)
		}
	}

	logger.Errorf("MalsyncClient", "All retries exhausted for request to %s", url)
	return nil, errors.New("failed to make request to Malsync API after max retries")
}

func GetAnimeByMALID(malID int) (*types.MalsyncAnimeResponse, error) {
	url := fmt.Sprintf("%s/anime/%d", malsyncAPIBaseURL, malID)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("MalsyncClient", "GetAnimeByMALID failed for MAL ID %d: %v", malID, err)
		return nil, errors.New("failed to fetch anime data from Malsync API")
	}

	// Handle 404 case where makeRequest returns nil, nil
	if bytes == nil {
		return nil, nil
	}

	var response types.MalsyncAnimeResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("MalsyncClient", "Failed to unmarshal response for MAL ID %d: %v", malID, err)
		return nil, errors.New("failed to parse anime data from Malsync API")
	}

	if response.ID == 0 {
		logger.Errorf("MalsyncClient", "Received empty response for MAL ID %d", malID)
		return nil, fmt.Errorf("received empty response for MAL ID %d", malID)
	}

	return &response, nil
}
