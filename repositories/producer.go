package repositories

import (
	"errors"
	"metachan/entities"
	"metachan/utils/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateOrUpdateProducer(producer *entities.Producer) error {
	result := DB.Clauses(clause.OnConflict{
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

func BatchCreateProducers(producers []entities.Producer) error {
	if len(producers) == 0 {
		return nil
	}

	result := DB.Session(&gorm.Session{FullSaveAssociations: true}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "mal_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"url", "favorites", "count", "established", "about", "image_id",
		}),
	}).CreateInBatches(&producers, 100)

	if result.Error != nil {
		logger.Errorf("Producer", "Failed to batch create producers: %v", result.Error)
		return errors.New("failed to batch create producers")
	}

	return nil
}
