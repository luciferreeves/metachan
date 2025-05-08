package anime

import (
	"crypto/tls"
	"fmt"
	"metachan/types"
	"metachan/utils/api/anilist"
	"metachan/utils/api/jikan"
	"metachan/utils/api/malsync"
	"metachan/utils/api/tmdb"
	api "metachan/utils/api/tvdb"
	"metachan/utils/logger"
	"net/http"
	"strings"
	"time"
)

// generateBasicEpisodes creates a basic list of episode data from Jikan episodes
func generateBasicEpisodes(episodes []jikan.JikanAnimeEpisode) []types.AnimeSingleEpisode {
	var animeEpisodes []types.AnimeSingleEpisode

	for _, episode := range episodes {
		animeEpisodes = append(animeEpisodes, types.AnimeSingleEpisode{
			Titles: types.EpisodeTitles{
				English:  episode.Title,
				Japanese: episode.TitleJapanese,
				Romaji:   episode.TitleRomaji,
			},
			Aired:        episode.Aired,
			Score:        episode.Score,
			Filler:       episode.Filler,
			Recap:        episode.Recap,
			ForumURL:     episode.ForumURL,
			URL:          episode.URL,
			Description:  "No description available",
			ThumbnailURL: "",
			// Stream field removed
		})
	}
	return animeEpisodes
}

// getEpisodeCount determines the highest episode count from different sources
func getEpisodeCount(malAnime *jikan.JikanAnimeResponse, anilistAnime *anilist.AnilistAnimeResponse) int {
	if anilistAnime == nil {
		return malAnime.Data.Episodes
	}

	streamingScheduleLength := len(anilistAnime.Data.Media.AiringSchedule.Nodes)
	episodes := max(malAnime.Data.Episodes, anilistAnime.Data.Media.Episodes)
	episodes = max(episodes, streamingScheduleLength)

	return episodes
}

// sortSeasonsByAirDate sorts the seasons array chronologically by air date
func sortSeasonsByAirDate(seasons *[]types.AnimeSeason) {
	// First, collect seasons with valid dates
	seasonsWithDates := make([]types.AnimeSeason, 0)
	seasonsWithoutDates := make([]types.AnimeSeason, 0)

	for _, season := range *seasons {
		if season.AiringStatus.From.Year > 0 {
			seasonsWithDates = append(seasonsWithDates, season)
		} else {
			seasonsWithoutDates = append(seasonsWithoutDates, season)
		}
	}

	// Sort seasons with dates
	if len(seasonsWithDates) > 0 {
		sortedSeasons := make([]types.AnimeSeason, len(seasonsWithDates))
		copy(sortedSeasons, seasonsWithDates)

		for i := 0; i < len(sortedSeasons)-1; i++ {
			for j := i + 1; j < len(sortedSeasons); j++ {
				a := sortedSeasons[i].AiringStatus.From
				b := sortedSeasons[j].AiringStatus.From

				// Compare years
				if a.Year > b.Year {
					sortedSeasons[i], sortedSeasons[j] = sortedSeasons[j], sortedSeasons[i]
				} else if a.Year == b.Year {
					// Compare months if years are equal
					if a.Month > b.Month {
						sortedSeasons[i], sortedSeasons[j] = sortedSeasons[j], sortedSeasons[i]
					} else if a.Month == b.Month {
						// Compare days if months are equal
						if a.Day > b.Day {
							sortedSeasons[i], sortedSeasons[j] = sortedSeasons[j], sortedSeasons[i]
						}
					}
				}
			}
		}

		// Combine sorted dates with no-date seasons
		result := append(sortedSeasons, seasonsWithoutDates...)
		*seasons = result
	}
}

// generateGenres converts Jikan genre structures to our format
func generateGenres(genres, explicitGenres []jikan.JikanGenericMALStructure) []types.AnimeGenres {
	var animeGenres []types.AnimeGenres

	// Add regular genres
	for _, genre := range genres {
		animeGenres = append(animeGenres, types.AnimeGenres{
			Name:    genre.Name,
			GenreID: genre.MALID,
			URL:     genre.URL,
		})
	}

	// Add explicit genres if any
	for _, genre := range explicitGenres {
		animeGenres = append(animeGenres, types.AnimeGenres{
			Name:    genre.Name,
			GenreID: genre.MALID,
			URL:     genre.URL,
		})
	}

	return animeGenres
}

// generateStudios converts Jikan studio structures to our format
func generateStudios(studios []jikan.JikanGenericMALStructure) []types.AnimeStudio {
	var animeStudios []types.AnimeStudio

	for _, studio := range studios {
		animeStudios = append(animeStudios, types.AnimeStudio{
			Name:     studio.Name,
			StudioID: studio.MALID,
			URL:      studio.URL,
		})
	}

	return animeStudios
}

// generateProducers converts Jikan producer structures to our format
func generateProducers(producers []jikan.JikanGenericMALStructure) []types.AnimeProducer {
	var animeProducers []types.AnimeProducer

	for _, producer := range producers {
		animeProducers = append(animeProducers, types.AnimeProducer{
			Name:       producer.Name,
			ProducerID: producer.MALID,
			URL:        producer.URL,
		})
	}

	return animeProducers
}

// generateLicensors converts Jikan licensor structures to our format
func generateLicensors(licensors []jikan.JikanGenericMALStructure) []types.AnimeLicensor {
	var animeLicensors []types.AnimeLicensor

	for _, licensor := range licensors {
		animeLicensors = append(animeLicensors, types.AnimeLicensor{
			Name:       licensor.Name,
			ProducerID: licensor.MALID,
			URL:        licensor.URL,
		})
	}

	return animeLicensors
}

// getAnimeCharacters processes character data from Jikan
func getAnimeCharacters(characterResponse *jikan.JikanAnimeCharacterResponse) []types.AnimeCharacter {
	var characters []types.AnimeCharacter

	for _, entry := range characterResponse.Data {
		character := types.AnimeCharacter{
			MALID:    entry.Character.MALID,
			URL:      entry.Character.URL,
			ImageURL: entry.Character.Images.JPG.ImageURL,
			Name:     entry.Character.Name,
			Role:     entry.Role,
		}

		for _, va := range entry.VoiceActors {
			character.VoiceActors = append(character.VoiceActors, types.AnimeVoiceActor{
				MALID:    va.Person.MALID,
				URL:      va.Person.URL,
				Image:    va.Person.Images.JPG.ImageURL,
				Name:     va.Person.Name,
				Language: va.Language,
			})
		}

		characters = append(characters, character)
	}

	return characters
}

// getNextAiringEpisode extracts next airing episode data from AniList
func getNextAiringEpisode(anilistAnime *anilist.AnilistAnimeResponse) types.AnimeAiringEpisode {
	if anilistAnime == nil || anilistAnime.Data.Media.ID == 0 {
		logger.Log("No valid AniList data for next airing episode", logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return types.AnimeAiringEpisode{}
	}

	// NextAiringEpisode can be nil for completed anime
	nextEpisode := anilistAnime.Data.Media.NextAiringEpisode

	// Check if there is valid data
	if nextEpisode.AiringAt == 0 && nextEpisode.Episode == 0 {
		logger.Log(fmt.Sprintf("Anime ID %d has no next airing episode", anilistAnime.Data.Media.ID), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return types.AnimeAiringEpisode{}
	}

	logger.Log(fmt.Sprintf("Found next airing episode %d at timestamp %d (in %d seconds)",
		nextEpisode.Episode, nextEpisode.AiringAt, nextEpisode.TimeUntilAiring), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	return types.AnimeAiringEpisode{
		AiringAt:        nextEpisode.AiringAt,
		TimeUntilAiring: nextEpisode.TimeUntilAiring,
		Episode:         nextEpisode.Episode,
	}
}

// getAnimeSchedule extracts airing schedule data from AniList
func getAnimeSchedule(anilistAnime *anilist.AnilistAnimeResponse) []types.AnimeAiringEpisode {
	if anilistAnime == nil {
		return []types.AnimeAiringEpisode{}
	}

	var schedule []types.AnimeAiringEpisode

	// The nodes might be nil if there's no schedule
	if anilistAnime.Data.Media.AiringSchedule.Nodes == nil {
		logger.Log("No airing schedule found in AniList data", logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return []types.AnimeAiringEpisode{}
	}

	for _, node := range anilistAnime.Data.Media.AiringSchedule.Nodes {
		schedule = append(schedule, types.AnimeAiringEpisode{
			AiringAt:        node.AiringAt,
			TimeUntilAiring: node.TimeUntilAiring,
			Episode:         node.Episode,
		})
	}

	logger.Log(fmt.Sprintf("Found %d episodes in airing schedule", len(schedule)), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	return schedule
}

// AttachEpisodeDescriptions enhances episode information with external data
// Imports the function from the anime utils to use in our service
var AttachEpisodeDescriptions = tmdb.AttachEpisodeDescriptions

// extractLogosFromMALSync extracts logo images from MALSync data
func extractLogosFromMALSync(malSyncData *malsync.MALSyncAnimeResponse) types.AnimeLogos {
	logos := types.AnimeLogos{}

	// Early return if no data
	if malSyncData == nil {
		return logos
	}

	// Check if Crunchyroll data exists in the MALSync response
	crunchyrollSites, exists := malSyncData.Sites["Crunchyroll"]
	if !exists || len(crunchyrollSites) == 0 {
		logger.Log("No Crunchyroll data found in MALSync response", logger.LogOptions{
			Level:  logger.Debug,
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
		logger.Log("No valid Crunchyroll URL found", logger.LogOptions{
			Level:  logger.Debug,
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

	logger.Log(fmt.Sprintf("Successfully generated logo URLs for series ID: %s", seriesID), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	return logos
}

// extractCrunchyrollSeriesID extracts the series ID from a Crunchyroll URL
func extractCrunchyrollSeriesID(crURL string) string {
	logger.Log(fmt.Sprintf("Attempting to extract series ID from URL: %s", crURL), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	// Direct series URL format
	if strings.Contains(crURL, "/series/") {
		parts := strings.Split(crURL, "/series/")
		if len(parts) < 2 {
			logger.Log("URL contains /series/ but couldn't extract ID part", logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AnimeAPI",
			})
			return ""
		}

		idParts := strings.Split(parts[1], "/")
		if len(idParts) < 1 {
			logger.Log("Couldn't extract ID from path segments", logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AnimeAPI",
			})
			return ""
		}

		logger.Log(fmt.Sprintf("Found series ID directly in URL: %s", idParts[0]), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return idParts[0]
	}

	// Need to follow redirect to get series ID
	logger.Log("URL doesn't contain /series/, following redirect...", logger.LogOptions{
		Level:  logger.Debug,
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
		logger.Log(fmt.Sprintf("Updated URL to HTTPS: %s", crURL), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
	}

	// Add User-Agent header to mimic a browser
	req, err := http.NewRequest("GET", crURL, nil)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to create request for Crunchyroll URL: %v", err), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return ""
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")

	resp, err := client.Do(req)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get Crunchyroll redirect: %v", err), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return ""
	}
	defer resp.Body.Close()

	// Log the status code and response headers for debugging
	logger.Log(fmt.Sprintf("Crunchyroll response status: %d %s", resp.StatusCode, resp.Status), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	for name, values := range resp.Header {
		logger.Log(fmt.Sprintf("Header %s: %s", name, strings.Join(values, ", ")), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
	}

	// Check for specific status codes for redirects
	if resp.StatusCode != http.StatusMovedPermanently &&
		resp.StatusCode != http.StatusFound &&
		resp.StatusCode != http.StatusTemporaryRedirect &&
		resp.StatusCode != http.StatusPermanentRedirect {
		logger.Log(fmt.Sprintf("Unexpected status code from Crunchyroll: %d", resp.StatusCode), logger.LogOptions{
			Level:  logger.Debug,
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
				logger.Log(fmt.Sprintf("Extracted potential series ID from original URL: %s", potentialId), logger.LogOptions{
					Level:  logger.Debug,
					Prefix: "AnimeAPI",
				})
				return potentialId
			}
		}
		return ""
	}

	redirectURL := resp.Header.Get("Location")
	if redirectURL == "" {
		logger.Log("No redirect URL found in Crunchyroll response", logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return ""
	}

	logger.Log(fmt.Sprintf("Found redirect URL: %s", redirectURL), logger.LogOptions{
		Level:  logger.Debug,
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

		logger.Log(fmt.Sprintf("Successfully extracted series ID from redirect: %s", idParts[0]), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return idParts[0]
	}

	// For multi-level redirects, try to follow one more time
	if strings.Contains(redirectURL, "crunchyroll.com") {
		logger.Log("Trying to follow one more redirect level...", logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return extractCrunchyrollSeriesID(redirectURL)
	}

	// As a fallback for older Crunchyroll URLs like fullmetal-alchemist-brotherhood
	// Use the last path segment as the ID
	urlParts := strings.Split(crURL, "/")
	if len(urlParts) > 0 {
		potentialId := urlParts[len(urlParts)-1]
		logger.Log(fmt.Sprintf("Using fallback: extracted ID from original URL: %s", potentialId), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		return potentialId
	}

	logger.Log("Could not extract series ID from Crunchyroll redirect URL", logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})
	return ""
}

// FindSeasonMappings finds all anime mappings that belong to the same series based on TVDB ID
var FindSeasonMappings = api.FindSeasonMappings
