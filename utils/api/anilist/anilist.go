package anilist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"metachan/utils/logger"
	"net/http"
)

// AniListClient provides methods for interacting with the AniList API
type AniListClient struct {
	client     *http.Client
	maxRetries int
}

// NewAniListClient creates a new AniList API client
func NewAniListClient() *AniListClient {
	return &AniListClient{
		client:     &http.Client{},
		maxRetries: 3,
	}
}

// GetAnime fetches anime details from AniList by ID using a simpler approach
func (c *AniListClient) GetAnime(anilistID int) (*AnilistAnimeResponse, error) {
	// Create a much simpler request with minimal formatting that might trigger Cloudflare
	query := `
	query ($id: Int) {
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
			startDate { year month day }
			endDate { year month day }
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
			trailer { id site thumbnail }
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
			tags { id name description category rank isGeneralSpoiler isMediaSpoiler isAdult }
			nextAiringEpisode { id airingAt timeUntilAiring episode }
			airingSchedule { nodes { id episode airingAt timeUntilAiring } }
			studios { edges { isMain node { id name } } }
			isAdult
		}
	}
	`

	// Create a simple JSON structure with variables
	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"id": anilistID,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Log the request for debugging
	logger.Log(fmt.Sprintf("Sending request to AniList for ID %d", anilistID), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AniList",
	})

	var resp *http.Response
	var lastErr error
	success := false

	for i := 0; i <= c.maxRetries && !success; i++ {
		req, err := http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Add User-Agent to make the request look more like a browser
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		resp, err = c.client.Do(req)
		if err != nil {
			lastErr = err
			logger.Log(fmt.Sprintf("AniList request attempt %d failed: %v", i+1, err), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AniList",
			})
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			lastErr = fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body[:n]))
			logger.Log(fmt.Sprintf("AniList returned non-200 status on attempt %d: %v", i+1, lastErr), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AniList",
			})
			continue
		}

		var anilistResponse AnilistAnimeResponse
		if err := json.NewDecoder(resp.Body).Decode(&anilistResponse); err != nil {
			lastErr = fmt.Errorf("failed to decode response: %w", err)
			continue
		}

		if anilistResponse.Data.Media.ID == 0 {
			lastErr = fmt.Errorf("no data found for Anilist ID %d", anilistID)
			continue
		}

		// Log cover image data for debugging
		if anilistResponse.Data.Media.CoverImage.ExtraLarge != "" {
			logger.Log(fmt.Sprintf("Found cover data - Color: %s, Image: %s",
				anilistResponse.Data.Media.CoverImage.Color,
				anilistResponse.Data.Media.CoverImage.ExtraLarge), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AniList",
			})
		}

		success = true
		return &anilistResponse, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}
