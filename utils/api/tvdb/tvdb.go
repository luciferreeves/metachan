package tvdb

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"metachan/config"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"time"
)

const (
	tvdbAPIBaseURL    = "https://api4.thetvdb.com/v4"
	tvdbLoginEndpoint = "/login"
	tvdbImageBaseURL  = "https://artworks.thetvdb.com"
	timeout           = 10 * time.Second
	episodesTimeout   = 15 * time.Second
	tokenExpiry       = 24 * time.Hour
	contentType       = "application/json"
	acceptHeader      = "application/json"
	noDescription     = "No description available"
	recapType         = "recap"
)

var (
	clientInstance = &client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
)

func authenticate() (string, error) {
	if clientInstance.token != "" && time.Now().Before(clientInstance.tokenExpiry) {
		return clientInstance.token, nil
	}

	if config.API.TVDBKey == "" {
		logger.Errorf("TVDB", "TVDB API key is not set")
		return "", errors.New("TVDB API key is not set")
	}

	logger.Debugf("TVDB", "Authenticating with TVDB API")

	authBody := map[string]string{"apikey": config.API.TVDBKey}
	jsonBody, err := json.Marshal(authBody)
	if err != nil {
		logger.Errorf("TVDB", "Failed to marshal auth body: %v", err)
		return "", errors.New("failed to marshal auth body")
	}

	req, err := http.NewRequest("POST", tvdbAPIBaseURL+tvdbLoginEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Errorf("TVDB", "Failed to create auth request: %v", err)
		return "", errors.New("failed to create auth request")
	}

	req.Header.Add("Content-Type", contentType)

	resp, err := clientInstance.httpClient.Do(req)
	if err != nil {
		logger.Errorf("TVDB", "Failed to authenticate: %v", err)
		return "", errors.New("failed to authenticate")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("TVDB", "Authentication failed with status: %d", resp.StatusCode)
		return "", errors.New("authentication failed")
	}

	var authResp types.TVDBAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		logger.Errorf("TVDB", "Failed to decode auth response: %v", err)
		return "", errors.New("failed to decode auth response")
	}

	if authResp.Data.Token == "" {
		logger.Errorf("TVDB", "No token received from TVDB")
		return "", errors.New("no token received from TVDB")
	}

	clientInstance.token = authResp.Data.Token
	clientInstance.tokenExpiry = time.Now().Add(tokenExpiry)

	logger.Successf("TVDB", "Successfully authenticated with TVDB")

	return clientInstance.token, nil
}

func GetSeriesEpisodes(tvdbID int) ([]types.TVDBEpisode, error) {
	token, err := authenticate()
	if err != nil {
		logger.Errorf("TVDB", "Failed to authenticate with TVDB for series %d: %v", tvdbID, err)
		return nil, errors.New("failed to authenticate with TVDB")
	}

	logger.Debugf("TVDB", "Fetching episodes for TVDB series %d", tvdbID)

	tempClient := &http.Client{Timeout: episodesTimeout}

	url := fmt.Sprintf("%s/series/%d/episodes/default", tvdbAPIBaseURL, tvdbID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorf("TVDB", "Failed to create request for series %d: %v", tvdbID, err)
		return nil, errors.New("failed to create request")
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", acceptHeader)

	resp, err := tempClient.Do(req)
	if err != nil {
		logger.Errorf("TVDB", "Failed to fetch episodes for series %d: %v", tvdbID, err)
		return nil, errors.New("failed to fetch episodes")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("TVDB", "Failed to fetch episodes with status: %d", resp.StatusCode)
		return nil, errors.New("failed to fetch episodes")
	}

	var episodesResp types.TVDBEpisodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&episodesResp); err != nil {
		logger.Errorf("TVDB", "Failed to decode episodes response for series %d: %v", tvdbID, err)
		return nil, errors.New("failed to decode episodes response")
	}

	logger.Successf("TVDB", "Successfully fetched %d episodes from TVDB for series %d", len(episodesResp.Data.Episodes), tvdbID)

	return episodesResp.Data.Episodes, nil
}

func EnrichEpisodesFromTVDB(anime *entities.Anime, tvdbEpisodes []types.TVDBEpisode) {
	if anime == nil || len(anime.Episodes) == 0 {
		return
	}

	malID := anime.MALID

	for i, ep := range tvdbEpisodes {
		if i >= len(anime.Episodes) {
			break
		}

		episode := &anime.Episodes[i]

		if ep.Name != "" {
			episode.Title = entities.EpisodeTitle{
				English:  ep.Name,
				Japanese: episode.Title.Japanese,
				Romaji:   episode.Title.Romaji,
			}
		}

		if ep.Image != "" {
			episode.ThumbnailURL = ep.Image
		}

		if ep.Overview != "" {
			episode.Description = ep.Overview
		} else {
			episode.Description = noDescription
		}

		if ep.Aired != "" {
			episode.Aired = ep.Aired
		}

		if ep.FinaleType != nil && *ep.FinaleType == recapType {
			episode.Recap = true
		}

		episode.EpisodeNumber = ep.Number
		episode.EpisodeLength = float64(ep.Runtime)

		titleForID := ep.Name
		if titleForID == "" {
			if episode.Title.English != "" {
				titleForID = episode.Title.English
			} else if episode.Title.Romaji != "" {
				titleForID = episode.Title.Romaji
			}
		}
		episode.EpisodeID = generateEpisodeID(malID, ep.Number, titleForID)
	}
}

func generateEpisodeID(malID int, episodeNumber int, title string) string {
	uniqueString := fmt.Sprintf("%d-%d-%s", malID, episodeNumber, title)
	hash := md5.Sum([]byte(uniqueString))
	return fmt.Sprintf("%x", hash)
}
