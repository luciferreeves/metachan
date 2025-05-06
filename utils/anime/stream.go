package anime

import (
	"encoding/json"
	"fmt"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	allanimeBaseURL = "https://api.allanime.day/api"
)

// AllAnimeClient handles communication with the AllAnime API
type AllAnimeClient struct {
	client  *http.Client
	headers http.Header
}

// StreamingSearchResult represents an anime search result from AllAnime
type StreamingSearchResult struct {
	ID          string  `json:"_id"`
	Name        string  `json:"name"`
	SubEpisodes int     `json:"sub_episodes"`
	DubEpisodes int     `json:"dub_episodes"`
	Similarity  float64 `json:"similarity"`
}

// NewAllAnimeClient creates a new client for accessing the AllAnime API
func NewAllAnimeClient() *AllAnimeClient {
	headers := http.Header{
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"},
		"Referer":    {"https://allmanga.to"},
	}

	return &AllAnimeClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		headers: headers,
	}
}

// calculateSimilarity determines how closely a title matches a query
func (c *AllAnimeClient) calculateSimilarity(query, title string) float64 {
	query = strings.ToLower(strings.TrimSpace(query))
	title = strings.ToLower(strings.TrimSpace(title))

	if query == title {
		return 1.0
	}

	if strings.Contains(title, query) {
		return 0.9
	}

	matches := 0
	queryRunes := []rune(query)
	titleRunes := []rune(title)

	for i := 0; i < len(queryRunes); i++ {
		for j := 0; j < len(titleRunes); j++ {
			if queryRunes[i] == titleRunes[j] {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(query))
}

// decodeURL decodes an encoded URL from AllAnime
func (c *AllAnimeClient) decodeURL(encodedString string) string {
	if !strings.HasPrefix(encodedString, "--") {
		return encodedString
	}

	encodedString = encodedString[2:]
	decodeMap := map[string]string{
		"01": "9", "08": "0", "05": "=", "0a": "2",
		"0b": "3", "0c": "4", "07": "?", "00": "8",
		"5c": "d", "0f": "7", "5e": "f", "17": "/",
		"54": "l", "09": "1", "48": "p", "4f": "w",
		"0e": "6", "5b": "c", "5d": "e", "0d": "5",
		"53": "k", "1e": "&", "5a": "b", "59": "a",
		"4a": "r", "4c": "t", "4e": "v", "57": "o",
		"51": "i",
	}

	var decoded strings.Builder
	for i := 0; i < len(encodedString); i += 2 {
		if i+2 <= len(encodedString) {
			pair := encodedString[i : i+2]
			if val, ok := decodeMap[pair]; ok {
				decoded.WriteString(val)
			}
		}
	}

	return decoded.String()
}

// processProviderURL processes provider URLs from AllAnime
func (c *AllAnimeClient) processProviderURL(urlStr string) string {
	baseURL := "https://allanime.day"

	if strings.HasPrefix(urlStr, "/") {
		urlStr = strings.Replace(urlStr, "/apivtwo/clock", "/apivtwo/clock.json", 1)
		return baseURL + urlStr
	}

	return urlStr
}

// getClockLink fetches a direct streaming link from a clock endpoint
func (c *AllAnimeClient) getClockLink(urlStr string) (string, error) {
	if strings.HasPrefix(urlStr, "/") {
		urlStr = "https://allanime.day" + urlStr
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	for key, values := range c.headers {
		req.Header[key] = values
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if links, ok := data["links"].([]interface{}); ok && len(links) > 0 {
		if link, ok := links[0].(map[string]interface{}); ok {
			if linkStr, ok := link["link"].(string); ok {
				return linkStr, nil
			}
		}
	}

	return "", fmt.Errorf("no valid link found")
}

// processSourceURL processes a streaming source URL from AllAnime
func (c *AllAnimeClient) processSourceURL(sourceURL, sourceType string) *types.AnimeStreamingSource {
	var decodedURL string
	if strings.HasPrefix(sourceURL, "--") {
		decodedURL = c.decodeURL(sourceURL)
	} else {
		decodedURL = strings.ReplaceAll(sourceURL, "\\u002F", "/")
	}

	processedURL := c.processProviderURL(decodedURL)

	// Check if it's a clock link
	if strings.Contains(processedURL, "/apivtwo/clock") {
		if directURL, err := c.getClockLink(processedURL); err == nil {
			return &types.AnimeStreamingSource{
				URL:    directURL,
				Server: getServerName(sourceType),
				Type:   "direct",
			}
		}
	}

	// Check if it's a direct stream link
	directPatterns := []string{"fast4speed.rsvp", "sharepoint.com", ".m3u8", ".mp4"}
	for _, pattern := range directPatterns {
		if strings.Contains(processedURL, pattern) {
			return &types.AnimeStreamingSource{
				URL:    processedURL,
				Server: getServerName(sourceType),
				Type:   "direct",
			}
		}
	}

	// Return as regular source if not direct
	return &types.AnimeStreamingSource{
		URL:    processedURL,
		Server: getServerName(sourceType),
		Type:   "embed",
	}
}

// getServerName maps AllAnime source types to readable server names
func getServerName(sourceType string) string {
	switch strings.ToLower(sourceType) {
	case "default":
		return "Maria"
	case "luf-mp4":
		return "Rose"
	case "s-mp4":
		return "Sina"
	default:
		return sourceType
	}
}

// searchAnime searches for anime by title on AllAnime
func (c *AllAnimeClient) searchAnime(query string) ([]StreamingSearchResult, error) {
	searchQuery := `
	query(
		$search: SearchInput
		$limit: Int
		$page: Int
		$countryOrigin: VaildCountryOriginEnumType
	) {
		shows(
			search: $search
			limit: $limit
			page: $page
			countryOrigin: $countryOrigin
		) {
			edges {
				_id
				name
				availableEpisodes
				__typename
			}
		}
	}
	`

	variables := map[string]interface{}{
		"search": map[string]interface{}{
			"allowAdult":   false,
			"allowUnknown": false,
			"query":        query,
		},
		"limit":         40,
		"page":          1,
		"countryOrigin": "ALL",
	}

	params := url.Values{}
	variablesJSON, _ := json.Marshal(variables)
	params.Set("variables", string(variablesJSON))
	params.Set("query", searchQuery)

	req, err := http.NewRequest("GET", allanimeBaseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	for key, values := range c.headers {
		req.Header[key] = values
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	shows := data["data"].(map[string]interface{})["shows"].(map[string]interface{})["edges"].([]interface{})
	results := make([]StreamingSearchResult, 0)

	for _, show := range shows {
		showMap := show.(map[string]interface{})
		episodes := showMap["availableEpisodes"].(map[string]interface{})

		result := StreamingSearchResult{
			ID:          showMap["_id"].(string),
			Name:        showMap["name"].(string),
			SubEpisodes: int(episodes["sub"].(float64)),
			DubEpisodes: int(episodes["dub"].(float64)),
		}
		result.Similarity = c.calculateSimilarity(query, result.Name)
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	return results, nil
}

// getEpisodesList gets the list of available episodes for an anime
func (c *AllAnimeClient) getEpisodesList(showID string, mode string) ([]string, error) {
	episodesQuery := `
	query ($showId: String!) {
		show(
			_id: $showId
		) {
			_id
			availableEpisodesDetail
		}
	}
	`

	variables := map[string]interface{}{
		"showId": showID,
	}

	params := url.Values{}
	variablesJSON, _ := json.Marshal(variables)
	params.Set("variables", string(variablesJSON))
	params.Set("query", episodesQuery)

	req, err := http.NewRequest("GET", allanimeBaseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	for key, values := range c.headers {
		req.Header[key] = values
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	showData := data["data"].(map[string]interface{})["show"].(map[string]interface{})
	episodesDetail := showData["availableEpisodesDetail"].(map[string]interface{})
	episodesList := episodesDetail[mode].([]interface{})

	result := make([]string, 0, len(episodesList))
	for _, ep := range episodesList {
		switch v := ep.(type) {
		case float64:
			result = append(result, fmt.Sprintf("%.0f", v))
		case string:
			result = append(result, v)
		default:
			result = append(result, fmt.Sprintf("%v", v))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		ni, _ := strconv.Atoi(result[i])
		nj, _ := strconv.Atoi(result[j])
		return ni < nj
	})

	return result, nil
}

// getEpisodeLinks gets streaming links for a specific episode
func (c *AllAnimeClient) getEpisodeLinks(showID, episode, mode string) ([]types.AnimeStreamingSource, error) {
	episodeQuery := `
	query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) {
		episode(
			showId: $showId
			translationType: $translationType
			episodeString: $episodeString
		) {
			episodeString
			sourceUrls
		}
	}
	`

	variables := map[string]interface{}{
		"showId":          showID,
		"translationType": mode,
		"episodeString":   episode,
	}

	params := url.Values{}
	variablesJSON, _ := json.Marshal(variables)
	params.Set("variables", string(variablesJSON))
	params.Set("query", episodeQuery)

	req, err := http.NewRequest("GET", allanimeBaseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	for key, values := range c.headers {
		req.Header[key] = values
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	episodeData := data["data"].(map[string]interface{})["episode"].(map[string]interface{})
	sourceUrls := episodeData["sourceUrls"].([]interface{})

	var links []types.AnimeStreamingSource
	for _, source := range sourceUrls {
		sourceMap := source.(map[string]interface{})
		if sourceURL, ok := sourceMap["sourceUrl"].(string); ok {
			sourceName := sourceMap["sourceName"].(string)
			sourceInfo := c.processSourceURL(sourceURL, sourceName)

			// Only add direct sources, matching the BubbleTea implementation
			if sourceInfo.Type == "direct" {
				links = append(links, *sourceInfo)
			}
		}
	}

	return links, nil
}

// GetStreamingSources fetches both sub and dub streaming sources for an anime episode
func GetStreamingSources(title string, episodeNumber int) (*types.AnimeStreaming, error) {
	client := NewAllAnimeClient()

	logger.Log(fmt.Sprintf("Searching for streaming sources for '%s' episode %d", title, episodeNumber), types.LogOptions{
		Level:  types.Info,
		Prefix: "Streaming",
	})

	// Search for the anime
	searchResults, err := client.searchAnime(title)
	if err != nil {
		return nil, fmt.Errorf("failed to search for anime: %w", err)
	}

	if len(searchResults) == 0 {
		return nil, fmt.Errorf("no streaming sources found for '%s'", title)
	}

	// Use the best match (first result)
	bestMatch := searchResults[0]
	logger.Log(fmt.Sprintf("Found anime '%s' with similarity %.2f", bestMatch.Name, bestMatch.Similarity), types.LogOptions{
		Level:  types.Debug,
		Prefix: "Streaming",
	})

	streaming := &types.AnimeStreaming{
		SkipTimes: []types.AnimeSkipTimes{},
		Sub:       []types.AnimeStreamingSource{},
		Dub:       []types.AnimeStreamingSource{},
	}

	// Convert episode number to string
	episodeStr := strconv.Itoa(episodeNumber)

	// Get sub episodes if available
	if bestMatch.SubEpisodes > 0 {
		episodes, err := client.getEpisodesList(bestMatch.ID, "sub")
		if err == nil && len(episodes) > 0 {
			// Find the closest episode
			var closestEpisode string
			for _, ep := range episodes {
				if ep == episodeStr {
					closestEpisode = ep
					break
				}
			}

			if closestEpisode != "" {
				subSources, err := client.getEpisodeLinks(bestMatch.ID, closestEpisode, "sub")
				if err == nil {
					streaming.Sub = subSources
					logger.Log(fmt.Sprintf("Found %d sub streaming sources for episode %s", len(subSources), closestEpisode), types.LogOptions{
						Level:  types.Debug,
						Prefix: "Streaming",
					})
				}
			}
		}
	}

	// Get dub episodes if available
	if bestMatch.DubEpisodes > 0 {
		episodes, err := client.getEpisodesList(bestMatch.ID, "dub")
		if err == nil && len(episodes) > 0 {
			// Find the closest episode
			var closestEpisode string
			for _, ep := range episodes {
				if ep == episodeStr {
					closestEpisode = ep
					break
				}
			}

			if closestEpisode != "" {
				dubSources, err := client.getEpisodeLinks(bestMatch.ID, closestEpisode, "dub")
				if err == nil {
					streaming.Dub = dubSources
					logger.Log(fmt.Sprintf("Found %d dub streaming sources for episode %s", len(dubSources), closestEpisode), types.LogOptions{
						Level:  types.Debug,
						Prefix: "Streaming",
					})
				}
			}
		}
	}

	// Check if we found any sources
	if len(streaming.Sub) == 0 && len(streaming.Dub) == 0 {
		return nil, fmt.Errorf("no streaming sources found for '%s' episode %d", title, episodeNumber)
	}

	logger.Log(fmt.Sprintf("Successfully retrieved streaming sources for '%s' episode %d: %d sub, %d dub",
		title, episodeNumber, len(streaming.Sub), len(streaming.Dub)), types.LogOptions{
		Level:  types.Success,
		Prefix: "Streaming",
	})

	return streaming, nil
}
