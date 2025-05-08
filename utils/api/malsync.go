package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"metachan/types"
	"net/http"
	"time"
)

const (
	malsyncAPIBaseURL = "https://api.malsync.moe/mal"
)

// MALSyncClient provides methods for interacting with the MALSync API
type MALSyncClient struct {
	client     *http.Client
	maxRetries int
}

// NewMALSyncClient creates a new client for the MALSync API
func NewMALSyncClient() *MALSyncClient {
	return &MALSyncClient{
		client: &http.Client{
			Timeout: 8 * time.Second, // Shorter timeout since this is a less critical API
		},
		maxRetries: 2,
	}
}

// GetAnimeByMALID fetches anime metadata from MALSync by MAL ID
func (c *MALSyncClient) GetAnimeByMALID(malID int) (*types.MALSyncAnimeResponse, error) {
	apiURL := fmt.Sprintf("%s/anime/%d", malsyncAPIBaseURL, malID)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request with retries
	var resp *http.Response
	var lastErr error
	success := false

	for i := 0; i <= c.maxRetries && !success; i++ {
		resp, err = c.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration((i+1)*300) * time.Millisecond) // Short backoff on error
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil, nil // Not found is not an error, just return nil
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			lastErr = fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
			time.Sleep(time.Duration((i+1)*300) * time.Millisecond)
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

	var malSyncResponse types.MALSyncAnimeResponse
	if err := json.Unmarshal(bodyBytes, &malSyncResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Simple validation
	if malSyncResponse.ID == 0 {
		return nil, fmt.Errorf("received empty response for MAL ID %d", malID)
	}

	return &malSyncResponse, nil
}
