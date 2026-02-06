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
		Preload("Characters").
		Preload("Characters.VoiceActors").
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

	return anime, nil
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
	}).Save(anime)

	if result.Error != nil {
		return fmt.Errorf("failed to save anime: %w", result.Error)
	}

	logger.Infof("Anime", "Saved anime (MAL ID: %d) with %d episodes, %d characters", anime.MALID, len(anime.Episodes), len(anime.Characters))
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
