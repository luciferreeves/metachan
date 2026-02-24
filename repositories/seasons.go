package repositories

import (
	"metachan/entities"
	"metachan/utils/logger"
)

func GetRelatedAnimeByTVDB(tvdbID int, excludeMALID int) ([]entities.Mapping, error) {
	var mappings []entities.Mapping

	result := DB.Where("tvdb = ? AND mal != ?", tvdbID, excludeMALID).Order("mal ASC").Find(&mappings)
	if result.Error != nil {
		logger.Errorf("Seasons", "Failed to get related anime by TVDB ID %d: %v", tvdbID, result.Error)
		return nil, result.Error
	}

	logger.Debugf("Seasons", "Found %d related anime via TVDB ID %d", len(mappings), tvdbID)
	return mappings, nil
}

func GetRelatedAnimeByTMDB(tmdbID int, excludeMALID int) ([]entities.Mapping, error) {
	var mappings []entities.Mapping

	result := DB.Where("tmdb = ? AND mal != ?", tmdbID, excludeMALID).Order("mal ASC").Find(&mappings)
	if result.Error != nil {
		logger.Errorf("Seasons", "Failed to get related anime by TMDB ID %d: %v", tmdbID, result.Error)
		return nil, result.Error
	}

	logger.Debugf("Seasons", "Found %d related anime via TMDB ID %d", len(mappings), tmdbID)
	return mappings, nil
}
