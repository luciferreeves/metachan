package anime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetAnimeDetails(animeMapping *entities.AnimeMapping) (*types.Anime, error) {
	malID := animeMapping.MAL

	anime, err := getAnimeViaJikan(malID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime details: %w", err)
	}
	var anilistAnime *types.AnilistAnimeResponse
	if animeMapping.Anilist != 0 {
		anilistAnime, err = getAnimeViaAnilist(animeMapping.Anilist)
		if err != nil {
			return nil, fmt.Errorf("failed to get anime details from Anilist: %w", err)
		}
	}

	episodes, err := getAnimeEpisodesViaJikan(malID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime episodes: %w", err)
	}

	episodeData, err := generateEpisodeDataWithDescriptions(
		episodes.Data,
		anime.Data.Title,
		anime.Data.TitleEnglish,
		animeMapping.TMDB,
	)

	animeDetails := &types.Anime{
		MALID: malID,
		Titles: types.AnimeTitles{
			Romaji:   anime.Data.Title,
			English:  anime.Data.TitleEnglish,
			Japanese: anime.Data.TitleJapanese,
			Synonyms: anime.Data.TitleSynonyms,
		},
		Synopsis: anime.Data.Synopsis,
		Type:     types.AniSyncType(animeMapping.Type),
		Source:   anime.Data.Source,
		Status:   anime.Data.Status,
		Duration: anime.Data.Duration,
		Episodes: types.AnimeEpisodes{
			Total:    getEpisodeCount(anime, anilistAnime),
			Aired:    len(episodes.Data),
			Episodes: episodeData,
		},
		Mappings: types.AnimeMappings{
			AniDB:          animeMapping.AniDB,
			Anilist:        animeMapping.Anilist,
			AnimeCountdown: animeMapping.AnimeCountdown,
			AnimePlanet:    animeMapping.AnimePlanet,
			AniSearch:      animeMapping.AniSearch,
			IMDB:           animeMapping.IMDB,
			Kitsu:          animeMapping.Kitsu,
			LiveChart:      animeMapping.LiveChart,
			NotifyMoe:      animeMapping.NotifyMoe,
			Simkl:          animeMapping.Simkl,
			TMDB:           animeMapping.TMDB,
			TVDB:           animeMapping.TVDB,
		},
	}
	return animeDetails, nil
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
		WaitForJikanRequest()

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

func getAnimeViaAnilist(anilistID int) (*types.AnilistAnimeResponse, error) {
	graphQLQuery := fmt.Sprintf(`query {
									Media(id: %d) {
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
										updatedAt
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
												middle
												last
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
										nextAiringEpisode {
										id
										airingAt
										timeUntilAiring
										episode
										}
										airingSchedule {
										nodes {
											id
											episode
											airingAt
											timeUntilAiring
										}
										}
										trends {
										nodes {
											date
											trending
											popularity
											inProgress
										}
										}
										externalLinks {
										id
										url
										site
										}
										streamingEpisodes {
										title
										thumbnail
										url
										site
										}
										rankings {
										id
										rank
										type
										format
										year
										season
										allTime
										context
										}
										stats {
										scoreDistribution {
											score
											amount
										}
										statusDistribution {
											status
											amount
										}
										}
										siteUrl
									}
								}`, anilistID)

	// Remove debug print that can cause issues with large queries
	// fmt.Printf("GraphQL Query: %s\n", graphQLQuery)

	apiURL := "https://graphql.anilist.co"

	// Escape quotes in the query to make valid JSON
	escapedQuery := strings.Replace(graphQLQuery, `"`, `\"`, -1)
	escapedQuery = strings.Replace(escapedQuery, "\n", " ", -1)
	escapedQuery = strings.Replace(escapedQuery, "\t", "", -1)

	// Create the JSON payload
	jsonData := []byte(fmt.Sprintf(`{"query": "%s"}`, escapedQuery))

	// Create a request with the proper body
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get anime data: %s - %s", resp.Status, string(bodyBytes))
	}

	var anilistResponse types.AnilistAnimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&anilistResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if anilistResponse.Data.Media.ID == 0 {
		return nil, fmt.Errorf("no data found for Anilist ID %d", anilistID)
	}
	return &anilistResponse, nil
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
			WaitForJikanRequest()

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

func getEpisodeCount(malAnime *types.JikanAnimeResponse, anilistAnime *types.AnilistAnimeResponse) int {
	streamingScheduleLength := len(anilistAnime.Data.Media.AiringSchedule.Nodes)
	episodes := max(malAnime.Data.Episodes, anilistAnime.Data.Media.Episodes)
	episodes = max(episodes, streamingScheduleLength)

	return episodes
}

func generateEpisodeData(episodes []types.JikanAnimeEpisode) ([]types.AnimeSingleEpisode, error) {
	var AnimeEpisodes []types.AnimeSingleEpisode

	for _, episode := range episodes {
		AnimeEpisodes = append(AnimeEpisodes, types.AnimeSingleEpisode{
			Titles: types.EpisodeTitles{
				English:  episode.Title,
				Japanese: episode.TitleJapanese,
				Romaji:   episode.TitleRomaji,
			},
			Aired:       episode.Aired,
			Score:       episode.Score,
			Filler:      episode.Filler,
			Recap:       episode.Recap,
			ForumURL:    episode.ForumURL,
			URL:         episode.URL,
			Description: "No description available",
		})
	}
	return AnimeEpisodes, nil
}

func generateEpisodeDataWithDescriptions(episodes []types.JikanAnimeEpisode, title string, alternativeTitle string, tmdbID int) ([]types.AnimeSingleEpisode, error) {
	// First create basic episode data
	basicEpisodes, err := generateEpisodeData(episodes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate basic episode data: %w", err)
	}

	// Then enrich with descriptions - this won't fail, just return original episodes if there's an issue
	enrichedEpisodes := AttachEpisodeDescriptions(title, basicEpisodes, alternativeTitle, tmdbID)
	return enrichedEpisodes, nil
}
