package repositories

import (
	"errors"
	"fmt"
	"metachan/entities"
	"metachan/enums"
	"metachan/utils/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAnime[T idType](maptype enums.MappingType, id T) (entities.Anime, error) {
	var anime entities.Anime

	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		logger.Errorf("Anime", "Failed to get anime mapping: %v", err)
		return entities.Anime{}, errors.New("anime not found")
	}

	result := DB.
		Preload("Mapping").
		Preload("Title").
		Preload("Images").
		Preload("Covers").
		Preload("Logos").
		Preload("Scores").
		Preload("AiringStatus").
		Preload("AiringStatus.From").
		Preload("AiringStatus.To").
		Preload("Broadcast").
		Preload("NextAiring").
		Preload("Genres").
		Preload("Producers").
		Preload("Producers.Image").
		Preload("Producers.Titles").
		Preload("Producers.ExternalURLs").
		Preload("Studios").
		Preload("Studios.Image").
		Preload("Studios.Titles").
		Preload("Studios.ExternalURLs").
		Preload("Licensors").
		Preload("Licensors.Image").
		Preload("Licensors.Titles").
		Preload("Licensors.ExternalURLs").
		Preload("Episodes").
		Preload("Episodes.Title").
		Preload("Episodes.SkipTimes").
		Preload("Episodes.StreamInfo").
		Preload("Episodes.StreamInfo.SubSources").
		Preload("Episodes.StreamInfo.DubSources").
		Preload("Schedule").
		Preload("Seasons").
		Preload("Seasons.Title").
		Preload("Seasons.Images").
		Preload("Seasons.Scores").
		Preload("Seasons.AiringStatus").
		Preload("Seasons.AiringStatus.From").
		Preload("Seasons.AiringStatus.To").
		Where("mapping_id = ?", mapping.ID).
		First(&anime)

	if result.Error != nil {
		logger.Errorf("Anime", "Failed to get anime details: %v", result.Error)
		return entities.Anime{}, errors.New("anime not found")
	}

	loadAnimeCharacters(&anime)

	return anime, nil
}

func loadAnimeCharacters(anime *entities.Anime) {
	var rows []struct {
		CharacterID uint
		Role        string
	}
	DB.Table("anime_characters").
		Select("character_id, role").
		Where("anime_id = ?", anime.ID).
		Scan(&rows)

	for _, row := range rows {
		var char entities.Character
		if err := DB.First(&char, row.CharacterID).Error; err != nil {
			continue
		}
		DB.Preload("VoiceActor").
			Where("character_id = ?", char.ID).
			Find(&char.VoiceActors)
		char.Role = row.Role
		anime.Characters = append(anime.Characters, char)
	}
}

func SaveAnimeCharacters(animeID uint, characters []entities.Character) error {
	for i := range characters {
		char := &characters[i]

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"url", "image_url", "name"}),
		}).Create(char)
		if char.ID == 0 {
			DB.Where("mal_id = ?", char.MALID).First(char)
		}

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "anime_id"}, {Name: "character_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role"}),
		}).Create(&entities.AnimeCharacter{
			AnimeID:     animeID,
			CharacterID: char.ID,
			Role:        char.Role,
		})

		for _, cva := range char.VoiceActors {
			va := cva.VoiceActor
			if va == nil {
				continue
			}

			DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "mal_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"url", "image", "name"}),
			}).Create(va)
			if va.ID == 0 {
				DB.Where("mal_id = ?", va.MALID).First(va)
			}

			DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "character_id"}, {Name: "voice_actor_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"language"}),
			}).Create(&entities.CharacterVoiceActor{
				CharacterID:  char.ID,
				VoiceActorID: va.ID,
				Language:     cva.Language,
			})
		}
	}
	return nil
}

func CreateOrUpdateAnime(anime *entities.Anime) error {
	if anime == nil {
		return fmt.Errorf("anime is nil")
	}

	var existingAnime entities.Anime
	result := DB.Where("mal_id = ?", anime.MALID).First(&existingAnime)
	if result.Error == nil {
		anime.ID = existingAnime.ID
	}

	result = DB.Session(&gorm.Session{FullSaveAssociations: true}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Omit("Characters", "Episodes").Save(anime)

	if result.Error != nil {
		return fmt.Errorf("failed to save anime: %w", result.Error)
	}

	logger.Infof("Anime", "Saved anime (MAL ID: %d) with %d episodes, %d characters", anime.MALID, len(anime.Episodes), len(anime.Characters))
	return nil
}

func SaveAnimeEpisodes(animeID uint, episodes []entities.Episode) error {
	for i := range episodes {
		ep := &episodes[i]
		ep.AnimeID = animeID

		var existing entities.Episode
		if DB.Where("episode_id = ?", ep.EpisodeID).First(&existing).Error == nil {
			ep.ID = existing.ID
			ep.TitleID = existing.TitleID
			DB.Model(ep).Omit("SkipTimes", "StreamInfo", "Title").Updates(ep)
			if ep.Title != nil && existing.TitleID != 0 {
				ep.Title.ID = existing.TitleID
				DB.Save(ep.Title)
			}
		} else {
			DB.Session(&gorm.Session{FullSaveAssociations: true}).
				Omit("SkipTimes", "StreamInfo").
				Create(ep)
		}
	}
	return nil
}

func SaveEpisodeSkipTimes(episodeID string, skipTimes []entities.EpisodeSkipTime) error {
	if len(skipTimes) == 0 {
		return nil
	}

	DB.Where("episode_id = ?", episodeID).Delete(&entities.EpisodeSkipTime{})

	for i := range skipTimes {
		skipTimes[i].EpisodeID = episodeID
		if err := DB.Create(&skipTimes[i]).Error; err != nil {
			return fmt.Errorf("failed to save skip time: %w", err)
		}
	}

	return nil
}

func GetAllAnimeStubs() ([]animeStub, error) {
	var stubs []animeStub
	if err := DB.Model(&entities.Anime{}).Select("mal_id, updated_at").Scan(&stubs).Error; err != nil {
		return nil, err
	}
	return stubs, nil
}

func GetAiringAnime() ([]entities.Anime, error) {
	var anime []entities.Anime

	result := DB.
		Where("airing = ?", true).
		Preload("NextAiring").
		Preload("Schedule").
		Preload("Title").
		Find(&anime)

	if result.Error != nil {
		logger.Errorf("Anime", "Failed to fetch airing anime: %v", result.Error)
		return nil, errors.New("failed to fetch airing anime")
	}

	return anime, nil
}
