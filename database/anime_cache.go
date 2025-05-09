package database

import (
	"fmt"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	// CacheExpirationTime represents the duration after which the cache is considered stale
	CacheExpirationTime = 24 * time.Hour
)

// GetCachedAnimeByMALID retrieves an anime from the cache by MAL ID
func GetCachedAnimeByMALID(malID int) (*entities.CachedAnime, error) {
	var anime entities.CachedAnime

	// Query the anime with all its relationships
	err := DB.
		Preload("Images").
		Preload("Logos").
		Preload("Covers").
		Preload("Scores").
		Preload("AiringStatus").
		Preload("AiringStatus.From").
		Preload("AiringStatus.To").
		Preload("Broadcast").
		Preload("NextAiringEpisode").
		Preload("Genres").
		Preload("Producers").
		Preload("Studios").
		Preload("Licensors").
		Preload("Episodes").
		Preload("Episodes.Titles").
		Preload("Characters").
		Preload("Characters.VoiceActors").
		Preload("AiringSchedule").
		Preload("Seasons").
		Preload("Seasons.Images").
		Preload("Seasons.Scores").
		Preload("Seasons.AiringStatus").
		Preload("Seasons.AiringStatus.From").
		Preload("Seasons.AiringStatus.To").
		Where("mal_id = ?", malID).
		First(&anime).Error

	if err != nil {
		return nil, err
	}

	return &anime, nil
}

// IsCacheValid checks if the cache is still valid based on the last update time
func IsCacheValid(anime *entities.CachedAnime) bool {
	if anime == nil {
		return false
	}

	// Check if the cache has expired
	return time.Since(anime.LastUpdated) < CacheExpirationTime
}

// SaveAnimeToCache saves or updates an anime in the cache
func SaveAnimeToCache(animeData *types.Anime) error {
	// For SQLite, add retry logic to handle database locks
	var maxRetries = 5
	var retryDelay = 500 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := saveAnimeToCacheWithRetry(animeData)
		if err == nil {
			logger.Log(fmt.Sprintf("Successfully saved anime (MAL ID: %d) to cache", animeData.MALID), logger.LogOptions{
				Level:  logger.Success,
				Prefix: "AnimeCache",
			})
			return nil
		}

		// Check if it's a database lock error
		if strings.Contains(err.Error(), "database is locked") {
			logger.Log(fmt.Sprintf("Database locked (attempt %d/%d) for MAL ID %d: %v. Retrying in %v...",
				attempt, maxRetries, animeData.MALID, err, retryDelay), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AnimeCache",
			})

			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		// Non-lock error, just return it
		return err
	}

	return fmt.Errorf("failed to save anime (MAL ID: %d) after %d retries: database is locked",
		animeData.MALID, maxRetries)
}

// saveAnimeToCacheWithRetry is the internal implementation of SaveAnimeToCache
func saveAnimeToCacheWithRetry(animeData *types.Anime) error {
	// Start a transaction
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// Check if anime already exists in cache
	var existingAnime entities.CachedAnime
	result := tx.Where("mal_id = ?", animeData.MALID).First(&existingAnime)

	// If exists, delete the existing record and all its relations to avoid duplicates
	if result.Error == nil {
		// First directly delete the record with raw SQL to bypass GORM's hooks
		// which might be causing issues with constraints
		if err := tx.Exec("DELETE FROM cached_animes WHERE mal_id = ?", animeData.MALID).Error; err != nil {
			logger.Log(fmt.Sprintf("Failed to delete existing anime with direct SQL: %v", err), logger.LogOptions{
				Level:  logger.Error,
				Prefix: "AnimeCache",
			})
			tx.Rollback()
			return err
		}

		// Then also try the standard deleteExistingAnimeCache to clean up related records
		if err := deleteExistingAnimeCache(tx, existingAnime.ID); err != nil {
			logger.Log(fmt.Sprintf("Warning: Issue with deleteExistingAnimeCache: %v", err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AnimeCache",
			})
			// Don't rollback here, we already deleted the main record which should address the constraint
		}
	}

	// Create new cached anime
	cachedAnime := convertToCachedAnime(animeData)

	// Clear the ID to ensure we're creating a new record
	cachedAnime.ID = 0

	// Save the main anime record
	if err := tx.Create(&cachedAnime).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Log at debug level to avoid duplicate success messages
	logger.Log(fmt.Sprintf("Successfully saved anime (MAL ID: %d) to cache", animeData.MALID), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeCache",
	})

	return nil
}

// deleteExistingAnimeCache deletes an anime and all its relations from the cache
func deleteExistingAnimeCache(tx *gorm.DB, animeID uint) error {
	// Define table structures with their foreign keys
	tableRelations := map[string]struct {
		Table      string
		ForeignKey string
		Special    bool // If true, needs special handling
	}{
		"cached_anime_voice_actors":    {"cached_anime_voice_actors", "character_id", true}, // Links to characters, not anime directly
		"cached_anime_characters":      {"cached_anime_characters", "anime_id", false},
		"cached_episode_titles":        {"cached_episode_titles", "episode_id", true}, // Links to episodes, not anime directly
		"cached_anime_single_episodes": {"cached_anime_single_episodes", "anime_id", false},
		"cached_next_episodes":         {"cached_next_episodes", "anime_id", false},
		"cached_schedule_episodes":     {"cached_schedule_episodes", "anime_id", false},
		"cached_anime_licensors":       {"cached_anime_licensors", "anime_id", false},
		"cached_anime_studios":         {"cached_anime_studios", "anime_id", false},
		"cached_anime_producers":       {"cached_anime_producers", "anime_id", false},
		"cached_anime_genres":          {"cached_anime_genres", "anime_id", false},
		"cached_anime_broadcasts":      {"cached_anime_broadcasts", "anime_id", false},
		"cached_airing_status_dates":   {"cached_airing_status_dates", "airing_status_id", true}, // Links to airing status, not anime directly
		"cached_airing_statuses":       {"cached_airing_statuses", "anime_id", false},
		"cached_anime_scores":          {"cached_anime_scores", "anime_id", false},
		"cached_anime_covers":          {"cached_anime_covers", "anime_id", false},
		"cached_anime_logos":           {"cached_anime_logos", "anime_id", false},
		"cached_anime_images":          {"cached_anime_images", "anime_id", false},
		"cached_anime_seasons":         {"cached_anime_seasons", "parent_anime_id", false}, // Uses parent_anime_id
		"cached_animes":                {"cached_animes", "id", true},                      // Uses id instead of anime_id
	}

	// First, find all the character IDs associated with this anime
	var characterIDs []uint
	if err := tx.Table("cached_anime_characters").Where("anime_id = ?", animeID).Pluck("id", &characterIDs).Error; err != nil {
		// If there's an error, it might be that the table doesn't exist yet
		logger.Log(fmt.Sprintf("Note: Could not find character IDs: %v", err), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeCache",
		})
	}

	// Find all episode IDs associated with this anime
	var episodeIDs []uint
	if err := tx.Table("cached_anime_single_episodes").Where("anime_id = ?", animeID).Pluck("id", &episodeIDs).Error; err != nil {
		logger.Log(fmt.Sprintf("Note: Could not find episode IDs: %v", err), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeCache",
		})
	}

	// Find all airing status IDs associated with this anime
	var airingStatusIDs []uint
	if err := tx.Table("cached_airing_statuses").Where("anime_id = ?", animeID).Pluck("id", &airingStatusIDs).Error; err != nil {
		logger.Log(fmt.Sprintf("Note: Could not find airing status IDs: %v", err), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeCache",
		})
	}

	// Delete voice actors by character IDs
	if len(characterIDs) > 0 {
		if err := tx.Where("character_id IN ?", characterIDs).Delete(&entities.CachedAnimeVoiceActor{}).Error; err != nil {
			logger.Log(fmt.Sprintf("Failed to delete voice actors: %v", err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AnimeCache",
			})
		}
	}

	// Delete episode titles by episode IDs
	if len(episodeIDs) > 0 {
		if err := tx.Where("episode_id IN ?", episodeIDs).Delete(&entities.CachedEpisodeTitles{}).Error; err != nil {
			logger.Log(fmt.Sprintf("Failed to delete episode titles: %v", err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AnimeCache",
			})
		}
	}

	// Delete airing status dates by airing status IDs
	if len(airingStatusIDs) > 0 {
		if err := tx.Where("airing_status_id IN ?", airingStatusIDs).Delete(&entities.CachedAiringStatusDates{}).Error; err != nil {
			logger.Log(fmt.Sprintf("Failed to delete airing status dates: %v", err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AnimeCache",
			})
		}
	}

	// Now delete the remaining tables with direct anime_id or parent_anime_id references
	for name, relation := range tableRelations {
		if relation.Special {
			continue // Skip special cases we've already handled
		}

		query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", relation.Table, relation.ForeignKey)
		if err := tx.Exec(query, animeID).Error; err != nil {
			logger.Log(fmt.Sprintf("Failed to delete from %s: %v", name, err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AnimeCache",
			})
			// Continue anyway - don't stop the entire delete operation
		}
	}

	// Finally delete the anime itself
	if err := tx.Where("id = ?", animeID).Delete(&entities.CachedAnime{}).Error; err != nil {
		return fmt.Errorf("failed to delete cached anime with ID %d: %w", animeID, err)
	}

	return nil
}

// ConvertToTypesAnime converts a cached anime to the types.Anime format
func ConvertToTypesAnime(cachedAnime *entities.CachedAnime) *types.Anime {
	if cachedAnime == nil {
		return nil
	}

	anime := &types.Anime{
		MALID: cachedAnime.MALID,
		Titles: types.AnimeTitles{
			Romaji:   cachedAnime.TitleRomaji,
			English:  cachedAnime.TitleEnglish,
			Japanese: cachedAnime.TitleJapanese,
			Synonyms: strings.Split(cachedAnime.TitleSynonyms, ","),
		},
		Synopsis: cachedAnime.Synopsis,
		Type:     types.AniSyncType(cachedAnime.Type),
		Source:   cachedAnime.Source,
		Airing:   cachedAnime.Airing,
		Status:   cachedAnime.Status,
		Duration: cachedAnime.Duration,
		Color:    cachedAnime.Color,
		Season:   cachedAnime.Season,
		Year:     cachedAnime.Year,
	}

	// Fill in Images
	if cachedAnime.Images != nil {
		anime.Images = types.AnimeImages{
			Small:    cachedAnime.Images.Small,
			Large:    cachedAnime.Images.Large,
			Original: cachedAnime.Images.Original,
		}
	}

	// Fill in Logos
	if cachedAnime.Logos != nil {
		anime.Logos = types.AnimeLogos{
			Small:    cachedAnime.Logos.Small,
			Medium:   cachedAnime.Logos.Medium,
			Large:    cachedAnime.Logos.Large,
			XLarge:   cachedAnime.Logos.XLarge,
			Original: cachedAnime.Logos.Original,
		}
	}

	// Fill in Covers
	if cachedAnime.Covers != nil {
		anime.Covers = types.AnimeImages{
			Small:    cachedAnime.Covers.Small,
			Large:    cachedAnime.Covers.Large,
			Original: cachedAnime.Covers.Original,
		}
	}

	// Fill in Scores
	if cachedAnime.Scores != nil {
		anime.Scores = types.AnimeScores{
			Score:      cachedAnime.Scores.Score,
			ScoredBy:   cachedAnime.Scores.ScoredBy,
			Rank:       cachedAnime.Scores.Rank,
			Popularity: cachedAnime.Scores.Popularity,
			Members:    cachedAnime.Scores.Members,
			Favorites:  cachedAnime.Scores.Favorites,
		}
	}

	// Fill in AiringStatus
	if cachedAnime.AiringStatus != nil {
		airingStatus := types.AiringStatus{
			String: cachedAnime.AiringStatus.String,
		}

		if cachedAnime.AiringStatus.From != nil {
			airingStatus.From = types.AiringStatusDates{
				Day:    cachedAnime.AiringStatus.From.Day,
				Month:  cachedAnime.AiringStatus.From.Month,
				Year:   cachedAnime.AiringStatus.From.Year,
				String: cachedAnime.AiringStatus.From.String,
			}
		}

		if cachedAnime.AiringStatus.To != nil {
			airingStatus.To = types.AiringStatusDates{
				Day:    cachedAnime.AiringStatus.To.Day,
				Month:  cachedAnime.AiringStatus.To.Month,
				Year:   cachedAnime.AiringStatus.To.Year,
				String: cachedAnime.AiringStatus.To.String,
			}
		}

		anime.AiringStatus = airingStatus
	}

	// Fill in Broadcast
	if cachedAnime.Broadcast != nil {
		anime.Broadcast = types.AnimeBroadcast{
			Day:      cachedAnime.Broadcast.Day,
			Time:     cachedAnime.Broadcast.Time,
			Timezone: cachedAnime.Broadcast.Timezone,
			String:   cachedAnime.Broadcast.String,
		}
	}

	// Convert genres
	if len(cachedAnime.Genres) > 0 {
		anime.Genres = make([]types.AnimeGenres, len(cachedAnime.Genres))
		for i, genre := range cachedAnime.Genres {
			anime.Genres[i] = types.AnimeGenres{
				Name:    genre.Name,
				GenreID: genre.GenreID,
				URL:     genre.URL,
			}
		}
	}

	// Convert producers
	if len(cachedAnime.Producers) > 0 {
		anime.Producers = make([]types.AnimeProducer, len(cachedAnime.Producers))
		for i, producer := range cachedAnime.Producers {
			anime.Producers[i] = types.AnimeProducer{
				Name:       producer.Name,
				ProducerID: producer.ProducerID,
				URL:        producer.URL,
			}
		}
	}

	// Convert studios
	if len(cachedAnime.Studios) > 0 {
		anime.Studios = make([]types.AnimeStudio, len(cachedAnime.Studios))
		for i, studio := range cachedAnime.Studios {
			anime.Studios[i] = types.AnimeStudio{
				Name:     studio.Name,
				StudioID: studio.StudioID,
				URL:      studio.URL,
			}
		}
	}

	// Convert licensors
	if len(cachedAnime.Licensors) > 0 {
		anime.Licensors = make([]types.AnimeLicensor, len(cachedAnime.Licensors))
		for i, licensor := range cachedAnime.Licensors {
			anime.Licensors[i] = types.AnimeLicensor{
				Name:       licensor.Name,
				ProducerID: licensor.ProducerID,
				URL:        licensor.URL,
			}
		}
	}

	// Fill in NextAiringEpisode
	if cachedAnime.NextAiringEpisode != nil && cachedAnime.NextAiringEpisode.AiringAt > 0 && cachedAnime.NextAiringEpisode.Episode > 0 {
		anime.NextAiringEpisode = types.AnimeAiringEpisode{
			AiringAt: cachedAnime.NextAiringEpisode.AiringAt,
			Episode:  cachedAnime.NextAiringEpisode.Episode,
		}
	}

	// Convert airing schedule
	if len(cachedAnime.AiringSchedule) > 0 {
		anime.AiringSchedule = make([]types.AnimeAiringEpisode, len(cachedAnime.AiringSchedule))
		for i, episode := range cachedAnime.AiringSchedule {
			anime.AiringSchedule[i] = types.AnimeAiringEpisode{
				AiringAt: episode.AiringAt,
				Episode:  episode.Episode,
			}
		}
	}

	// Convert episodes
	if len(cachedAnime.Episodes) > 0 {
		anime.Episodes.Episodes = make([]types.AnimeSingleEpisode, len(cachedAnime.Episodes))

		var subCount, dubCount int
		for i, episode := range cachedAnime.Episodes {
			episodeData := types.AnimeSingleEpisode{
				Description:  episode.Description,
				Aired:        episode.Aired,
				Score:        episode.Score,
				Filler:       episode.Filler,
				Recap:        episode.Recap,
				ForumURL:     episode.ForumURL,
				URL:          episode.URL,
				ThumbnailURL: episode.ThumbnailURL,
			}

			if episode.Titles != nil {
				episodeData.Titles = types.EpisodeTitles{
					English:  episode.Titles.English,
					Japanese: episode.Titles.Japanese,
					Romaji:   episode.Titles.Romaji,
				}
			}

			anime.Episodes.Episodes[i] = episodeData
		}

		anime.Episodes.Total = len(cachedAnime.Episodes)
		anime.Episodes.Aired = len(cachedAnime.Episodes)
		anime.Episodes.Subbed = subCount
		anime.Episodes.Dubbed = dubCount
	}

	// Convert characters
	if len(cachedAnime.Characters) > 0 {
		anime.Characters = make([]types.AnimeCharacter, len(cachedAnime.Characters))
		for i, character := range cachedAnime.Characters {
			animeCharacter := types.AnimeCharacter{
				MALID:    character.MALID,
				URL:      character.URL,
				ImageURL: character.ImageURL,
				Name:     character.Name,
				Role:     character.Role,
			}

			if len(character.VoiceActors) > 0 {
				animeCharacter.VoiceActors = make([]types.AnimeVoiceActor, len(character.VoiceActors))
				for j, va := range character.VoiceActors {
					animeCharacter.VoiceActors[j] = types.AnimeVoiceActor{
						MALID:    va.MALID,
						URL:      va.URL,
						Image:    va.Image,
						Name:     va.Name,
						Language: va.Language,
					}
				}
			}

			anime.Characters[i] = animeCharacter
		}
	}

	// Convert seasons
	if len(cachedAnime.Seasons) > 0 {
		anime.Seasons = make([]types.AnimeSeason, len(cachedAnime.Seasons))
		for i, season := range cachedAnime.Seasons {
			animeSeason := types.AnimeSeason{
				MALID:    season.MALID,
				Synopsis: season.Synopsis,
				Type:     types.AniSyncType(season.Type),
				Source:   season.Source,
				Airing:   season.Airing,
				Status:   season.Status,
				Duration: season.Duration,
				Season:   season.Season,
				Year:     season.Year,
				Current:  season.Current,
				Titles: types.AnimeTitles{
					Romaji:   season.TitleRomaji,
					English:  season.TitleEnglish,
					Japanese: season.TitleJapanese,
					Synonyms: strings.Split(season.TitleSynonyms, ","),
				},
			}

			if season.Images != nil {
				animeSeason.Images = types.AnimeImages{
					Small:    season.Images.Small,
					Large:    season.Images.Large,
					Original: season.Images.Original,
				}
			}

			if season.Scores != nil {
				animeSeason.Scores = types.AnimeScores{
					Score:      season.Scores.Score,
					ScoredBy:   season.Scores.ScoredBy,
					Rank:       season.Scores.Rank,
					Popularity: season.Scores.Popularity,
					Members:    season.Scores.Members,
					Favorites:  season.Scores.Favorites,
				}
			}

			if season.AiringStatus != nil {
				airingStatus := types.AiringStatus{
					String: season.AiringStatus.String,
				}

				if season.AiringStatus.From != nil {
					airingStatus.From = types.AiringStatusDates{
						Day:    season.AiringStatus.From.Day,
						Month:  season.AiringStatus.From.Month,
						Year:   season.AiringStatus.From.Year,
						String: season.AiringStatus.From.String,
					}
				}

				if season.AiringStatus.To != nil {
					airingStatus.To = types.AiringStatusDates{
						Day:    season.AiringStatus.To.Day,
						Month:  season.AiringStatus.To.Month,
						Year:   season.AiringStatus.To.Year,
						String: season.AiringStatus.To.String,
					}
				}

				animeSeason.AiringStatus = airingStatus
			}

			anime.Seasons[i] = animeSeason
		}
	}

	return anime
}

// convertToCachedAnime converts from types.Anime to entities.CachedAnime
func convertToCachedAnime(animeData *types.Anime) *entities.CachedAnime {
	if animeData == nil {
		return nil
	}

	cachedAnime := &entities.CachedAnime{
		MALID:         animeData.MALID,
		TitleRomaji:   animeData.Titles.Romaji,
		TitleEnglish:  animeData.Titles.English,
		TitleJapanese: animeData.Titles.Japanese,
		TitleSynonyms: strings.Join(animeData.Titles.Synonyms, ","),
		Synopsis:      animeData.Synopsis,
		Type:          string(animeData.Type),
		Source:        animeData.Source,
		Airing:        animeData.Airing,
		Status:        animeData.Status,
		Duration:      animeData.Duration,
		Color:         animeData.Color,
		Season:        animeData.Season,
		Year:          animeData.Year,
		LastUpdated:   time.Now(),
	}

	// Add Images
	cachedAnime.Images = &entities.CachedAnimeImages{
		Small:    animeData.Images.Small,
		Large:    animeData.Images.Large,
		Original: animeData.Images.Original,
	}

	// Add Logos
	cachedAnime.Logos = &entities.CachedAnimeLogos{
		Small:    animeData.Logos.Small,
		Medium:   animeData.Logos.Medium,
		Large:    animeData.Logos.Large,
		XLarge:   animeData.Logos.XLarge,
		Original: animeData.Logos.Original,
	}

	// Add Covers
	cachedAnime.Covers = &entities.CachedAnimeCovers{
		Small:    animeData.Covers.Small,
		Large:    animeData.Covers.Large,
		Original: animeData.Covers.Original,
	}

	// Add Scores
	cachedAnime.Scores = &entities.CachedAnimeScores{
		Score:      animeData.Scores.Score,
		ScoredBy:   animeData.Scores.ScoredBy,
		Rank:       animeData.Scores.Rank,
		Popularity: animeData.Scores.Popularity,
		Members:    animeData.Scores.Members,
		Favorites:  animeData.Scores.Favorites,
	}

	// Add AiringStatus
	if animeData.AiringStatus.From.Year > 0 || animeData.AiringStatus.From.String != "" {
		cachedAnime.AiringStatus = &entities.CachedAiringStatus{
			String: animeData.AiringStatus.String,
		}

		if animeData.AiringStatus.From.Year > 0 || animeData.AiringStatus.From.String != "" {
			cachedAnime.AiringStatus.From = &entities.CachedAiringStatusDates{
				Day:    animeData.AiringStatus.From.Day,
				Month:  animeData.AiringStatus.From.Month,
				Year:   animeData.AiringStatus.From.Year,
				String: animeData.AiringStatus.From.String,
			}
		}

		if animeData.AiringStatus.To.Year > 0 || animeData.AiringStatus.To.String != "" {
			cachedAnime.AiringStatus.To = &entities.CachedAiringStatusDates{
				Day:    animeData.AiringStatus.To.Day,
				Month:  animeData.AiringStatus.To.Month,
				Year:   animeData.AiringStatus.To.Year,
				String: animeData.AiringStatus.To.String,
			}
		}
	}

	// Add Broadcast
	if animeData.Broadcast.Day != "" || animeData.Broadcast.Time != "" {
		cachedAnime.Broadcast = &entities.CachedAnimeBroadcast{
			Day:      animeData.Broadcast.Day,
			Time:     animeData.Broadcast.Time,
			Timezone: animeData.Broadcast.Timezone,
			String:   animeData.Broadcast.String,
		}
	}

	// Add Genres
	if len(animeData.Genres) > 0 {
		cachedAnime.Genres = make([]entities.CachedAnimeGenre, len(animeData.Genres))
		for i, genre := range animeData.Genres {
			cachedAnime.Genres[i] = entities.CachedAnimeGenre{
				Name:    genre.Name,
				GenreID: genre.GenreID,
				URL:     genre.URL,
			}
		}
	}

	// Add Producers
	if len(animeData.Producers) > 0 {
		cachedAnime.Producers = make([]entities.CachedAnimeProducer, len(animeData.Producers))
		for i, producer := range animeData.Producers {
			cachedAnime.Producers[i] = entities.CachedAnimeProducer{
				Name:       producer.Name,
				ProducerID: producer.ProducerID,
				URL:        producer.URL,
			}
		}
	}

	// Add Studios
	if len(animeData.Studios) > 0 {
		cachedAnime.Studios = make([]entities.CachedAnimeStudio, len(animeData.Studios))
		for i, studio := range animeData.Studios {
			cachedAnime.Studios[i] = entities.CachedAnimeStudio{
				Name:     studio.Name,
				StudioID: studio.StudioID,
				URL:      studio.URL,
			}
		}
	}

	// Add Licensors
	if len(animeData.Licensors) > 0 {
		cachedAnime.Licensors = make([]entities.CachedAnimeLicensor, len(animeData.Licensors))
		for i, licensor := range animeData.Licensors {
			cachedAnime.Licensors[i] = entities.CachedAnimeLicensor{
				Name:       licensor.Name,
				ProducerID: licensor.ProducerID,
				URL:        licensor.URL,
			}
		}
	}

	// Get the current timestamp
	currentTime := time.Now().Unix()

	// Determine the next airing episode from the schedule
	var nextEpisode *types.AnimeAiringEpisode

	// Process next airing episode data - first check if there's a valid next airing episode directly provided
	if animeData.NextAiringEpisode.AiringAt > 0 && animeData.NextAiringEpisode.Episode > 0 {
		// Check if it's still in the future
		if int64(animeData.NextAiringEpisode.AiringAt) > currentTime {
			nextEpisode = &types.AnimeAiringEpisode{
				AiringAt: animeData.NextAiringEpisode.AiringAt,
				Episode:  animeData.NextAiringEpisode.Episode,
			}
		}
	}

	// If we don't have a valid next episode yet, or the one we have has already aired,
	// scan the schedule to find the actual next episode
	if (nextEpisode == nil || nextEpisode.AiringAt == 0) && len(animeData.AiringSchedule) > 0 {
		// Sort the schedule by airing time
		sortedSchedule := make([]types.AnimeAiringEpisode, len(animeData.AiringSchedule))
		copy(sortedSchedule, animeData.AiringSchedule)

		// Sort by airing time
		for i := 0; i < len(sortedSchedule)-1; i++ {
			for j := i + 1; j < len(sortedSchedule); j++ {
				if sortedSchedule[i].AiringAt > sortedSchedule[j].AiringAt {
					sortedSchedule[i], sortedSchedule[j] = sortedSchedule[j], sortedSchedule[i]
				}
			}
		}

		// Find the first episode that hasn't aired yet
		for _, episode := range sortedSchedule {
			if int64(episode.AiringAt) > currentTime {
				nextEpisode = &types.AnimeAiringEpisode{
					AiringAt: episode.AiringAt,
					Episode:  episode.Episode,
				}
				break
			}
		}
	}

	// Add NextAiringEpisode if we found a valid next episode
	if nextEpisode != nil && nextEpisode.AiringAt > 0 {
		cachedAnime.NextAiringEpisode = &entities.CachedNextEpisode{
			AiringAt: nextEpisode.AiringAt,
			Episode:  nextEpisode.Episode,
		}
		logger.Log(fmt.Sprintf("Set next airing episode for %s (ID: %d): Episode %d at %s",
			cachedAnime.TitleRomaji, cachedAnime.MALID,
			nextEpisode.Episode,
			time.Unix(int64(nextEpisode.AiringAt), 0).Format(time.RFC3339)),
			logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AnimeCache",
			})
	}

	// Add AiringSchedule
	if len(animeData.AiringSchedule) > 0 {
		cachedAnime.AiringSchedule = make([]entities.CachedScheduleEpisode, len(animeData.AiringSchedule))
		for i, episode := range animeData.AiringSchedule {
			cachedAnime.AiringSchedule[i] = entities.CachedScheduleEpisode{
				AiringAt: episode.AiringAt,
				Episode:  episode.Episode,
			}
		}
	}

	// Add Episodes
	if len(animeData.Episodes.Episodes) > 0 {
		cachedAnime.Episodes = make([]entities.CachedAnimeSingleEpisode, len(animeData.Episodes.Episodes))
		for i, episode := range animeData.Episodes.Episodes {
			// Create episode titles
			titles := &entities.CachedEpisodeTitles{
				English:  episode.Titles.English,
				Japanese: episode.Titles.Japanese,
				Romaji:   episode.Titles.Romaji,
			}

			cachedAnime.Episodes[i] = entities.CachedAnimeSingleEpisode{
				Description:  episode.Description,
				Aired:        episode.Aired,
				Score:        episode.Score,
				Filler:       episode.Filler,
				Recap:        episode.Recap,
				ForumURL:     episode.ForumURL,
				URL:          episode.URL,
				ThumbnailURL: episode.ThumbnailURL,
				Titles:       titles,
			}
		}
	}

	// Add Characters
	if len(animeData.Characters) > 0 {
		cachedAnime.Characters = make([]entities.CachedAnimeCharacter, len(animeData.Characters))
		for i, character := range animeData.Characters {
			char := entities.CachedAnimeCharacter{
				MALID:    character.MALID,
				URL:      character.URL,
				ImageURL: character.ImageURL,
				Name:     character.Name,
				Role:     character.Role,
			}

			if len(character.VoiceActors) > 0 {
				char.VoiceActors = make([]entities.CachedAnimeVoiceActor, len(character.VoiceActors))
				for j, va := range character.VoiceActors {
					char.VoiceActors[j] = entities.CachedAnimeVoiceActor{
						MALID:    va.MALID,
						URL:      va.URL,
						Image:    va.Image,
						Name:     va.Name,
						Language: va.Language,
					}
				}
			}
			cachedAnime.Characters[i] = char
		}
	}

	// Add Seasons
	if len(animeData.Seasons) > 0 {
		cachedAnime.Seasons = make([]entities.CachedAnimeSeason, len(animeData.Seasons))
		for i, season := range animeData.Seasons {
			cachedSeason := entities.CachedAnimeSeason{
				ParentAnimeID: cachedAnime.ID,
				MALID:         season.MALID,
				TitleRomaji:   season.Titles.Romaji,
				TitleEnglish:  season.Titles.English,
				TitleJapanese: season.Titles.Japanese,
				TitleSynonyms: strings.Join(season.Titles.Synonyms, ","),
				Synopsis:      season.Synopsis,
				Type:          string(season.Type),
				Source:        season.Source,
				Airing:        season.Airing,
				Status:        season.Status,
				Duration:      season.Duration,
				Season:        season.Season,
				Year:          season.Year,
				Current:       season.Current,
			}

			// Add Images
			cachedSeason.Images = &entities.CachedAnimeImages{
				Small:    season.Images.Small,
				Large:    season.Images.Large,
				Original: season.Images.Original,
			}

			// Add Scores
			cachedSeason.Scores = &entities.CachedAnimeScores{
				Score:      season.Scores.Score,
				ScoredBy:   season.Scores.ScoredBy,
				Rank:       season.Scores.Rank,
				Popularity: season.Scores.Popularity,
				Members:    season.Scores.Members,
				Favorites:  season.Scores.Favorites,
			}

			// Add AiringStatus
			if season.AiringStatus.From.Year > 0 || season.AiringStatus.From.String != "" {
				cachedSeason.AiringStatus = &entities.CachedAiringStatus{
					String: season.AiringStatus.String,
				}

				if season.AiringStatus.From.Year > 0 || season.AiringStatus.From.String != "" {
					cachedSeason.AiringStatus.From = &entities.CachedAiringStatusDates{
						Day:    season.AiringStatus.From.Day,
						Month:  season.AiringStatus.From.Month,
						Year:   season.AiringStatus.From.Year,
						String: season.AiringStatus.From.String,
					}
				}

				if season.AiringStatus.To.Year > 0 || season.AiringStatus.To.String != "" {
					cachedSeason.AiringStatus.To = &entities.CachedAiringStatusDates{
						Day:    season.AiringStatus.To.Day,
						Month:  season.AiringStatus.To.Month,
						Year:   season.AiringStatus.To.Year,
						String: season.AiringStatus.To.String,
					}
				}
			}

			cachedAnime.Seasons[i] = cachedSeason
		}
	}

	return cachedAnime
}
