package anime

import (
	"encoding/json"
	"fmt"
	"io"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"time"
)

func getAnimeViaMalSync(malID int) (*types.MALSyncAnimeResponse, error) {
	apiURL := fmt.Sprintf("https://api.malsync.moe/mal/anime/%d", malID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{
		Timeout: 10 * time.Second, // Add timeout to prevent hanging requests
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get anime data: %s - %s", resp.Status, string(bodyBytes))
	}

	var malSyncResponse types.MALSyncAnimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&malSyncResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &malSyncResponse, nil
}

func extractLogosFromMALSync(malSyncResponse *types.MALSyncAnimeResponse) types.AnimeLogos {
	logos := types.AnimeLogos{}

	// Check if Crunchyroll data exists in the MALSync response
	crunchyrollSites, exists := malSyncResponse.Sites["Crunchyroll"]
	if !exists || len(crunchyrollSites) == 0 {
		logger.Log("No Crunchyroll data found in MALSync response", types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return logos
	}

	// Get the Crunchyroll URL from any of the entries
	crURL := ""
	for _, site := range crunchyrollSites {
		crURL = site.URL
		break // Take the first URL
	}

	if crURL == "" {
		logger.Log("No valid Crunchyroll URL found", types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return logos
	}

	// Extract series ID from URL
	seriesID := extractCrunchyrollSeriesID(crURL)
	if seriesID == "" {
		return logos
	}

	// Define logo sizes
	logoSizes := map[string]int{
		"Small":    320,
		"Medium":   480,
		"Large":    600,
		"XLarge":   800,
		"Original": 1200,
	}

	// Generate logo URLs
	logos.Small = fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Small"], seriesID)
	logos.Medium = fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Medium"], seriesID)
	logos.Large = fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Large"], seriesID)
	logos.XLarge = fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["XLarge"], seriesID)
	logos.Original = fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Original"], seriesID)

	logger.Log(fmt.Sprintf("Successfully generated logo URLs for series ID: %s", seriesID), types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})

	return logos
}
