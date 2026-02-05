package repositories

import (
	"errors"
	"metachan/database"
	"metachan/entities"
	"metachan/utils/logger"

	"gorm.io/gorm/clause"
)

func CreateOrUpdateProducer(producer *entities.Producer) error {
	result := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "mal_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"url", "favorites", "count", "established", "about", "image_id",
		}),
	}).Create(producer)

	if result.Error != nil {
		logger.Errorf("Producer", "Failed to create or update producer: %v", result.Error)
		return errors.New("failed to create or update producer")
	}

	return nil
}
