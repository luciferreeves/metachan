package repositories

import (
	"errors"
	"metachan/entities"
	"metachan/utils/logger"

	"gorm.io/gorm/clause"
)

func CreateOrUpdateGenre(genre *entities.Genre) error {
	result := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "genre_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "url", "count"}),
	}).Create(genre)

	if result.Error != nil {
		logger.Errorf("Genre", "Failed to create or update genre: %v", result.Error)
		return errors.New("failed to create or update genre")
	}

	return nil
}
