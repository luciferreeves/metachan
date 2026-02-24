package jikan

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
	jikanAPIBaseURL = "https://api.jikan.moe/v4"
	rateLimitPerSec = 3
	rateLimitPerMin = 60
	contextTimeout  = 60 * time.Second
	timeout         = 15 * time.Second
	maxRetries      = 3
	backoffDuration = 1 * time.Second
)

var (
	rateLimiter = ratelimit.NewMultiLimiter(
		ratelimit.NewRateLimiter(rateLimitPerSec, time.Second),
		ratelimit.NewRateLimiter(rateLimitPerMin, time.Minute),
	)
	clientInstance = &client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		backoff:    backoffDuration,
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

	logger.Debugf("JikanClient", "%s for %s (attempt %d/%d)", reason, url, *retries, c.maxRetries)
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
			logger.Errorf("JikanClient", "Failed to create request: %v", err)
			return nil, errors.New("failed to create request to Jikan API")
		}

		response, err = c.httpClient.Do(request)
		if err != nil {
			if !c.handleRetry(&retries, url, fmt.Sprintf("Request failed: %v", err), 0) {
				logger.Errorf("JikanClient", "All retries exhausted for request to %s: %v", url, err)
				return nil, errors.New("failed to make request to Jikan API after max retries")
			}
			continue
		}

		defer response.Body.Close()

		switch response.StatusCode {
		case http.StatusTooManyRequests:
			retryAfter := c.getRetryAfterDuration(response)
			if !c.handleRetry(&retries, url, "Rate limited", retryAfter) {
				logger.Errorf("JikanClient", "All retries exhausted for request to %s", url)
				return nil, errors.New("failed to make request to Jikan API after max retries")
			}
		case http.StatusOK:
			bytes, err := io.ReadAll(response.Body)

			if err != nil {
				logger.Errorf("JikanClient", "Failed to read response body from %s: %v", url, err)
				return nil, errors.New("failed to read response from Jikan API")
			}

			return bytes, nil
		case http.StatusNotFound:
			logger.Warnf("JikanClient", "Resource not found: %s", url)
			return nil, errors.New("resource not found")
		default:
			if response.StatusCode >= 400 && response.StatusCode < 500 {
				logger.Warnf("JikanClient", "Client error %d for %s", response.StatusCode, url)
				return nil, fmt.Errorf("client error: status %d", response.StatusCode)
			}

			retries++
			backoffDuration := c.getBackOffDuration(retries)

			logger.Warnf("JikanClient", "Request to %s returned status %d (attempt %d/%d)", url, response.StatusCode, retries, c.maxRetries)

			time.Sleep(backoffDuration)
		}
	}

	logger.Errorf("JikanClient", "All retries exhausted for request to %s", url)
	return nil, errors.New("failed to make request to Jikan API after max retries")
}

func GetAnimeByMALID(id int) (*types.JikanAnimeResponse, error) {
	url := fmt.Sprintf("%s/anime/%d/full", jikanAPIBaseURL, id)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("JikanClient", "GetAnimeByMALID failed for ID %d: %v", id, err)
		return nil, errors.New("failed to fetch anime data from Jikan API")
	}

	var response types.JikanAnimeResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("JikanClient", "Failed to unmarshal response for ID %d: %v", id, err)
		return nil, errors.New("failed to parse anime data from Jikan API")
	}

	return &response, nil
}

func GetAnimeEpisodesByMALID(id int) (*types.JikanAnimeEpisodeResponse, error) {
	url := fmt.Sprintf("%s/anime/%d/episodes", jikanAPIBaseURL, id)

	page := 1
	hasNextPage := true

	response := &types.JikanAnimeEpisodeResponse{
		Pagination: types.JikanGenericPaginationEntity{},
		Data:       []types.JikanAnimeSingleEpisode{},
	}

	for hasNextPage {
		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
		pageURL := fmt.Sprintf("%s?page=%d", url, page)

		bytes, err := clientInstance.makeRequest(ctx, pageURL)

		cancel()

		if err != nil {
			logger.Errorf("JikanClient", "GetAnimeEpisodesByMALID failed for ID %d on page %d: %v", id, page, err)
			return nil, errors.New("failed to fetch anime episodes from Jikan API")
		}

		var pageResponse types.JikanAnimeEpisodeResponse

		if err := json.Unmarshal(bytes, &pageResponse); err != nil {
			logger.Errorf("JikanClient", "Failed to unmarshal episodes response for ID %d on page %d: %v", id, page, err)
			return nil, errors.New("failed to parse anime episodes from Jikan API")
		}

		if response.Pagination.LastVisiblePage == 0 {
			response.Pagination = pageResponse.Pagination
		}

		response.Data = append(response.Data, pageResponse.Data...)
		hasNextPage = pageResponse.Pagination.HasNextPage
		page++
	}

	return response, nil
}

func GetAnimeCharactersByMALID(id int) (*types.JikanAnimeCharacterResponse, error) {
	url := fmt.Sprintf("%s/anime/%d/characters", jikanAPIBaseURL, id)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("JikanClient", "GetAnimeCharactersByMALID failed for ID %d: %v", id, err)
		return nil, errors.New("failed to fetch anime characters from Jikan API")
	}

	var response types.JikanAnimeCharacterResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("JikanClient", "Failed to unmarshal characters response for ID %d: %v", id, err)
		return nil, errors.New("failed to parse anime characters from Jikan API")
	}

	return &response, nil
}

func GetAnimeGenres() (*types.JikanGenresResponse, error) {
	url := fmt.Sprintf("%s/genres/anime", jikanAPIBaseURL)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("JikanClient", "GetAnimeGenres failed: %v", err)
		return nil, errors.New("failed to fetch anime genres from Jikan API")
	}

	var response types.JikanGenresResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("JikanClient", "Failed to unmarshal genres response: %v", err)
		return nil, errors.New("failed to parse anime genres from Jikan API")
	}

	return &response, nil
}

func GetAnimeByGenre(genreID int, page int, limit int) (*types.JikanAnimeSearchResponse, error) {
	url := fmt.Sprintf("%s/anime?genres=%d&page=%d&limit=%d", jikanAPIBaseURL, genreID, page, limit)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("JikanClient", "GetAnimeByGenre failed for genre %d: %v", genreID, err)
		return nil, errors.New("failed to fetch anime by genre from Jikan API")
	}

	var response types.JikanAnimeSearchResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("JikanClient", "Failed to unmarshal anime by genre response for genre %d: %v", genreID, err)
		return nil, errors.New("failed to parse anime by genre from Jikan API")
	}

	return &response, nil
}

func GetAnimeProducers() (*types.JikanProducersResponse, error) {
	url := fmt.Sprintf("%s/producers", jikanAPIBaseURL)
	page := 1
	hasNextPage := true

	response := &types.JikanProducersResponse{
		Data:       []types.JikanSingleProducer{},
		Pagination: types.JikanGenericPaginationEntity{},
	}

	for hasNextPage {
		pageURL := fmt.Sprintf("%s?page=%d", url, page)

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
		bytes, err := clientInstance.makeRequest(ctx, pageURL)
		cancel()
		if err != nil {
			logger.Errorf("JikanClient", "GetAnimeProducers failed on page %d: %v", page, err)
			return nil, errors.New("failed to fetch anime producers from Jikan API")
		}

		var pageResponse types.JikanProducersResponse
		if err := json.Unmarshal(bytes, &pageResponse); err != nil {
			logger.Errorf("JikanClient", "Failed to unmarshal producers response on page %d: %v", page, err)
			return nil, errors.New("failed to parse anime producers from Jikan API")
		}

		if response.Pagination.LastVisiblePage == 0 {
			response.Pagination = pageResponse.Pagination
		}

		logger.Infof("JikanClient", "Fetched page (%d/%d) - %d producers", page, response.Pagination.LastVisiblePage, len(pageResponse.Data))

		response.Data = append(response.Data, pageResponse.Data...)
		hasNextPage = pageResponse.Pagination.HasNextPage
		page++
	}

	return response, nil
}

func GetProducerByID(producerID int) (*types.JikanSingleProducerResponse, error) {
	url := fmt.Sprintf("%s/producers/%d/full", jikanAPIBaseURL, producerID)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

	defer cancel()
	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("JikanClient", "GetProducerByID failed for ID %d: %v", producerID, err)
		return nil, errors.New("failed to fetch producer data from Jikan API")
	}

	var response types.JikanSingleProducerResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("JikanClient", "Failed to unmarshal producer response for ID %d: %v", producerID, err)
		return nil, errors.New("failed to parse producer data from Jikan API")
	}
	return &response, nil
}

func GetCharacterByMALID(id int) (*types.JikanCharacterFullResponse, error) {
	url := fmt.Sprintf("%s/characters/%d/full", jikanAPIBaseURL, id)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("JikanClient", "GetCharacterByMALID failed for ID %d: %v", id, err)
		return nil, errors.New("failed to fetch character data from Jikan API")
	}

	var response types.JikanCharacterFullResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("JikanClient", "Failed to unmarshal character response for ID %d: %v", id, err)
		return nil, errors.New("failed to parse character data from Jikan API")
	}
	return &response, nil
}

func GetPersonByMALID(id int) (*types.JikanPersonFullResponse, error) {
	url := fmt.Sprintf("%s/people/%d/full", jikanAPIBaseURL, id)
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, url)
	if err != nil {
		logger.Errorf("JikanClient", "GetPersonByMALID failed for ID %d: %v", id, err)
		return nil, errors.New("failed to fetch person data from Jikan API")
	}

	var response types.JikanPersonFullResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("JikanClient", "Failed to unmarshal person response for ID %d: %v", id, err)
		return nil, errors.New("failed to parse person data from Jikan API")
	}
	return &response, nil
}
