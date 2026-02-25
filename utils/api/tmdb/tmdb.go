package tmdb

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"metachan/config"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
	"net/http"
	"strings"
	"time"
)

const (
	tmdbAPIBaseURL          = "https://api.themoviedb.org/3"
	tmdbImageBaseURL        = "https://image.tmdb.org/t/p/"
	searchTVEndpoint        = "/search/tv"
	searchMovieEndpoint     = "/search/movie"
	tvDetailsEndpoint       = "/tv/%d"
	seasonDetailsEndpoint   = "/tv/%d/season/%d"
	movieDetailsEndpoint    = "/movie/%d"
	timeout                 = 5 * time.Second
	rateLimitWait           = 5 * time.Second
	maxRetries              = 10
	maxEnrichmentDuration   = 10 * time.Second
	thumbnailSize           = "w300"
	backdropSize            = "w780"
	acceptHeader            = "application/json"
	connectionResetError    = "connection reset"
	noDescription           = "No description available"
	tvAnimation             = "TV Animation"
	seasonSuffix            = ": Season"
	seasonWord              = "Season"
	partWord                = "Part"
	courWord                = "Cour"
	countryPriorityJP       = "JP"
	episodeCountFlexibility = 2
)

var (
	clientInstance = &client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
)

func makeRequest(req *http.Request) (*http.Response, error) {
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, lastErr = clientInstance.httpClient.Do(req)

		if lastErr == nil {
			if resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				logger.Warnf("TMDB", "TMDB rate limited (attempt %d/%d): waiting %d seconds", attempt+1, maxRetries, int(rateLimitWait.Seconds()))
				time.Sleep(rateLimitWait)
				continue
			}
			return resp, nil
		}

		if strings.Contains(lastErr.Error(), connectionResetError) {
			logger.Debugf("TMDB", "TMDB connection reset (attempt %d/%d): retrying immediately", attempt+1, maxRetries)
			continue
		}

		logger.Debugf("TMDB", "TMDB request error (attempt %d/%d): %v", attempt+1, maxRetries, lastErr)
	}

	logger.Errorf("TMDB", "Failed after %d retry attempts: %v", maxRetries, lastErr)
	return nil, errors.New("failed after max retry attempts")
}

func normalizeTitle(title string) string {
	if title == "" {
		return ""
	}

	normalized := title
	normalized = strings.Replace(normalized, tvAnimation, "", -1)
	normalized = strings.Replace(normalized, seasonSuffix, "", -1)
	normalized = strings.Replace(normalized, seasonWord, "", -1)
	normalized = strings.Replace(normalized, partWord, "", -1)
	normalized = strings.Replace(normalized, courWord, "", -1)

	if colonIndex := strings.Index(normalized, ":"); colonIndex > 0 {
		normalized = normalized[:colonIndex]
	}

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

func searchTVShowsByTitle(title string, alternativeTitle string, isAdult bool, countryPriority string) ([]types.TMDBShowResult, error) {
	if config.API.TMDBReadToken == "" {
		logger.Errorf("TMDB", "TMDB is not initialized")
		return nil, errors.New("TMDB is not initialized")
	}

	query := normalizeTitle(title)
	if query == "" && alternativeTitle != "" {
		query = normalizeTitle(alternativeTitle)
	}

	logger.Debugf("TMDB", "Searching TMDB for TV show: %s", query)

	apiURL := tmdbAPIBaseURL + searchTVEndpoint
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Errorf("TMDB", "Failed to create request: %v", err)
		return nil, errors.New("failed to create request")
	}

	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.API.TMDBReadToken))
	req.Header.Add("Accept", acceptHeader)

	resp, err := makeRequest(req)
	if err != nil {
		logger.Errorf("TMDB", "Failed to search TV shows: %v", err)
		return nil, errors.New("failed to search TV shows")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("TMDB", "Failed to search TV shows: %s", resp.Status)
		return nil, errors.New("failed to search TV shows")
	}

	var searchResponse types.TMDBSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		logger.Errorf("TMDB", "Failed to decode response: %v", err)
		return nil, errors.New("failed to decode response")
	}

	var filteredResults []types.TMDBShowResult
	for _, show := range searchResponse.Results {
		if (isAdult && show.Adult) || (!isAdult && !show.Adult) {
			filteredResults = append(filteredResults, show)
		}
	}

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

		filteredResults = append(prioritizedResults, otherResults...)
	}

	if len(filteredResults) == 0 {
		logger.Warnf("TMDB", "No TMDB shows found for: %s", query)
	} else {
		logger.Debugf("TMDB", "Found %d TMDB shows for: %s", len(filteredResults), query)
	}

	return filteredResults, nil
}

func getTVShowDetails(showID int) (*types.TMDBShowDetails, error) {
	if config.API.TMDBReadToken == "" {
		logger.Errorf("TMDB", "TMDB is not initialized")
		return nil, errors.New("TMDB is not initialized")
	}

	apiURL := fmt.Sprintf(tmdbAPIBaseURL+tvDetailsEndpoint, showID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Errorf("TMDB", "Failed to create request: %v", err)
		return nil, errors.New("failed to create request")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.API.TMDBReadToken))
	req.Header.Add("Accept", acceptHeader)

	resp, err := makeRequest(req)
	if err != nil {
		logger.Errorf("TMDB", "Failed to get TV show details: %v", err)
		return nil, errors.New("failed to get TV show details")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("TMDB", "Failed to get TV show details: %s", resp.Status)
		return nil, errors.New("failed to get TV show details")
	}

	details := &types.TMDBShowDetails{}
	if err := json.NewDecoder(resp.Body).Decode(details); err != nil {
		logger.Errorf("TMDB", "Failed to decode response: %v", err)
		return nil, errors.New("failed to decode response")
	}

	return details, nil
}

func getSeasonDetails(showID, seasonNumber int) (*types.TMDBSeasonDetails, error) {
	if config.API.TMDBReadToken == "" {
		logger.Errorf("TMDB", "TMDB is not initialized")
		return nil, errors.New("TMDB is not initialized")
	}

	apiURL := fmt.Sprintf(tmdbAPIBaseURL+seasonDetailsEndpoint, showID, seasonNumber)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Errorf("TMDB", "Failed to create request: %v", err)
		return nil, errors.New("failed to create request")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.API.TMDBReadToken))
	req.Header.Add("Accept", acceptHeader)

	resp, err := makeRequest(req)
	if err != nil {
		logger.Errorf("TMDB", "Failed to get season details: %v", err)
		return nil, errors.New("failed to get season details")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("TMDB", "Failed to get season details: %s", resp.Status)
		return nil, errors.New("failed to get season details")
	}

	details := &types.TMDBSeasonDetails{}
	if err := json.NewDecoder(resp.Body).Decode(details); err != nil {
		logger.Errorf("TMDB", "Failed to decode response: %v", err)
		return nil, errors.New("failed to decode response")
	}

	return details, nil
}

func findBestSeason(shows []types.TMDBShowResult, title string, episodeCount int, airDate string) (int, int, error) {
	for _, show := range shows {
		showDetails, err := getTVShowDetails(show.ID)
		if err != nil {
			logger.Warnf("TMDB", "Failed to get details for show %d: %v", show.ID, err)
			continue
		}

		for _, season := range showDetails.Seasons {
			if season.SeasonNumber == 0 {
				continue
			}

			episodeCountMatches := season.EpisodeCount == episodeCount ||
				(episodeCount > 0 && season.EpisodeCount >= episodeCount-episodeCountFlexibility &&
					season.EpisodeCount <= episodeCount+episodeCountFlexibility)

			airDateMatches := false
			if airDate != "" && season.AirDate != "" {
				animeYear := airDate[:4]
				seasonYear := season.AirDate[:4]
				airDateMatches = animeYear == seasonYear
			}

			if episodeCountMatches || airDateMatches {
				logger.Infof("TMDB", "Found matching season for \"%s\": Show ID %d, Season %d", title, show.ID, season.SeasonNumber)
				return show.ID, season.SeasonNumber, nil
			}
		}
	}

	logger.Warnf("TMDB", "Could not find matching season for: %s", title)
	return 0, 0, errors.New("could not find matching season")
}

func AttachEpisodeDescriptions(anime *entities.Anime) error {
	if config.API.TMDBReadToken == "" {
		logger.Warnf("TMDB", "TMDB is not configured, skipping episode description enrichment")
		return errors.New("TMDB is not configured")
	}

	if anime == nil || len(anime.Episodes) == 0 {
		return nil
	}

	title := anime.Title.Romaji
	alternativeTitle := anime.Title.English

	tmdbID := 0
	malID := anime.MALID
	if anime.Mapping != nil {
		tmdbID = anime.Mapping.TMDB
	}

	episodes := make([]*entities.Episode, len(anime.Episodes))
	for i := range anime.Episodes {
		episodes[i] = &anime.Episodes[i]
	}

	logger.Infof("TMDB", "Enriching episodes for: %s", title)

	var showID int
	var seasonNumber int
	var err error

	startTime := time.Now()

	if tmdbID > 0 {
		showID = tmdbID

		if time.Since(startTime) > maxEnrichmentDuration {
			logger.Warnf("TMDB", "TMDB enrichment timed out")
			return errors.New("TMDB enrichment timed out")
		}

		showDetails, err := getTVShowDetails(showID)
		if err != nil {
			logger.Warnf("TMDB", "Failed to get TMDB show details for ID %d: %v", tmdbID, err)
			return errors.New("failed to get TMDB show details")
		}

		seasonNumber = 1
		bestMatchScore := 0

		for _, season := range showDetails.Seasons {
			if season.SeasonNumber == 0 {
				continue
			}

			matchScore := 0

			if math.Abs(float64(season.EpisodeCount-len(episodes))) <= episodeCountFlexibility {
				matchScore += 2
			}

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

		logger.Infof("TMDB", "Using TMDB ID %d with season %d", showID, seasonNumber)
	} else {
		if time.Since(startTime) > maxEnrichmentDuration {
			logger.Warnf("TMDB", "TMDB enrichment timed out")
			return errors.New("TMDB enrichment timed out")
		}

		shows, err := searchTVShowsByTitle(title, alternativeTitle, false, countryPriorityJP)
		if err != nil {
			logger.Warnf("TMDB", "Failed to search TV shows: %v", err)
			return errors.New("failed to search TMDB shows")
		}

		if len(shows) == 0 {
			logger.Warnf("TMDB", "No TV shows found for: %s", title)
			return errors.New("no TMDB shows found")
		}

		airDate := ""
		if len(episodes) > 0 && episodes[0].Aired != "" {
			airDate = episodes[0].Aired
		}

		if time.Since(startTime) > maxEnrichmentDuration {
			logger.Warnf("TMDB", "TMDB enrichment timed out")
			return errors.New("TMDB enrichment timed out")
		}

		showID, seasonNumber, err = findBestSeason(shows, title, len(episodes), airDate)
		if err != nil {
			logger.Warnf("TMDB", "Failed to find best season: %v", err)
			return errors.New("failed to find best season")
		}
	}

	if time.Since(startTime) > maxEnrichmentDuration {
		logger.Warnf("TMDB", "TMDB enrichment timed out")
		return errors.New("TMDB enrichment timed out")
	}

	seasonDetails, err := getSeasonDetails(showID, seasonNumber)
	if err != nil {
		logger.Warnf("TMDB", "Failed to get season details: %v", err)
		return errors.New("failed to get season details")
	}

	tmdbEpisodes := seasonDetails.Episodes

	for i, episode := range episodes {
		if i < len(tmdbEpisodes) {
			if tmdbEpisodes[i].Overview != "" {
				episode.Description = tmdbEpisodes[i].Overview
			} else {
				episode.Description = noDescription
			}

			if tmdbEpisodes[i].StillPath != "" {
				episode.ThumbnailURL = tmdbImageBaseURL + thumbnailSize + tmdbEpisodes[i].StillPath
			}

			episode.EpisodeNumber = tmdbEpisodes[i].EpisodeNumber

			titleForID := ""
			if episode.Title.English != "" {
				titleForID = episode.Title.English
			} else if episode.Title.Romaji != "" {
				titleForID = episode.Title.Romaji
			}
			if titleForID == "" && tmdbEpisodes[i].Name != "" {
				titleForID = tmdbEpisodes[i].Name
			}
			episode.EpisodeID = generateEpisodeID(malID, episode.EpisodeNumber, titleForID)
		} else {
			episode.Description = noDescription
		}
	}

	thumbnailCount := 0
	for _, ep := range episodes {
		if ep.ThumbnailURL != "" {
			thumbnailCount++
		}
	}

	logger.Successf("TMDB", "Successfully enriched %d episodes with descriptions and %d with thumbnails for: %s", len(episodes), thumbnailCount, title)

	return nil
}

func searchMoviesByTitle(title string, alternativeTitle string) ([]types.TMDBMovieResult, error) {
	if config.API.TMDBReadToken == "" {
		logger.Errorf("TMDB", "TMDB is not initialized")
		return nil, errors.New("TMDB is not initialized")
	}

	query := normalizeTitle(title)
	if query == "" && alternativeTitle != "" {
		query = normalizeTitle(alternativeTitle)
	}

	logger.Debugf("TMDB", "Searching TMDB for movie: %s", query)

	apiURL := tmdbAPIBaseURL + searchMovieEndpoint
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Errorf("TMDB", "Failed to create request: %v", err)
		return nil, errors.New("failed to create request")
	}

	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.API.TMDBReadToken))
	req.Header.Add("Accept", acceptHeader)

	resp, err := makeRequest(req)
	if err != nil {
		logger.Errorf("TMDB", "Failed to search movies: %v", err)
		return nil, errors.New("failed to search movies")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("TMDB", "Search failed with status: %d", resp.StatusCode)
		return nil, errors.New("search failed")
	}

	var searchResp types.TMDBMovieSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		logger.Errorf("TMDB", "Failed to decode response: %v", err)
		return nil, errors.New("failed to decode response")
	}

	logger.Debugf("TMDB", "Found %d movie results for: %s", len(searchResp.Results), query)

	return searchResp.Results, nil
}

func getMovieDetails(movieID int) (*types.TMDBMovieDetails, error) {
	if config.API.TMDBReadToken == "" {
		logger.Errorf("TMDB", "TMDB is not initialized")
		return nil, errors.New("TMDB is not initialized")
	}

	apiURL := fmt.Sprintf(tmdbAPIBaseURL+movieDetailsEndpoint, movieID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Errorf("TMDB", "Failed to create request: %v", err)
		return nil, errors.New("failed to create request")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.API.TMDBReadToken))
	req.Header.Add("Accept", acceptHeader)

	resp, err := makeRequest(req)
	if err != nil {
		logger.Errorf("TMDB", "Failed to fetch movie details: %v", err)
		return nil, errors.New("failed to fetch movie details")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("TMDB", "Request failed with status: %d", resp.StatusCode)
		return nil, errors.New("request failed")
	}

	var movieDetails types.TMDBMovieDetails
	if err := json.NewDecoder(resp.Body).Decode(&movieDetails); err != nil {
		logger.Errorf("TMDB", "Failed to decode response: %v", err)
		return nil, errors.New("failed to decode response")
	}

	return &movieDetails, nil
}

func EnrichEpisodeFromMovie(anime *entities.Anime) error {
	if anime == nil || len(anime.Episodes) == 0 {
		return nil
	}

	episode := &anime.Episodes[0]

	title := anime.Title.Romaji
	alternativeTitle := anime.Title.English
	japaneseTitle := anime.Title.Japanese

	tmdbID := 0
	malID := anime.MALID
	if anime.Mapping != nil {
		tmdbID = anime.Mapping.TMDB
	}

	animeScore := anime.Scores.Score

	logger.Debugf("TMDB", "Fetching movie episode data for: %s", title)

	var movieID int
	var err error

	if tmdbID > 0 {
		movieID = tmdbID
		logger.Debugf("TMDB", "Using provided TMDB movie ID: %d", movieID)
	} else {
		movies, err := searchMoviesByTitle(title, alternativeTitle)
		if err != nil || len(movies) == 0 {
			logger.Warnf("TMDB", "Failed to find movie on TMDB: %v", err)
			return errors.New("movie not found on TMDB")
		}

		movieID = movies[0].ID
		logger.Debugf("TMDB", "Found TMDB movie ID: %d for title: %s", movieID, title)
	}

	movieDetails, err := getMovieDetails(movieID)
	if err != nil {
		logger.Warnf("TMDB", "Failed to fetch movie details: %v", err)
		return err
	}

	backdropURL := ""
	if movieDetails.BackdropPath != "" {
		backdropURL = tmdbImageBaseURL + backdropSize + movieDetails.BackdropPath
	} else if movieDetails.PosterPath != "" {
		backdropURL = tmdbImageBaseURL + backdropSize + movieDetails.PosterPath
	}

	description := movieDetails.Overview
	if description == "" {
		description = noDescription
	}

	episode.Title = entities.EpisodeTitle{
		English:  movieDetails.Title,
		Japanese: japaneseTitle,
		Romaji:   title,
	}

	movieScore := float64(int((animeScore/2.0)*100)) / 100

	malURL := ""
	forumURL := ""
	if malID > 0 {
		malURL = fmt.Sprintf("https://myanimelist.net/anime/%d", malID)
		forumURL = fmt.Sprintf("https://myanimelist.net/anime/%d/forum", malID)
	}

	episode.EpisodeID = generateEpisodeID(malID, 1, movieDetails.Title)
	episode.Description = description
	episode.ThumbnailURL = backdropURL
	episode.Aired = movieDetails.ReleaseDate
	episode.Score = movieScore
	episode.Filler = false
	episode.Recap = false
	episode.ForumURL = forumURL
	episode.URL = malURL
	episode.EpisodeNumber = 1
	episode.EpisodeLength = float64(movieDetails.Runtime)

	logger.Successf("TMDB", "Successfully created episode from movie: %s", title)

	return nil
}

func generateEpisodeID(malID int, episodeNumber int, title string) string {
	uniqueString := fmt.Sprintf("%d-%d-%s", malID, episodeNumber, title)
	hash := md5.Sum([]byte(uniqueString))
	return fmt.Sprintf("%x", hash)
}
