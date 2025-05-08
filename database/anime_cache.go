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
		if err := deleteExistingAnimeCache(tx, existingAnime.ID); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Create new cached anime
	cachedAnime := convertToCachedAnime(animeData)

	// Save the main anime record
	if err := tx.Create(&cachedAnime).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	logger.Log(fmt.Sprintf("Successfully saved anime (MAL ID: %d) to cache", animeData.MALID), logger.LogOptions{
		Level:  logger.Success,
		Prefix: "AnimeCache",
	})

	return nil
}

// deleteExistingAnimeCache deletes an anime and all its relations from the cache
func deleteExistingAnimeCache(tx *gorm.DB, animeID uint) error {
	// Delete related entities in order to avoid foreign key constraints
	tables := []string{
		"cached_anime_voice_actors",
		"cached_anime_characters",
		"cached_episode_titles",
		"cached_anime_single_episodes",
		"cached_airing_episodes",
		"cached_anime_licensors",
		"cached_anime_studios",
		"cached_anime_producers",
		"cached_anime_genres",
		"cached_anime_broadcasts",
		"cached_airing_status_dates",
		"cached_airing_statuses",
		"cached_anime_scores",
		"cached_anime_covers",
		"cached_anime_logos",
		"cached_anime_images",
		"cached_anime_seasons",
		"cached_animes",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE anime_id = ?", table)
		if table == "cached_animes" {
			query = "DELETE FROM cached_animes WHERE id = ?"
		}

		if err := tx.Exec(query, animeID).Error; err != nil {
			return err
		}
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
	if cachedAnime.NextAiringEpisode != nil && cachedAnime.NextAiringEpisode.IsNext {
		anime.NextAiringEpisode = types.AnimeAiringEpisode{
			AiringAt:        cachedAnime.NextAiringEpisode.AiringAt,
			TimeUntilAiring: cachedAnime.NextAiringEpisode.TimeUntilAiring,
			Episode:         cachedAnime.NextAiringEpisode.Episode,
		}
	}

	// Convert airing schedule
	if len(cachedAnime.AiringSchedule) > 0 {
		anime.AiringSchedule = make([]types.AnimeAiringEpisode, len(cachedAnime.AiringSchedule))
		for i, episode := range cachedAnime.AiringSchedule {
			anime.AiringSchedule[i] = types.AnimeAiringEpisode{
				AiringAt:        episode.AiringAt,
				TimeUntilAiring: episode.TimeUntilAiring,
				Episode:         episode.Episode,
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

	// Create the anime with basic fields
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

	// Add Logos if they exist
	cachedAnime.Logos = &entities.CachedAnimeLogos{
		Small:    animeData.Logos.Small,
		Medium:   animeData.Logos.Medium,
		Large:    animeData.Logos.Large,
		XLarge:   animeData.Logos.XLarge,
		Original: animeData.Logos.Original,
	}

	// Add Covers if they exist
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

	// Add AiringStatus with From and To dates
	fromDates := &entities.CachedAiringStatusDates{
		Day:    animeData.AiringStatus.From.Day,
		Month:  animeData.AiringStatus.From.Month,
		Year:   animeData.AiringStatus.From.Year,
		String: animeData.AiringStatus.From.String,
	}

	toDates := &entities.CachedAiringStatusDates{
		Day:    animeData.AiringStatus.To.Day,
		Month:  animeData.AiringStatus.To.Month,
		Year:   animeData.AiringStatus.To.Year,
		String: animeData.AiringStatus.To.String,
	}

	cachedAnime.AiringStatus = &entities.CachedAiringStatus{
		String: animeData.AiringStatus.String,
		From:   fromDates,
		To:     toDates,
	}

	// Add Broadcast
	cachedAnime.Broadcast = &entities.CachedAnimeBroadcast{
		Day:      animeData.Broadcast.Day,
		Time:     animeData.Broadcast.Time,
		Timezone: animeData.Broadcast.Timezone,
		String:   animeData.Broadcast.String,
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

	// Add NextAiringEpisode if available
	if animeData.NextAiringEpisode.Episode > 0 {
		cachedAnime.NextAiringEpisode = &entities.CachedAiringEpisode{
			AiringAt:        animeData.NextAiringEpisode.AiringAt,
			TimeUntilAiring: animeData.NextAiringEpisode.TimeUntilAiring,
			Episode:         animeData.NextAiringEpisode.Episode,
			IsNext:          true,
		}
	}

	// Add AiringSchedule
	if len(animeData.AiringSchedule) > 0 {
		cachedAnime.AiringSchedule = make([]entities.CachedAiringEpisode, len(animeData.AiringSchedule))
		for i, episode := range animeData.AiringSchedule {
			cachedAnime.AiringSchedule[i] = entities.CachedAiringEpisode{
				AiringAt:        episode.AiringAt,
				TimeUntilAiring: episode.TimeUntilAiring,
				Episode:         episode.Episode,
				IsNext:          false, // Only the dedicated next episode is marked true
			}
		}
	}

	// Add Episodes
	if len(animeData.Episodes.Episodes) > 0 {
		cachedAnime.Episodes = make([]entities.CachedAnimeSingleEpisode, len(animeData.Episodes.Episodes))
		for i, episode := range animeData.Episodes.Episodes {
			episodeTitles := &entities.CachedEpisodeTitles{
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
				Titles:       episodeTitles,
			}
		}
	}

	// Add Characters
	if len(animeData.Characters) > 0 {
		cachedAnime.Characters = make([]entities.CachedAnimeCharacter, len(animeData.Characters))
		for i, character := range animeData.Characters {
			cachedCharacter := entities.CachedAnimeCharacter{
				MALID:    character.MALID,
				URL:      character.URL,
				ImageURL: character.ImageURL,
				Name:     character.Name,
				Role:     character.Role,
			}

			if len(character.VoiceActors) > 0 {
				cachedCharacter.VoiceActors = make([]entities.CachedAnimeVoiceActor, len(character.VoiceActors))
				for j, va := range character.VoiceActors {
					cachedCharacter.VoiceActors[j] = entities.CachedAnimeVoiceActor{
						MALID:    va.MALID,
						URL:      va.URL,
						Image:    va.Image,
						Name:     va.Name,
						Language: va.Language,
					}
				}
			}

			cachedAnime.Characters[i] = cachedCharacter
		}
	}

	// Add Seasons
	if len(animeData.Seasons) > 0 {
		cachedAnime.Seasons = make([]entities.CachedAnimeSeason, len(animeData.Seasons))
		for i, season := range animeData.Seasons {
			// Create the related entities first to get their IDs
			seasonImages := &entities.CachedAnimeImages{
				Small:    season.Images.Small,
				Large:    season.Images.Large,
				Original: season.Images.Original,
			}

			seasonScores := &entities.CachedAnimeScores{
				Score:      season.Scores.Score,
				ScoredBy:   season.Scores.ScoredBy,
				Rank:       season.Scores.Rank,
				Popularity: season.Scores.Popularity,
				Members:    season.Scores.Members,
				Favorites:  season.Scores.Favorites,
			}

			// Create airing status dates
			seasonFromDates := &entities.CachedAiringStatusDates{
				Day:    season.AiringStatus.From.Day,
				Month:  season.AiringStatus.From.Month,
				Year:   season.AiringStatus.From.Year,
				String: season.AiringStatus.From.String,
			}

			seasonToDates := &entities.CachedAiringStatusDates{
				Day:    season.AiringStatus.To.Day,
				Month:  season.AiringStatus.To.Month,
				Year:   season.AiringStatus.To.Year,
				String: season.AiringStatus.To.String,
			}

			// Create airing status
			seasonAiringStatus := &entities.CachedAiringStatus{
				String: season.AiringStatus.String,
				From:   seasonFromDates,
				To:     seasonToDates,
			}

			// Create the season with references to related entities
			cachedSeason := entities.CachedAnimeSeason{
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

				// Add references to related entities
				Images:       seasonImages,
				Scores:       seasonScores,
				AiringStatus: seasonAiringStatus,
			}

			cachedAnime.Seasons[i] = cachedSeason
		}
	}

	return cachedAnime
}
