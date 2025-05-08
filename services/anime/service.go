package anime

import (
	"fmt"
	"metachan/database"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/api/anilist"
	"metachan/utils/api/jikan"
	"metachan/utils/api/malsync"
	"metachan/utils/api/streaming"
	"metachan/utils/concurrency"
	"metachan/utils/logger"
	"strings"
	"time"
)

// Service provides high-level operations for anime data
type Service struct {
	jikanClient     *jikan.JikanClient
	streamingClient *streaming.AllAnimeClient
	anilistClient   *anilist.AniListClient
	malsyncClient   *malsync.MALSyncClient
}

// NewService creates a new anime service
func NewService() *Service {
	return &Service{
		jikanClient:     jikan.NewJikanClient(),
		streamingClient: streaming.NewAllAnimeClient(),
		anilistClient:   anilist.NewAniListClient(),
		malsyncClient:   malsync.NewMALSyncClient(),
	}
}

// GetAnimeDetailsWithSource fetches comprehensive anime details with source information
func (s *Service) GetAnimeDetailsWithSource(mapping *entities.AnimeMapping, source string) (*types.Anime, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Log(fmt.Sprintf("GetAnimeDetails (%s) execution time: %s", source, duration), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
	}()

	malID := mapping.MAL

	// For updater source, always fetch fresh data and skip cache
	if source != "updater" {
		// First, check if we have a valid cached version
		cachedAnime, err := database.GetCachedAnimeByMALID(malID)
		if err == nil && database.IsCacheValid(cachedAnime) {
			logger.Log(fmt.Sprintf("Cache hit for anime (MAL ID: %d), returning cached data", malID), logger.LogOptions{
				Level:  logger.Info,
				Prefix: "AnimeCache",
			})

			// Convert the cached anime to the types.Anime format
			return database.ConvertToTypesAnime(cachedAnime), nil
		}
	} else {
		logger.Log(fmt.Sprintf("Bypassing cache for anime (MAL ID: %d) - source: %s", malID, source), logger.LogOptions{
			Level:  logger.Info,
			Prefix: "AnimeAPI",
		})
	}

	// Rest of the implementation is the same as GetAnimeDetails
	logger.Log(fmt.Sprintf("Cache miss for anime (MAL ID: %d), fetching fresh data", malID), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AnimeAPI",
	})

	// Create the different types of functions for proper Go generic type inference
	animeFunc := func() (*jikan.JikanAnimeResponse, error) {
		return s.jikanClient.GetFullAnime(malID)
	}

	episodesFunc := func() (*jikan.JikanAnimeEpisodeResponse, error) {
		return s.jikanClient.GetAnimeEpisodes(malID)
	}

	charactersFunc := func() (*jikan.JikanAnimeCharacterResponse, error) {
		return s.jikanClient.GetAnimeCharacters(malID)
	}

	fetchStartTime := time.Now()
	// Use separate results variables for each type
	animeResult := concurrency.Parallel(animeFunc)[0]
	episodesResult := concurrency.Parallel(episodesFunc)[0]
	charactersResult := concurrency.Parallel(charactersFunc)[0]
	logger.Log(fmt.Sprintf("Initial parallel API fetch time: %s", time.Since(fetchStartTime)), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	// Extract results and handle errors
	anime := animeResult.Value
	if animeResult.Error != nil {
		return nil, fmt.Errorf("failed to get anime details: %w", animeResult.Error)
	}

	episodes := episodesResult.Value
	if episodesResult.Error != nil {
		return nil, fmt.Errorf("failed to get anime episodes: %w", episodesResult.Error)
	}

	characterResponse := charactersResult.Value
	if charactersResult.Error != nil {
		return nil, fmt.Errorf("failed to get anime characters: %w", charactersResult.Error)
	}

	// Get Anilist and MALSync data in parallel if available
	var anilistAnime *anilist.AnilistAnimeResponse
	var malSyncData *malsync.MALSyncAnimeResponse

	anilistStartTime := time.Now()
	if mapping.Anilist != 0 {
		// We need separate functions for each type for proper type inference
		anilistFunc := func() (*anilist.AnilistAnimeResponse, error) {
			return s.anilistClient.GetAnime(mapping.Anilist)
		}

		malsyncFunc := func() (*malsync.MALSyncAnimeResponse, error) {
			return s.malsyncClient.GetAnimeByMALID(malID)
		}

		// Execute them separately to avoid type errors
		anilistResult := concurrency.Parallel(anilistFunc)[0]
		malsyncResult := concurrency.Parallel(malsyncFunc)[0]

		// Extract AniList result
		if anilistResult.Error == nil {
			anilistAnime = anilistResult.Value
			logger.Log(fmt.Sprintf("Successfully fetched AniList data for ID %d", mapping.Anilist), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AnimeAPI",
			})
		} else {
			logger.Log(fmt.Sprintf("Failed to fetch AniList data: %v", anilistResult.Error), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AnimeAPI",
			})
		}

		// Extract MALSync result
		if malsyncResult.Error == nil {
			malSyncData = malsyncResult.Value
		} else {
			logger.Log(fmt.Sprintf("Failed to fetch MALSync data: %v", malsyncResult.Error), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AnimeAPI",
			})
		}
	} else {
		logger.Log(fmt.Sprintf("No AniList ID available for MAL ID %d", malID), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		// If no AniList ID, just fetch MALSync data
		malSyncData, _ = s.malsyncClient.GetAnimeByMALID(malID)
	}
	logger.Log(fmt.Sprintf("AniList and MALSync fetch time: %s", time.Since(anilistStartTime)), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	// Process episode data in parallel with seasons and other operations
	episodeDataChan := make(chan []types.AnimeSingleEpisode, 1)
	subbedCountChan := make(chan int, 1)
	dubbedCountChan := make(chan int, 1)
	tmdbErrorChan := make(chan error, 1)

	episodeProcessingStartTime := time.Now()
	go func() {
		defer close(episodeDataChan)
		defer close(subbedCountChan)
		defer close(dubbedCountChan)
		defer close(tmdbErrorChan)

		basicEpisodes := generateBasicEpisodes(episodes.Data)
		logger.Log(fmt.Sprintf("Generated basic episodes: %d", len(basicEpisodes)), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})

		// Enrich episodes with TMDB data
		logger.Log(fmt.Sprintf("Starting enrichEpisodes for %d episodes", len(basicEpisodes)), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
		enrichStart := time.Now()

		enrichedEpisodes, tmdbErr := AttachEpisodeDescriptions(anime.Data.Title, basicEpisodes, anime.Data.TitleEnglish, mapping.TMDB)
		tmdbErrorChan <- tmdbErr

		// Get subbed and dubbed episode counts in bulk with a single API call (much faster)
		subCount, dubCount := 0, 0
		searchTitle := anime.Data.Title

		startStreamingCheck := time.Now()
		logger.Log("Fetching streaming episode counts...", logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})

		var err error
		subCount, dubCount, err = s.streamingClient.GetStreamingCounts(searchTitle)

		// If the first title fails, try with English title
		if err != nil && anime.Data.TitleEnglish != "" {
			englishTitle := strings.TrimPrefix(anime.Data.TitleEnglish, "English: ")
			subCount, dubCount, err = s.streamingClient.GetStreamingCounts(englishTitle)
		}

		// If English title fails, try with Japanese title
		if err != nil && anime.Data.TitleJapanese != "" {
			subCount, dubCount, err = s.streamingClient.GetStreamingCounts(anime.Data.TitleJapanese)

			// Log the final error if all attempts failed
			if err != nil {
				logger.Log(fmt.Sprintf("Failed to fetch streaming counts after all attempts: %v", err), logger.LogOptions{
					Level:  logger.Warn,
					Prefix: "AnimeAPI",
				})
			}
		}

		logger.Log(fmt.Sprintf("Streaming count check took %s. Subbed: %d, Dubbed: %d",
			time.Since(startStreamingCheck), subCount, dubCount), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})

		logger.Log(fmt.Sprintf("enrichEpisodes execution time: %s", time.Since(enrichStart)), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})

		episodeDataChan <- enrichedEpisodes
		subbedCountChan <- subCount
		dubbedCountChan <- dubCount
	}()

	// Get seasons information if TVDB ID is available
	seasonsStartTime := time.Now()
	var seasons []types.AnimeSeason
	if mapping.TVDB != 0 {
		logger.Log(fmt.Sprintf("Finding season mappings for TVDB ID %d", mapping.TVDB), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "TVDB",
		})
		seasonMappings, err := FindSeasonMappings(mapping.TVDB)
		if err == nil && len(seasonMappings) > 0 {
			logger.Log(fmt.Sprintf("Found %d season mappings for TVDB ID %d", len(seasonMappings), mapping.TVDB), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "TVDB",
			})
			seasons = s.getSeasonDetails(&seasonMappings, malID)
		}
	}
	logger.Log(fmt.Sprintf("Seasons fetch time: %s", time.Since(seasonsStartTime)), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	// Get logos data from MALSync data
	logos := extractLogosFromMALSync(malSyncData)

	// Extract character data
	characters := getAnimeCharacters(characterResponse)

	// Extract episode count, next airing episode, and schedule
	var nextAiringEpisode types.AnimeAiringEpisode
	var schedule []types.AnimeAiringEpisode

	if anilistAnime != nil {
		nextAiringEpisode = getNextAiringEpisode(anilistAnime)
		schedule = getAnimeSchedule(anilistAnime)
	}

	// Wait for episode data to complete
	logger.Log(fmt.Sprintf("Waiting for episode data processing (started %s ago)", time.Since(episodeProcessingStartTime)), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})
	episodeWaitStartTime := time.Now()
	episodeData := <-episodeDataChan
	subbedCount := <-subbedCountChan
	dubbedCount := <-dubbedCountChan
	tmdbError := <-tmdbErrorChan
	logger.Log(fmt.Sprintf("Episode data wait time: %s (total episode processing time: %s)",
		time.Since(episodeWaitStartTime),
		time.Since(episodeProcessingStartTime)), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeAPI",
	})

	// Assemble final anime details
	animeDetails := &types.Anime{
		MALID: malID,
		Titles: types.AnimeTitles{
			Romaji:   anime.Data.Title,
			English:  anime.Data.TitleEnglish,
			Japanese: anime.Data.TitleJapanese,
			Synonyms: anime.Data.TitleSynonyms,
		},
		Synopsis: anime.Data.Synopsis,
		Type:     types.AniSyncType(mapping.Type),
		Source:   anime.Data.Source,
		Airing:   anime.Data.Airing,
		Status:   anime.Data.Status,
		AiringStatus: types.AiringStatus{
			From: types.AiringStatusDates{
				Day:    anime.Data.Aired.Prop.From.Day,
				Month:  anime.Data.Aired.Prop.From.Month,
				Year:   anime.Data.Aired.Prop.From.Year,
				String: anime.Data.Aired.From,
			},
			To: types.AiringStatusDates{
				Day:    anime.Data.Aired.Prop.To.Day,
				Month:  anime.Data.Aired.Prop.To.Month,
				Year:   anime.Data.Aired.Prop.To.Year,
				String: anime.Data.Aired.To,
			},
			String: anime.Data.Aired.String,
		},
		Duration: anime.Data.Duration,
		Images: types.AnimeImages{
			Small:    anime.Data.Images.JPG.SmallImageURL,
			Large:    anime.Data.Images.JPG.LargeImageURL,
			Original: anime.Data.Images.JPG.ImageURL,
		},
		Logos:  logos,
		Covers: types.AnimeImages{},
		Color:  "",
		Genres: generateGenres(anime.Data.Genres, anime.Data.ExplicitGenres),
		Scores: types.AnimeScores{
			Score:      anime.Data.Score,
			ScoredBy:   anime.Data.ScoredBy,
			Rank:       anime.Data.Rank,
			Popularity: anime.Data.Popularity,
			Members:    anime.Data.Members,
			Favorites:  anime.Data.Favorites,
		},
		Season: anime.Data.Season,
		Year:   anime.Data.Year,
		Broadcast: types.AnimeBroadcast{
			Day:      anime.Data.Broadcast.Day,
			Time:     anime.Data.Broadcast.Time,
			Timezone: anime.Data.Broadcast.Timezone,
			String:   anime.Data.Broadcast.String,
		},
		Producers: generateProducers(anime.Data.Producers),
		Studios:   generateStudios(anime.Data.Studios),
		Licensors: generateLicensors(anime.Data.Licensors),
		Seasons:   seasons,
		Episodes: types.AnimeEpisodes{
			Total:    getEpisodeCount(anime, anilistAnime),
			Aired:    len(episodes.Data),
			Subbed:   subbedCount,
			Dubbed:   dubbedCount,
			Episodes: episodeData,
		},
		NextAiringEpisode: nextAiringEpisode,
		AiringSchedule:    schedule,
		Characters:        characters,
		Mappings: types.AnimeMappings{
			AniDB:          mapping.AniDB,
			Anilist:        mapping.Anilist,
			AnimeCountdown: mapping.AnimeCountdown,
			AnimePlanet:    mapping.AnimePlanet,
			AniSearch:      mapping.AniSearch,
			IMDB:           mapping.IMDB,
			Kitsu:          mapping.Kitsu,
			LiveChart:      mapping.LiveChart,
			NotifyMoe:      mapping.NotifyMoe,
			Simkl:          mapping.Simkl,
			TMDB:           mapping.TMDB,
			TVDB:           mapping.TVDB,
		},
	}

	// Add AniList cover images and color if available
	if anilistAnime != nil && anilistAnime.Data.Media.ID > 0 {
		logger.Log("Setting covers and color from AniList data", logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})

		// Create debug logs for the data
		coverImage := anilistAnime.Data.Media.CoverImage

		// Explicitly set the cover images, ensuring we don't have empty values
		animeDetails.Covers = types.AnimeImages{
			Small:    coverImage.Medium,
			Large:    coverImage.Large,
			Original: coverImage.ExtraLarge,
		}

		// For color, also make sure it's not empty
		if coverImage.Color != "" {
			animeDetails.Color = coverImage.Color
			logger.Log(fmt.Sprintf("Set color to: %s", coverImage.Color), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AnimeAPI",
			})
		}
	} else {
		logger.Log("No valid AniList data available for covers and color", logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeAPI",
		})
	}

	// Save the anime to cache only if TMDB didn't fail
	if tmdbError == nil {
		go func() {
			if err := database.SaveAnimeToCache(animeDetails); err != nil {
				logger.Log(fmt.Sprintf("Failed to save anime to cache: %v", err), logger.LogOptions{
					Level:  logger.Error,
					Prefix: "AnimeCache",
				})
			} else {
				logger.Log(fmt.Sprintf("Successfully saved anime (MAL ID: %d) to cache", malID), logger.LogOptions{
					Level:  logger.Debug,
					Prefix: "AnimeCache",
				})
			}
		}()
	} else {
		logger.Log(fmt.Sprintf("Skipping anime cache due to TMDB error: %v", tmdbError), logger.LogOptions{
			Level:  logger.Warn,
			Prefix: "AnimeCache",
		})
	}

	return animeDetails, nil
}

// GetAnimeDetails fetches comprehensive anime details
func (s *Service) GetAnimeDetails(mapping *entities.AnimeMapping) (*types.Anime, error) {
	return s.GetAnimeDetailsWithSource(mapping, "api")
}

// getSeasonDetails fetches details for anime seasons
func (s *Service) getSeasonDetails(mappings *[]entities.AnimeMapping, currentMALID int) []types.AnimeSeason {
	// Helper function to fetch anime details for a single mapping
	fetchSeason := func(mapping entities.AnimeMapping, isCurrent bool) (types.AnimeSeason, error) {
		anime, err := s.jikanClient.GetAnime(mapping.MAL)
		if err != nil {
			return types.AnimeSeason{}, err
		}

		return types.AnimeSeason{
			MALID: mapping.MAL,
			Titles: types.AnimeTitles{
				English:  anime.Data.TitleEnglish,
				Japanese: anime.Data.TitleJapanese,
				Romaji:   anime.Data.Title,
				Synonyms: anime.Data.TitleSynonyms,
			},
			Synopsis: anime.Data.Synopsis,
			Type:     types.AniSyncType(mapping.Type),
			Source:   anime.Data.Source,
			Airing:   anime.Data.Airing,
			Status:   anime.Data.Status,
			AiringStatus: types.AiringStatus{
				From: types.AiringStatusDates{
					Day:    anime.Data.Aired.Prop.From.Day,
					Month:  anime.Data.Aired.Prop.From.Month,
					Year:   anime.Data.Aired.Prop.From.Year,
					String: anime.Data.Aired.From,
				},
				To: types.AiringStatusDates{
					Day:    anime.Data.Aired.Prop.To.Day,
					Month:  anime.Data.Aired.Prop.To.Month,
					Year:   anime.Data.Aired.Prop.To.Year,
					String: anime.Data.Aired.To,
				},
				String: anime.Data.Aired.String,
			},
			Duration: anime.Data.Duration,
			Images: types.AnimeImages{
				Small:    anime.Data.Images.JPG.SmallImageURL,
				Large:    anime.Data.Images.JPG.LargeImageURL,
				Original: anime.Data.Images.JPG.ImageURL,
			},
			Scores: types.AnimeScores{
				Score:      anime.Data.Score,
				ScoredBy:   anime.Data.ScoredBy,
				Rank:       anime.Data.Rank,
				Popularity: anime.Data.Popularity,
				Members:    anime.Data.Members,
				Favorites:  anime.Data.Favorites,
			},
			Season:  anime.Data.Season,
			Year:    anime.Data.Year,
			Current: isCurrent,
		}, nil
	}

	// Fetch all seasons in parallel
	seasonFunctions := make([]func() (types.AnimeSeason, error), len(*mappings))

	for i, mapping := range *mappings {
		mapping := mapping // Capture variable for closure
		isCurrent := mapping.MAL == currentMALID

		seasonFunctions[i] = func() (types.AnimeSeason, error) {
			return fetchSeason(mapping, isCurrent)
		}
	}

	// Execute in parallel
	results := concurrency.Parallel(seasonFunctions...)

	// Extract successful results
	var seasons []types.AnimeSeason
	for _, result := range results {
		if result.Error == nil {
			seasons = append(seasons, result.Value)
		}
	}

	// Sort seasons chronologically by air date
	if len(seasons) > 1 {
		sortSeasonsByAirDate(&seasons)
	}

	return seasons
}
