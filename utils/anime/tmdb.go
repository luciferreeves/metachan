package anime

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"metachan/config"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// makeRequestWithRetries executes an HTTP request with retries for handling temporary network failures
func makeRequestWithRetries(req *http.Request, maxRetries int) (*http.Response, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter for retries
			backoffTime := time.Duration(math.Pow(1.5, float64(attempt))) * time.Second

			// Updated jitter calculation without using deprecated rand.Seed
			jitter := time.Duration(rand.Int31n(500)) * time.Millisecond

			sleepTime := backoffTime + jitter

			logger.Log(fmt.Sprintf("TMDB request retry %d/%d after %v due to: %v",
				attempt, maxRetries, sleepTime, lastErr), types.LogOptions{
				Level:  types.Debug,
				Prefix: "TMDB",
			})

			time.Sleep(sleepTime)

			// Create a fresh request to avoid any issues with reusing the same request
			newReq, err := http.NewRequest(req.Method, req.URL.String(), nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create new request for retry: %w", err)
			}

			// Copy all headers from the original request
			for key, values := range req.Header {
				for _, value := range values {
					newReq.Header.Add(key, value)
				}
			}

			// Set the new retry request as our active request
			req = newReq
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			// Check if this is a network error that might be temporary
			if strings.Contains(err.Error(), "connection reset by peer") ||
				strings.Contains(err.Error(), "EOF") ||
				strings.Contains(err.Error(), "connection refused") ||
				strings.Contains(err.Error(), "timeout") {
				// These are retryable errors
				continue
			}
			// Other errors are not retryable
			return nil, err
		}

		// If we got a server error (5xx), retry
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			lastErr = fmt.Errorf("server error: %s", resp.Status)
			resp.Body.Close() // Make sure we close the body before we retry
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// normalizeTitle cleans up the anime title for better matching with TMDB
func normalizeTitle(title string) string {
	// Handle empty titles
	if title == "" {
		return ""
	}

	// Remove common suffixes and prefixes
	normalized := title
	normalized = strings.Replace(normalized, "TV Animation", "", -1)
	normalized = strings.Replace(normalized, ": Season", "", -1)
	normalized = strings.Replace(normalized, "Season", "", -1)
	normalized = strings.Replace(normalized, "Part", "", -1)
	normalized = strings.Replace(normalized, "Cour", "", -1)

	// Handle patterns like "Dr. Stone: Stone Wars" -> "Dr. Stone"
	if colonIndex := strings.Index(normalized, ":"); colonIndex > 0 {
		normalized = normalized[:colonIndex]
	}

	// Remove parentheses and text inside them
	for {
		openParen := strings.Index(normalized, "(")
		if openParen == -1 {
			break
		}
		closeParen := strings.Index(normalized, ")")
		if closeParen == -1 || closeParen < openParen {
			break
		}
		normalized = normalized[:openParen] + normalized[closeParen+1:]
	}

	return strings.TrimSpace(normalized)
}

// searchTVShowsByTitle searches for TV shows on TMDB by title
func searchTVShowsByTitle(title string, alternativeTitle string, isAdult bool, countryPriority string) ([]types.TMDBShowResult, error) {
	if config.Config.TMDB.ReadAccessToken == "" {
		return nil, fmt.Errorf("TMDB is not initialized")
	}

	// Normalize the title
	query := normalizeTitle(title)
	if query == "" && alternativeTitle != "" {
		query = normalizeTitle(alternativeTitle)
	}

	logger.Log(fmt.Sprintf("Searching TMDB for TV show: %s", query), types.LogOptions{
		Level:  types.Debug,
		Prefix: "TMDB",
	})

	apiURL := "https://api.themoviedb.org/3/search/tv"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Config.TMDB.ReadAccessToken))
	req.Header.Add("Accept", "application/json")

	// Use our retry mechanism (3 retries)
	resp, err := makeRequestWithRetries(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to search TV shows: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search TV shows: %s", resp.Status)
	}

	// Parse response
	var searchResponse types.TMDBSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := searchResponse.Results

	// Filter results if needed
	var filteredResults []types.TMDBShowResult
	for _, show := range results {
		if (isAdult && show.Adult) || (!isAdult && !show.Adult) {
			filteredResults = append(filteredResults, show)
		}
	}

	// Sort by country priority if specified
	if countryPriority != "" && len(filteredResults) > 0 {
		var prioritizedResults []types.TMDBShowResult
		var otherResults []types.TMDBShowResult

		for _, show := range filteredResults {
			hasPriority := false
			for _, country := range show.OriginCountry {
				if country == countryPriority {
					hasPriority = true
					break
				}
			}

			if hasPriority {
				prioritizedResults = append(prioritizedResults, show)
			} else {
				otherResults = append(otherResults, show)
			}
		}

		// Combine the results with prioritized ones first
		filteredResults = append(prioritizedResults, otherResults...)
	}

	if len(filteredResults) == 0 {
		logger.Log(fmt.Sprintf("No TMDB shows found for: %s", query), types.LogOptions{
			Level:  types.Warn,
			Prefix: "TMDB",
		})
	} else {
		logger.Log(fmt.Sprintf("Found %d TMDB shows for: %s", len(filteredResults), query), types.LogOptions{
			Level:  types.Debug,
			Prefix: "TMDB",
		})
	}

	return filteredResults, nil
}

// getTVShowDetails gets details for a TV show from TMDB
func getTVShowDetails(showID int) (*types.TMDBShowDetails, error) {
	if config.Config.TMDB.ReadAccessToken == "" {
		return nil, fmt.Errorf("TMDB is not initialized")
	}

	apiURL := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d", showID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Config.TMDB.ReadAccessToken))
	req.Header.Add("Accept", "application/json")

	// Use our retry mechanism (3 retries)
	resp, err := makeRequestWithRetries(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to get TV show details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get TV show details: %s", resp.Status)
	}

	// Parse response
	var showDetails types.TMDBShowDetails
	if err := json.NewDecoder(resp.Body).Decode(&showDetails); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &showDetails, nil
}

// getSeasonDetails gets details for a TV season from TMDB
func getSeasonDetails(showID, seasonNumber int) (*types.TMDBSeasonDetails, error) {
	if config.Config.TMDB.ReadAccessToken == "" {
		return nil, fmt.Errorf("TMDB is not initialized")
	}

	apiURL := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/season/%d", showID, seasonNumber)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Config.TMDB.ReadAccessToken))
	req.Header.Add("Accept", "application/json")

	// Use our retry mechanism (3 retries)
	resp, err := makeRequestWithRetries(req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to get season details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get season details: %s", resp.Status)
	}

	// Parse response
	var seasonDetails types.TMDBSeasonDetails
	if err := json.NewDecoder(resp.Body).Decode(&seasonDetails); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &seasonDetails, nil
}

// findBestSeason finds the best matching season for an anime
func findBestSeason(shows []types.TMDBShowResult, title string, episodeCount int, airDate string) (int, int, error) {
	for _, show := range shows {
		showDetails, err := getTVShowDetails(show.ID)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to get details for show %d: %v", show.ID, err), types.LogOptions{
				Level:  types.Warn,
				Prefix: "TMDB",
			})
			continue
		}

		for _, season := range showDetails.Seasons {
			// Skip season 0 (usually specials)
			if season.SeasonNumber == 0 {
				continue
			}

			// Check if episode count matches (with some flexibility)
			episodeCountMatches := season.EpisodeCount == episodeCount ||
				(episodeCount > 0 && season.EpisodeCount >= episodeCount-2 &&
					season.EpisodeCount <= episodeCount+2)

			// Check if air dates are close
			airDateMatches := false
			if airDate != "" && season.AirDate != "" {
				// Simple year comparison
				animeYear := airDate[:4]
				seasonYear := season.AirDate[:4]
				airDateMatches = animeYear == seasonYear
			}

			// If either count or air date matches, consider it a potential match
			if episodeCountMatches || airDateMatches {
				logger.Log(fmt.Sprintf("Found matching season for \"%s\": Show ID %d, Season %d",
					title, show.ID, season.SeasonNumber), types.LogOptions{
					Level:  types.Info,
					Prefix: "TMDB",
				})
				return show.ID, season.SeasonNumber, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("could not find matching season for: %s", title)
}

// AttachEpisodeDescriptions enriches anime episodes with descriptions and thumbnails from TMDB
func AttachEpisodeDescriptions(title string, episodes []types.AnimeSingleEpisode, alternativeTitle string, tmdbID int) []types.AnimeSingleEpisode {
	if config.Config.TMDB.ReadAccessToken == "" {
		logger.Log("TMDB is not configured, skipping episode description enrichment", types.LogOptions{
			Level:  types.Warn,
			Prefix: "TMDB",
		})
		return episodes
	}

	if len(episodes) == 0 {
		return episodes
	}

	logger.Log(fmt.Sprintf("Enriching episodes for: %s", title), types.LogOptions{
		Level:  types.Info,
		Prefix: "TMDB",
	})

	var showID int
	var seasonNumber int
	var err error

	// If we have a TMDB ID, use it directly
	if tmdbID > 0 {
		showID = tmdbID

		// Try to get show details and find the best season
		showDetails, err := getTVShowDetails(showID)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to get TMDB show details for ID %d: %v", tmdbID, err), types.LogOptions{
				Level:  types.Warn,
				Prefix: "TMDB",
			})
			return episodes
		}

		// Find the best matching season - prefer the first season if we can't determine
		seasonNumber = 1
		bestMatchScore := 0

		for _, season := range showDetails.Seasons {
			if season.SeasonNumber == 0 {
				continue // Skip specials
			}

			matchScore := 0

			// Check episode count similarity
			if math.Abs(float64(season.EpisodeCount-len(episodes))) <= 2 {
				matchScore += 2
			}

			// Check air date if available
			if len(episodes) > 0 && episodes[0].Aired != "" && season.AirDate != "" {
				animeYear := episodes[0].Aired[:4]
				seasonYear := season.AirDate[:4]
				if animeYear == seasonYear {
					matchScore += 1
				}
			}

			if matchScore > bestMatchScore {
				bestMatchScore = matchScore
				seasonNumber = season.SeasonNumber
			}
		}

		logger.Log(fmt.Sprintf("Using TMDB ID %d with season %d", showID, seasonNumber), types.LogOptions{
			Level:  types.Info,
			Prefix: "TMDB",
		})
	} else {
		// Search for the TV show on TMDB if we don't have a direct ID
		shows, err := searchTVShowsByTitle(title, alternativeTitle, false, "JP")
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to search TV shows: %v", err), types.LogOptions{
				Level:  types.Warn,
				Prefix: "TMDB",
			})
			return episodes
		}

		if len(shows) == 0 {
			logger.Log(fmt.Sprintf("No TV shows found for: %s", title), types.LogOptions{
				Level:  types.Warn,
				Prefix: "TMDB",
			})
			return episodes
		}

		// Find the best matching season
		airDate := ""
		if len(episodes) > 0 && episodes[0].Aired != "" {
			airDate = episodes[0].Aired
		}

		showID, seasonNumber, err = findBestSeason(shows, title, len(episodes), airDate)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to find best season: %v", err), types.LogOptions{
				Level:  types.Warn,
				Prefix: "TMDB",
			})
			return episodes
		}
	}

	// Get season details with episode information
	seasonDetails, err := getSeasonDetails(showID, seasonNumber)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get season details: %v", err), types.LogOptions{
			Level:  types.Warn,
			Prefix: "TMDB",
		})
		return episodes
	}

	// Enrich episodes with descriptions and thumbnails
	tmdbEpisodes := seasonDetails.Episodes
	enrichedEpisodes := make([]types.AnimeSingleEpisode, len(episodes))
	copy(enrichedEpisodes, episodes)

	// The base URL for TMDB images
	const tmdbImageBaseURL = "https://image.tmdb.org/t/p/"
	const thumbnailSize = "w300" // Use w300 size for episode thumbnails

	for i := range enrichedEpisodes {
		if i < len(tmdbEpisodes) {
			// Only add description if it's not empty
			if tmdbEpisodes[i].Overview != "" {
				enrichedEpisodes[i].Description = tmdbEpisodes[i].Overview
			} else {
				enrichedEpisodes[i].Description = "No description available"
			}

			// Add thumbnail URL if available
			if tmdbEpisodes[i].StillPath != "" {
				enrichedEpisodes[i].ThumbnailURL = tmdbImageBaseURL + thumbnailSize + tmdbEpisodes[i].StillPath
			}
		} else {
			enrichedEpisodes[i].Description = "No description available"
		}
	}

	thumbnailCount := 0
	for _, ep := range enrichedEpisodes {
		if ep.ThumbnailURL != "" {
			thumbnailCount++
		}
	}

	logger.Log(fmt.Sprintf("Successfully enriched %d episodes with descriptions and %d with thumbnails for: %s",
		len(enrichedEpisodes), thumbnailCount, title), types.LogOptions{
		Level:  types.Success,
		Prefix: "TMDB",
	})

	return enrichedEpisodes
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
			Aired:        episode.Aired,
			Score:        episode.Score,
			Filler:       episode.Filler,
			Recap:        episode.Recap,
			ForumURL:     episode.ForumURL,
			URL:          episode.URL,
			Description:  "No description available",
			ThumbnailURL: "",
			Stream: types.AnimeStreaming{
				SkipTimes: []types.AnimeSkipTimes{},
			},
		})
	}
	return AnimeEpisodes, nil
}

func generateEpisodeDataWithDescriptions(episodes []types.JikanAnimeEpisode, title string, alternativeTitle string, tmdbID int) ([]types.AnimeSingleEpisode, error) {
	basicEpisodes, err := generateEpisodeData(episodes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate basic episode data: %w", err)
	}

	enrichedEpisodes := AttachEpisodeDescriptions(title, basicEpisodes, alternativeTitle, tmdbID)

	// Get the MAL ID from the first episode URL if available
	var malID int
	if len(episodes) > 0 && episodes[0].URL != "" {
		// Extract MAL ID from URL like "https://myanimelist.net/anime/1735/Naruto__Shippuuden/episode/1"
		parts := strings.Split(episodes[0].URL, "/")
		for i, part := range parts {
			if part == "anime" && i+1 < len(parts) {
				// Try to parse the next part as an integer
				id, err := strconv.Atoi(parts[i+1])
				if err == nil {
					malID = id
					break
				}
			}
		}
	}

	if malID > 0 {
		logger.Log(fmt.Sprintf("Fetching skip times for anime with MAL ID %d", malID), types.LogOptions{
			Level:  types.Info,
			Prefix: "AniSkip",
		})

		// Process each episode to add skip times
		for i := range enrichedEpisodes {
			// Episode numbers in the API are 1-indexed
			episodeNumber := i + 1
			skipTimes, err := getAnimeEpisodeSkipTimes(malID, episodeNumber)

			if err != nil {
				logger.Log(fmt.Sprintf("Failed to get skip times for episode %d: %v", episodeNumber, err), types.LogOptions{
					Level:  types.Debug,
					Prefix: "AniSkip",
				})
				continue
			}

			if len(skipTimes) > 0 {
				enrichedEpisodes[i].Stream.SkipTimes = skipTimes
				logger.Log(fmt.Sprintf("Added %d skip times for episode %d", len(skipTimes), episodeNumber), types.LogOptions{
					Level:  types.Debug,
					Prefix: "AniSkip",
				})
			}
		}

		// Count how many episodes have skip times
		skipTimeCount := 0
		for _, ep := range enrichedEpisodes {
			if len(ep.Stream.SkipTimes) > 0 {
				skipTimeCount++
			}
		}

		if skipTimeCount > 0 {
			logger.Log(fmt.Sprintf("Successfully added skip times to %d/%d episodes for: %s",
				skipTimeCount, len(enrichedEpisodes), title), types.LogOptions{
				Level:  types.Success,
				Prefix: "AniSkip",
			})
		} else {
			logger.Log(fmt.Sprintf("No skip times found for any episodes of: %s", title), types.LogOptions{
				Level:  types.Warn,
				Prefix: "AniSkip",
			})
		}
	} else {
		logger.Log(fmt.Sprintf("Could not determine MAL ID for skip times for: %s", title), types.LogOptions{
			Level:  types.Warn,
			Prefix: "AniSkip",
		})
	}

	// Add streaming sources to episodes
	logger.Log(fmt.Sprintf("Fetching streaming sources for anime: %s", title), types.LogOptions{
		Level:  types.Info,
		Prefix: "Streaming",
	})

	// Prioritize original titles for searching - first romaji, then Japanese, then English
	// This better matches how anime streaming sites catalog their content
	searchTitle := title // Default to romaji title

	// Process all episodes to add streaming sources
	for i := range enrichedEpisodes {
		episodeNumber := i + 1

		streaming, err := GetStreamingSources(searchTitle, episodeNumber)
		if err != nil {
			// If search fails with romaji title, try with Japanese title if available
			if enrichedEpisodes[i].Titles.Japanese != "" {
				logger.Log(fmt.Sprintf("Retrying search with Japanese title for episode %d", episodeNumber), types.LogOptions{
					Level:  types.Debug,
					Prefix: "Streaming",
				})
				streaming, err = GetStreamingSources(enrichedEpisodes[i].Titles.Japanese, episodeNumber)
			}

			// If both fail and English title is available, try with that
			if err != nil && alternativeTitle != "" {
				logger.Log(fmt.Sprintf("Retrying search with English title for episode %d", episodeNumber), types.LogOptions{
					Level:  types.Debug,
					Prefix: "Streaming",
				})

				englishTitle := strings.TrimPrefix(alternativeTitle, "English: ")

				streaming, err = GetStreamingSources(englishTitle, episodeNumber)
			}

			if err != nil {
				logger.Log(fmt.Sprintf("Failed to get streaming sources for episode %d: %v", episodeNumber, err), types.LogOptions{
					Level:  types.Debug,
					Prefix: "Streaming",
				})
				continue
			}
		}

		// Keep the skip times which were already added
		streaming.SkipTimes = enrichedEpisodes[i].Stream.SkipTimes

		// Update the streaming sources
		enrichedEpisodes[i].Stream = *streaming

		// Add a small delay to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	// Count how many episodes have streaming sources
	streamingSubCount := 0
	streamingDubCount := 0

	for _, ep := range enrichedEpisodes {
		if len(ep.Stream.Sub) > 0 {
			streamingSubCount++
		}
		if len(ep.Stream.Dub) > 0 {
			streamingDubCount++
		}
	}

	if streamingSubCount > 0 || streamingDubCount > 0 {
		logger.Log(fmt.Sprintf("Successfully added streaming sources to episodes for: %s (SUB: %d/%d, DUB: %d/%d)",
			title, streamingSubCount, len(enrichedEpisodes), streamingDubCount, len(enrichedEpisodes)), types.LogOptions{
			Level:  types.Success,
			Prefix: "Streaming",
		})
	} else {
		logger.Log(fmt.Sprintf("No streaming sources found for any episodes of: %s", title), types.LogOptions{
			Level:  types.Warn,
			Prefix: "Streaming",
		})
	}

	return enrichedEpisodes, nil
}
