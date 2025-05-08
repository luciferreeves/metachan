package anime

import (
	"fmt"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/api"
	"metachan/utils/concurrency"
	"metachan/utils/logger"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Service provides high-level operations for anime data
type Service struct {
	jikanClient     *api.JikanClient
	streamingClient *api.AllAnimeClient
	anilistClient   *api.AniListClient
	malsyncClient   *api.MALSyncClient
}

// NewService creates a new anime service
func NewService() *Service {
	return &Service{
		jikanClient:     api.NewJikanClient(),
		streamingClient: api.NewAllAnimeClient(),
		anilistClient:   api.NewAniListClient(),
		malsyncClient:   api.NewMALSyncClient(),
	}
}

// GetAnimeDetails fetches comprehensive anime details
func (s *Service) GetAnimeDetails(mapping *entities.AnimeMapping) (*types.Anime, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Log(fmt.Sprintf("GetAnimeDetails total execution time: %s", duration), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
	}()

	malID := mapping.MAL

	// Create the different types of functions for proper Go generic type inference
	animeFunc := func() (*types.JikanAnimeResponse, error) {
		return s.jikanClient.GetFullAnime(malID)
	}

	episodesFunc := func() (*types.JikanAnimeEpisodeResponse, error) {
		return s.jikanClient.GetAnimeEpisodes(malID)
	}

	charactersFunc := func() (*types.JikanAnimeCharacterResponse, error) {
		return s.jikanClient.GetAnimeCharacters(malID)
	}

	fetchStartTime := time.Now()
	// Use separate results variables for each type
	animeResult := concurrency.Parallel(animeFunc)[0]
	episodesResult := concurrency.Parallel(episodesFunc)[0]
	charactersResult := concurrency.Parallel(charactersFunc)[0]
	logger.Log(fmt.Sprintf("Initial parallel API fetch time: %s", time.Since(fetchStartTime)), types.LogOptions{
		Level:  types.Debug,
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
	var anilistAnime *types.AnilistAnimeResponse
	var malSyncData *types.MALSyncAnimeResponse

	anilistStartTime := time.Now()
	if mapping.Anilist != 0 {
		// We need separate functions for each type for proper type inference
		anilistFunc := func() (*types.AnilistAnimeResponse, error) {
			return s.anilistClient.GetAnime(mapping.Anilist)
		}

		malsyncFunc := func() (*types.MALSyncAnimeResponse, error) {
			return s.malsyncClient.GetAnimeByMALID(malID)
		}

		// Execute them separately to avoid type errors
		anilistResult := concurrency.Parallel(anilistFunc)[0]
		malsyncResult := concurrency.Parallel(malsyncFunc)[0]

		// Extract AniList result
		if anilistResult.Error == nil {
			anilistAnime = anilistResult.Value
			logger.Log(fmt.Sprintf("Successfully fetched AniList data for ID %d", mapping.Anilist), types.LogOptions{
				Level:  types.Debug,
				Prefix: "AnimeAPI",
			})
		} else {
			logger.Log(fmt.Sprintf("Failed to fetch AniList data: %v", anilistResult.Error), types.LogOptions{
				Level:  types.Warn,
				Prefix: "AnimeAPI",
			})
		}

		// Extract MALSync result
		if malsyncResult.Error == nil {
			malSyncData = malsyncResult.Value
		} else {
			logger.Log(fmt.Sprintf("Failed to fetch MALSync data: %v", malsyncResult.Error), types.LogOptions{
				Level:  types.Warn,
				Prefix: "AnimeAPI",
			})
		}
	} else {
		logger.Log(fmt.Sprintf("No AniList ID available for MAL ID %d", malID), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		// If no AniList ID, just fetch MALSync data
		malSyncData, _ = s.malsyncClient.GetAnimeByMALID(malID)
	}
	logger.Log(fmt.Sprintf("AniList and MALSync fetch time: %s", time.Since(anilistStartTime)), types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})

	// Process episode data in parallel with seasons and other operations
	episodeDataChan := make(chan []types.AnimeSingleEpisode, 1)
	subbedCountChan := make(chan int, 1)
	dubbedCountChan := make(chan int, 1)

	episodeProcessingStartTime := time.Now()
	go func() {
		defer close(episodeDataChan)
		defer close(subbedCountChan)
		defer close(dubbedCountChan)

		basicEpisodes := generateBasicEpisodes(episodes.Data)
		logger.Log(fmt.Sprintf("Generated basic episodes: %d", len(basicEpisodes)), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})

		// Enrich episodes with TMDB data
		logger.Log(fmt.Sprintf("Starting enrichEpisodes for %d episodes", len(basicEpisodes)), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		enrichStart := time.Now()

		enrichedEpisodes := AttachEpisodeDescriptions(anime.Data.Title, basicEpisodes, anime.Data.TitleEnglish, mapping.TMDB)

		// Get subbed and dubbed episode counts in bulk with a single API call (much faster)
		subCount, dubCount := 0, 0
		searchTitle := anime.Data.Title

		startStreamingCheck := time.Now()
		logger.Log("Fetching streaming episode counts...", types.LogOptions{
			Level:  types.Debug,
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
		}

		logger.Log(fmt.Sprintf("Streaming count check took %s. Subbed: %d, Dubbed: %d",
			time.Since(startStreamingCheck), subCount, dubCount), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})

		logger.Log(fmt.Sprintf("enrichEpisodes execution time: %s", time.Since(enrichStart)), types.LogOptions{
			Level:  types.Debug,
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
		logger.Log(fmt.Sprintf("Finding season mappings for TVDB ID %d", mapping.TVDB), types.LogOptions{
			Level:  types.Debug,
			Prefix: "TVDB",
		})
		seasonMappings, err := FindSeasonMappings(mapping.TVDB)
		if err == nil && len(seasonMappings) > 0 {
			logger.Log(fmt.Sprintf("Found %d season mappings for TVDB ID %d", len(seasonMappings), mapping.TVDB), types.LogOptions{
				Level:  types.Debug,
				Prefix: "TVDB",
			})
			seasons = s.getSeasonDetails(&seasonMappings, malID)
		}
	}
	logger.Log(fmt.Sprintf("Seasons fetch time: %s", time.Since(seasonsStartTime)), types.LogOptions{
		Level:  types.Debug,
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
	logger.Log(fmt.Sprintf("Waiting for episode data processing (started %s ago)", time.Since(episodeProcessingStartTime)), types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})
	episodeWaitStartTime := time.Now()
	episodeData := <-episodeDataChan
	subbedCount := <-subbedCountChan
	dubbedCount := <-dubbedCountChan
	logger.Log(fmt.Sprintf("Episode data wait time: %s (total episode processing time: %s)",
		time.Since(episodeWaitStartTime),
		time.Since(episodeProcessingStartTime)), types.LogOptions{
		Level:  types.Debug,
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
		logger.Log("Setting covers and color from AniList data", types.LogOptions{
			Level:  types.Debug,
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
			logger.Log(fmt.Sprintf("Set color to: %s", coverImage.Color), types.LogOptions{
				Level:  types.Debug,
				Prefix: "AnimeAPI",
			})
		}
	} else {
		logger.Log("No valid AniList data available for covers and color", types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
	}

	return animeDetails, nil
}

// enrichEpisodes adds streaming sources to episodes in parallel
func (s *Service) enrichEpisodes(
	episodes []types.AnimeSingleEpisode,
	title string,
	alternativeTitle string,
	tmdbID int,
	malID int,
) []types.AnimeSingleEpisode {
	enrichStart := time.Now()
	defer func() {
		logger.Log(fmt.Sprintf("Full enrichEpisodes function time: %s", time.Since(enrichStart)), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
	}()

	// We can still enrich with TMDB data if needed
	tmdbStart := time.Now()
	enrichedEpisodes := AttachEpisodeDescriptions(title, episodes, alternativeTitle, tmdbID)
	logger.Log(fmt.Sprintf("TMDB episode descriptions fetch time: %s", time.Since(tmdbStart)), types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})

	// Early return if no episodes
	if len(enrichedEpisodes) == 0 {
		return enrichedEpisodes
	}

	// Process streaming sources in parallel - use a reasonable number of workers
	const maxConcurrentStreaming = 8
	streamingChan := make(chan types.EpisodeStreamingResult, len(enrichedEpisodes))
	streamingStart := time.Now()

	// Create a worker pool
	var wg sync.WaitGroup
	workerCh := make(chan int, maxConcurrentStreaming)

	// Track stats for streaming fetches
	var streamingSuccessCount int32
	var streamingFailCount int32
	var streamingRetryCount int32

	// Count of subbed and dubbed episodes
	var subbedCount int32
	var dubbedCount int32

	for i := 0; i < len(enrichedEpisodes); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Get a worker slot or block
			workerCh <- 1
			defer func() { <-workerCh }()

			episodeStart := time.Now()

			// Try to get streaming sources
			episodeNumber := idx + 1
			searchTitle := title

			streaming, err := s.streamingClient.GetStreamingSources(searchTitle, episodeNumber)
			retryCount := 0

			// If search fails with romaji title, try with Japanese title if available
			if err != nil && enrichedEpisodes[idx].Titles.Japanese != "" {
				retryCount++
				atomic.AddInt32(&streamingRetryCount, 1)
				streaming, err = s.streamingClient.GetStreamingSources(enrichedEpisodes[idx].Titles.Japanese, episodeNumber)

				// If still fails and English title is available, try with that
				if err != nil && alternativeTitle != "" {
					retryCount++
					atomic.AddInt32(&streamingRetryCount, 1)
					englishTitle := strings.TrimPrefix(alternativeTitle, "English: ")
					streaming, err = s.streamingClient.GetStreamingSources(englishTitle, episodeNumber)
				}
			}

			// Send the result back whether successful or not
			if err != nil {
				atomic.AddInt32(&streamingFailCount, 1)
				streamingChan <- types.EpisodeStreamingResult{
					EpisodeNumber: episodeNumber,
					Streaming: &types.AnimeStreaming{
						Sub: []types.AnimeStreamingSource{},
						Dub: []types.AnimeStreamingSource{},
					},
				}
			} else {
				atomic.AddInt32(&streamingSuccessCount, 1)
				// Update subbed and dubbed counts based on what we found
				if len(streaming.Sub) > 0 {
					atomic.AddInt32(&subbedCount, 1)
				}
				if len(streaming.Dub) > 0 {
					atomic.AddInt32(&dubbedCount, 1)
				}
				streamingChan <- types.EpisodeStreamingResult{
					EpisodeNumber: episodeNumber,
					Streaming:     streaming,
				}
			}

			if episodeNumber%5 == 0 || episodeNumber == len(enrichedEpisodes) {
				logger.Log(fmt.Sprintf("Episode %d/%d streaming fetch time: %s (retries: %d)",
					episodeNumber, len(enrichedEpisodes), time.Since(episodeStart), retryCount), types.LogOptions{
					Level:  types.Debug,
					Prefix: "AnimeAPI",
				})
			}
		}(i)
	}

	// Wait for all workers to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(streamingChan)
		logger.Log(fmt.Sprintf("All streaming workers finished. Success: %d, Failed: %d, Retries: %d, Subbed: %d, Dubbed: %d",
			atomic.LoadInt32(&streamingSuccessCount),
			atomic.LoadInt32(&streamingFailCount),
			atomic.LoadInt32(&streamingRetryCount),
			atomic.LoadInt32(&subbedCount),
			atomic.LoadInt32(&dubbedCount)), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
	}()

	// Collect results
	resultCount := 0
	for range streamingChan {
		resultCount++
		// Simply count the received results

		// Log progress periodically
		if resultCount%10 == 0 || resultCount == len(enrichedEpisodes) {
			logger.Log(fmt.Sprintf("Processed %d/%d streaming results (%s elapsed)",
				resultCount, len(enrichedEpisodes), time.Since(streamingStart)), types.LogOptions{
				Level:  types.Debug,
				Prefix: "AnimeAPI",
			})
		}
	}

	logger.Log(fmt.Sprintf("Total streaming fetch time for %d episodes: %s, Subbed: %d, Dubbed: %d",
		len(enrichedEpisodes), time.Since(streamingStart),
		atomic.LoadInt32(&subbedCount),
		atomic.LoadInt32(&dubbedCount)), types.LogOptions{
		Level:  types.Debug,
		Prefix: "AnimeAPI",
	})

	// Return the enriched episodes and the counts
	return enrichedEpisodes
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
