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
	result := DB.Preload("Episodes").
		Preload("Episodes.Titles").
		Preload("Characters").
		Preload("Characters.VoiceActors").
		Preload("AiringSchedule").
		Preload("NextAiringEpisode").
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
