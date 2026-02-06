package anilist

import (
	"bytes"
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
	anilistAPIBaseURL = "https://graphql.anilist.co"
	contextTimeout    = 60 * time.Second
	timeout           = 15 * time.Second
	maxRetries        = 3
	backoffDuration   = 1 * time.Second
	contentType       = "application/json"
	acceptHeader      = "application/json"
	userAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

var (
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

func (c *client) handleRetry(retries *int, reason string, retryAfter time.Duration) bool {
	*retries++
	if *retries >= c.maxRetries {
		return false
	}

	backoffDuration := c.getBackOffDuration(*retries)
	if retryAfter > backoffDuration {
		backoffDuration = retryAfter
	}

	logger.Warnf("AnilistClient", "%s (attempt %d/%d)", reason, *retries, c.maxRetries)
	time.Sleep(backoffDuration)
	return true
}

func (c *client) makeRequest(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	var response *http.Response
	var retries int

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		logger.Errorf("AnilistClient", "Failed to marshal request body: %v", err)
		return nil, errors.New("failed to create request to Anilist API")
	}

	for retries < c.maxRetries {
		request, err := http.NewRequestWithContext(ctx, "POST", anilistAPIBaseURL, bytes.NewBuffer(jsonData))
		if err != nil {
			logger.Errorf("AnilistClient", "Failed to create request: %v", err)
			return nil, errors.New("failed to create request to Anilist API")
		}

		request.Header.Set("Content-Type", contentType)
		request.Header.Set("Accept", acceptHeader)
		request.Header.Set("User-Agent", userAgent)

		response, err = c.httpClient.Do(request)
		if err != nil {
			if !c.handleRetry(&retries, fmt.Sprintf("Request failed: %v", err), 0) {
				logger.Errorf("AnilistClient", "All retries exhausted for request: %v", err)
				return nil, errors.New("failed to make request to Anilist API after max retries")
			}
			continue
		}

		defer response.Body.Close()

		switch response.StatusCode {
		case http.StatusTooManyRequests:
			retryAfter := c.getRetryAfterDuration(response)
			if !c.handleRetry(&retries, "Rate limited", retryAfter) {
				logger.Errorf("AnilistClient", "All retries exhausted for request")
				return nil, errors.New("failed to make request to Anilist API after max retries")
			}
		case http.StatusOK:
			bytes, err := io.ReadAll(response.Body)

			if err != nil {
				logger.Errorf("AnilistClient", "Failed to read response body: %v", err)
				return nil, errors.New("failed to read response from Anilist API")
			}

			return bytes, nil
		default:
			retries++
			backoffDuration := c.getBackOffDuration(retries)

			logger.Warnf("AnilistClient", "Request returned status %d (attempt %d/%d)", response.StatusCode, retries, c.maxRetries)

			time.Sleep(backoffDuration)
		}
	}

	logger.Errorf("AnilistClient", "All retries exhausted for request")
	return nil, errors.New("failed to make request to Anilist API after max retries")
}

func GetAnimeByAnilistID(id int) (*types.AnilistAnimeResponse, error) {
	query := `
	query($id: Int) {
		Media(id: $id, type: ANIME) {
			id
			idMal
			title {
				romaji
				english
				native
				userPreferred
			}
			type
			format
			status
			description
			startDate {
				year
				month
				day
			}
			endDate {
				year
				month
				day
			}
			season
			seasonYear
			episodes
			duration
			chapters
			volumes
			countryOfOrigin
			isLicensed
			source
			hashtag
			trailer {
				id
				site
				thumbnail
			}
			coverImage {
				extraLarge
				large
				medium
				color
			}
			bannerImage
			genres
			synonyms
			averageScore
			meanScore
			popularity
			isLocked
			trending
			favourites
			tags {
				id
				name
				description
				category
				rank
				isGeneralSpoiler
				isMediaSpoiler
				isAdult
			}
			relations {
				edges {
					id
					relationType
					node {
						id
						title {
							romaji
							english
							native
							userPreferred
						}
						format
						type
						status
						coverImage {
							extraLarge
							large
							medium
							color
						}
						bannerImage
					}
				}
			}
			characters {
				edges {
					role
					node {
						id
						name {
							first
							last
							middle
							full
							native
							userPreferred
						}
						image {
							large
							medium
						}
						description
						age
					}
				}
			}
			staff {
				edges {
					role
					node {
						id
						name {
							first
							last
							middle
							full
							native
							userPreferred
						}
						image {
							large
							medium
						}
						description
						primaryOccupations
						gender
						age
						languageV2
					}
				}
			}
			studios {
				edges {
					isMain
					node {
						id
						name
					}
				}
			}
			isAdult
			nextAiringEpisode { id airingAt episode timeUntilAiring }
			airingSchedule { nodes { id episode airingAt timeUntilAiring } }
			trends { nodes { date trending popularity inProgress } }
			externalLinks { id url site }
			streamingEpisodes { title thumbnail url site }
			rankings { id rank type format year season allTime context }
			stats {
				scoreDistribution { score amount }
				statusDistribution { status amount }
			}
			siteUrl
		}
	}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	bytes, err := clientInstance.makeRequest(ctx, query, variables)
	if err != nil {
		logger.Errorf("AnilistClient", "GetAnime failed for ID %d: %v", id, err)
		return nil, errors.New("failed to fetch anime data from Anilist API")
	}

	var response types.AnilistAnimeResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		logger.Errorf("AnilistClient", "Failed to unmarshal response for ID %d: %v", id, err)
		return nil, errors.New("failed to parse anime data from Anilist API")
	}

	if response.Data.Media.ID == 0 {
		logger.Errorf("AnilistClient", "No data found for Anilist ID %d", id)
		return nil, errors.New("no data found")
	}

	return &response, nil
}
