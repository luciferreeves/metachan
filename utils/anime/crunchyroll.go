package anime

import (
	"crypto/tls"
	"fmt"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"strings"
	"time"
)

func extractCrunchyrollSeriesID(crURL string) string {
	logger.Log(fmt.Sprintf("Attempting to extract series ID from URL: %s", crURL), types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})

	// Direct series URL format
	if strings.Contains(crURL, "/series/") {
		parts := strings.Split(crURL, "/series/")
		if len(parts) < 2 {
			logger.Log("URL contains /series/ but couldn't extract ID part", types.LogOptions{
				Level:  types.Debug,
				Prefix: "AnimeAPI",
			})
			return ""
		}

		idParts := strings.Split(parts[1], "/")
		if len(idParts) < 1 {
			logger.Log("Couldn't extract ID from path segments", types.LogOptions{
				Level:  types.Debug,
				Prefix: "AnimeAPI",
			})
			return ""
		}

		logger.Log(fmt.Sprintf("Found series ID directly in URL: %s", idParts[0]), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return idParts[0]
	}

	// Need to follow redirect to get series ID
	logger.Log("URL doesn't contain /series/, following redirect...", types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})

	// Create a transport that uses modern TLS settings
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects, just capture the Location header
			return http.ErrUseLastResponse
		},
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	// Update HTTP to HTTPS for Crunchyroll URLs if needed
	if strings.HasPrefix(crURL, "http://www.crunchyroll.com") {
		crURL = strings.Replace(crURL, "http://", "https://", 1)
		logger.Log(fmt.Sprintf("Updated URL to HTTPS: %s", crURL), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
	}

	// Add User-Agent header to mimic a browser
	req, err := http.NewRequest("GET", crURL, nil)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to create request for Crunchyroll URL: %v", err), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return ""
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")

	resp, err := client.Do(req)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get Crunchyroll redirect: %v", err), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return ""
	}
	defer resp.Body.Close()

	// Log the status code and response headers for debugging
	logger.Log(fmt.Sprintf("Crunchyroll response status: %d %s", resp.StatusCode, resp.Status), types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})

	for name, values := range resp.Header {
		logger.Log(fmt.Sprintf("Header %s: %s", name, strings.Join(values, ", ")), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
	}

	// Check for specific status codes for redirects
	if resp.StatusCode != http.StatusMovedPermanently &&
		resp.StatusCode != http.StatusFound &&
		resp.StatusCode != http.StatusTemporaryRedirect &&
		resp.StatusCode != http.StatusPermanentRedirect {
		logger.Log(fmt.Sprintf("Unexpected status code from Crunchyroll: %d", resp.StatusCode), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})

		// If we got a 200 OK, maybe Crunchyroll served the page directly
		// Try to extract the series ID from the URL itself as a fallback
		if resp.StatusCode == http.StatusOK && strings.Contains(crURL, "crunchyroll.com") {
			// For URLs like http://www.crunchyroll.com/fullmetal-alchemist-brotherhood
			// Extract the last part as a potential identifier
			urlParts := strings.Split(crURL, "/")
			if len(urlParts) > 0 {
				potentialId := urlParts[len(urlParts)-1]
				logger.Log(fmt.Sprintf("Extracted potential series ID from original URL: %s", potentialId), types.LogOptions{
					Level:  types.Debug,
					Prefix: "AnimeAPI",
				})
				return potentialId
			}
		}
		return ""
	}

	redirectURL := resp.Header.Get("Location")
	if redirectURL == "" {
		logger.Log("No redirect URL found in Crunchyroll response", types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return ""
	}

	logger.Log(fmt.Sprintf("Found redirect URL: %s", redirectURL), types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})

	// Extract series ID from redirect URL
	if strings.Contains(redirectURL, "/series/") {
		parts := strings.Split(redirectURL, "/series/")
		if len(parts) < 2 {
			return ""
		}

		idParts := strings.Split(parts[1], "/")
		if len(idParts) < 1 {
			return ""
		}

		logger.Log(fmt.Sprintf("Successfully extracted series ID from redirect: %s", idParts[0]), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return idParts[0]
	}

	// For multi-level redirects, try to follow one more time
	if strings.Contains(redirectURL, "crunchyroll.com") {
		logger.Log("Trying to follow one more redirect level...", types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return extractCrunchyrollSeriesID(redirectURL)
	}

	// As a fallback for older Crunchyroll URLs like fullmetal-alchemist-brotherhood
	// Use the last path segment as the ID
	urlParts := strings.Split(crURL, "/")
	if len(urlParts) > 0 {
		potentialId := urlParts[len(urlParts)-1]
		logger.Log(fmt.Sprintf("Using fallback: extracted ID from original URL: %s", potentialId), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		return potentialId
	}

	logger.Log("Could not extract series ID from Crunchyroll redirect URL", types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})
	return ""
}
