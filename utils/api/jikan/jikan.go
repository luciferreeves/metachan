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
)

var (
	rateLimiter = ratelimit.NewMultiLimiter(
		ratelimit.NewRateLimiter(rateLimitPerSec, time.Second),
		ratelimit.NewRateLimiter(rateLimitPerMin, time.Minute),
	)
	clientInstance = &client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
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

	logger.Warnf("JikanClient", "%s for %s (attempt %d/%d)", reason, url, *retries, c.maxRetries)
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
		default:
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
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	page := 1
	hasNextPage := true

	response := &types.JikanProducersResponse{
		Data:       []types.JikanSingleProducer{},
		Pagination: types.JikanGenericPaginationEntity{},
	}

	defer cancel()

	for hasNextPage {
		pageURL := fmt.Sprintf("%s?page=%d", url, page)

		bytes, err := clientInstance.makeRequest(ctx, pageURL)
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

// var (
// 	// Global Jikan rate limiters
// 	jikanPerSecLimiter = ratelimit.NewRateLimiter(3, time.Second)
// 	jikanPerMinLimiter = ratelimit.NewRateLimiter(60, time.Minute)
// 	jikanLimiter       = ratelimit.NewMultiLimiter(jikanPerSecLimiter, jikanPerMinLimiter)
// )

// // JikanClient provides methods to interact with the Jikan API
// type JikanClient struct {
// 	client      *http.Client
// 	maxRetries  int
// 	baseBackoff time.Duration
// }

// // NewJikanClient creates a new Jikan API client
// func NewJikanClient() *JikanClient {
// 	return &JikanClient{
// 		client: &http.Client{
// 			Timeout: 15 * time.Second,
// 		},
// 		maxRetries:  3,
// 		baseBackoff: 1 * time.Second,
// 	}
// }

// // WaitForRateLimit waits until a request can be made according to rate limiting rules
// func (c *JikanClient) WaitForRateLimit() {
// 	jikanLimiter.Wait()
// }

// // makeRequest makes an HTTP request with retries and proper error handling
// func (c *JikanClient) makeRequest(ctx context.Context, url string) ([]byte, error) {
// 	var bodyBytes []byte
// 	var statusCode int

// 	retries := 0
// 	for retries <= c.maxRetries {
// 		// Wait for rate limiter before attempting request
// 		c.WaitForRateLimit()

// 		// Create the request with timeout context
// 		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to create request: %w", err)
// 		}

// 		// Execute the request
// 		resp, err := c.client.Do(req)
// 		if err != nil {
// 			if retries < c.maxRetries {
// 				retries++
// 				backoffTime := time.Duration(float64(c.baseBackoff) * math.Pow(2, float64(retries-1)))
// 				time.Sleep(backoffTime)
// 				continue
// 			}
// 			return nil, fmt.Errorf("failed to execute request after %d retries: %w", c.maxRetries, err)
// 		}
// 		defer resp.Body.Close()

// 		statusCode = resp.StatusCode

// 		// Handle rate limiting with exponential backoff
// 		if statusCode == http.StatusTooManyRequests {
// 			if retries < c.maxRetries {
// 				retries++
// 				backoffTime := time.Duration(float64(c.baseBackoff) * math.Pow(1.5, float64(retries-1)))

// 				// Respect Retry-After header if available
// 				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
// 					if seconds, err := strconv.Atoi(retryAfter); err == nil {
// 						backoffTime = time.Duration(seconds) * time.Second
// 					}
// 				}

// 				time.Sleep(backoffTime)
// 				continue
// 			}
// 			return nil, fmt.Errorf("rate limited after %d retries", c.maxRetries)
// 		} else if statusCode != http.StatusOK {
// 			if retries < c.maxRetries {
// 				retries++
// 				backoffTime := time.Duration(float64(c.baseBackoff) * math.Pow(2, float64(retries-1)))
// 				time.Sleep(backoffTime)
// 				continue
// 			}
// 			return nil, fmt.Errorf("request failed with status: %d", statusCode)
// 		}

// 		// Limit response body size to prevent memory issues
// 		bodyBytes, err = io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
// 		if err != nil {
// 			if retries < c.maxRetries {
// 				retries++
// 				backoffTime := time.Duration(float64(c.baseBackoff) * math.Pow(2, float64(retries-1)))
// 				time.Sleep(backoffTime)
// 				continue
// 			}
// 			return nil, fmt.Errorf("failed to read response body: %w", err)
// 		}

// 		// Success, break the retry loop
// 		return bodyBytes, nil
// 	}

// 	return nil, fmt.Errorf("exhausted all retries with status code: %d", statusCode)
// }

// // GetAnime fetches basic anime information by MAL ID
// func (c *JikanClient) GetAnime(malID int) (*JikanAnimeResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d", malID)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get anime data: %w", err)
// 	}

// 	var animeResponse JikanAnimeResponse
// 	if err := json.Unmarshal(bodyBytes, &animeResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode response: %w", err)
// 	}

// 	if animeResponse.Data.MALID == 0 {
// 		return nil, fmt.Errorf("no data found for MAL ID %d", malID)
// 	}

// 	return &animeResponse, nil
// }

// // GetFullAnime fetches detailed anime information by MAL ID
// func (c *JikanClient) GetFullAnime(malID int) (*JikanAnimeResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/full", malID)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		// Fallback to curl if HTTP client fails
// 		var curlErr error
// 		bodyBytes, curlErr = c.makeRequestWithCurl(apiURL)
// 		if curlErr != nil {
// 			return nil, fmt.Errorf("failed to get anime full data via HTTP (%w) and curl (%v)", err, curlErr)
// 		}
// 	}

// 	var animeResponse JikanAnimeResponse
// 	if err := json.Unmarshal(bodyBytes, &animeResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode response: %w", err)
// 	}

// 	if animeResponse.Data.MALID == 0 {
// 		return nil, fmt.Errorf("no data found for MAL ID %d", malID)
// 	}

// 	return &animeResponse, nil
// }

// // makeRequestWithCurl uses curl as a fallback when Go HTTP client fails
// func (c *JikanClient) makeRequestWithCurl(url string) ([]byte, error) {
// 	c.WaitForRateLimit()

// 	cmd := exec.Command("curl", "-s", "-H", "Accept: application/json", url)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return nil, fmt.Errorf("curl command failed: %w", err)
// 	}

// 	return output, nil
// }

// // GetAnimeEpisodes fetches all episodes for an anime by MAL ID
// func (c *JikanClient) GetAnimeEpisodes(malID int) (*JikanAnimeEpisodeResponse, error) {
// 	result := JikanAnimeEpisodeResponse{
// 		Data: []JikanAnimeEpisode{},
// 	}

// 	maxPages := 25 // Safety limit to avoid excessive requests
// 	page := 1
// 	maxAttempts := 15 // Maximum number of attempts across all pages
// 	totalAttempts := 0

// 	for page <= maxPages && totalAttempts < maxAttempts {
// 		apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/episodes?page=%d", malID, page)

// 		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

// 		totalAttempts++

// 		bodyBytes, err := c.makeRequest(ctx, apiURL)
// 		cancel()

// 		if err != nil {
// 			// If we have some episodes already, return them rather than failing
// 			if len(result.Data) > 0 {
// 				result.Pagination.HasNextPage = false
// 				break
// 			}
// 			return nil, fmt.Errorf("failed to get anime episodes page %d: %w", page, err)
// 		}

// 		var pageResponse JikanAnimeEpisodeResponse
// 		if err := json.Unmarshal(bodyBytes, &pageResponse); err != nil {
// 			// Return what we have if we got some pages successfully
// 			if len(result.Data) > 0 {
// 				result.Pagination.HasNextPage = false
// 				break
// 			}
// 			return nil, fmt.Errorf("failed to decode episodes response: %w", err)
// 		}

// 		// Append episodes from this page
// 		result.Data = append(result.Data, pageResponse.Data...)
// 		result.Pagination = pageResponse.Pagination

// 		// Check if we need to fetch more pages
// 		if !pageResponse.Pagination.HasNextPage {
// 			break
// 		}

// 		page++
// 	}

// 	return &result, nil
// }

// // GetAnimeCharacters fetches all characters for an anime by MAL ID
// func (c *JikanClient) GetAnimeCharacters(malID int) (*JikanAnimeCharacterResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/characters", malID)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get anime characters: %w", err)
// 	}

// 	var characterResponse JikanAnimeCharacterResponse
// 	if err := json.Unmarshal(bodyBytes, &characterResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode characters response: %w", err)
// 	}

// 	return &characterResponse, nil
// }

// // GetAnimeGenres fetches all anime genres from MAL
// func (c *JikanClient) GetAnimeGenres() (*JikanGenresResponse, error) {
// 	apiURL := "https://api.jikan.moe/v4/genres/anime"

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get anime genres: %w", err)
// 	}

// 	var genresResponse JikanGenresResponse
// 	if err := json.Unmarshal(bodyBytes, &genresResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode genres response: %w", err)
// 	}

// 	return &genresResponse, nil
// }

// // GetAnimeByGenre fetches paginated anime list for a specific genre
// func (c *JikanClient) GetAnimeByGenre(genreID int, page int, limit int) (*JikanAnimeListResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime?genres=%d&page=%d&limit=%d", genreID, page, limit)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get anime by genre: %w", err)
// 	}

// 	var listResponse JikanAnimeListResponse
// 	if err := json.Unmarshal(bodyBytes, &listResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode anime list response: %w", err)
// 	}

// 	return &listResponse, nil
// }

// // GetAnimeProducers fetches all producers from Jikan API (paginated)
// func (c *JikanClient) GetAnimeProducers(page int) (*JikanProducersFullResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/producers?page=%d", page)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get producers: %w", err)
// 	}

// 	var response JikanProducersFullResponse
// 	if err := json.Unmarshal(bodyBytes, &response); err != nil {
// 		return nil, fmt.Errorf("failed to decode producers response: %w", err)
// 	}

// 	return &response, nil
// }

// // GetProducerExternal fetches external URLs for a specific producer
// func (c *JikanClient) GetProducerExternal(producerID int) (*JikanProducerExternalResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/producers/%d/external", producerID)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get producer external: %w", err)
// 	}

// 	var response JikanProducerExternalResponse
// 	if err := json.Unmarshal(bodyBytes, &response); err != nil {
// 		return nil, fmt.Errorf("failed to decode producer external response: %w", err)
// 	}

// 	return &response, nil
// }

// // GetAnimeByProducer fetches paginated anime list by producer ID
// func (c *JikanClient) GetAnimeByProducer(producerID int, page int, limit int) (*JikanAnimeListResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime?producers=%d&page=%d&limit=%d", producerID, page, limit)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get anime by producer: %w", err)
// 	}

// 	var listResponse JikanAnimeListResponse
// 	if err := json.Unmarshal(bodyBytes, &listResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode anime list response: %w", err)
// 	}

// 	return &listResponse, nil
// }

// // GetAnimeByStudio fetches paginated anime list by studio ID (uses producers endpoint)
// func (c *JikanClient) GetAnimeByStudio(studioID int, page int, limit int) (*JikanAnimeListResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime?producers=%d&page=%d&limit=%d", studioID, page, limit)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get anime by studio: %w", err)
// 	}

// 	var listResponse JikanAnimeListResponse
// 	if err := json.Unmarshal(bodyBytes, &listResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode anime list response: %w", err)
// 	}

// 	return &listResponse, nil
// }

// // GetAnimeByLicensor fetches paginated anime list by licensor ID (uses producers endpoint)
// func (c *JikanClient) GetAnimeByLicensor(licensorID int, page int, limit int) (*JikanAnimeListResponse, error) {
// 	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime?producers=%d&page=%d&limit=%d", licensorID, page, limit)

// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	bodyBytes, err := c.makeRequest(ctx, apiURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get anime by licensor: %w", err)
// 	}

// 	var listResponse JikanAnimeListResponse
// 	if err := json.Unmarshal(bodyBytes, &listResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode anime list response: %w", err)
// 	}

// 	return &listResponse, nil
// }
