package repositories

import (
	"errors"
	"fmt"
	"metachan/entities"
	"metachan/enums"
	"metachan/utils/logger"

	"gorm.io/gorm/clause"
)

func GetAnimeMapping[T idType](maptype enums.MappingType, id T) (entities.Mapping, error) {
	var mapping entities.Mapping

	result := DB.Where(fmt.Sprintf("%s = ?", maptype), id).First(&mapping)

	if result.Error != nil {
		logger.Errorf("Mapping", "Failed to get mapping for %s with ID %v: %v", maptype, id, result.Error)
		return entities.Mapping{}, errors.New("mapping not found")
	}

	return mapping, nil
}

func CreateOrUpdateMapping(mapping *entities.Mapping) error {
	result := DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "mal"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"ani_db", "anilist", "anime_countdown", "anime_planet",
			"ani_search", "imdb", "kitsu", "live_chart", "notify_moe",
			"simkl", "tmdb", "tvdb", "type", "mal_anilist_composite",
		}),
	}).Create(mapping)

	if result.Error != nil {
		logger.Errorf("Mapping", "Failed to create or update mapping: %v", result.Error)
		return errors.New("failed to create or update mapping")
	}

	return nil
}

func GetAllMappings() ([]entities.Mapping, error) {
	var mappings []entities.Mapping

	result := DB.Find(&mappings)
	if result.Error != nil {
		logger.Errorf("Mapping", "Failed to fetch all mappings: %v", result.Error)
		return nil, errors.New("failed to fetch mappings")
	}

	return mappings, nil
}
