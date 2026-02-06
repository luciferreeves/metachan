package streaming

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"metachan/types"
	"metachan/utils/logger"
	"metachan/utils/mappers"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	allanimeBaseURL   = "https://api.allanime.day/api"
	allanimeDay       = "https://allanime.day"
	allanimeReferer   = "https://allmanga.to"
	clockPath         = "/apivtwo/clock"
	clockJSONPath     = "/apivtwo/clock.json"
	urlPrefix         = "--"
	unicodeSlash      = "\\u002F"
	userAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"
	timeout           = 10 * time.Second
	maxRetries        = 3
	backoffDuration   = 1 * time.Second
	perfectMatch      = 1.0
	partialMatch      = 0.9
	specialMatch      = 2.0
	serverMaria       = "Maria"
	serverSina        = "Sina"
	serverRose        = "Rose"
	serverTypeMP4     = "s-mp4"
	serverTypeLufMP4  = "luf-mp4"
	serverTypeDefault = "default"
	sourceTypeDirect  = "direct"
	sourceTypeEmbed   = "embed"
	sourceTypeHLS     = "HLS"
	sourceTypeMP4     = "MP4"
	patternSharepoint = "sharepoint.com"
	patternM3U8       = ".m3u8"
	patternMP4        = ".mp4"
	searchLimit       = 40
	searchPage        = 1
	countryOrigin     = "ALL"
)

var (
	clientInstance = &client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		headers: http.Header{
			"User-Agent": {userAgent},
			"Referer":    {allanimeReferer},
		},
		maxRetries: maxRetries,
		backoff:    backoffDuration,
	}
)

func calculateSimilarity(query, title string) float64 {
	queryLower := strings.ToLower(query)
	titleLower := strings.ToLower(title)

	if queryLower == titleLower {
		return perfectMatch
	}

	if strings.Contains(titleLower, queryLower) {
		return partialMatch
	}

	queryWords := strings.Fields(queryLower)
	titleWords := strings.Fields(titleLower)

	matchCount := 0
	for _, qw := range queryWords {
		for _, tw := range titleWords {
			if qw == tw || strings.Contains(tw, qw) || strings.Contains(qw, tw) {
				matchCount++
				break
			}
		}
	}

	if len(queryWords) == 0 {
		return 0
	}

	return float64(matchCount) / float64(len(queryWords))
}

func decodeURL(encodedString string) string {
	if !strings.HasPrefix(encodedString, urlPrefix) {
		return encodedString
	}

	encodedString = encodedString[len(urlPrefix):]
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

func processProviderURL(urlStr string) string {
	if strings.HasPrefix(urlStr, "/") {
		urlStr = strings.Replace(urlStr, clockPath, clockJSONPath, 1)
		return allanimeDay + urlStr
	}

	return urlStr
}

func getClockLink(urlStr string) (string, error) {
	if strings.HasPrefix(urlStr, "/") {
		urlStr = allanimeDay + urlStr
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	maps.Copy(req.Header, clientInstance.headers)

	resp, err := clientInstance.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if links, ok := data["links"].([]any); ok && len(links) > 0 {
		if link, ok := links[0].(map[string]any); ok {
			if linkStr, ok := link["link"].(string); ok {
				return linkStr, nil
			}
		}
	}

	return "", errors.New("no valid link found")
}

func processSourceURL(sourceURL, sourceType string) *types.StreamAnimeStreamingSource {
	var decodedURL string
	if strings.HasPrefix(sourceURL, urlPrefix) {
		decodedURL = decodeURL(sourceURL)
	} else {
		decodedURL = strings.ReplaceAll(sourceURL, unicodeSlash, "/")
	}

	processedURL := processProviderURL(decodedURL)

	if strings.Contains(processedURL, clockPath) {
		if directURL, err := getClockLink(processedURL); err == nil {
			return &types.StreamAnimeStreamingSource{
				URL:    directURL,
				Server: getServerName(sourceType),
				Type:   sourceTypeDirect,
			}
		}
	}

	directPatterns := []string{patternSharepoint, patternM3U8, patternMP4}
	for _, pattern := range directPatterns {
		if strings.Contains(processedURL, pattern) {
			return &types.StreamAnimeStreamingSource{
				URL:    processedURL,
				Server: getServerName(sourceType),
				Type:   sourceTypeDirect,
			}
		}
	}

	return &types.StreamAnimeStreamingSource{
		URL:    processedURL,
		Server: getServerName(sourceType),
		Type:   sourceTypeEmbed,
	}
}

func getServerName(sourceType string) string {
	switch strings.ToLower(sourceType) {
	case serverTypeMP4:
		return serverMaria
	case serverTypeLufMP4:
		return serverSina
	case serverTypeDefault:
		return serverRose
	default:
		return sourceType
	}
}

func SearchAnime(query string) ([]types.StreamSearchResult, error) {
	specialID, hasSpecialMapping := mappers.GetSpecialAnimeID(query)

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

	variables := map[string]any{
		"search": map[string]any{
			"allowAdult":   false,
			"allowUnknown": false,
			"query":        query,
		},
		"limit":         searchLimit,
		"page":          searchPage,
		"countryOrigin": countryOrigin,
	}

	params := url.Values{}
	variablesJSON, _ := json.Marshal(variables)
	params.Set("variables", string(variablesJSON))
	params.Set("query", searchQuery)

	req, err := http.NewRequest("GET", allanimeBaseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	maps.Copy(req.Header, clientInstance.headers)

	resp, err := clientInstance.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	shows := data["data"].(map[string]any)["shows"].(map[string]any)["edges"].([]any)
	results := make([]types.StreamSearchResult, 0, len(shows))

	for _, show := range shows {
		showMap := show.(map[string]any)
		episodes := showMap["availableEpisodes"].(map[string]any)
		result := types.StreamSearchResult{
			ID:          showMap["_id"].(string),
			Name:        showMap["name"].(string),
			SubEpisodes: int(episodes["sub"].(float64)),
			DubEpisodes: int(episodes["dub"].(float64)),
			Similarity:  calculateSimilarity(query, showMap["name"].(string)),
		}

		if hasSpecialMapping && result.ID == specialID {
			result.Similarity = specialMatch
		}

		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	return results, nil
}

func GetEpisodesList(showID string, mode string) ([]string, error) {
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

	variables := map[string]any{
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

	maps.Copy(req.Header, clientInstance.headers)

	resp, err := clientInstance.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	showData := data["data"].(map[string]any)["show"].(map[string]any)
	episodesDetail := showData["availableEpisodesDetail"].(map[string]any)
	episodesList := episodesDetail[mode].([]any)

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

func GetEpisodeLinks(showID, episode, mode string) ([]types.StreamAnimeStreamingSource, error) {
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

	variables := map[string]any{
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

	maps.Copy(req.Header, clientInstance.headers)

	resp, err := clientInstance.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	episodeData := data["data"].(map[string]any)["episode"].(map[string]any)
	sourceUrls := episodeData["sourceUrls"].([]any)

	var links []types.StreamAnimeStreamingSource
	for _, source := range sourceUrls {
		sourceMap := source.(map[string]any)
		if sourceURL, ok := sourceMap["sourceUrl"].(string); ok {
			sourceName := sourceMap["sourceName"].(string)
			sourceInfo := processSourceURL(sourceURL, sourceName)

			if sourceInfo.Type == sourceTypeDirect {
				if strings.HasSuffix(sourceInfo.URL, patternM3U8) {
					sourceInfo.Type = sourceTypeHLS
				} else {
					sourceInfo.Type = sourceTypeMP4
				}
				links = append(links, *sourceInfo)
			}
		}
	}

	return links, nil
}

func GetStreamingSources(title string, episodeNumber int) (*types.StreamAnimeStreaming, error) {
	logger.Debugf("Streaming", "Fetching streaming sources for '%s' episode %d", title, episodeNumber)

	searchResults, err := SearchAnime(title)
	if err != nil {
		logger.Errorf("Streaming", "Failed to search anime '%s': %v", title, err)
		return nil, errors.New("failed to search for anime")
	}

	if len(searchResults) == 0 {
		logger.Warnf("Streaming", "No streaming sources found for '%s'", title)
		return nil, errors.New("no streaming sources found")
	}

	bestMatch := searchResults[0]
	logger.Debugf("Streaming", "Best match: '%s' (ID: %s, Sub: %d, Dub: %d)", bestMatch.Name, bestMatch.ID, bestMatch.SubEpisodes, bestMatch.DubEpisodes)

	streaming := &types.StreamAnimeStreaming{
		Sub: []types.StreamAnimeStreamingSource{},
		Dub: []types.StreamAnimeStreamingSource{},
	}

	if bestMatch.SubEpisodes > 0 {
		episodes, err := GetEpisodesList(bestMatch.ID, "sub")
		if err == nil && len(episodes) > 0 {
			episodeStr := fmt.Sprintf("%d", episodeNumber)
			var closestEpisode string

			for _, ep := range episodes {
				if ep == episodeStr {
					closestEpisode = ep
					break
				}
			}

			if closestEpisode != "" {
				subSources, err := GetEpisodeLinks(bestMatch.ID, closestEpisode, "sub")
				if err == nil {
					streaming.Sub = subSources
					logger.Debugf("Streaming", "Found %d sub sources for episode %d", len(subSources), episodeNumber)
				} else {
					logger.Warnf("Streaming", "Failed to get sub sources: %v", err)
				}
			}
		}
	}

	if bestMatch.DubEpisodes > 0 {
		episodes, err := GetEpisodesList(bestMatch.ID, "dub")
		if err == nil && len(episodes) > 0 {
			episodeStr := fmt.Sprintf("%d", episodeNumber)
			var closestEpisode string

			for _, ep := range episodes {
				if ep == episodeStr {
					closestEpisode = ep
					break
				}
			}

			if closestEpisode != "" {
				dubSources, err := GetEpisodeLinks(bestMatch.ID, closestEpisode, "dub")
				if err == nil {
					streaming.Dub = dubSources
					logger.Debugf("Streaming", "Found %d dub sources for episode %d", len(dubSources), episodeNumber)
				} else {
					logger.Warnf("Streaming", "Failed to get dub sources: %v", err)
				}
			}
		}
	}

	logger.Infof("Streaming", "Successfully fetched streaming sources for episode %d (Sub: %d, Dub: %d)", episodeNumber, len(streaming.Sub), len(streaming.Dub))
	return streaming, nil
}

func GetStreamingCounts(title string) (int, int, error) {
	searchResults, err := SearchAnime(title)
	if err != nil {
		logger.Errorf("Streaming", "Failed to search anime '%s': %v", title, err)
		return 0, 0, errors.New("failed to search for anime")
	}

	if len(searchResults) == 0 {
		logger.Warnf("Streaming", "No results found for '%s'", title)
		return 0, 0, errors.New("no results found")
	}

	bestMatch := searchResults[0]

	return bestMatch.SubEpisodes, bestMatch.DubEpisodes, nil
}
