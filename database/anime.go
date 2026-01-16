package database

import (
	"fmt"
	"metachan/entities"
	"metachan/types"
	"strings"
	"time"

	"gorm.io/gorm"
)

func GetAnimeByMALID(malID int) (*types.Anime, error) {
	var anime entities.Anime
	result := DB.Preload("Images").
		Preload("Logos").
		Preload("Covers").
		Preload("Scores").
		Preload("AiringStatus").
		Preload("AiringStatus.From").
		Preload("AiringStatus.To").
		Preload("Broadcast").
		Preload("Genres").
		Preload("Producers").
		Preload("Studios").
		Preload("Licensors").
		Preload("Episodes").
		Preload("Episodes.Titles").
		Preload("Characters").
		Preload("Characters.VoiceActors").
		Preload("AiringSchedule").
		Preload("NextAiringEpisode").
		Preload("Seasons").
		Preload("Seasons.Images").
		Preload("Seasons.Scores").
		Preload("Seasons.AiringStatus").
		Preload("Seasons.AiringStatus.From").
		Preload("Seasons.AiringStatus.To").
		Where("mal_id = ?", malID).First(&anime)

	if result.Error != nil {
		return nil, result.Error
	}

	return ConvertToTypesAnime(&anime), nil
}

func SaveAnimeToDatabase(animeData *types.Anime) error {
	if animeData == nil {
		return fmt.Errorf("anime data is nil")
	}

	var tx *gorm.DB = DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingAnime entities.Anime
	result := tx.Where("mal_id = ?", animeData.MALID).First(&existingAnime)
	if result.Error == nil {
		if err := tx.Delete(&existingAnime).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	anime := &entities.Anime{
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
		SubbedCount:   animeData.Episodes.Subbed,
		DubbedCount:   animeData.Episodes.Dubbed,
		TotalEpisodes: animeData.Episodes.Total,
		AiredEpisodes: animeData.Episodes.Aired,
		LastUpdated:   time.Now(),
	}

	// Save images
	if animeData.Images.Small != "" || animeData.Images.Large != "" || animeData.Images.Original != "" {
		anime.Images = &entities.AnimeImages{
			Small:    animeData.Images.Small,
			Large:    animeData.Images.Large,
			Original: animeData.Images.Original,
		}
	}

	// Save logos
	if animeData.Logos.Small != "" || animeData.Logos.Medium != "" || animeData.Logos.Large != "" {
		anime.Logos = &entities.AnimeLogos{
			Small:    animeData.Logos.Small,
			Medium:   animeData.Logos.Medium,
			Large:    animeData.Logos.Large,
			XLarge:   animeData.Logos.XLarge,
			Original: animeData.Logos.Original,
		}
	}

	// Save covers
	if animeData.Covers.Small != "" || animeData.Covers.Large != "" || animeData.Covers.Original != "" {
		anime.Covers = &entities.AnimeCovers{
			Small:    animeData.Covers.Small,
			Large:    animeData.Covers.Large,
			Original: animeData.Covers.Original,
		}
	}

	// Save scores
	if animeData.Scores.Score > 0 || animeData.Scores.ScoredBy > 0 {
		anime.Scores = &entities.AnimeScores{
			Score:      animeData.Scores.Score,
			ScoredBy:   animeData.Scores.ScoredBy,
			Rank:       animeData.Scores.Rank,
			Popularity: animeData.Scores.Popularity,
			Members:    animeData.Scores.Members,
			Favorites:  animeData.Scores.Favorites,
		}
	}

	// Save airing status
	if animeData.AiringStatus.String != "" {
		airingStatus := &entities.AiringStatus{
			String: animeData.AiringStatus.String,
		}

		if animeData.AiringStatus.From.Year > 0 {
			airingStatus.From = &entities.AiringStatusDates{
				Day:    animeData.AiringStatus.From.Day,
				Month:  animeData.AiringStatus.From.Month,
				Year:   animeData.AiringStatus.From.Year,
				String: animeData.AiringStatus.From.String,
			}
		}

		if animeData.AiringStatus.To.Year > 0 {
			airingStatus.To = &entities.AiringStatusDates{
				Day:    animeData.AiringStatus.To.Day,
				Month:  animeData.AiringStatus.To.Month,
				Year:   animeData.AiringStatus.To.Year,
				String: animeData.AiringStatus.To.String,
			}
		}

		anime.AiringStatus = airingStatus
	}

	// Save broadcast info
	if animeData.Broadcast.String != "" {
		anime.Broadcast = &entities.AnimeBroadcast{
			Day:      animeData.Broadcast.Day,
			Time:     animeData.Broadcast.Time,
			Timezone: animeData.Broadcast.Timezone,
			String:   animeData.Broadcast.String,
		}
	}

	// Save genres
	if len(animeData.Genres) > 0 {
		anime.Genres = make([]entities.AnimeGenre, len(animeData.Genres))
		for i, genre := range animeData.Genres {
			anime.Genres[i] = entities.AnimeGenre{
				Name:    genre.Name,
				GenreID: genre.GenreID,
				URL:     genre.URL,
			}
		}
	}

	// Save producers
	if len(animeData.Producers) > 0 {
		anime.Producers = make([]entities.AnimeProducer, len(animeData.Producers))
		for i, producer := range animeData.Producers {
			anime.Producers[i] = entities.AnimeProducer{
				Name:       producer.Name,
				ProducerID: producer.ProducerID,
				URL:        producer.URL,
			}
		}
	}

	// Save studios
	if len(animeData.Studios) > 0 {
		anime.Studios = make([]entities.AnimeStudio, len(animeData.Studios))
		for i, studio := range animeData.Studios {
			anime.Studios[i] = entities.AnimeStudio{
				Name:     studio.Name,
				StudioID: studio.StudioID,
				URL:      studio.URL,
			}
		}
	}

	// Save licensors
	if len(animeData.Licensors) > 0 {
		anime.Licensors = make([]entities.AnimeLicensor, len(animeData.Licensors))
		for i, licensor := range animeData.Licensors {
			anime.Licensors[i] = entities.AnimeLicensor{
				Name:       licensor.Name,
				ProducerID: licensor.ProducerID,
				URL:        licensor.URL,
			}
		}
	}

	// Save seasons
	if len(animeData.Seasons) > 0 {
		anime.Seasons = make([]entities.AnimeSeason, len(animeData.Seasons))
		for i, season := range animeData.Seasons {
			animeSeason := entities.AnimeSeason{
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

			// Save season images
			if season.Images.Small != "" || season.Images.Large != "" {
				animeSeason.Images = &entities.AnimeImages{
					Small:    season.Images.Small,
					Large:    season.Images.Large,
					Original: season.Images.Original,
				}
			}

			// Save season scores
			if season.Scores.Score > 0 {
				animeSeason.Scores = &entities.AnimeScores{
					Score:      season.Scores.Score,
					ScoredBy:   season.Scores.ScoredBy,
					Rank:       season.Scores.Rank,
					Popularity: season.Scores.Popularity,
					Members:    season.Scores.Members,
					Favorites:  season.Scores.Favorites,
				}
			}

			anime.Seasons[i] = animeSeason
		}
	}

	if len(animeData.Episodes.Episodes) > 0 {
		anime.Episodes = make([]entities.AnimeSingleEpisode, len(animeData.Episodes.Episodes))
		for i, episode := range animeData.Episodes.Episodes {
			titles := &entities.EpisodeTitles{
				English:  episode.Titles.English,
				Japanese: episode.Titles.Japanese,
				Romaji:   episode.Titles.Romaji,
			}

			anime.Episodes[i] = entities.AnimeSingleEpisode{
				EpisodeID:    episode.ID,
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

	// Save characters data
	if len(animeData.Characters) > 0 {
		anime.Characters = make([]entities.AnimeCharacter, len(animeData.Characters))
		for i, character := range animeData.Characters {
			anime.Characters[i] = entities.AnimeCharacter{
				MALID:    character.MALID,
				URL:      character.URL,
				ImageURL: character.ImageURL,
				Name:     character.Name,
				Role:     character.Role,
			}

			// Save voice actors for this character
			if len(character.VoiceActors) > 0 {
				anime.Characters[i].VoiceActors = make([]entities.AnimeVoiceActor, len(character.VoiceActors))
				for j, va := range character.VoiceActors {
					anime.Characters[i].VoiceActors[j] = entities.AnimeVoiceActor{
						MALID:    va.MALID,
						URL:      va.URL,
						Image:    va.Image,
						Name:     va.Name,
						Language: va.Language,
					}
				}
			}
		}
	}

	// Save airing schedule data
	if len(animeData.AiringSchedule) > 0 {
		anime.AiringSchedule = make([]entities.ScheduleEpisode, len(animeData.AiringSchedule))
		for i, schedule := range animeData.AiringSchedule {
			anime.AiringSchedule[i] = entities.ScheduleEpisode{
				AiringAt: schedule.AiringAt,
				Episode:  schedule.Episode,
				IsNext:   false, // We'll set this based on next airing episode if available
			}
		}
	}

	// Set next airing episode data
	if animeData.NextAiringEpisode.Episode > 0 {
		anime.NextAiringEpisode = &entities.NextEpisode{
			AiringAt: animeData.NextAiringEpisode.AiringAt,
			Episode:  animeData.NextAiringEpisode.Episode,
		}

		// Mark the next airing episode in the schedule as IsNext
		for i := range anime.AiringSchedule {
			if anime.AiringSchedule[i].Episode == animeData.NextAiringEpisode.Episode &&
				anime.AiringSchedule[i].AiringAt == animeData.NextAiringEpisode.AiringAt {
				anime.AiringSchedule[i].IsNext = true
				break
			}
		}
	}

	if err := tx.Create(anime).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func ConvertToTypesAnime(anime *entities.Anime) *types.Anime {
	if anime == nil {
		return nil
	}

	result := &types.Anime{
		MALID: anime.MALID,
		Titles: types.AnimeTitles{
			Romaji:   anime.TitleRomaji,
			English:  anime.TitleEnglish,
			Japanese: anime.TitleJapanese,
			Synonyms: strings.Split(anime.TitleSynonyms, ","),
		},
		Synopsis: anime.Synopsis,
		Type:     types.AniSyncType(anime.Type),
		Source:   anime.Source,
		Airing:   anime.Airing,
		Status:   anime.Status,
		Duration: anime.Duration,
		Color:    anime.Color,
		Season:   anime.Season,
		Year:     anime.Year,
		Episodes: types.AnimeEpisodes{
			Total:  anime.TotalEpisodes,
			Aired:  anime.AiredEpisodes,
			Subbed: anime.SubbedCount,
			Dubbed: anime.DubbedCount,
		},
	}

	// Convert images
	if anime.Images != nil {
		result.Images = types.AnimeImages{
			Small:    anime.Images.Small,
			Large:    anime.Images.Large,
			Original: anime.Images.Original,
		}
	}

	// Convert logos
	if anime.Logos != nil {
		result.Logos = types.AnimeLogos{
			Small:    anime.Logos.Small,
			Medium:   anime.Logos.Medium,
			Large:    anime.Logos.Large,
			XLarge:   anime.Logos.XLarge,
			Original: anime.Logos.Original,
		}
	}

	// Convert covers
	if anime.Covers != nil {
		result.Covers = types.AnimeImages{
			Small:    anime.Covers.Small,
			Large:    anime.Covers.Large,
			Original: anime.Covers.Original,
		}
	}

	// Convert scores
	if anime.Scores != nil {
		result.Scores = types.AnimeScores{
			Score:      anime.Scores.Score,
			ScoredBy:   anime.Scores.ScoredBy,
			Rank:       anime.Scores.Rank,
			Popularity: anime.Scores.Popularity,
			Members:    anime.Scores.Members,
			Favorites:  anime.Scores.Favorites,
		}
	}

	// Convert airing status
	if anime.AiringStatus != nil {
		result.AiringStatus = types.AiringStatus{
			String: anime.AiringStatus.String,
		}

		if anime.AiringStatus.From != nil {
			result.AiringStatus.From = types.AiringStatusDates{
				Day:    anime.AiringStatus.From.Day,
				Month:  anime.AiringStatus.From.Month,
				Year:   anime.AiringStatus.From.Year,
				String: anime.AiringStatus.From.String,
			}
		}

		if anime.AiringStatus.To != nil {
			result.AiringStatus.To = types.AiringStatusDates{
				Day:    anime.AiringStatus.To.Day,
				Month:  anime.AiringStatus.To.Month,
				Year:   anime.AiringStatus.To.Year,
				String: anime.AiringStatus.To.String,
			}
		}
	}

	// Convert broadcast
	if anime.Broadcast != nil {
		result.Broadcast = types.AnimeBroadcast{
			Day:      anime.Broadcast.Day,
			Time:     anime.Broadcast.Time,
			Timezone: anime.Broadcast.Timezone,
			String:   anime.Broadcast.String,
		}
	}

	// Convert genres
	if len(anime.Genres) > 0 {
		result.Genres = make([]types.AnimeGenres, len(anime.Genres))
		for i, genre := range anime.Genres {
			result.Genres[i] = types.AnimeGenres{
				Name:    genre.Name,
				GenreID: genre.GenreID,
				URL:     genre.URL,
			}
		}
	}

	// Convert producers
	if len(anime.Producers) > 0 {
		result.Producers = make([]types.AnimeProducer, len(anime.Producers))
		for i, producer := range anime.Producers {
			result.Producers[i] = types.AnimeProducer{
				Name:       producer.Name,
				ProducerID: producer.ProducerID,
				URL:        producer.URL,
			}
		}
	}

	// Convert studios
	if len(anime.Studios) > 0 {
		result.Studios = make([]types.AnimeStudio, len(anime.Studios))
		for i, studio := range anime.Studios {
			result.Studios[i] = types.AnimeStudio{
				Name:     studio.Name,
				StudioID: studio.StudioID,
				URL:      studio.URL,
			}
		}
	}

	// Convert licensors
	if len(anime.Licensors) > 0 {
		result.Licensors = make([]types.AnimeLicensor, len(anime.Licensors))
		for i, licensor := range anime.Licensors {
			result.Licensors[i] = types.AnimeLicensor{
				Name:       licensor.Name,
				ProducerID: licensor.ProducerID,
				URL:        licensor.URL,
			}
		}
	}

	// Convert seasons
	if len(anime.Seasons) > 0 {
		result.Seasons = make([]types.AnimeSeason, len(anime.Seasons))
		for i, season := range anime.Seasons {
			result.Seasons[i] = types.AnimeSeason{
				MALID: season.MALID,
				Titles: types.AnimeTitles{
					Romaji:   season.TitleRomaji,
					English:  season.TitleEnglish,
					Japanese: season.TitleJapanese,
					Synonyms: strings.Split(season.TitleSynonyms, ","),
				},
				Synopsis: season.Synopsis,
				Type:     types.AniSyncType(season.Type),
				Source:   season.Source,
				Airing:   season.Airing,
				Status:   season.Status,
				Duration: season.Duration,
				Season:   season.Season,
				Year:     season.Year,
				Current:  season.Current,
			}

			// Convert season images
			if season.Images != nil {
				result.Seasons[i].Images = types.AnimeImages{
					Small:    season.Images.Small,
					Large:    season.Images.Large,
					Original: season.Images.Original,
				}
			}

			// Convert season scores
			if season.Scores != nil {
				result.Seasons[i].Scores = types.AnimeScores{
					Score:      season.Scores.Score,
					ScoredBy:   season.Scores.ScoredBy,
					Rank:       season.Scores.Rank,
					Popularity: season.Scores.Popularity,
					Members:    season.Scores.Members,
					Favorites:  season.Scores.Favorites,
				}
			}

			// Convert season airing status
			if season.AiringStatus != nil {
				result.Seasons[i].AiringStatus = types.AiringStatus{
					String: season.AiringStatus.String,
				}

				if season.AiringStatus.From != nil {
					result.Seasons[i].AiringStatus.From = types.AiringStatusDates{
						Day:    season.AiringStatus.From.Day,
						Month:  season.AiringStatus.From.Month,
						Year:   season.AiringStatus.From.Year,
						String: season.AiringStatus.From.String,
					}
				}

				if season.AiringStatus.To != nil {
					result.Seasons[i].AiringStatus.To = types.AiringStatusDates{
						Day:    season.AiringStatus.To.Day,
						Month:  season.AiringStatus.To.Month,
						Year:   season.AiringStatus.To.Year,
						String: season.AiringStatus.To.String,
					}
				}
			}
		}
	}

	if len(anime.Episodes) > 0 {
		result.Episodes.Episodes = make([]types.AnimeSingleEpisode, len(anime.Episodes))
		for i, episode := range anime.Episodes {
			episodeData := types.AnimeSingleEpisode{
				ID:           episode.EpisodeID,
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

			result.Episodes.Episodes[i] = episodeData
		}
	}

	// Convert characters
	if len(anime.Characters) > 0 {
		result.Characters = make([]types.AnimeCharacter, len(anime.Characters))
		for i, character := range anime.Characters {
			result.Characters[i] = types.AnimeCharacter{
				MALID:    character.MALID,
				URL:      character.URL,
				ImageURL: character.ImageURL,
				Name:     character.Name,
				Role:     character.Role,
			}

			// Convert voice actors for this character
			if len(character.VoiceActors) > 0 {
				result.Characters[i].VoiceActors = make([]types.AnimeVoiceActor, len(character.VoiceActors))
				for j, va := range character.VoiceActors {
					result.Characters[i].VoiceActors[j] = types.AnimeVoiceActor{
						MALID:    va.MALID,
						URL:      va.URL,
						Image:    va.Image,
						Name:     va.Name,
						Language: va.Language,
					}
				}
			}
		}
	}

	// Convert airing schedule
	if len(anime.AiringSchedule) > 0 {
		result.AiringSchedule = make([]types.AnimeAiringEpisode, len(anime.AiringSchedule))
		for i, schedule := range anime.AiringSchedule {
			result.AiringSchedule[i] = types.AnimeAiringEpisode{
				AiringAt: schedule.AiringAt,
				Episode:  schedule.Episode,
			}
		}
	}

	// Convert next airing episode
	if anime.NextAiringEpisode != nil {
		result.NextAiringEpisode = types.AnimeAiringEpisode{
			AiringAt: anime.NextAiringEpisode.AiringAt,
			Episode:  anime.NextAiringEpisode.Episode,
		}
	}

	var mapping entities.AnimeMapping
	if err := DB.Where("mal = ?", anime.MALID).First(&mapping).Error; err == nil {
		result.Mappings = types.AnimeMappings{
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
		}
	}

	return result
}

func GetAnimeMappingViaMALID(malID int) (*entities.AnimeMapping, error) {
	var mapping entities.AnimeMapping
	result := DB.Where("mal = ?", malID).First(&mapping)
	if result.Error != nil {
		return nil, result.Error
	}
	return &mapping, nil
}

func GetAnimeMappingViaAnilistID(anilistID int) (*entities.AnimeMapping, error) {
	var mapping entities.AnimeMapping
	result := DB.Where("anilist = ?", anilistID).First(&mapping)
	if result.Error != nil {
		return nil, result.Error
	}
	return &mapping, nil
}

func GetAnimeMappingsByTVDBID(tvdbID int) ([]entities.AnimeMapping, error) {
	var mappings []entities.AnimeMapping
	result := DB.Where("tvdb = ?", tvdbID).Find(&mappings)
	if result.Error != nil {
		return nil, result.Error
	}
	return mappings, nil
}

// GetEpisodeStreaming retrieves cached streaming data for an episode
func GetEpisodeStreaming(episodeID string, animeID uint) (*entities.EpisodeStreaming, error) {
	var streaming entities.EpisodeStreaming
	result := DB.Preload("SubSources").
		Preload("DubSources").
		Where("episode_id = ? AND anime_id = ?", episodeID, animeID).
		First(&streaming)

	if result.Error != nil {
		return nil, result.Error
	}

	// Check if data is stale (older than 7 days)
	if time.Since(streaming.LastFetch) > 7*24*time.Hour {
		return nil, fmt.Errorf("streaming data is stale")
	}

	return &streaming, nil
}

// SaveEpisodeStreaming saves streaming data to the database
func SaveEpisodeStreaming(episodeID string, animeID uint, subSources, dubSources []types.AnimeStreamingSource) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete existing streaming data for this episode
	var existing entities.EpisodeStreaming
	if err := tx.Where("episode_id = ? AND anime_id = ?", episodeID, animeID).First(&existing).Error; err == nil {
		if err := tx.Delete(&existing).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Create new streaming record
	streaming := &entities.EpisodeStreaming{
		EpisodeID: episodeID,
		AnimeID:   animeID,
		LastFetch: time.Now(),
	}

	// Save the main record first
	if err := tx.Create(streaming).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Save sub sources
	for _, source := range subSources {
		subSource := entities.EpisodeStreamingSource{
			EpisodeStreamingID: streaming.ID,
			URL:                source.URL,
			Server:             source.Server,
			Type:               source.Type,
		}
		if err := tx.Create(&subSource).Error; err != nil {
			tx.Rollback()
			return err
		}
		streaming.SubSources = append(streaming.SubSources, subSource)
	}

	// Save dub sources
	for _, source := range dubSources {
		dubSource := entities.EpisodeStreamingSource{
			EpisodeStreamingID: streaming.ID,
			URL:                source.URL,
			Server:             source.Server,
			Type:               source.Type,
		}
		if err := tx.Create(&dubSource).Error; err != nil {
			tx.Rollback()
			return err
		}
		streaming.DubSources = append(streaming.DubSources, dubSource)
	}

	return tx.Commit().Error
}
