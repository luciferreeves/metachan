package tmdb

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"metachan/config"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"strings"
	"time"
)

const (
	MAX_RETRIES = 10
)

// makeSimpleRequest executes a simple HTTP request with retries for both connection and rate limit errors
func makeSimpleRequest(req *http.Request) (*http.Response, error) {
	// Create a simple HTTP client with a short timeout
	client := &http.Client{
		Timeout: 5 * time.Second, // Reduced timeout for faster failure
	}

	// Do retries for up to MAX_RETRIES attempts
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt < MAX_RETRIES; attempt++ {
		// Log the attempt
		// Execute the request
		resp, lastErr = client.Do(req)

		// If successful, check for rate limiting
		if lastErr == nil {
			// If we got rate limited (429), wait and retry
			if resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()

				logger.Log(fmt.Sprintf("TMDB rate limited (attempt %d/%d): waiting 5 seconds", attempt+1, MAX_RETRIES), logger.LogOptions{
					Level:  logger.Warn,
					Prefix: "TMDB",
				})

				// Wait for 5 seconds before retrying for rate limits
				time.Sleep(5 * time.Second)
				continue
			}

			// Any other status code (including success) should be returned
			return resp, nil
		}

		// Check if this is a connection reset error for immediate retry
		if strings.Contains(lastErr.Error(), "connection reset") {
			logger.Log(fmt.Sprintf("TMDB connection reset (attempt %d/%d): retrying immediately", attempt+1, MAX_RETRIES), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "TMDB",
			})
			continue
		}

		// Log the error
		logger.Log(fmt.Sprintf("TMDB request error (attempt %d/%d): %v", attempt+1, MAX_RETRIES, lastErr), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "TMDB",
		})
	}

	// All attempts failed, return the last error
	return nil, fmt.Errorf("failed after %d retry attempts: %w", MAX_RETRIES, lastErr)
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
func searchTVShowsByTitle(title string, alternativeTitle string, isAdult bool, countryPriority string) ([]TMDBShowResult, error) {
	if config.Config.TMDB.ReadAccessToken == "" {
		return nil, fmt.Errorf("TMDB is not initialized")
	}

	// Normalize the title
	query := normalizeTitle(title)
	if query == "" && alternativeTitle != "" {
		query = normalizeTitle(alternativeTitle)
	}

	logger.Log(fmt.Sprintf("Searching TMDB for TV show: %s", query), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "TMDB",
	})

	// Create request
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

	// Make the simple request
	resp, err := makeSimpleRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search TV shows: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search TV shows: %s", resp.Status)
	}

	// Parse response
	var searchResponse TMDBSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Filter results if needed
	var filteredResults []TMDBShowResult
	for _, show := range searchResponse.Results {
		if (isAdult && show.Adult) || (!isAdult && !show.Adult) {
			filteredResults = append(filteredResults, show)
		}
	}

	// Sort by country priority if specified
	if countryPriority != "" && len(filteredResults) > 0 {
		var prioritizedResults []TMDBShowResult
		var otherResults []TMDBShowResult

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
		logger.Log(fmt.Sprintf("No TMDB shows found for: %s", query), logger.LogOptions{
			Level:  logger.Warn,
			Prefix: "TMDB",
		})
	} else {
		logger.Log(fmt.Sprintf("Found %d TMDB shows for: %s", len(filteredResults), query), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "TMDB",
		})
	}

	return filteredResults, nil
}

// getTVShowDetails gets details for a TV show from TMDB
func getTVShowDetails(showID int) (*TMDBShowDetails, error) {
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

	// Make the simple request
	resp, err := makeSimpleRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get TV show details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get TV show details: %s", resp.Status)
	}

	// Parse response
	details := &TMDBShowDetails{}
	if err := json.NewDecoder(resp.Body).Decode(details); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return details, nil
}

// getSeasonDetails gets details for a TV season from TMDB
func getSeasonDetails(showID, seasonNumber int) (*TMDBSeasonDetails, error) {
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

	// Make the simple request
	resp, err := makeSimpleRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get season details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get season details: %s", resp.Status)
	}

	// Parse response
	details := &TMDBSeasonDetails{}
	if err := json.NewDecoder(resp.Body).Decode(details); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return details, nil
}

// findBestSeason finds the best matching season for an anime
func findBestSeason(shows []TMDBShowResult, title string, episodeCount int, airDate string) (int, int, error) {
	for _, show := range shows {
		showDetails, err := getTVShowDetails(show.ID)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to get details for show %d: %v", show.ID, err), logger.LogOptions{
				Level:  logger.Warn,
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
					title, show.ID, season.SeasonNumber), logger.LogOptions{
					Level:  logger.Info,
					Prefix: "TMDB",
				})
				return show.ID, season.SeasonNumber, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("could not find matching season for: %s", title)
}

// AttachEpisodeDescriptions enriches anime episodes with descriptions and thumbnails from TMDB
func AttachEpisodeDescriptions(title string, episodes []types.AnimeSingleEpisode, alternativeTitle string, tmdbID int) ([]types.AnimeSingleEpisode, error) {
	if config.Config.TMDB.ReadAccessToken == "" {
		logger.Log("TMDB is not configured, skipping episode description enrichment", logger.LogOptions{
			Level:  logger.Warn,
			Prefix: "TMDB",
		})
		return episodes, fmt.Errorf("TMDB is not configured")
	}

	if len(episodes) == 0 {
		return episodes, nil
	}

	logger.Log(fmt.Sprintf("Enriching episodes for: %s", title), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "TMDB",
	})

	var showID int
	var seasonNumber int
	var err error

	// Use a short timeout for the entire operation
	startTime := time.Now()
	maxDuration := 10 * time.Second

	// If we have a TMDB ID, use it directly
	if tmdbID > 0 {
		showID = tmdbID

		// Check if we've exceeded the timeout
		if time.Since(startTime) > maxDuration {
			logger.Log("TMDB enrichment timed out", logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "TMDB",
			})
			return episodes, fmt.Errorf("TMDB enrichment timed out")
		}

		// Try to get show details and find the best season
		showDetails, err := getTVShowDetails(showID)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to get TMDB show details for ID %d: %v", tmdbID, err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "TMDB",
			})
			return episodes, fmt.Errorf("failed to get TMDB show details: %w", err)
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

		logger.Log(fmt.Sprintf("Using TMDB ID %d with season %d", showID, seasonNumber), logger.LogOptions{
			Level:  logger.Info,
			Prefix: "TMDB",
		})
	} else {
		// Check if we've exceeded the timeout
		if time.Since(startTime) > maxDuration {
			logger.Log("TMDB enrichment timed out", logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "TMDB",
			})
			return episodes, fmt.Errorf("TMDB enrichment timed out")
		}

		// Search for the TV show on TMDB if we don't have a direct ID
		shows, err := searchTVShowsByTitle(title, alternativeTitle, false, "JP")
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to search TV shows: %v", err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "TMDB",
			})
			return episodes, fmt.Errorf("failed to search TMDB shows: %w", err)
		}

		if len(shows) == 0 {
			logger.Log(fmt.Sprintf("No TV shows found for: %s", title), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "TMDB",
			})
			return episodes, fmt.Errorf("no TMDB shows found for: %s", title)
		}

		// Find the best matching season
		airDate := ""
		if len(episodes) > 0 && episodes[0].Aired != "" {
			airDate = episodes[0].Aired
		}

		// Check if we've exceeded the timeout
		if time.Since(startTime) > maxDuration {
			logger.Log("TMDB enrichment timed out", logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "TMDB",
			})
			return episodes, fmt.Errorf("TMDB enrichment timed out")
		}

		showID, seasonNumber, err = findBestSeason(shows, title, len(episodes), airDate)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to find best season: %v", err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "TMDB",
			})
			return episodes, fmt.Errorf("failed to find best season: %w", err)
		}
	}

	// Check if we've exceeded the timeout
	if time.Since(startTime) > maxDuration {
		logger.Log("TMDB enrichment timed out", logger.LogOptions{
			Level:  logger.Warn,
			Prefix: "TMDB",
		})
		return episodes, fmt.Errorf("TMDB enrichment timed out")
	}

	// Get season details with episode information
	seasonDetails, err := getSeasonDetails(showID, seasonNumber)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get season details: %v", err), logger.LogOptions{
			Level:  logger.Warn,
			Prefix: "TMDB",
		})
		return episodes, fmt.Errorf("failed to get season details: %w", err)
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
		len(enrichedEpisodes), thumbnailCount, title), logger.LogOptions{
		Level:  logger.Success,
		Prefix: "TMDB",
	})

	return enrichedEpisodes, nil
}

// searchMoviesByTitle searches for movies on TMDB by title
func searchMoviesByTitle(title string, alternativeTitle string) ([]TMDBMovieResult, error) {
	if config.Config.TMDB.ReadAccessToken == "" {
		return nil, fmt.Errorf("TMDB is not initialized")
	}

	query := normalizeTitle(title)
	if query == "" && alternativeTitle != "" {
		query = normalizeTitle(alternativeTitle)
	}

	logger.Log(fmt.Sprintf("Searching TMDB for movie: %s", query), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "TMDB",
	})

	apiURL := "https://api.themoviedb.org/3/search/movie"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Config.TMDB.ReadAccessToken))
	req.Header.Add("Accept", "application/json")

	resp, err := makeSimpleRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	var searchResp TMDBMovieSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Log(fmt.Sprintf("Found %d movie results for: %s", len(searchResp.Results), query), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "TMDB",
	})

	return searchResp.Results, nil
}

// getMovieDetails fetches details for a specific movie
func getMovieDetails(movieID int) (*TMDBMovieDetails, error) {
	if config.Config.TMDB.ReadAccessToken == "" {
		return nil, fmt.Errorf("TMDB is not initialized")
	}

	apiURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d", movieID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Config.TMDB.ReadAccessToken))
	req.Header.Add("Accept", "application/json")

	resp, err := makeSimpleRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var movieDetails TMDBMovieDetails
	if err := json.NewDecoder(resp.Body).Decode(&movieDetails); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &movieDetails, nil
}

// GetMovieAsEpisode fetches movie details and returns it as a single episode
func GetMovieAsEpisode(title string, alternativeTitle string, tmdbID int, malID int, japaneseTitle string, animeScore float64) ([]types.AnimeSingleEpisode, error) {
	logger.Log(fmt.Sprintf("Fetching movie episode data for: %s", title), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "TMDB",
	})

	var movieID int
	var err error

	// If TMDB ID is provided, use it directly
	if tmdbID > 0 {
		movieID = tmdbID
		logger.Log(fmt.Sprintf("Using provided TMDB movie ID: %d", movieID), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "TMDB",
		})
	} else {
		// Search for the movie
		movies, err := searchMoviesByTitle(title, alternativeTitle)
		if err != nil || len(movies) == 0 {
			logger.Log(fmt.Sprintf("Failed to find movie on TMDB: %v", err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "TMDB",
			})
			return nil, fmt.Errorf("movie not found on TMDB")
		}

		// Use the first result
		movieID = movies[0].ID
		logger.Log(fmt.Sprintf("Found TMDB movie ID: %d for title: %s", movieID, title), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "TMDB",
		})
	}

	// Get movie details
	movieDetails, err := getMovieDetails(movieID)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to fetch movie details: %v", err), logger.LogOptions{
			Level:  logger.Warn,
			Prefix: "TMDB",
		})
		return nil, err
	}

	// Convert movie to a single episode format
	const tmdbImageBaseURL = "https://image.tmdb.org/t/p/"
	const backdropSize = "w780"

	backdropURL := ""
	if movieDetails.BackdropPath != "" {
		backdropURL = tmdbImageBaseURL + backdropSize + movieDetails.BackdropPath
	} else if movieDetails.PosterPath != "" {
		backdropURL = tmdbImageBaseURL + backdropSize + movieDetails.PosterPath
	}

	description := movieDetails.Overview
	if description == "" {
		description = "No description available"
	}

	// Create titles structure with English title from TMDB and Japanese/Romaji from MAL
	titles := types.EpisodeTitles{
		English:  movieDetails.Title,
		Japanese: japaneseTitle,
		Romaji:   title,
	}

	// Calculate score out of 5 (half of MAL score out of 10), rounded to 2 decimal points
	movieScore := float64(int((animeScore/2.0)*100)) / 100

	// Generate MAL URLs
	malURL := ""
	forumURL := ""
	if malID > 0 {
		malURL = fmt.Sprintf("https://myanimelist.net/anime/%d", malID)
		forumURL = fmt.Sprintf("https://myanimelist.net/anime/%d/forum", malID)
	}

	episode := types.AnimeSingleEpisode{
		ID:           generateEpisodeID(titles),
		Titles:       titles,
		Description:  description,
		ThumbnailURL: backdropURL,
		Aired:        movieDetails.ReleaseDate,
		Score:        movieScore,
		Filler:       false,
		Recap:        false,
		ForumURL:     forumURL,
		URL:          malURL,
	}

	logger.Log(fmt.Sprintf("Successfully created episode from movie: %s", title), logger.LogOptions{
		Level:  logger.Success,
		Prefix: "TMDB",
	})

	return []types.AnimeSingleEpisode{episode}, nil
}

// generateEpisodeID creates a unique episode ID from titles
func generateEpisodeID(titles types.EpisodeTitles) string {
	var title string
	if titles.English != "" {
		title = titles.English
	} else if titles.Romaji != "" {
		title = titles.Romaji
	} else {
		title = titles.Japanese
	}

	// MD5 hash for ID generation to match Jikan episode IDs
	hash := md5.Sum([]byte(title))
	return fmt.Sprintf("%x", hash)
}
