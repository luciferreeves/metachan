package anime

import (
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

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
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

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
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

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
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

// AttachEpisodeDescriptions enriches anime episodes with descriptions from TMDB
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

	// Enrich episodes with descriptions
	tmdbEpisodes := seasonDetails.Episodes
	enrichedEpisodes := make([]types.AnimeSingleEpisode, len(episodes))
	copy(enrichedEpisodes, episodes)

	for i := range enrichedEpisodes {
		if i < len(tmdbEpisodes) {
			// Only add description if it's not empty
			if tmdbEpisodes[i].Overview != "" {
				enrichedEpisodes[i].Description = tmdbEpisodes[i].Overview
			} else {
				enrichedEpisodes[i].Description = "No description available"
			}
		} else {
			enrichedEpisodes[i].Description = "No description available"
		}
	}

	logger.Log(fmt.Sprintf("Successfully enriched %d episodes with descriptions for: %s",
		len(enrichedEpisodes), title), types.LogOptions{
		Level:  types.Success,
		Prefix: "TMDB",
	})

	return enrichedEpisodes
}
