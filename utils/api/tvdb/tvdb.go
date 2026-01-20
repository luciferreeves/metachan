package tvdb

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"metachan/config"
	"metachan/database"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"time"
)

var tvdbToken string
var tvdbTokenExpiry time.Time

// authenticateTVDB authenticates with TVDB API and returns a token
func authenticateTVDB() (string, error) {
	// Check if we have a valid token
	if tvdbToken != "" && time.Now().Before(tvdbTokenExpiry) {
		return tvdbToken, nil
	}

	if config.Config.TVDB.APIKey == "" {
		return "", fmt.Errorf("TVDB API key is not set")
	}

	logger.Log("Authenticating with TVDB API", logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "TVDB",
	})

	client := &http.Client{Timeout: 10 * time.Second}

	// Create request body with apikey
	authBody := map[string]string{"apikey": config.Config.TVDB.APIKey}
	jsonBody, err := json.Marshal(authBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api4.thetvdb.com/v4/login", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var authResp TVDBAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResp.Data.Token == "" {
		return "", fmt.Errorf("no token received from TVDB")
	}

	// Store token and set expiry (TVDB tokens typically last 30 days, but we'll refresh after 24 hours to be safe)
	tvdbToken = authResp.Data.Token
	tvdbTokenExpiry = time.Now().Add(24 * time.Hour)

	logger.Log("Successfully authenticated with TVDB", logger.LogOptions{
		Level:  logger.Success,
		Prefix: "TVDB",
	})

	return tvdbToken, nil
}

// GetSeriesEpisodes fetches all episodes for a TVDB series
func GetSeriesEpisodes(tvdbID int) ([]TVDBEpisode, error) {
	token, err := authenticateTVDB()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with TVDB: %w", err)
	}

	logger.Log(fmt.Sprintf("Fetching episodes for TVDB series %d", tvdbID), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "TVDB",
	})

	client := &http.Client{Timeout: 15 * time.Second}

	// TVDB v4 API endpoint for episodes
	url := fmt.Sprintf("https://api4.thetvdb.com/v4/series/%d/episodes/default", tvdbID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch episodes with status: %d", resp.StatusCode)
	}

	var episodesResp TVDBEpisodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&episodesResp); err != nil {
		return nil, fmt.Errorf("failed to decode episodes response: %w", err)
	}

	logger.Log(fmt.Sprintf("Successfully fetched %d episodes from TVDB for series %d", len(episodesResp.Data.Episodes), tvdbID), logger.LogOptions{
		Level:  logger.Success,
		Prefix: "TVDB",
	})

	return episodesResp.Data.Episodes, nil
}

// ConvertTVDBEpisodesToAnimeEpisodes converts TVDB episodes to anime episode format
func ConvertTVDBEpisodesToAnimeEpisodes(tvdbEpisodes []TVDBEpisode) []types.AnimeSingleEpisode {
	var animeEpisodes []types.AnimeSingleEpisode

	const tvdbImageBaseURL = "https://artworks.thetvdb.com"

	for _, ep := range tvdbEpisodes {
		// Generate episode ID from name
		titles := types.EpisodeTitles{
			English:  ep.Name,
			Japanese: "",
			Romaji:   "",
		}

		thumbnailURL := ""
		if ep.Image != "" {
			thumbnailURL = ep.Image
		}

		description := ep.Overview
		if description == "" {
			description = "No description available"
		}

		isRecap := false
		if ep.FinaleType != nil && *ep.FinaleType == "recap" {
			isRecap = true
		}

		animeEpisodes = append(animeEpisodes, types.AnimeSingleEpisode{
			ID:           generateEpisodeID(titles),
			Titles:       titles,
			Description:  description,
			Aired:        ep.Aired,
			ThumbnailURL: thumbnailURL,
			Score:        0,
			Filler:       false,
			Recap:        isRecap,
			ForumURL:     "",
			URL:          "",
		})
	}

	return animeEpisodes
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

// FindSeasonMappings finds all anime mappings that belong to the same series based on TVDB ID
func FindSeasonMappings(tvdbID int) ([]entities.AnimeMapping, error) {
	logger.Log(fmt.Sprintf("Finding season mappings for TVDB ID %d", tvdbID), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "TVDB",
	})

	// Use our database function to find all mappings with the same TVDB ID
	mappings, err := database.GetAnimeMappingsByTVDBID(tvdbID)
	if err != nil {
		return nil, fmt.Errorf("failed to get season mappings: %w", err)
	}

	if len(mappings) == 0 {
		logger.Log(fmt.Sprintf("No season mappings found for TVDB ID %d", tvdbID), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "TVDB",
		})
	} else {
		logger.Log(fmt.Sprintf("Found %d season mappings for TVDB ID %d", len(mappings), tvdbID), logger.LogOptions{
			Level:  logger.Info,
			Prefix: "TVDB",
		})
	}

	return mappings, nil
}
